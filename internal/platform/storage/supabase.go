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