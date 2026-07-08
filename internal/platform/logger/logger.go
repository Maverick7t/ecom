package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/Maverick7t/ecom/internal/platform/config"
)

type contextKey string

const (
	TraceIdKey contextKey = "trace_id"
	UserIDKey  contextKey = "user_id"
)

func NewLogger(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}

	if cfg.IsDev() {
		opts.Level = slog.LevelDebug
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIdKey, traceID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

func TraceIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(TraceIdKey).(string)
	return v
}

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}
