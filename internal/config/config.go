package config

// Config holds the complete application configuration.
type Config struct {
	Simulation SimulationConfig `yaml:"simulation"`
	Metrics    []MetricConfig   `yaml:"metrics"`
	Export     ExportConfig     `yaml:"export"`
	Settings   SettingsConfig   `yaml:"settings"`
}
