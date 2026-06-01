package otlp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

// ServerConfig configures the bundled OTLP receiver listeners.
type ServerConfig struct {
	GRPCAddr     string // e.g. ":4317" (empty disables the gRPC receiver)
	HTTPAddr     string // e.g. ":4318" (empty disables the HTTP receiver)
	MaxRecvBytes int    // 0 => default (4 MiB)
}

// Server runs the OTLP/gRPC and OTLP/HTTP receivers on TLS listeners. It is the
// inbound OTLP surface: TLS-only, authenticated, tenant-scoped, untrusted input
// (CLAUDE.md §7 guardrail 12).
type Server struct {
	cfg  ServerConfig
	tls  *tls.Config
	auth Authenticator
	sink Sink
	log  *slog.Logger
}

// NewServer validates the receiver configuration and returns a runnable Server.
func NewServer(cfg ServerConfig, tlsCfg *tls.Config, auth Authenticator, sink Sink, log *slog.Logger) (*Server, error) {
	if tlsCfg == nil {
		return nil, errors.New("otlp: TLS config required (the receiver is TLS-only)")
	}
	if auth == nil || sink == nil {
		return nil, errors.New("otlp: authenticator and sink are required")
	}
	if cfg.GRPCAddr == "" && cfg.HTTPAddr == "" {
		return nil, errors.New("otlp: a gRPC or HTTP address is required")
	}
	if log == nil {
		log = slog.Default()
	}
	return &Server{cfg: cfg, tls: tlsCfg, auth: auth, sink: sink, log: log}, nil
}

// Run starts the configured listeners and blocks until ctx is canceled or a
// listener fails.
func (s *Server) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	if s.cfg.GRPCAddr != "" {
		grpcSrv, err := NewGRPCServer(s.tls, s.auth, s.sink, s.cfg.MaxRecvBytes)
		if err != nil {
			return err
		}
		lis, err := net.Listen("tcp", s.cfg.GRPCAddr)
		if err != nil {
			return fmt.Errorf("otlp: grpc listen %q: %w", s.cfg.GRPCAddr, err)
		}
		s.log.Info("otlp grpc receiver listening", "addr", s.cfg.GRPCAddr)
		g.Go(func() error { return grpcSrv.Serve(lis) })
		g.Go(func() error {
			<-gctx.Done()
			grpcSrv.GracefulStop()
			return nil
		})
	}

	if s.cfg.HTTPAddr != "" {
		mux := http.NewServeMux()
		mux.Handle("/v1/metrics", MetricsHTTPHandler(s.auth, s.sink, int64(s.cfg.MaxRecvBytes)))
		httpSrv := &http.Server{
			Addr:              s.cfg.HTTPAddr,
			Handler:           mux,
			TLSConfig:         s.tls,
			ReadHeaderTimeout: 10 * time.Second,
		}
		s.log.Info("otlp http receiver listening", "addr", s.cfg.HTTPAddr)
		g.Go(func() error {
			if err := httpSrv.ListenAndServeTLS("", ""); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		})
		g.Go(func() error {
			<-gctx.Done()
			sctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			return httpSrv.Shutdown(sctx)
		})
	}

	return g.Wait()
}
