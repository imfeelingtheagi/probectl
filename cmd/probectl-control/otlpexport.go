// SPDX-License-Identifier: LicenseRef-probectl-TBD

package main

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/imfeelingtheagi/probectl/internal/config"
	"github.com/imfeelingtheagi/probectl/internal/otel/otlp"
	"github.com/imfeelingtheagi/probectl/internal/pipeline"
)

// buildOTLPExporter constructs the configured OTLP export client (ARCH-007).
// A remote (non-loopback) endpoint MUST be TLS — Insecure is refused for it
// (guardrail 12); loopback may be plaintext for a co-located collector.
func buildOTLPExporter(cfg *config.Config) (pipeline.MetricsExporter, error) {
	insecure := cfg.OTLPExportInsecure
	if insecure && !isLoopbackEndpoint(cfg.OTLPExportEndpoint) {
		return nil, fmt.Errorf("PROBECTL_OTLP_EXPORT_INSECURE is only allowed for a loopback endpoint, not %q (guardrail 12)", cfg.OTLPExportEndpoint)
	}
	ec := otlp.ExporterConfig{
		Endpoint: cfg.OTLPExportEndpoint,
		Token:    cfg.OTLPExportToken,
		Insecure: insecure,
	}
	if !insecure {
		ec.TLS = &tls.Config{MinVersion: tls.VersionTLS12} // system roots; verified
	}
	if cfg.OTLPExportProtocol == "http" {
		return otlp.NewHTTPExporter(ec)
	}
	return otlp.NewGRPCExporter(ec)
}

// isLoopbackEndpoint reports whether the endpoint host is loopback/localhost.
func isLoopbackEndpoint(ep string) bool {
	e := ep
	for _, p := range []string{"https://", "http://", "grpc://"} {
		e = strings.TrimPrefix(e, p)
	}
	host := e
	if i := strings.IndexAny(host, ":/"); i >= 0 {
		host = host[:i]
	}
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}
