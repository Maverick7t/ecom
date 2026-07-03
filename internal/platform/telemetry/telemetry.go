package platform

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

type Telemetry struct {
	TraceProvider *sdktrace.TracerProvider
	Tracer        trace.Tracer
}

func NewTelemetry(ctx context.Context, cfg *Config, logger *slog.Logger) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTELServiceName),
			semconv.DeploymentEnvironment(cfg.AppEnv),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("telemetry initialized", slog.String("endpoint", cfg.OTELEndpoint))

	return &Telemetry{
		TraceProvider: tp,
		Tracer:        tp.Tracer(cfg.OTELServiceName),
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.TraceProvider.Shutdown(ctx)
}
