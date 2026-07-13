package catalog

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"

	"github.com/Maverick7t/ecom/internal/jobs/reviews"
	"github.com/Maverick7t/ecom/internal/platform/database/dbgen"
)

// CatalogIngestionArgs is the root job enqueued by cmd/seed.
// Open item #2 (resolved): one category per run, terminates after
// Limit successfully filtered products.
type CatalogIngestionArgs struct {
	SourcePath string `json:"source_path"` // local path or R2-mounted path to metadata.jsonl.gz
	Category   string `json:"category"`    // required, matches metadata main_category/categories
	Limit      int    `json:"limit"`       // default 50000
}

func (CatalogIngestionArgs) Kind() string { return "catalog_ingestion" }

// metadataRecord mirrors the Amazon Reviews 2023 metadata JSONL schema.
// NOT independently verified in this session — confirm field names
// against your actual metadata.jsonl.gz before a full run.
type metadataRecord struct {
	MainCategory string   `json:"main_category"`
	Categories   []string `json:"categories"`
	Title        string   `json:"title"`
	Store        string   `json:"store"` // used as brand
	Description  []string `json:"description"`
	Price        *string  `json:"price"` // dataset stores price as string or null
	ParentAsin   string   `json:"parent_asin"`
	Images       []struct {
		Large string `json:"large"`
	} `json:"images"`
}

type Worker struct {
	river.WorkerDefaults[CatalogIngestionArgs]
	pool    *pgxpool.Pool
	queries *dbgen.Queries
	logger  *slog.Logger
}

func NewWorker(pool *pgxpool.Pool, queries *dbgen.Queries, logger *slog.Logger) *Worker {
	return &Worker{pool: pool, queries: queries, logger: logger}
}

const checkpointInterval = 1000

func (w *Worker) Work(ctx context.Context, job *river.Job[CatalogIngestionArgs]) error {
	args := job.Args
	if strings.TrimSpace(args.Category) == "" {
		return fmt.Errorf("catalog_ingestion: category is required")
	}
	if args.Limit <= 0 {
		args.Limit = 50000
	}

	f, err := os.Open(args.SourcePath)
	if err != nil {
		return fmt.Errorf("open source file %s: %w", args.SourcePath, err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("open gzip reader: %w", err)
	}
	defer gz.Close()

	syncRunID, err := w.queries.CreateSyncRun(ctx, "catalog_ingestion")
	if err != nil {
		return fmt.Errorf("create sync_run: %w", err)
	}

	categoryID, err := w.queries.GetOrCreateCategory(ctx, dbgen.GetOrCreateCategoryParams{
		Slug: slugify(args.Category),
		Name: args.Category,
	})
	if err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("get or create category: %w", err))
	}

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024)

	var recordsIn, recordsOut int
	riverClient := river.ClientFromContext[pgx.Tx](ctx)
	batchDate := time.Now().UTC().Format("2006-01-02")

	for scanner.Scan() {
		if recordsOut >= args.Limit {
			break
		}
		recordsIn++

		var rec metadataRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			w.logger.Warn("skip malformed record", slog.Int("line", recordsIn), slog.Any("error", err))
			continue
		}

		if !matchesCategory(rec.MainCategory, rec.Categories, args.Category) {
			continue
		}
		if rec.ParentAsin == "" || rec.Title == "" {
			w.logger.Warn("skip invalid record — missing required field", slog.String("parent_asin", rec.ParentAsin))
			continue
		}

		title := sanitize(rec.Title)
		description := sanitize(strings.Join(rec.Description, " "))

		var imageURL *string
		if len(rec.Images) > 0 && rec.Images[0].Large != "" {
			imageURL = &rec.Images[0].Large
		}

		product, err := w.queries.UpsertProduct(ctx, dbgen.UpsertProductParams{
			ParentAsin:  rec.ParentAsin,
			Title:       title,
			Brand:       nilIfEmpty(rec.Store),
			Description: nilIfEmpty(description),
			Price:       rec.Price,
			Currency:    strPtr("USD"),
			ImageUrl:    imageURL,
			ProductType: nilIfEmpty(rec.MainCategory),
			Condition:   strPtr("New"),
		})
		if err != nil {
			w.logger.Error("upsert product failed", slog.String("parent_asin", rec.ParentAsin), slog.Any("error", err))
			continue
		}

		if err := w.queries.LinkProductCategory(ctx, dbgen.LinkProductCategoryParams{
			ProductID:  product.ID,
			CategoryID: categoryID,
		}); err != nil {
			w.logger.Error("link category failed", slog.String("parent_asin", rec.ParentAsin), slog.Any("error", err))
		}

		recordsOut++

		if recordsOut%checkpointInterval == 0 {
			if err := w.queries.UpdateSyncRunProgress(ctx, dbgen.UpdateSyncRunProgressParams{
				ID:         syncRunID,
				RecordsIn:  int32(recordsIn),
				RecordsOut: int32(recordsOut),
			}); err != nil {
				w.logger.Error("checkpoint failed", slog.Any("error", err))
			}
		}

		// idempotency key: product_id + source_batch_date (execution_phase 2.3)
		if _, err := riverClient.Insert(ctx, reviews.ReviewIngestionArgs{
			ProductID:       product.ID.String(),
			SourceASIN:      product.ParentAsin,
			SourceBatchDate: batchDate,
		}, &river.InsertOpts{
			UniqueOpts: river.UniqueOpts{ByArgs: true},
		}); err != nil {
			w.logger.Error("enqueue review_ingestion failed", slog.String("parent_asin", rec.ParentAsin), slog.Any("error", err))
		}
	}

	if err := scanner.Err(); err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("scan source file: %w", err))
	}

	return w.queries.CompleteSyncRun(ctx, dbgen.CompleteSyncRunParams{
		ID:         syncRunID,
		Status:     "COMPLETED",
		RecordsIn:  int32(recordsIn),
		RecordsOut: int32(recordsOut),
	})
}

func (w *Worker) failRun(ctx context.Context, syncRunID uuid.UUID, cause error) error {
	if err := w.queries.CompleteSyncRun(ctx, dbgen.CompleteSyncRunParams{
		ID:           syncRunID,
		Status:       "FAILED",
		ErrorMessage: nilIfEmpty(cause.Error()),
	}); err != nil {
		w.logger.Error("failed to mark sync_run FAILED", slog.Any("error", err))
	}
	return cause
}

func matchesCategory(mainCat string, categories []string, target string) bool {
	if strings.EqualFold(mainCat, target) {
		return true
	}
	for _, c := range categories {
		if strings.EqualFold(c, target) {
			return true
		}
	}
	return false
}

func sanitize(s string) string {
	return strings.TrimSpace(html.UnescapeString(s))
}

func slugify(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strPtr(s string) *string { return &s }
