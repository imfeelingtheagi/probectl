package pipeline

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/proto"

	"github.com/imfeelingtheagi/netctl/internal/bus"
	resultv1 "github.com/imfeelingtheagi/netctl/internal/gen/netctl/result/v1"
	"github.com/imfeelingtheagi/netctl/internal/store/tsdb"
)

// DefaultGroup is the consumer-group name for the control-plane result pipeline.
const DefaultGroup = "netctl-control"

// Consumer drains result messages from the bus and writes them to the TSDB.
type Consumer struct {
	bus   bus.Bus
	tsdb  tsdb.Writer
	group string
	log   *slog.Logger
}

// NewConsumer builds the result-pipeline consumer.
func NewConsumer(b bus.Bus, w tsdb.Writer, group string, log *slog.Logger) *Consumer {
	if group == "" {
		group = DefaultGroup
	}
	return &Consumer{bus: b, tsdb: w, group: group, log: log}
}

// Run subscribes to the network-results topic and writes each result to the TSDB
// until ctx is canceled. It blocks.
func (c *Consumer) Run(ctx context.Context) error {
	c.log.Info("result pipeline consumer starting", "topic", bus.NetworkResultsTopic, "group", c.group)
	return c.bus.Subscribe(ctx, bus.NetworkResultsTopic, c.group, c.handle)
}

// handle decodes one result and writes its series. Malformed messages and
// transient write failures are logged and dropped (best-effort) rather than
// blocking the stream; durable retry/redelivery is a later hardening step.
func (c *Consumer) handle(ctx context.Context, msg bus.Message) error {
	var r resultv1.Result
	if err := proto.Unmarshal(msg.Value, &r); err != nil {
		c.log.Error("dropping malformed result", "error", err.Error())
		return nil
	}
	if err := c.tsdb.Write(ctx, ResultToSeries(&r)); err != nil {
		c.log.Error("tsdb write failed", "tenant_id", r.GetTenantId(), "agent_id", r.GetAgentId(), "error", err.Error())
		return nil
	}
	return nil
}
