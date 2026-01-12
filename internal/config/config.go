package config

import "time"

// Config holds the complete application configuration.
type Config struct {
	Server  ServerConfig            `yaml:"server"`
	Clock   ClockConfig             `yaml:"clock"`
	Sources map[string]SourceConfig `yaml:"sources"`
	Values  map[string]ValueConfig  `yaml:"values"`
	Metrics []MetricConfig          `yaml:"metrics"`
}

// ServerConfig defines HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

// ClockConfig defines timing settings.
type ClockConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// SourceConfig defines a simv source.
type SourceConfig struct {
	Type string `yaml:"type"`
	Min  int    `yaml:"min,omitempty"`
	Max  int    `yaml:"max,omitempty"`
}

// ValueConfig defines a simv value with transforms or derivation.
type ValueConfig struct {
	Source     string   `yaml:"source,omitempty"`
	Transforms []string `yaml:"transforms,omitempty"`
	CloneFrom  string   `yaml:"clone_from,omitempty"`
	Wrap       string   `yaml:"wrap,omitempty"`
	ResetValue int      `yaml:"reset_value,omitempty"`
}

// MetricConfig defines a Prometheus metric.
type MetricConfig struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Help  string `yaml:"help"`
	Value string `yaml:"value"`
}
