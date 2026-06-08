// SPDX-License-Identifier: LicenseRef-probectl-TBD

package objectstore

import (
	"context"
	"errors"
	"testing"
)

func TestTenantKey(t *testing.T) {
	if got := TenantKey("t1", "browser", "run-9.png"); got != "tenant/t1/browser/run-9.png" {
		t.Fatalf("TenantKey = %q", got)
	}
}

func TestValidKey(t *testing.T) {
	for _, bad := range []string{"", "/abs", "a/../../etc/passwd", "x\x00y", "..", "a/.."} {
		if err := validKey(bad); err == nil {
			t.Fatalf("validKey(%q) should reject", bad)
		}
	}
	for _, ok := range []string{"a", "tenant/t1/browser/r.png", "a/b/c"} {
		if err := validKey(ok); err != nil {
			t.Fatalf("validKey(%q) should accept: %v", ok, err)
		}
	}
}

// runStoreSuite exercises the Store contract against any implementation.
func runStoreSuite(t *testing.T, s Store) {
	t.Helper()
	ctx := context.Background()
	key := TenantKey("t1", "browser", "login-1.png")

	if _, _, err := s.Stat(ctx, key); err != nil {
		t.Fatalf("stat missing: %v", err)
	}
	if _, err := s.Get(ctx, key); !errors.Is(err, ErrNotFound) {
		t.Fatalf("get missing should be ErrNotFound, got %v", err)
	}

	data := []byte("\x89PNG fake screenshot bytes")
	if err := s.Put(ctx, key, "image/png", data); err != nil {
		t.Fatalf("put: %v", err)
	}
	o, err := s.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(o.Data) != string(data) || o.ContentType != "image/png" || o.Size != int64(len(data)) {
		t.Fatalf("round-trip mismatch: %+v", o)
	}
	if size, exists, err := s.Stat(ctx, key); err != nil || !exists || size != int64(len(data)) {
		t.Fatalf("stat: size=%d exists=%v err=%v", size, exists, err)
	}

	// Tenant isolation: a different tenant's key is a different object.
	other := TenantKey("t2", "browser", "login-1.png")
	if _, exists, _ := s.Stat(ctx, other); exists {
		t.Fatal("tenant t2 must not see tenant t1's object")
	}

	// Default content type when empty.
	_ = s.Put(ctx, "tenant/t1/x", "", []byte("x"))
	if o, _ := s.Get(ctx, "tenant/t1/x"); o.ContentType != "application/octet-stream" {
		t.Fatalf("default content type: %q", o.ContentType)
	}

	// Path traversal is rejected.
	if err := s.Put(ctx, "tenant/t1/../../escape", "text/plain", []byte("x")); err == nil {
		t.Fatal("traversal key must be rejected")
	}
}

func TestMemStore(t *testing.T) { runStoreSuite(t, NewMemory()) }

func TestFSStore(t *testing.T) {
	s, err := NewFS(t.TempDir())
	if err != nil {
		t.Fatalf("new fs: %v", err)
	}
	runStoreSuite(t, s)
}

func TestFSStorePersistsAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	s1, _ := NewFS(dir)
	_ = s1.Put(context.Background(), "tenant/t1/a.txt", "text/plain", []byte("hello"))

	s2, _ := NewFS(dir) // a fresh handle on the same dir
	o, err := s2.Get(context.Background(), "tenant/t1/a.txt")
	if err != nil || string(o.Data) != "hello" || o.ContentType != "text/plain" {
		t.Fatalf("persisted get: %+v err=%v", o, err)
	}
}
