package exporter

import (
	"context"
	"fmt"

	"github.com/neox5/obsbox/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// createMeterProvider creates an OTEL meter provider with OTLP exporter.
func createMeterProvider(
	cfg *config.OTELExportConfig,
	res *resource.Resource,
) (*sdkmetric.MeterProvider, error) {
	// Create OTLP HTTP exporter
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		otlpmetrichttp.WithInsecure(), // TODO: Add TLS support later
	}

	// Add custom headers
	if len(cfg.Headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(cfg.Headers))
	}

	exporter, err := otlpmetrichttp.New(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Create periodic reader with push interval
	reader := sdkmetric.NewPeriodicReader(
		exporter,
		sdkmetric.WithInterval(cfg.Interval.Push),
	)

	// Create meter provider
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(reader),
	)

	return meterProvider, nil
}
