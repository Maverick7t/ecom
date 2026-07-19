package storage
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Maverick7t/ecom/internal/platform/config"
)

type SupabaseStorage struct {
	baseURL    string // e.g. https://<project>.supabase.co/storage/v1
	bucket     string
	serviceKey string
	httpClient *http.Client
}

func NewSupabaseStorage(cfg *config.Config) *SupabaseStorage {
	return &SupabaseStorage{
		baseURL:    cfg.SupabaseURL + "/storage/v1",
		bucket:     cfg.SupabaseStorageBucket,
		serviceKey: cfg.SupabaseServiceRoleKey,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SupabaseStorage) Upload(ctx context.Context, path string, data []byte, contentType string) error {
	url := fmt.Sprintf("%s/object/%s/%s", s.baseURL, s.bucket, path)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build storage request: %w", err)
	}