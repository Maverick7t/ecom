package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/YOURUSERNAME/ecom/internal/api/handlers"
	"github.com/YOURUSERNAME/ecom/internal/api/middleware"
	"github.com/YOURUSERNAME/ecom/internal/platform"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(cfg *platform.Config, db *pgxpool.Pool, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.CORS(cfg))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(chimiddleware.Timeout(30 * time.Second))
	r.Use(middleware.RateLimiter(cfg.RateLimitRPM))

	healthHandler := handlers.NewHealthHandler(db, logger)
	r.Get("/health", healthHandler.Health)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/search", placeholder("search"))
		r.Get("/products/{id}", placeholder("product detail"))
		r.Get("/products/{id}/summary", placeholder("product summary"))
		r.Post("/compare", placeholder("compare"))
		r.Get("/categories", placeholder("categories"))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", placeholder("login"))
			r.Post("/logout", placeholder("logout"))
			r.Get("/me", placeholder("me"))
		})
	})

}
