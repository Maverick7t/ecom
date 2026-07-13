package catalog


import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
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

type CatalogIngestionJob struct {
	SourcePath string `json:"source_path"`
	Category string `json:"category"`
	Limit int `json:"limit"`
}

func (CatalogIngestionArgs) Kind() string { return "catalog_ingestion"}
type metadataRecord struct {
	MainCategory string `json:"main_category"`
	Categories []string `json:"categories"`
	Title string `json:"title"`
	Store string `json:"store"`
	Description []string `json:"description"`
	Price *string `json:"price"`
	ParentAsin string `json:"parent_asin"`
	Images []struct {
		Large string `json:"large"`
	} 'json:"images"'
}

type Worker struct {
	river.WorkerDefaults[CatalogIngestionArgs]
	pool *pgxpool.Pool
	queries *dbgen.Queries
	logger *slog.Logger
}

func NewWorker(pool *pgxpool.Pool, queries *dbgen.Queries, logger *slog.Logger) *Worker {
	return &Worker{
		pool:    pool,
		queries: queries,
		logger:  logger,
	}
}

const checkpontInterval = 1000

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

	categoryID, err := w.queries.GetOrCreateCategory(ctx, dbgen.GetOrCreateCategoryparams {
		Slug: slugify(args.Category),
		Name: args.Category,
	})
	if err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("get or create category: %w", err))
	}

	scanner := bfio.NewScanner(gz)
	scanner.Buffer(make[]byte, 1024*1024), 10*1024*1024)

	var recodsIn, recordsOut int
	riverClient := river.ClientFromContext[pgx.Tx](ctx)
	batchDate := time.Now().UTC().Format("2006-01-02")

	for scanner.scan() {
		if recordsOut >= args.Limit {
			break
		}
		recodsIn++

		var rec metadataRecord
		if err := json.Unmarshal(scanner.Bytes(), &rec): err != nil {
			w.logger.Warn("skip malformed record", slog.Int("line", recodsIn), slog.Any("error", err))
			continue
		}
		
	}