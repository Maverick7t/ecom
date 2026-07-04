package middleware

import (
	"net/http"

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
