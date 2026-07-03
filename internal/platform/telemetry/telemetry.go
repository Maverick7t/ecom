package platform

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptrachttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	TraceProvider *sdktrace.TracerProvider
	Tracer 		  trace.Tracer
}

func NewTelemetry(ctx context.Context, cfg *Config, logger *slog.Logger) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTELServiceName),
			semconv.DeploymentEnviourment(cfg.AppEnv),
		),
	)
		))