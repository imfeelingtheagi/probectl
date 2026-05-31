package bus

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kfake"
)

func TestMemoryPublishSubscribe(t *testing.T) {
	b := NewMemory()
	defer b.Close()
	testPubSub(t, b)
}

// TestKafkaPublishSubscribe exercises the real Kafka client path against an
// in-process kfake broker (Kafka protocol, no JVM/Docker needed).
func TestKafkaPublishSubscribe(t *testing.T) {
	cluster, err := kfake.NewCluster(kfake.SeedTopics(1, NetworkResultsTopic))
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Close()

	b, err := NewKafka(cluster.ListenAddrs())
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	testPubSub(t, b)
}

// testPubSub publishes three messages and asserts the subscriber receives them.
func testPubSub(t *testing.T, b Bus) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	got := make(chan byte, 16)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = b.Subscribe(ctx, NetworkResultsTopic, "test-group", func(_ context.Context, m Message) error {
			if len(m.Value) > 0 {
				got <- m.Value[0]
			}
			return nil
		})
	}()

	// Let the subscriber register / join the consumer group.
	time.Sleep(500 * time.Millisecond)
	for i := byte(0); i < 3; i++ {
		if err := b.Publish(ctx, NetworkResultsTopic, []byte("tenant-1"), []byte{i}); err != nil {
			t.Fatalf("publish %d: %v", i, err)
		}
	}

	received := map[byte]bool{}
	timeout := time.After(25 * time.Second)
	for len(received) < 3 {
		select {
		case v := <-got:
			received[v] = true
		case <-timeout:
			t.Fatalf("received only %d/3 messages", len(received))
		}
	}
	cancel()
	wg.Wait()

	for i := byte(0); i < 3; i++ {
		if !received[i] {
			t.Errorf("missing message %d", i)
		}
	}
}
