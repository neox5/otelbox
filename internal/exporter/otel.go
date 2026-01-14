package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/metric"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// OTELExporter pushes metrics to an OTEL collector.
type OTELExporter struct {
	config        *config.OTELExportConfig
	metrics       *metric.Registry
	meterProvider *sdkmetric.MeterProvider
	cancelFunc    context.CancelFunc
}

// NewOTELExporter creates a new OTEL exporter.
func NewOTELExporter(cfg *config.OTELExportConfig, metrics *metric.Registry) (*OTELExporter, error) {
	// Create resource with configured attributes
	attrs := make([]attribute.KeyValue, 0, len(cfg.Resource))
	for k, v := range cfg.Resource {
		attrs = append(attrs, attribute.String(k, v))
	}
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(attrs...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

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

	return &OTELExporter{
		config:        cfg,
		metrics:       metrics,
		meterProvider: meterProvider,
	}, nil
}

// Start begins periodic metric export.
func (e *OTELExporter) Start(ctx context.Context) error {
	slog.Info("starting otel exporter",
		"endpoint", e.config.Endpoint,
		"read_interval", e.config.Interval.Read,
		"push_interval", e.config.Interval.Push,
	)

	// Create cancellable context for read loop
	readCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	// Start periodic value reading
	ticker := time.NewTicker(e.config.Interval.Read)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Trigger value reads (will happen through collector scrape)
			e.metrics.Read()
		case <-readCtx.Done():
			return nil
		}
	}
}

// Stop gracefully stops the exporter.
func (e *OTELExporter) Stop() error {
	slog.Info("shutting down otel exporter")

	if e.cancelFunc != nil {
		e.cancelFunc()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.meterProvider.Shutdown(ctx)
}
