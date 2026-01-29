package config

// RawSettingsConfig holds general application settings
type RawSettingsConfig struct {
	Seed            *uint64                  `yaml:"seed,omitempty"`
	InternalMetrics RawInternalMetricsConfig `yaml:"internal_metrics"`
}

// RawInternalMetricsConfig controls obsbox's self-monitoring metrics
type RawInternalMetricsConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
}
