// SPDX-License-Identifier: LicenseRef-probectl-TBD

package pipeline

import (
	"context"
	"errors"
	"testing"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	"google.golang.org/protobuf/proto"

	"github.com/imfeelingtheagi/probectl/internal/bus"
)

type fakeExporter struct {
	calls int
	fail  bool
}

func (f *fakeExporter) ExportMetrics(_ context.Context, _ *colmetricspb.ExportMetricsServiceRequest) error {
	f.calls++
	if f.fail {
		return errors.New("collector down")
	}
	return nil
}

// ARCH-007: a successful export is counted; an export failure returns an error
// (so the at-least-once consumer redelivers) and is counted as failed — never a
// silent drop. A malformed payload is dropped without erroring the stream.
func TestOTLPExportConsumer(t *testing.T) {
	payload, _ := proto.Marshal(&colmetricspb.ExportMetricsServiceRequest{
		ResourceMetrics: []*metricspb.ResourceMetrics{{}},
	})

	// Success path.
	exp := &fakeExporter{}
	c := NewOTLPExportConsumer(bus.NewMemory(), exp, testLogger())
	if err := c.handle(context.Background(), bus.Message{Value: payload}); err != nil {
		t.Fatalf("export: %v", err)
	}
	if c.Exported() != 1 || c.Failed() != 0 {
		t.Fatalf("success counters: exported=%d failed=%d", c.Exported(), c.Failed())
	}

	// Failure path → error returned (redelivery) + counted.
	expF := &fakeExporter{fail: true}
	cf := NewOTLPExportConsumer(bus.NewMemory(), expF, testLogger())
	if err := cf.handle(context.Background(), bus.Message{Value: payload}); err == nil {
		t.Fatal("export failure must return an error so the record redelivers")
	}
	if cf.Failed() != 1 {
		t.Fatalf("failure not counted: %d", cf.Failed())
	}

	// Malformed payload is dropped, not errored.
	if err := c.handle(context.Background(), bus.Message{Value: []byte("garbage")}); err != nil {
		t.Fatalf("malformed payload must not error the stream: %v", err)
	}
}
