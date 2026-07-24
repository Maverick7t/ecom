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
	reviewsSource := flag.String("reviews-source", "", "path to review.jsonl.gz for this category (optional — omit to skip reviews_ingestion)")
	flag.Parse()

	if *source == "" || *category == "" {
		slog.Error("usage: seed --source=path/to/metadata.jsonl.gz --category=Electronics --limit=50000 [--reviews-source=path/to/review.jsonl.gz]")
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
	//
	// Note: only catalog.Worker needs registering here. The downstream
	// reviews_ingestion job is inserted at runtime by catalog_ingestion
	// itself via river.ClientFromContext, which resolves to the River
	// client running inside cmd/api (the one with Start() called and
	// reviews.Worker registered) — not this insert-only client.
	riverClient, err := queue.NewClient(db, workers, log)
	if err != nil {
		log.Error("river client error", slog.Any("error", err))
		os.Exit(1)
	}

	if *reviewsSource == "" {
		log.Warn("no --reviews-source provided — reviews_ingestion will be skipped after catalog_ingestion completes")
	}

	job, err := riverClient.Insert(ctx, catalog.CatalogIngestionArgs{
		SourcePath:        *source,
		Category:          *category,
		Limit:             *limit,
		ReviewsSourcePath: *reviewsSource,
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
		slog.Bool("reviews_ingestion_will_follow", *reviewsSource != ""),
	)
}
