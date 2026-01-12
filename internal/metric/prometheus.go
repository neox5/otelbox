package metric

import (
	"fmt"
	"log/slog"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/generator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Registry manages Prometheus metrics and their updates.
type Registry struct {
	registry *prometheus.Registry
	updaters []func()
}

// New creates a Prometheus registry and wires metrics to generator values.
func New(cfg *config.Config, gen *generator.Generator) (*Registry, error) {
	reg := prometheus.NewRegistry()
	factory := promauto.With(reg)

	var updaters []func()

	for _, metricCfg := range cfg.Metrics {
		val, exists := gen.GetValue(metricCfg.Value)
		if !exists {
			return nil, fmt.Errorf("value %q not found for metric %q", metricCfg.Value, metricCfg.Name)
		}

		switch metricCfg.Type {
		case "counter":
			counter := factory.NewCounter(prometheus.CounterOpts{
				Name: metricCfg.Name,
				Help: metricCfg.Help,
			})
			// Counter update function
			updaters = append(updaters, func() {
				current := float64(val.Value())
				counter.Add(current)
			})

		case "gauge":
			gauge := factory.NewGauge(prometheus.GaugeOpts{
				Name: metricCfg.Name,
				Help: metricCfg.Help,
			})
			// Gauge update function
			updaters = append(updaters, func() {
				gauge.Set(float64(val.Value()))
			})

		default:
			return nil, fmt.Errorf("unsupported metric type: %s", metricCfg.Type)
		}

		slog.Info("registered metric", "name", metricCfg.Name, "type", metricCfg.Type)
	}

	return &Registry{
		registry: reg,
		updaters: updaters,
	}, nil
}

// Update executes all metric update functions.
// This is currently unused but will be needed for push-based updates.
func (r *Registry) Update() {
	for _, update := range r.updaters {
		update()
	}
}

// PrometheusRegistry returns the underlying Prometheus registry.
func (r *Registry) PrometheusRegistry() *prometheus.Registry {
	return r.registry
}
