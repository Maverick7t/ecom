package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Maverick7t/ecom/internal/platform/config"

	"go.opentelemetry.io/otel"
	stdouttrace "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
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

func NewTelemetry(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Telemetry, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTELServiceName),
			semconv.DeploymentEnvironment(cfg.AppEnv),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	exporter, err := stdouttrace.New()
	if err != nil {
		return nil, fmt.Errorf("create stdout exporter: %w", err)
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
