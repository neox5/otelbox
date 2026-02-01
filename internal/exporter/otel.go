package exporter

import (
	"context"
	"log/slog"
	"time"

	"github.com/neox5/otelbox/internal/config"
	"github.com/neox5/otelbox/internal/metric"
	"github.com/neox5/simv/value"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

// OTELExporter pushes metrics to an OTEL collector.
type OTELExporter struct {
	config        *config.OTELExportConfig
	meterProvider *sdkmetric.MeterProvider
	meter         otelmetric.Meter
	instruments   []instrument
}

// instrument holds an OTEL observable instrument and its value reference.
type instrument struct {
	counter    otelmetric.Int64ObservableCounter
	gauge      otelmetric.Int64ObservableGauge
	value      *value.Value[int]
	attributes []attribute.KeyValue
}

// NewOTELExporter creates a new OTEL exporter.
func NewOTELExporter(
	cfg *config.OTELExportConfig,
	metrics *metric.Registry,
) (*OTELExporter, error) {
	// Create resource
	res, err := createOTELResource(cfg.Resource)
	if err != nil {
		return nil, err
	}

	// Create meter provider
	meterProvider, err := createMeterProvider(cfg, res)
	if err != nil {
		return nil, err
	}

	// Create meter
	meter := meterProvider.Meter("otelbox")

	// Create exporter
	e := &OTELExporter{
		config:        cfg,
		meterProvider: meterProvider,
		meter:         meter,
	}

	// Register instruments
	if err := registerOTELInstruments(e, metrics); err != nil {
		return nil, err
	}

	return e, nil
}

// Start begins periodic metric export.
// Blocks until context is cancelled, then shuts down gracefully.
func (e *OTELExporter) Start(ctx context.Context) error {
	slog.Info("starting otel exporter",
		"transport", e.config.Transport,
		"endpoint", e.config.GetEndpoint(),
		"push_interval", e.config.Interval.Push,
	)

	// Wait for context cancellation
	<-ctx.Done()

	// Shutdown meter provider
	slog.Info("shutting down otel exporter")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return e.meterProvider.Shutdown(shutdownCtx)
}
