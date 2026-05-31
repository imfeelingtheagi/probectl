package bus

import (
	"context"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Kafka is a franz-go-backed Bus (pure Go, CGO-free). TLS in transit is supported
// by passing kgo.DialTLSConfig through extra options (default-on in regulated
// deploy profiles; the dev stack is plaintext). CLAUDE.md §7 guardrail 12.
type Kafka struct {
	producer *kgo.Client
	brokers  []string
	extra    []kgo.Opt
}

// NewKafka creates a Kafka bus seeded with brokers.
func NewKafka(brokers []string, extra ...kgo.Opt) (*Kafka, error) {
	opts := append([]kgo.Opt{kgo.SeedBrokers(brokers...)}, extra...)
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("bus: kafka producer: %w", err)
	}
	return &Kafka{producer: cl, brokers: brokers, extra: extra}, nil
}

// Publish produces value to topic synchronously, keyed by key.
func (k *Kafka) Publish(ctx context.Context, topic string, key, value []byte) error {
	res := k.producer.ProduceSync(ctx, &kgo.Record{Topic: topic, Key: key, Value: value})
	return res.FirstErr()
}

// Subscribe consumes topic in a consumer group until ctx is canceled. franz-go
// auto-commits offsets, so delivery is at-least-once.
func (k *Kafka) Subscribe(ctx context.Context, topic, group string, handler Handler) error {
	opts := append([]kgo.Opt{
		kgo.SeedBrokers(k.brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
		// A brand-new group reads from the start so no buffered results are lost;
		// an established group resumes from its committed offset.
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	}, k.extra...)
	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return fmt.Errorf("bus: kafka consumer: %w", err)
	}
	defer cl.Close()

	for ctx.Err() == nil {
		fetches := cl.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return nil
		}
		fetches.EachRecord(func(r *kgo.Record) {
			_ = handler(ctx, Message{Topic: r.Topic, Key: r.Key, Value: r.Value})
		})
	}
	return nil
}

// Close closes the producer.
func (k *Kafka) Close() error {
	k.producer.Close()
	return nil
}
