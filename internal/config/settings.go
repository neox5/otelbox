package config

import "fmt"

// SettingsConfig holds general application settings.
type SettingsConfig struct {
	InternalMetrics InternalMetricsConfig `yaml:"internal_metrics"`
}

// InternalMetricsConfig controls obsbox's self-monitoring metrics.
type InternalMetricsConfig struct {
	Enabled bool         `yaml:"enabled"`
	Format  NamingFormat `yaml:"format"`
}

// NamingFormat defines the naming convention for internal metrics.
type NamingFormat string

const (
	// NamingFormatNative uses each exporter's native convention
	// (underscore for Prometheus, dot for OTEL)
	NamingFormatNative NamingFormat = "native"

	// NamingFormatUnderscore forces underscore-separated names
	NamingFormatUnderscore NamingFormat = "underscore"

	// NamingFormatDot forces dot-separated names
	NamingFormatDot NamingFormat = "dot"
)

// Validate applies defaults and validates settings configuration.
func (s *SettingsConfig) Validate() error {
	// Apply defaults
	if s.InternalMetrics.Format == "" {
		s.InternalMetrics.Format = NamingFormatNative
	}

	// Validate format value
	switch s.InternalMetrics.Format {
	case NamingFormatNative, NamingFormatUnderscore, NamingFormatDot:
		return nil
	default:
		return fmt.Errorf("invalid naming format: %s (must be native, underscore, or dot)", s.InternalMetrics.Format)
	}
}
