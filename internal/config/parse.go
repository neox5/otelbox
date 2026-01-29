package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Parse reads and parses a YAML configuration file
func Parse(path string) (*RawConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var raw RawConfig
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := Validate(&raw); err != nil {
		return nil, err
	}

	return &raw, nil
}
