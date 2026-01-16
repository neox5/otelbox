package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/metric"
	"github.com/neox5/simv/value"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Internal metric name definitions (both formats hardcoded)
const (
	otelExportsTotalUnderscore        = "obsbox_otel_exports_total"
	otelExportsTotalDot               = "obsbox.otel.exports.total"
	otelExportFailuresTotalUnderscore = "obsbox_otel_export_failures_total"
	otelExportFailuresTotalDot        = "obsbox.otel.export.failures.total"
	otelExportDurationUnderscore      = "obsbox_otel_export_duration_seconds"
	otelExportDurationDot             = "obsbox.otel.export.duration.seconds"
)

// OTELExporter pushes metrics to an OTEL collector.
type OTELExporter struct {
	config        *config.OTELExportConfig
	meterProvider *sdkmetric.MeterProvider
	meter         otelmetric.Meter
	instruments   []instrument
	cancelFunc    context.CancelFunc

	// Internal metrics
	exportsTotal        otelmetric.Int64Counter
	exportFailuresTotal otelmetric.Int64Counter
	exportDuration      otelmetric.Float64Histogram
}

// instrument holds an OTEL observable instrument and its value reference.
type instrument struct {
	counter    otelmetric.Int64ObservableCounter
	gauge      otelmetric.Int64ObservableGauge
	value      value.Value[int]
	attributes []attribute.KeyValue
}

// NewOTELExporter creates a new OTEL exporter.
func NewOTELExporter(
	cfg *config.OTELExportConfig,
	metrics *metric.Registry,
	internalMetricsEnabled bool,
	namingFormat config.NamingFormat,
) (*OTELExporter, error) {
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

	// Create meter
	meter := meterProvider.Meter("obsbox")

	e := &OTELExporter{
		config:        cfg,
		meterProvider: meterProvider,
		meter:         meter,
	}

	// Register internal metrics if enabled
	if internalMetricsEnabled {
		// Select names based on format
		exportsName := otelExportsTotalDot
		failuresName := otelExportFailuresTotalDot
		durationName := otelExportDurationDot

		if namingFormat == config.NamingFormatUnderscore {
			exportsName = otelExportsTotalUnderscore
			failuresName = otelExportFailuresTotalUnderscore
			durationName = otelExportDurationUnderscore
		}
		// native format uses dot for OTEL

		e.exportsTotal, err = meter.Int64Counter(exportsName,
			otelmetric.WithDescription("Total number of export attempts"))
		if err != nil {
			return nil, fmt.Errorf("failed to create exports counter: %w", err)
		}

		e.exportFailuresTotal, err = meter.Int64Counter(failuresName,
			otelmetric.WithDescription("Total number of failed exports"))
		if err != nil {
			return nil, fmt.Errorf("failed to create failures counter: %w", err)
		}

		e.exportDuration, err = meter.Float64Histogram(durationName,
			otelmetric.WithDescription("Duration of export operations in seconds"))
		if err != nil {
			return nil, fmt.Errorf("failed to create duration histogram: %w", err)
		}

		slog.Info("registered otel internal metrics",
			"format", namingFormat,
			"exports_total", exportsName,
			"export_failures_total", failuresName,
			"export_duration", durationName)
	}

	// Register instruments for each metric
	var instruments []instrument
	for _, m := range metrics.Metrics() {
		// Convert attributes map to OTEL attributes
		attrs := make([]attribute.KeyValue, 0, len(m.Attributes))
		for key, val := range m.Attributes {
			attrs = append(attrs, attribute.String(key, val))
		}

		inst := instrument{
			value:      m.Value,
			attributes: attrs,
		}

		switch m.Type {
		case metric.MetricTypeCounter:
			counter, err := meter.Int64ObservableCounter(
				m.OTELName,
				otelmetric.WithDescription(m.Description),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create counter %q: %w", m.OTELName, err)
			}
			inst.counter = counter

		case metric.MetricTypeGauge:
			gauge, err := meter.Int64ObservableGauge(
				m.OTELName,
				otelmetric.WithDescription(m.Description),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create gauge %q: %w", m.OTELName, err)
			}
			inst.gauge = gauge
		}

		instruments = append(instruments, inst)
		slog.Info("registered otel metric",
			"name", m.OTELName,
			"type", m.Type,
			"attributes", len(attrs))
	}

	e.instruments = instruments

	// Collect all observables for callback registration
	var observables []otelmetric.Observable
	for _, inst := range instruments {
		if inst.counter != nil {
			observables = append(observables, inst.counter)
		}
		if inst.gauge != nil {
			observables = append(observables, inst.gauge)
		}
	}

	// Register callback with attributes
	_, err = meter.RegisterCallback(
		func(ctx context.Context, observer otelmetric.Observer) error {
			for _, inst := range instruments {
				val := int64(inst.value.Value()) // Triggers reset_on_read if configured
				if inst.counter != nil {
					observer.ObserveInt64(inst.counter, val,
						otelmetric.WithAttributes(inst.attributes...))
				}
				if inst.gauge != nil {
					observer.ObserveInt64(inst.gauge, val,
						otelmetric.WithAttributes(inst.attributes...))
				}
			}
			return nil
		},
		observables...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register callback: %w", err)
	}

	return e, nil
}

// Start begins periodic metric export.
func (e *OTELExporter) Start(ctx context.Context) error {
	slog.Info("starting otel exporter",
		"endpoint", e.config.Endpoint,
		"push_interval", e.config.Interval.Push,
	)

	// Create cancellable context
	readCtx, cancel := context.WithCancel(ctx)
	e.cancelFunc = cancel

	// Periodic reader handles push automatically
	// Instrument each push cycle if internal metrics enabled
	if e.exportsTotal != nil {
		// Note: The actual instrumentation of the export cycle happens
		// inside the periodic reader's callback, which we don't control directly.
		// For now, we'll track this through the reader's success/failure,
		// which would require modifying the SDK or wrapping the exporter.
		// This is a limitation we'll document and potentially address later.
		slog.Info("otel internal metrics registered but export cycle instrumentation limited by SDK")
	}

	// Wait for context cancellation
	<-readCtx.Done()
	return nil
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
