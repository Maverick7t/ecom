package main

import (
	"log/slog"
	"os"

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

}
