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
	"github.com/Maverick7t/ecom/internal/platform/database/dbgen"
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

	queries := dbgen.New(db)
	workers := river.NewWorkers()
	river.AddWorker(workers, catalog.NewWorker(db, queries, log))

	// Insert-only client: never call Start(). River still requires the
	// job kind registered in Workers to validate Insert() calls.
	riverClient, err := queue.NewClient(db, workers, log)
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
