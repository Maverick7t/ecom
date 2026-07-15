package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/riverqueue/river"

	"github.com/Maverick7t/ecom/internal/jobs/catalog"
	"github.com/Maverick7t/ecom/internal/platform/config"
	"github.com/Maverick7t/ecom/internal/platform/database"
	"github.com/Maverick7t/ecom/internal/platform/logger"
	"github.com/Maverick7t/ecom/internal/platform/queue"
)

func main() {
	_ = godotenv.Load(".env.local")

	source := flag.String("source", "", "path to metadata.jsonl.gz (required)")
	category := flag.String("category", "", "single category to ingest, e.g. Electronics (required)")
	limit := flag.Int("limit", 50000, "max products to ingest")
	flag.Parse()

	if *source == "" || *category == "" {
		slog.Error("usage: seed --source=path/to/metadata.jsonl.gz --category=Electronics --limit=50000")
		os.Exit(1)
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger.NewLogger(cfg)

	db, err := database.NewDB(ctx, cfg, log)
	if err != nil {
		log.Error("database error", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// Insert-only client: no workers registered, never call Start.
	// VERIFY: confirm river.NewClient accepts an empty river.Workers{}
	// for insert-only usage against your installed River version.
	riverClient, err := queue.NewClient(db, river.NewWorkers(), log)
	if err != nil {
		log.Error("river client error", slog.Any("error", err))
		os.Exit(1)
	}

	job, err := riverClient.Insert(ctx, catalog.CatalogIngestionArgs{
		SourcePath: *source,
		Category:   *category,
		Limit:      *limit,
	}, &river.InsertOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	})
	if err != nil {
		log.Error("enqueue catalog_ingestion failed", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("catalog_ingestion enqueued",
		slog.Int64("job_id", job.Job.ID),
		slog.String("category", *category),
		slog.Int("limit", *limit),
	)
}
