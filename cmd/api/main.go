package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/riverqueue/river"

	"github.com/Maverick7t/ecom/internal/api"
	"github.com/Maverick7t/ecom/internal/jobs/catalog"
	"github.com/Maverick7t/ecom/internal/platform/config"
	"github.com/Maverick7t/ecom/internal/platform/database"
	"github.com/Maverick7t/ecom/internal/platform/database/dbgen"
	"github.com/Maverick7t/ecom/internal/platform/logger"
	"github.com/Maverick7t/ecom/internal/platform/queue"
	"github.com/Maverick7t/ecom/internal/platform/telemetry"
)

func main() {
	_ = godotenv.Load(".env.local")

	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config error", slog.Any("error", err))
		os.Exit(1)
	}

	log := logger.NewLogger(cfg)

	tel, err := telemetry.NewTelemetry(ctx, cfg, log)
	if err != nil {
		log.Error("telemetry error", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := database.NewDB(ctx, cfg, log)
	if err != nil {
		log.Error("database error", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// ---- River job queue ----
	queries := dbgen.New(db)

	workers := river.NewWorkers()
	river.AddWorker(workers, catalog.NewWorker(db, queries, log))
	// review, feature, embedding, summary workers registered here as each is built

	riverClient, err := queue.NewClient(db, workers, log)
	if err != nil {
		log.Error("river client error", slog.Any("error", err))
		os.Exit(1)
	}

	riverCtx, riverCancel := context.WithCancel(ctx)
	go queue.Start(riverCtx, riverClient, log)

	// ---- HTTP server ----
	srv := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      api.NewRouter(cfg, db, log),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("server starting", slog.String("addr", srv.Addr), slog.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	_ = srv.Shutdown(shutdownCtx)

	riverCancel()
	if err := riverClient.Stop(shutdownCtx); err != nil {
		log.Error("river client stop error", slog.Any("error", err))
	}

	_ = tel.Shutdown(shutdownCtx)
	log.Info("stopped")
}
