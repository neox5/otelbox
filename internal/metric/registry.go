package metric

import (
	"fmt"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/generator"
)

// Registry holds protocol-agnostic metric definitions.
type Registry struct {
	metrics []Descriptor
}

// New creates a registry from configuration.
func New(cfg *config.Config, gen *generator.Generator) (*Registry, error) {
	var metrics []Descriptor

	for _, metricCfg := range cfg.Metrics {
		val, exists := gen.GetValue(metricCfg.Value)
		if !exists {
			return nil, fmt.Errorf("value %q not found for metric %q", metricCfg.Value, metricCfg.Name.GetPrometheusName())
		}

		metrics = append(metrics, Descriptor{
			PrometheusName: metricCfg.Name.GetPrometheusName(),
			OTELName:       metricCfg.Name.GetOTELName(),
			Type:           MetricType(metricCfg.Type),
			Description:    metricCfg.Description,
			Attributes:     metricCfg.Attributes,
			Value:          val,
		})
	}

	return &Registry{metrics: metrics}, nil
}

// Metrics returns all registered metric descriptors.
func (r *Registry) Metrics() []Descriptor {
	return r.metrics
}
