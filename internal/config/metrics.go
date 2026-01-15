package config

import (
	"fmt"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v4"
)

var attributeNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// MetricNameConfig supports both short and full forms for metric names.
type MetricNameConfig struct {
	Simple     string // Short form - same name for both protocols
	Prometheus string // Full form - Prometheus-specific name
	OTEL       string // Full form - OTEL-specific name
}

// UnmarshalYAML handles both string and object forms for metric names.
func (m *MetricNameConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try string form first (short form)
	var simple string
	if err := value.Decode(&simple); err == nil {
		m.Simple = simple
		return nil
	}

	// Try full form (object)
	type nameConfig struct {
		Prometheus string `yaml:"prometheus"`
		OTEL       string `yaml:"otel"`
	}
	var full nameConfig
	if err := value.Decode(&full); err != nil {
		return err
	}
	m.Prometheus = full.Prometheus
	m.OTEL = full.OTEL
	return nil
}

// GetPrometheusName returns the Prometheus metric name.
func (m *MetricNameConfig) GetPrometheusName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.Prometheus
}

// GetOTELName returns the OTEL metric name.
func (m *MetricNameConfig) GetOTELName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.OTEL
}

// MetricConfig defines a metric with protocol-specific naming and attributes.
type MetricConfig struct {
	Name        MetricNameConfig  `yaml:"name"`
	Type        string            `yaml:"type"`
	Description string            `yaml:"description"` // Primary field
	Value       string            `yaml:"value"`
	Attributes  map[string]string `yaml:"attributes"` // Primary field
}

// UnmarshalYAML handles aliasing for description and attributes.
func (m *MetricConfig) UnmarshalYAML(value *yaml.Node) error {
	type metricConfig MetricConfig
	var mc metricConfig

	if err := value.Decode(&mc); err != nil {
		return err
	}

	*m = MetricConfig(mc)

	// Check raw node for aliases
	var raw map[string]interface{}
	if err := value.Decode(&raw); err != nil {
		return err
	}

	// Handle alias: help -> description
	if help, ok := raw["help"].(string); ok && m.Description == "" {
		m.Description = help
	}

	// Handle alias: labels -> attributes
	if labels, ok := raw["labels"].(map[string]interface{}); ok && m.Attributes == nil {
		m.Attributes = make(map[string]string)
		for k, v := range labels {
			if str, ok := v.(string); ok {
				m.Attributes[k] = str
			}
		}
	}

	// Error if both forms specified
	_, hasHelp := raw["help"]
	_, hasDesc := raw["description"]
	if hasHelp && hasDesc {
		return fmt.Errorf("cannot specify both 'help' and 'description'")
	}

	_, hasLabels := raw["labels"]
	_, hasAttrs := raw["attributes"]
	if hasLabels && hasAttrs {
		return fmt.Errorf("cannot specify both 'labels' and 'attributes'")
	}

	return nil
}

// validateMetrics validates metric configuration.
func validateMetrics(cfg *Config) error {
	promEnabled := cfg.Export.Prometheus != nil && cfg.Export.Prometheus.Enabled
	otelEnabled := cfg.Export.OTEL != nil && cfg.Export.OTEL.Enabled

	for _, metric := range cfg.Metrics {
		// Validate value reference
		if _, exists := cfg.Simulation.Values[metric.Value]; !exists {
			return fmt.Errorf("metric %q references unknown value %q", metric.Name.GetPrometheusName(), metric.Value)
		}

		// Validate name requirements based on enabled exporters
		if promEnabled && metric.Name.GetPrometheusName() == "" {
			return fmt.Errorf("metric missing prometheus name but Prometheus exporter enabled")
		}
		if otelEnabled && metric.Name.GetOTELName() == "" {
			return fmt.Errorf("metric missing otel name but OTEL exporter enabled")
		}

		// Validate attribute names and values
		for key, val := range metric.Attributes {
			if !isValidAttributeName(key) {
				return fmt.Errorf("invalid attribute name %q in metric %q", key, metric.Name.GetPrometheusName())
			}
			if val == "" {
				return fmt.Errorf("empty attribute value for key %q in metric %q", key, metric.Name.GetPrometheusName())
			}
		}
	}

	return nil
}

// isValidAttributeName checks if an attribute name follows conventions.
func isValidAttributeName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Cannot start with __
	if strings.HasPrefix(name, "__") {
		return false
	}
	// Must match [a-zA-Z_][a-zA-Z0-9_]*
	return attributeNameRegex.MatchString(name)
}
