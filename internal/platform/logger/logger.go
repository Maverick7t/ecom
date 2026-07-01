package platform

import (
	"log/slog"
	"os"
)

type contextKey string

const (
	TraceIdKey contextKey = "trace_id"
	UserIDKey  contextKey = "user_id"
)

func NewLogger(cfg *Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if cfg.IsDev() {
		opts.Level = slog.LevelDebug
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
