package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/YOURUSERNAME/product-intelligence/internal/api"
)

type HealthHandler struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewHealthHandler(db *pgxpool.Pool, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{db: db, logger: logger}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]string{}

	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("health db ping failed", slog.Any("error", err))
		checks["database"] = "unhealthy"
		api.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":  "unhealthy",
			"checks":  checks,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
		return
	}