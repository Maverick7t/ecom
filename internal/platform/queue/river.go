package queue

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// NewClient builds a River client bound to the given pgx pool.
// Workers must be registered by each job package (internal/jobs/*) before
// this is called — see RegisterWorkers.
func NewClient(pool *pgxpool.Pool, workers *river.Workers, logger *slog.Logger) (*river.Client[pgx.Tx], error) {
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5},
		},
		Workers: workers,
		Logger:  slog.New(logger.Handler()),
	})
	if err != nil {
		return nil, fmt.Errorf("create river client: %w", err)
	}
	return client, nil
}

// Start begins processing jobs. Call in a goroutine from main.go.
func Start(ctx context.Context, client *river.Client[pgx.Tx], logger *slog.Logger) {
	if err := client.Start(ctx); err != nil {
		logger.Error("river client stopped", slog.Any("error", err))
	}
}
