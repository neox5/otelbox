package config

import "go.yaml.in/yaml/v4"

// RawMetricConfig with polymorphic value field
type RawMetricConfig struct {
	Name        RawMetricNameConfig `yaml:"name"`
	Type        string              `yaml:"type"`
	Description string              `yaml:"description"`
	Value       RawValueReference   `yaml:"value"`
	Attributes  map[string]string   `yaml:"attributes,omitempty"`
}

// RawMetricDefinition - always full object, no self-reference
type RawMetricDefinition struct {
	Type       string             `yaml:"type"`
	Value      *RawValueReference `yaml:"value,omitempty"`
	Attributes map[string]string  `yaml:"attributes,omitempty"`
}

// RawMetricNameConfig supports both short and full forms for metric names
type RawMetricNameConfig struct {
	Simple     string
	Prometheus string
	OTEL       string
}

// UnmarshalYAML handles both string and object forms for metric names
func (m *RawMetricNameConfig) UnmarshalYAML(value *yaml.Node) error {
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

// GetPrometheusName returns the Prometheus metric name
func (m *RawMetricNameConfig) GetPrometheusName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.Prometheus
}

// GetOTELName returns the OTEL metric name
func (m *RawMetricNameConfig) GetOTELName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.OTEL
}
