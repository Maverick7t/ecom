package response

import (
	"encoding/json"
	"net/http"

	"github.com/Maverick7t/ecom/internal/platform/logger"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

const (
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeInternal      = "INTERNAL_ERROR"
	ErrorCodeBadRequest    = "BAD_REQUEST"
	ErrorCodeUnauthorized  = "UNAUTHORIZED"
	ErrorCodeAIUnavailable = "AI_UNAVAILABLE"
	ErrorCodeRateLimited   = "RATE_LIMITED"
)

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	traceID := logger.TraceIDFromContext(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{Code: code, Message: message, TraceID: traceID})
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func NotFound(w http.ResponseWriter, r *http.Request, entity string) {
	writeError(w, r, http.StatusNotFound, ErrorCodeNotFound, entity+" not found")
}

func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	writeError(w, r, http.StatusBadRequest, ErrorCodeBadRequest, message)
}

func Unauthorized(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "authentication required")
}

func InternalError(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusInternalServerError, ErrorCodeInternal, "an unexpected error occurred")
}

func AIUnavailable(w http.ResponseWriter, r *http.Request) {
	writeError(w, r, http.StatusServiceUnavailable, ErrorCodeAIUnavailable, "AI service is currently unavailable")
}
