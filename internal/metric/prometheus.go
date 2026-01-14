package metric

import (
	"fmt"
	"log/slog"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/generator"
	"github.com/prometheus/client_golang/prometheus"
)

// Registry manages Prometheus metrics collection.
type Registry struct {
	registry  *prometheus.Registry
	collector *Collector
}

// New creates a Prometheus registry with a custom collector.
func New(cfg *config.Config, gen *generator.Generator) (*Registry, error) {
	reg := prometheus.NewRegistry()

	var metrics []metricDescriptor

	for _, metricCfg := range cfg.Metrics {
		val, exists := gen.GetValue(metricCfg.Value)
		if !exists {
			return nil, fmt.Errorf("value %q not found for metric %q", metricCfg.Value, metricCfg.Name)
		}

		var valueType prometheus.ValueType
		switch metricCfg.Type {
		case "counter":
			valueType = prometheus.CounterValue
		case "gauge":
			valueType = prometheus.GaugeValue
		default:
			return nil, fmt.Errorf("unsupported metric type: %s", metricCfg.Type)
		}

		desc := prometheus.NewDesc(
			metricCfg.Name,
			metricCfg.Help,
			nil, // no labels for now
			nil, // no constant labels
		)

		metrics = append(metrics, metricDescriptor{
			desc:      desc,
			valueType: valueType,
			value:     val,
		})

		slog.Info("registered metric", "name", metricCfg.Name, "type", metricCfg.Type)
	}

	collector := NewCollector(metrics)
	reg.MustRegister(collector)

	return &Registry{
		registry:  reg,
		collector: collector,
	}, nil
}

// PrometheusRegistry returns the underlying Prometheus registry.
func (r *Registry) PrometheusRegistry() *prometheus.Registry {
	return r.registry
}

// Read triggers value reads (used by OTEL exporter).
func (r *Registry) Read() {
	// Trigger collector to read values
	// This happens automatically during Prometheus scrape,
	// but OTEL needs explicit read triggers
	for _, m := range r.collector.metrics {
		_ = m.value.Value()
	}
}
