package platform

import (
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	TraceProvider *sdktrace.TracerProvider
	Tracer        trace.Tracer
}
