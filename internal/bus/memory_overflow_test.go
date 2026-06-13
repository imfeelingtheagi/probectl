// SPDX-License-Identifier: LicenseRef-probectl-TBD

package bus

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Under the DROP overflow policy, a stuck subscriber cannot deadlock the
// publisher in a burst (U-079): publishes past the buffer are dropped and
// counted, never blocked.
func TestMemoryDropOverflowDoesNotBlock(t *testing.T) {
	m := NewMemory(WithBuffer(2), WithOverflowDrop())
	defer m.Close()

	// A subscriber that never drains its channel (registered but stuck).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	subscribed := make(chan struct{})
	go func() {
		_ = m.Subscribe(ctx, "t", "g", func(context.Context, Message) error {
			<-ctx.Done() // block forever inside the handler — channel fills
			return nil
		})
	}()
	// Let the subscriber register.
	for i := 0; i < 5000 && m.subscriberCount("t") == 0; i++ { // ~5s: -race-safe (cf. TestPolicyLifecycle)
		time.Sleep(time.Millisecond)
	}
	close(subscribed)

	// Publish far more than the buffer; with the drop policy this must return
	// promptly (no deadlock) and count the overflow.
	done := make(chan error, 1)
	go func() {
		for i := 0; i < 1000; i++ {
			if err := m.Publish(context.Background(), "t", nil, []byte("x")); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("publish error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("drop policy deadlocked the publisher")
	}
	if m.Dropped() == 0 {
		t.Fatal("overflow under the drop policy must be counted")
	}
}

// The block policy (default) keeps every message: a draining subscriber
// receives all of them, and Dropped stays zero.
func TestMemoryBlockPolicyLosesNothing(t *testing.T) {
	m := NewMemory(WithBuffer(4)) // default = block
	defer m.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n = 200
	got := make(chan int, 1)
	go func() {
		count := 0
		_ = m.Subscribe(ctx, "t", "g", func(_ context.Context, _ Message) error {
			count++
			if count == n {
				got <- count
			}
			return nil
		})
	}()
	for i := 0; i < 5000 && m.subscriberCount("t") == 0; i++ { // ~5s: -race-safe (cf. TestPolicyLifecycle)
		time.Sleep(time.Millisecond)
	}
	for i := 0; i < n; i++ {
		if err := m.Publish(context.Background(), "t", nil, []byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case c := <-got:
		if c != n {
			t.Fatalf("received %d, want %d", c, n)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("block policy lost messages")
	}
	if m.Dropped() != 0 {
		t.Fatalf("block policy dropped %d", m.Dropped())
	}
}

// TestMemoryDefaultPolicyOneStuckSubDoesNotStarveOthers is the SCALE-006
// behavioral acceptance: with the now-default drop policy, one stuck subscriber
// cannot block the producer or starve a second, draining subscriber on the same
// topic.
func TestMemoryDefaultPolicyOneStuckSubDoesNotStarveOthers(t *testing.T) {
	// The shipped default is drop (config default flipped to "drop" in SCALE-006,
	// wired via WithOverflowDrop in main.go). Model that here.
	m := NewMemory(WithBuffer(2), WithOverflowDrop())
	defer m.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscriber A: stuck forever (never drains).
	go func() {
		_ = m.Subscribe(ctx, "t", "stuck", func(context.Context, Message) error {
			<-ctx.Done()
			return nil
		})
	}()
	// Subscriber B: drains normally and counts what it receives.
	const n = 500
	var received atomic.Uint64
	go func() {
		_ = m.Subscribe(ctx, "t", "drainer", func(context.Context, Message) error {
			received.Add(1)
			return nil
		})
	}()
	for i := 0; i < 5000 && m.subscriberCount("t") < 2; i++ { // ~5s: -race-safe
		time.Sleep(time.Millisecond)
	}
	if m.subscriberCount("t") < 2 {
		t.Fatal("both subscribers did not register")
	}

	// The producer must never block on the stuck subscriber — the whole point of
	// SCALE-006's default. With block-on-full this Publish loop would wedge on
	// the stuck subscriber's full channel.
	done := make(chan struct{})
	go func() {
		for i := 0; i < n; i++ {
			// Small spacing so the drainer's bounded channel makes progress
			// rather than the burst out-racing its single consumer goroutine.
			if err := m.Publish(context.Background(), "t", nil, []byte("x")); err != nil {
				return
			}
			time.Sleep(100 * time.Microsecond)
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("producer blocked by a stuck subscriber under the default policy (SCALE-006)")
	}
	// The draining subscriber must make real progress (it is NOT blocked behind
	// the stuck one). Under the drop policy it may shed a few on a burst, so we
	// assert substantial progress rather than an exact count.
	deadline := time.Now().Add(2 * time.Second)
	for received.Load() < n/2 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if got := received.Load(); got < n/2 {
		t.Fatalf("draining subscriber starved by the stuck subscriber: received %d of %d", got, n)
	}
}
