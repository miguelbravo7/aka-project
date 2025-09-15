package telemetry

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Meter          metric.Meter
}

func NewTelemetry(serviceName string) (*Telemetry, error) {
	return NewTelemetryWithMeterProvider(serviceName, sdkmetric.NewMeterProvider())
}

func NewTelemetryWithMeterProvider(serviceName string, mp *sdkmetric.MeterProvider) (*Telemetry, error) {
	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	otel.SetMeterProvider(mp)

	meter := mp.Meter("aka-project")

	return &Telemetry{
		TracerProvider: tp,
		MeterProvider:  mp,
		Meter:          meter,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) {
	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			log.Printf("failed shutting down tracer: %v", err)
		}
	}
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			log.Printf("failed shutting down meter provider: %v", err)
		}
	}
}
