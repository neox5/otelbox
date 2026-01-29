package config

// RawConfig represents unparsed YAML structure
type RawConfig struct {
	Templates RawTemplates      `yaml:"templates"`
	Instances RawInstances      `yaml:"instances"`
	Metrics   []RawMetricConfig `yaml:"metrics"`
	Export    RawExportConfig   `yaml:"export"`
	Settings  RawSettingsConfig `yaml:"settings"`
}

// RawTemplates holds all template definitions
type RawTemplates struct {
	Clocks  map[string]RawClockReference   `yaml:"clocks,omitempty"`
	Sources map[string]RawSourceReference  `yaml:"sources,omitempty"`
	Values  map[string]RawValueReference   `yaml:"values,omitempty"`
	Metrics map[string]RawMetricDefinition `yaml:"metrics,omitempty"`
}

// RawInstances holds all instance definitions
type RawInstances struct {
	Clocks  map[string]RawClockReference  `yaml:"clocks,omitempty"`
	Sources map[string]RawSourceReference `yaml:"sources,omitempty"`
	Values  map[string]RawValueReference  `yaml:"values,omitempty"`
}
