package config

import (
	"time"

	"go.yaml.in/yaml/v4"
)

const (
	// Prometheus defaults
	DefaultPrometheusPort = 9090
	DefaultPrometheusPath = "/metrics"

	// OTEL defaults
	DefaultOTELReadInterval = 1 * time.Second
	DefaultOTELPushInterval = 1 * time.Second
	DefaultServiceName      = "obsbox"
	DefaultServiceVersion   = "dev"
)

// Config holds the complete application configuration.
type Config struct {
	Export     ExportConfig     `yaml:"export"`
	Simulation SimulationConfig `yaml:"simulation"`
	Metrics    []MetricConfig   `yaml:"metrics"`
}

// ExportConfig defines how metrics are exposed.
type ExportConfig struct {
	Prometheus *PrometheusExportConfig `yaml:"prometheus,omitempty"`
	OTEL       *OTELExportConfig       `yaml:"otel,omitempty"`
}

// PrometheusExportConfig defines Prometheus pull endpoint settings.
type PrometheusExportConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// OTELExportConfig defines OTEL push settings.
type OTELExportConfig struct {
	Enabled  bool              `yaml:"enabled"`
	Endpoint string            `yaml:"endpoint"`
	Interval IntervalConfig    `yaml:"interval"`
	Resource map[string]string `yaml:"resource,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
}

// IntervalConfig defines read and push intervals for OTEL.
type IntervalConfig struct {
	Read time.Duration
	Push time.Duration
}

// UnmarshalYAML handles both simple (10s) and detailed (read/push) forms.
func (i *IntervalConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try simple duration form first
	var simple time.Duration
	if err := value.Decode(&simple); err == nil {
		i.Read = simple
		i.Push = simple
		return nil
	}

	// Fall back to detailed form
	type intervalConfig struct {
		Read time.Duration `yaml:"read"`
		Push time.Duration `yaml:"push"`
	}
	var detailed intervalConfig
	if err := value.Decode(&detailed); err != nil {
		return err
	}
	i.Read = detailed.Read
	i.Push = detailed.Push
	return nil
}

// SimulationConfig defines the simulation domain configuration.
type SimulationConfig struct {
	Clocks  map[string]ClockConfig  `yaml:"clocks"`
	Sources map[string]SourceConfig `yaml:"sources"`
	Values  map[string]ValueConfig  `yaml:"values"`
}

// ClockConfig defines a clock.
type ClockConfig struct {
	Type     string        `yaml:"type"`
	Interval time.Duration `yaml:"interval"`
}

// SourceConfig defines a simv source.
type SourceConfig struct {
	Type  string `yaml:"type"`
	Clock string `yaml:"clock"`
	Min   int    `yaml:"min,omitempty"`
	Max   int    `yaml:"max,omitempty"`
}

// TransformConfig defines a transform with optional parameters.
type TransformConfig struct {
	Type    string                 `yaml:"type"`
	Options map[string]interface{} `yaml:"options,omitempty"`
}

// ResetConfig defines reset behavior for values.
type ResetConfig struct {
	Type  string `yaml:"type,omitempty"`
	Value int    `yaml:"value,omitempty"`
}

// UnmarshalYAML handles both string and object forms for reset config.
func (r *ResetConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try string form first (short form)
	var shortForm string
	if err := value.Decode(&shortForm); err == nil {
		r.Type = shortForm
		r.Value = 0 // default
		return nil
	}

	// Fall back to full form (object)
	type resetConfig ResetConfig // Avoid recursion
	var fullForm resetConfig
	if err := value.Decode(&fullForm); err != nil {
		return err
	}
	*r = ResetConfig(fullForm)
	return nil
}

// ValueConfig defines a simv value with transforms or derivation.
type ValueConfig struct {
	Source     string            `yaml:"source,omitempty"`
	Clone      string            `yaml:"clone,omitempty"`
	Transforms []TransformConfig `yaml:"transforms,omitempty"`
	Reset      ResetConfig       `yaml:"reset,omitempty"`
}

// MetricConfig defines a Prometheus metric.
type MetricConfig struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Help  string `yaml:"help"`
	Value string `yaml:"value"`
}
