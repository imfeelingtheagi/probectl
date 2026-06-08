// SPDX-License-Identifier: LicenseRef-probectl-TBD

package otlp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

// ExporterConfig configures an OTLP exporter (to an external collector).
type ExporterConfig struct {
	Endpoint string      // host:port (gRPC) or full URL (HTTP)
	Token    string      // bearer token, sent as "Authorization: Bearer <token>"
	TLS      *tls.Config // TLS for the connection (required unless Insecure)
	Insecure bool        // dev/test only: plaintext transport (never in production)
}

// GRPCExporter exports OTLP metrics over OTLP/gRPC.
type GRPCExporter struct {
	conn   *grpc.ClientConn
	client colmetricspb.MetricsServiceClient
	token  string
}

// NewGRPCExporter dials the collector. Extra dial options (e.g. a test dialer)
// are appended. TLS is required unless cfg.Insecure is set.
func NewGRPCExporter(cfg ExporterConfig, dialOpts ...grpc.DialOption) (*GRPCExporter, error) {
	var creds grpc.DialOption
	switch {
	case cfg.Insecure:
		creds = grpc.WithTransportCredentials(insecure.NewCredentials())
	case cfg.TLS != nil:
		creds = grpc.WithTransportCredentials(credentials.NewTLS(cfg.TLS))
	default:
		return nil, errors.New("otlp: gRPC exporter requires a TLS config (or Insecure for dev)")
	}
	conn, err := grpc.NewClient(cfg.Endpoint, append([]grpc.DialOption{creds}, dialOpts...)...)
	if err != nil {
		return nil, fmt.Errorf("otlp: dial %q: %w", cfg.Endpoint, err)
	}
	return &GRPCExporter{conn: conn, client: colmetricspb.NewMetricsServiceClient(conn), token: cfg.Token}, nil
}

// ExportMetrics sends an OTLP metrics request, attaching the bearer token.
func (e *GRPCExporter) ExportMetrics(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) error {
	if e.token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+e.token)
	}
	_, err := e.client.Export(ctx, req)
	return err
}

// Close releases the connection.
func (e *GRPCExporter) Close() error { return e.conn.Close() }

// HTTPExporter exports OTLP metrics over OTLP/HTTP (protobuf).
type HTTPExporter struct {
	url    string
	token  string
	client *http.Client
}

// NewHTTPExporter builds an OTLP/HTTP exporter. Production endpoints are https;
// cfg.TLS customizes the client's TLS (e.g. a private CA).
func NewHTTPExporter(cfg ExporterConfig) (*HTTPExporter, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("otlp: HTTP exporter requires an endpoint URL")
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.TLS != nil {
		tr.TLSClientConfig = cfg.TLS
	}
	return &HTTPExporter{url: cfg.Endpoint, token: cfg.Token, client: &http.Client{Transport: tr}}, nil
}

// ExportMetrics POSTs an OTLP metrics request as protobuf.
func (e *HTTPExporter) ExportMetrics(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) error {
	body, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("otlp: marshal: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, e.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	if e.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+e.token)
	}
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("otlp: export rejected: %s", resp.Status)
	}
	return nil
}
