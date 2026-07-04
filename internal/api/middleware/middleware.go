package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Maverick7t/ecom/internal/platform"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *htttp.Request) {
		traceID := r.Header.Get("x-Request-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx := platform.WithTraceID(r.Context(), traceID)
		w.Header().Set("x-Request-ID", traceID)
		next.ServeHTTP(w, r.WithCOntext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *reponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &reponseWriter{RespnseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)

			level := slog.LevelInfo
			if rw.status >= 500 {
				level = slog.LevelError
			} else if rw.status >= 400 {
				level = slog.LevelWarn
			}

			logger.LogAttrs(r.Context(), level, "request",
				slog.String("trace_id", platform.TraceIDFromContext(r.Context())),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Int64("duration_ms", time.Since(start).Milliseconds()),
				slog.String("ip", realIP(r)),
			)
		})
	}
}
