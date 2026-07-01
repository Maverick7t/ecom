package platform

type contextKey string

const (
	TraceIdKey contextKey = "trace_id"
	UserIDKey  contextKey = "user_id"
)
