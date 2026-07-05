package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Maverick7t/ecom/internal/api"
	"github.com/YOURUSERNAME/product-intelligence/internal/platform"
)

func main() {
	ctx := ocntext.Background()
	if err != nil {
		slog.Error("config error", slog.Any("error", err))
		os.Exit(1)
	}

	logger := platform.NewLogger(cfg)

	tel, err := platform.NewTelemetry(ctx, cfg, logger)
	if err != nil {
		logger.Error("telemetry error", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := platform.NewDB(ctx, cfg, logger)
	if err != nil {
		logger.Error("database error", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      api.NewRouter(cfg, db, logger),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

}
