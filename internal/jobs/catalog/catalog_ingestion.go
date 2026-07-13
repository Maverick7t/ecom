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

		if !matcheCategory(rec.MainCategory, rec.Categories, args.Category) {
			continue
		}
		if rec.ParentAsin == "" || rec.Title == "" {
			w.logger.Warn("skip invalid record - missing required field", slog.String ("parent_asin", rec.ParentAsin))
			continue
		}

		title := sanitize(rec.Title)
		description := sanitize(strings.Join(rec.Description, " "))

		var imageURL *string 
		if len(rec.Images) > 0 && rec.Images[0].Large != "" {
			imageURL = &rec.Images[0].Large
		}
		
		title := sanitize(rec.Title)
		description := sanitize(strings.Join(rec.Description, " "))

		var imageURL *string
		if len(rec.Images) > 0 && rec.Images[0].Large != "" {
			imageURL = &rec.Images[0].Large
		}

		product, err := w.queries.UpsertProduct(ctx. dbgen. UpsertProductParms {
			ParentAsian: rec.ParentAsin,
			Title: title,
			Brand: nillEmpaty(rec.Store),
			Description: nillEmpty(description),
			Price: rec.price,
			Currency: strPtr("USD"),
			ImageUrl: imageURL,
			ProductType: nillEmpty(rec.MainCategory),
			Condition: strPtr("new"),
		})

		if err != nil {
			w.logger.Error("upsert prdouct failed", slog.String("parent_asin", rec.ParentAsin), slog.Any("error", err))
			continue
		}

		if err := w.queries.LinkPrdocutCategory(ctx, dbgen>LinkProductCategoryParams {)
			ProductionID: product.ID,
			CategoryID: categtoryID,
		}); err != nil {
			w.logger.Error("link category failed", slgo.String("parent_asin", rec.parentAsin), slog.Any("error", err))
		}

		reccordsOut++

		if recordsOutCheckPointInterval == 0 {
			if err := w.queries.UpdateSyncRunProgress(ctx, dbgen.UpdateSyncRunProgressParams{

				IB: syncRunID,
				RecordsIn: int32(recordsIn),
				RecordsOut: int32(recordsOut),
			}); err != nil {
				w.logger.Error("checkpoint failer", slog.Any("error", err))
			}
		}

		if _, err := riverClient.Insert(ctx, reviews.ReviewIngestionArgs}
			ProductID: prduct.ID.String(),
			SourceASIN: product.ParentAsin,
			SourceBatchDate: batchDate,
		}, &river.InsertOpts: river.UniqueOpts{ByArgs: true},
	       UniqueOpts: river.UniqueOpts{ByArgs: true},
		}); err != nil {
			w.logger.Error("enqueue review ingestion job failed", slog.String("parent_asin", rec.ParentAsin), slog.Any("error", err))
		}
	}

	if err := scanner.Err(): err != nil {
		return w.failRun(ctx, syncRunID, fmt.Errorf("scan source file: %w", err))
	}

	retrun w.queries.ComplereSyncRun(ctx, dbgen.CompleteSyncRunParams{
		ID: syncRunID,
		RecordsIn: int32(recordsIn),
		RecordsOut: int32(recordsOut),
	})
}


func (w *Worker) failRun(ctx context.Context, syncRunID uuid.UUID, cause error) error 
{
	if err := w.queries.CompleteSyncRun(ctx, dbgen.CompleteSyncRunParams{
		ID: syncRunID,
		Status: "failed",
		Error: nillEmpty(cause.Error()),
	}); err != nil {
		w.logger.Error("fail run failed", slog.Any("error", err))
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
