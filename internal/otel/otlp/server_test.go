package otlp

import (
	"context"
	"crypto/tls"
	"testing"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
)

func TestNewServerValidates(t *testing.T) {
	auth := NewTokenAuthenticator(map[string]string{"tok": "tenant-a"})
	sink := SinkFunc(func(context.Context, string, *colmetricspb.ExportMetricsServiceRequest) error { return nil })
	tlsCfg := &tls.Config{MinVersion: tls.VersionTLS12}

	if _, err := NewServer(ServerConfig{GRPCAddr: ":0"}, nil, auth, sink, nil); err == nil {
		t.Error("nil TLS config should fail (TLS-only)")
	}
	if _, err := NewServer(ServerConfig{}, tlsCfg, auth, sink, nil); err == nil {
		t.Error("no listen address should fail")
	}
	if _, err := NewServer(ServerConfig{GRPCAddr: ":0"}, tlsCfg, nil, sink, nil); err == nil {
		t.Error("nil authenticator should fail")
	}
	if _, err := NewServer(ServerConfig{HTTPAddr: ":0"}, tlsCfg, auth, sink, nil); err != nil {
		t.Errorf("valid config rejected: %v", err)
	}
}
