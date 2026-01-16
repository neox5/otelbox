package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// validate orchestrates configuration validation.
func validate(cfg *Config) error {
	// Validate settings configuration
	if err := cfg.Settings.Validate(); err != nil {
		return err
	}

	// Validate export configuration
	if err := cfg.Export.Validate(); err != nil {
		return err
	}

	// Validate simulation configuration
	if err := cfg.Simulation.Validate(); err != nil {
		return err
	}

	// Validate metrics configuration
	if err := validateMetrics(cfg); err != nil {
		return err
	}

	return nil
}
