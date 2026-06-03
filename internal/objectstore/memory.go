package objectstore

import (
	"context"
	"sync"
)

// MemStore is an in-memory Store for tests and the lightweight/dev mode.
type MemStore struct {
	mu      sync.RWMutex
	objects map[string]Object
}

// NewMemory returns an empty in-memory store.
func NewMemory() *MemStore {
	return &MemStore{objects: make(map[string]Object)}
}

func (m *MemStore) Put(_ context.Context, key, contentType string, data []byte) error {
	if err := validKey(key); err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[key] = Object{Data: cp, ContentType: contentType, Size: int64(len(cp))}
	return nil
}

func (m *MemStore) Get(_ context.Context, key string) (Object, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.objects[key]
	if !ok {
		return Object{}, ErrNotFound
	}
	cp := make([]byte, len(o.Data))
	copy(cp, o.Data)
	return Object{Data: cp, ContentType: o.ContentType, Size: o.Size}, nil
}

func (m *MemStore) Stat(_ context.Context, key string) (int64, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.objects[key]
	if !ok {
		return 0, false, nil
	}
	return o.Size, true, nil
}

// Len reports the number of stored objects (test inspection).
func (m *MemStore) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.objects)
}
