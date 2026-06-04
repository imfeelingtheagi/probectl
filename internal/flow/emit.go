package flow

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"github.com/imfeelingtheagi/probectl/internal/bus"
	flowv1 "github.com/imfeelingtheagi/probectl/internal/gen/probectl/flow/v1"
)

// BusEmitter publishes FlowBatches to probectl.flow.events, tenant-keyed
// (pooled tenant-tagging, CLAUDE.md §6), mirroring the eBPF/endpoint emitters.
type BusEmitter struct {
	bus    bus.Bus
	tenant string
}

// NewBusEmitter returns an Emitter publishing to bus.FlowEventsTopic.
func NewBusEmitter(b bus.Bus, tenant string) *BusEmitter {
	return &BusEmitter{bus: b, tenant: tenant}
}

// Emit marshals the batch and publishes it. An empty batch is a no-op.
func (e *BusEmitter) Emit(ctx context.Context, recs []Record) error {
	if len(recs) == 0 {
		return nil
	}
	batch := &flowv1.FlowBatch{Flows: make([]*flowv1.FlowRecord, 0, len(recs))}
	for i := range recs {
		batch.Flows = append(batch.Flows, recs[i].ToProto())
	}
	value, err := proto.Marshal(batch)
	if err != nil {
		return fmt.Errorf("flow: marshal batch: %w", err)
	}
	return e.bus.Publish(ctx, bus.FlowEventsTopic, []byte(e.tenant), value)
}
