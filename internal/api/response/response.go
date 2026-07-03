package api

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
