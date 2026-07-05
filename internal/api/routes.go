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

	r.Route("/users", func(r chi.Router) {
		r.Get("/saved-products", placeholder("saved products"))
		r.Post("/saved-products", placeholder("save product"))
		r.Delete("/saved-products/{id}", placeholder("unsave product"))
		r.Get("/saved-searches", placeholder("saved searches"))
			r.Post("/saved-searches", placeholder("save search"))
			r.Delete("/saved-searches/{id}", placeholder("delete search"))
			r.Get("/collections", placeholder("collections"))
			r.Post("/collections", placeholder("create collection"))
			r.Post("/collections/{id}/products", placeholder("add to collection"))
			r.Get("/chat", placeholder("chat history"))
			r.Post("/chat", placeholder("chat"))
			r.Get("/cart", placeholder("cart"))
			r.Post("/cart", placeholder("add to cart"))
			r.Delete("/cart/{productId}", placeholder("remove from cart"))
			r.Post("/orders", placeholder("place order"))
			r.Get("/orders", placeholder("order history"))
			r.Get("/notifications", placeholder("notifications"))
			r.Get("/notifications/stream", placeholder("notifications SSE"))
		})
	})

	return r



}
