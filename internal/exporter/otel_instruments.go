package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"github.com/neox5/obsbox/internal/metric"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

// registerOTELInstruments creates and registers instruments for all metrics.
func registerOTELInstruments(e *OTELExporter, metrics *metric.Registry) error {
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
			counter, err := e.meter.Int64ObservableCounter(
				m.OTELName,
				otelmetric.WithDescription(m.Description),
			)
			if err != nil {
				return fmt.Errorf("failed to create counter %q: %w", m.OTELName, err)
			}
			inst.counter = counter

		case metric.MetricTypeGauge:
			gauge, err := e.meter.Int64ObservableGauge(
				m.OTELName,
				otelmetric.WithDescription(m.Description),
			)
			if err != nil {
				return fmt.Errorf("failed to create gauge %q: %w", m.OTELName, err)
			}
			inst.gauge = gauge
		}

		instruments = append(instruments, inst)

		// Extract and sort attribute key=value pairs for logging
		attrPairs := make([]string, len(attrs))
		for i, attr := range attrs {
			attrPairs[i] = fmt.Sprintf("%s=%s", attr.Key, attr.Value.AsString())
		}
		sort.Strings(attrPairs)

		slog.Info("registered otel metric",
			"name", m.OTELName,
			"type", m.Type,
			"attributes", fmt.Sprintf("[%s]", attrPairs))
	}

	e.instruments = instruments

	// Register callback
	if err := registerOTELCallback(e); err != nil {
		return err
	}

	return nil
}

// registerOTELCallback registers the observation callback for all instruments.
func registerOTELCallback(e *OTELExporter) error {
	// Collect all observables for callback registration
	var observables []otelmetric.Observable
	for _, inst := range e.instruments {
		if inst.counter != nil {
			observables = append(observables, inst.counter)
		}
		if inst.gauge != nil {
			observables = append(observables, inst.gauge)
		}
	}

	// Register callback with attributes
	_, err := e.meter.RegisterCallback(
		func(ctx context.Context, observer otelmetric.Observer) error {
			slog.Debug("otel push", "metrics", len(e.instruments))

			for _, inst := range e.instruments {
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
		return fmt.Errorf("failed to register callback: %w", err)
	}

	return nil
}
