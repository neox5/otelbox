package config

import (
	"time"

	"go.yaml.in/yaml/v4"
)

// RawExportConfig defines how metrics are exposed
type RawExportConfig struct {
	Prometheus *RawPrometheusExportConfig `yaml:"prometheus,omitempty"`
	OTEL       *RawOTELExportConfig       `yaml:"otel,omitempty"`
}

// RawPrometheusExportConfig defines Prometheus pull endpoint settings
type RawPrometheusExportConfig struct {
	Enabled bool   `yaml:"enabled"`
	Port    int    `yaml:"port"`
	Path    string `yaml:"path"`
}

// RawOTELExportConfig defines OTEL push settings
type RawOTELExportConfig struct {
	Enabled   bool              `yaml:"enabled"`
	Transport string            `yaml:"transport"`
	Host      string            `yaml:"host"`
	Port      int               `yaml:"port"`
	Interval  RawIntervalConfig `yaml:"interval"`
	Resource  map[string]string `yaml:"resource,omitempty"`
	Headers   map[string]string `yaml:"headers,omitempty"`
}

// RawIntervalConfig defines read and push intervals for OTEL
type RawIntervalConfig struct {
	Read time.Duration
	Push time.Duration
}

// UnmarshalYAML handles both simple (10s) and detailed (read/push) forms
func (i *RawIntervalConfig) UnmarshalYAML(value *yaml.Node) error {
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
