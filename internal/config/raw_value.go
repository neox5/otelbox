package config

import "go.yaml.in/yaml/v4"

// RawValueReference handles polymorphic value field (instance/template/inline)
type RawValueReference struct {
	Instance   string              `yaml:"instance,omitempty"`
	Template   string              `yaml:"template,omitempty"`
	Source     *RawSourceReference `yaml:"source,omitempty"`
	Transforms []TransformConfig   `yaml:"transforms,omitempty"`
	Reset      ResetConfig         `yaml:"reset,omitempty"`
}

// TransformConfig defines a transform operation
type TransformConfig struct {
	Type string
}

// UnmarshalYAML handles both string and object forms for transforms
func (t *TransformConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try string form first (shorthand)
	var simple string
	if err := value.Decode(&simple); err == nil {
		t.Type = simple
		return nil
	}

	// Fall back to object form
	type transformConfig struct {
		Type string `yaml:"type"`
	}
	var full transformConfig
	if err := value.Decode(&full); err != nil {
		return err
	}
	t.Type = full.Type
	return nil
}

// ResetConfig defines reset behavior
type ResetConfig struct {
	Type  string
	Value int
}

// UnmarshalYAML handles both string and object forms for reset
func (r *ResetConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try string form first (shorthand)
	var simple string
	if err := value.Decode(&simple); err == nil {
		r.Type = simple
		r.Value = 0 // Default value for shorthand
		return nil
	}

	// Fall back to object form
	type resetConfig struct {
		Type  string `yaml:"type"`
		Value int    `yaml:"value"`
	}
	var full resetConfig
	if err := value.Decode(&full); err != nil {
		return err
	}
	r.Type = full.Type
	r.Value = full.Value
	return nil
}
