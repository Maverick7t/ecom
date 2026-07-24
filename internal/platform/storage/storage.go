package storage

import "context"

// Storage abstracts the raw-review blob sink. LocalStorage is used in dev
// (writes to disk, zero egress cost on pipeline reruns). SupabaseStorage
// is used in prod. Same interface, swapped once at wiring time in
// cmd/api/main.go based on cfg.IsProd().
type Storage interface {
	Upload(ctx context.Context, path string, data []byte, contentType string) error
}
