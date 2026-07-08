package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/Maverick7t/ecom/internal/api/response"
	"github.com/jackc/pgx/v5/pgxpool"
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
		response.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"checks": checks,
			"time":   time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	checks["database"] = "healthy"
	response.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"checks": checks,
		"time":   time.Now().UTC().Format(time.RFC3339),
	})
}
