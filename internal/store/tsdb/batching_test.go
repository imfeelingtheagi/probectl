// SPDX-License-Identifier: LicenseRef-probectl-TBD

package tsdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type countingWriter struct {
	mu       sync.Mutex
	calls    int
	series   int
	failNext bool
}

func (c *countingWriter) Write(_ context.Context, s []Series) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	c.series += len(s)
	if c.failNext {
		return errors.New("remote-write down")
	}
	return nil
}
func (c *countingWriter) Close() error { return nil }

// SCALE-001: concurrent Writes within the window coalesce into ONE underlying
// remote-write request, and every caller still gets that request's result (so
// per-message DLQ attribution is preserved).
func TestBatchingWriterCoalesces(t *testing.T) {
	under := &countingWriter{}
	bw := NewBatchingWriter(under, 500, 50*time.Millisecond)

	const n = 20
	var wg sync.WaitGroup
	var oks atomic.Int64
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := bw.Write(context.Background(), []Series{{Metric: "m", Value: 1, TimeMillis: 1}}); err == nil {
				oks.Add(1)
			}
		}()
	}
	wg.Wait()

	if oks.Load() != n {
		t.Fatalf("callers succeeded = %d, want %d", oks.Load(), n)
	}
	under.mu.Lock()
	calls, series := under.calls, under.series
	under.mu.Unlock()
	if calls >= n {
		t.Fatalf("no coalescing: %d underlying writes for %d callers (want far fewer)", calls, n)
	}
	if series != n {
		t.Fatalf("series lost in coalescing: wrote %d, want %d", series, n)
	}
}

// SCALE-001: a size-cap trigger flushes without waiting for the timer.
func TestBatchingWriterSizeTrigger(t *testing.T) {
	under := &countingWriter{}
	bw := NewBatchingWriter(under, 3, time.Hour) // huge wait → only size can flush
	done := make(chan error, 1)
	go func() { done <- bw.Write(context.Background(), []Series{{}, {}, {}}) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("write: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("size-cap did not trigger a flush (still waiting on the timer)")
	}
}

// SCALE-001: the underlying error reaches the caller (so the consumer DLQs it).
func TestBatchingWriterPropagatesError(t *testing.T) {
	under := &countingWriter{failNext: true}
	bw := NewBatchingWriter(under, 500, 10*time.Millisecond)
	if err := bw.Write(context.Background(), []Series{{Metric: "m"}}); err == nil {
		t.Fatal("a failed flush must surface to the caller for DLQ attribution")
	}
}
