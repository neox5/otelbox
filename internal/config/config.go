package config

// Config holds the complete resolved application configuration
type Config struct {
	Instances InstanceRegistry
	Metrics   []MetricConfig
	Export    ExportConfig
	Settings  SettingsConfig
}

// InstanceRegistry holds resolved instance configurations
type InstanceRegistry struct {
	Clocks  map[string]ClockConfig
	Sources map[string]SourceConfig
	Values  map[string]ValueConfig
}
