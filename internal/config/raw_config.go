package config

// RawConfig represents unparsed YAML structure
type RawConfig struct {
	Iterators []RawIterator     `yaml:"iterators,omitempty"`
	Templates RawTemplates      `yaml:"templates"`
	Instances RawInstances      `yaml:"instances"`
	Metrics   []RawMetricConfig `yaml:"metrics"`
	Export    RawExportConfig   `yaml:"export"`
	Settings  RawSettingsConfig `yaml:"settings"`
}

// RawTemplates holds all template definitions
type RawTemplates struct {
	Clocks  []RawClockReference  `yaml:"clocks,omitempty"`
	Sources []RawSourceReference `yaml:"sources,omitempty"`
	Values  []RawValueReference  `yaml:"values,omitempty"`
}

// RawInstances holds all instance definitions
type RawInstances struct {
	Clocks  []RawClockReference  `yaml:"clocks,omitempty"`
	Sources []RawSourceReference `yaml:"sources,omitempty"`
	Values  []RawValueReference  `yaml:"values,omitempty"`
}
