package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStorage writes blobs to disk instead of making a network call.
// Wired in when cfg.AppEnv != "prod" so repeated pipeline reruns during
// development cost zero Supabase egress.
type LocalStorage struct {
	rootDir string
}

func NewLocalStorage(rootDir string) (*LocalStorage, error) {
	if rootDir == "" {
		rootDir = "./tmp/reviews"
	}
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, fmt.Errorf("create local storage root %q: %w", rootDir, err)
	}
	return &LocalStorage{rootDir: rootDir}, nil
}

func (s *LocalStorage) Upload(_ context.Context, path string, data []byte, _ string) error {
	fullPath := filepath.Join(s.rootDir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return fmt.Errorf("create dir for %s: %w", path, err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return fmt.Errorf("write local blob %s: %w", path, err)
	}
	return nil
}
