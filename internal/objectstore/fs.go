package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FSStore is a filesystem-backed Store. Each object is a file under root; its
// content type is kept in a sibling ".meta" file. Suitable for single-node /
// air-gapped deploys; swap for S3/MinIO at scale.
type FSStore struct {
	root string
}

// NewFS returns a filesystem store rooted at dir (created if missing).
func NewFS(dir string) (*FSStore, error) {
	if dir == "" {
		return nil, errors.New("objectstore: empty root dir")
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, fmt.Errorf("objectstore: create root: %w", err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	return &FSStore{root: abs}, nil
}

// path resolves key to an absolute path confined to root.
func (s *FSStore) path(key string) (string, error) {
	if err := validKey(key); err != nil {
		return "", err
	}
	p := filepath.Join(s.root, filepath.FromSlash(key))
	// Defense-in-depth: ensure the join stayed under root.
	if p != s.root && !strings.HasPrefix(p, s.root+string(os.PathSeparator)) {
		return "", errors.New("objectstore: key escapes root")
	}
	return p, nil
}

func (s *FSStore) Put(_ context.Context, key, contentType string, data []byte) error {
	p, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(p, data, 0o600); err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return os.WriteFile(p+".meta", []byte(contentType), 0o600)
}

func (s *FSStore) Get(_ context.Context, key string) (Object, error) {
	p, err := s.path(key)
	if err != nil {
		return Object{}, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, fs.ErrNotExist) {
		return Object{}, ErrNotFound
	}
	if err != nil {
		return Object{}, err
	}
	ct := "application/octet-stream"
	if meta, err := os.ReadFile(p + ".meta"); err == nil {
		ct = strings.TrimSpace(string(meta))
	}
	return Object{Data: data, ContentType: ct, Size: int64(len(data))}, nil
}

func (s *FSStore) Stat(_ context.Context, key string) (int64, bool, error) {
	p, err := s.path(key)
	if err != nil {
		return 0, false, err
	}
	fi, err := os.Stat(p)
	if errors.Is(err, fs.ErrNotExist) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return fi.Size(), true, nil
}
