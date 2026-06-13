// SPDX-License-Identifier: LicenseRef-probectl-TBD

package pipeline

import (
	"context"
	"log/slog"
	"sync/atomic"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	"google.golang.org/protobuf/proto"

	"github.com/imfeelingtheagi/probectl/internal/bus"
)

// MetricsExporter forwards an OTLP metrics request to an external collector.
// internal/otel/otlp.{GRPC,HTTP}Exporter implement it; the interface keeps the
// pipeline decoupled from the exporter transport.
type MetricsExporter interface {
	ExportMetrics(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) error
}

// OTLPExportConsumer drains the ingested OTLP-metrics topic and re-exports each
// (already tenant-stamped) batch to an external collector (ARCH-007). This is
// the config-driven export pipeline that makes the dormant exporter live: the
// platform can fan ingested OTLP back out to a customer's own backend without
// a separate collector hop. A failed export is logged + counted and the record
// is left UNMARKED so the at-least-once consumer redelivers it (no silent loss).
type OTLPExportConsumer struct {
	bus      bus.Bus
	exporter MetricsExporter
	group    string
	log      *slog.Logger
	exported atomic.Uint64
	failed   atomic.Uint64
}

// NewOTLPExportConsumer builds the consumer over a non-nil exporter.
func NewOTLPExportConsumer(b bus.Bus, exp MetricsExporter, log *slog.Logger) *OTLPExportConsumer {
	if log == nil {
		log = slog.Default()
	}
	return &OTLPExportConsumer{bus: b, exporter: exp, group: DefaultGroup + "-otlp-export", log: log}
}

// Exported / Failed report cumulative export outcomes (observability).
func (c *OTLPExportConsumer) Exported() uint64 { return c.exported.Load() }
func (c *OTLPExportConsumer) Failed() uint64   { return c.failed.Load() }

// Run subscribes until ctx is canceled. It blocks.
func (c *OTLPExportConsumer) Run(ctx context.Context) error {
	c.log.Info("otlp export consumer starting", "topic", bus.OTLPMetricsTopic, "group", c.group)
	return c.bus.Subscribe(ctx, bus.OTLPMetricsTopic, c.group, c.handle)
}

func (c *OTLPExportConsumer) handle(ctx context.Context, msg bus.Message) error {
	var req colmetricspb.ExportMetricsServiceRequest
	if err := proto.Unmarshal(msg.Value, &req); err != nil {
		c.log.Warn("otlp-export: skipping malformed metrics payload", "error", err)
		return nil // poison message: drop (counted as handled), never wedge
	}
	if err := c.exporter.ExportMetrics(ctx, &req); err != nil {
		c.failed.Add(1)
		c.log.Error("otlp-export: forward to external collector failed (will redeliver)", "error", err.Error())
		return err // leave uncommitted → at-least-once redelivery
	}
	c.exported.Add(1)
	return nil
}
