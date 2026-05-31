package bus

import (
	"context"
	"errors"
	"sync"
)

var errClosed = errors.New("bus: closed")

// Memory is an in-process Bus for the lightweight (<5 agent) mode and tests. It
// is a live pub/sub: a subscriber receives messages published after it
// subscribes, which matches the pipeline (the consumer starts at boot, before
// agents connect). Messages are not persisted.
type Memory struct {
	mu     sync.Mutex
	subs   map[string][]chan Message
	closed bool
}

// NewMemory returns an in-memory bus.
func NewMemory() *Memory {
	return &Memory{subs: make(map[string][]chan Message)}
}

// Publish delivers value to every current subscriber of topic.
func (m *Memory) Publish(ctx context.Context, topic string, key, value []byte) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errClosed
	}
	chans := append([]chan Message(nil), m.subs[topic]...)
	m.mu.Unlock()

	msg := Message{Topic: topic, Key: key, Value: value}
	for _, ch := range chans {
		select {
		case ch <- msg:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// Subscribe delivers topic messages to handler until ctx is canceled.
func (m *Memory) Subscribe(ctx context.Context, topic, _ string, handler Handler) error {
	ch := make(chan Message, 1024)
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return errClosed
	}
	m.subs[topic] = append(m.subs[topic], ch)
	m.mu.Unlock()
	defer m.removeSub(topic, ch)

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			// Handler errors are the handler's concern (it retries/logs); the
			// in-memory bus does not redeliver.
			_ = handler(ctx, msg)
		}
	}
}

// Close marks the bus closed.
func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *Memory) removeSub(topic string, ch chan Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	subs := m.subs[topic]
	for i, c := range subs {
		if c == ch {
			m.subs[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
}
