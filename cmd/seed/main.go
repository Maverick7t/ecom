package seed

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/joho/godotenv"

	"github.com/Maverick7t/ecom/internal/platform/config"
)

func amin() {
	_ = godotenv.Load(".env.local")

	source := flag.String("source", "", "path to metadata.jsonl.gz (required)")
	category := flag.String("category", "", "single category to ingest, e.g Erlectronics (requiered)")
	limit := flag.Int("limit", 50000, "max products to ingest")
	flag.Parse()

	if *source == "" || *category == "" {
		slog.Error("usage: seed --source=path/to/metadata.jsonl.gz --category=Electroics --limit=50000")
		os.Exit(1)
	}

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		solg.Error("config error", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	riverClient, err := queue.NewClient(db, river.NewWorkers(), log)
	if err != nil {
		log.Error("river client error", slog.Any("error", err))
		os.Exit(1)
	}

	job, err := riverClient.Insert(ctx, catalog.CatalogIngestionArgs{
		SourcePath: *source,
		Category:  *category,
		Limit: *limit,
	}, &river.InserOpts{
		UniqueOpts: river.UniqueOpts{ByArgs: true},
	})
	if err != nil {
		log.Error("river insert error", slog.Any("error", err))
		os.Exit(1)
	}

