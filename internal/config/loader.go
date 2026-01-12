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

// validate checks configuration consistency.
func validate(cfg *Config) error {
	// Validate server
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Server.Path == "" {
		return fmt.Errorf("server path cannot be empty")
	}

	// Validate clock
	if cfg.Clock.Interval <= 0 {
		return fmt.Errorf("clock interval must be positive")
	}

	// Validate sources exist
	if len(cfg.Sources) == 0 {
		return fmt.Errorf("at least one source must be defined")
	}

	// Validate values reference valid sources or clones
	for name, val := range cfg.Values {
		if val.Source != "" {
			if _, exists := cfg.Sources[val.Source]; !exists {
				return fmt.Errorf("value %q references unknown source %q", name, val.Source)
			}
		}
		if val.CloneFrom != "" {
			if _, exists := cfg.Values[val.CloneFrom]; !exists {
				return fmt.Errorf("value %q references unknown clone_from %q", name, val.CloneFrom)
			}
		}
		if val.Source == "" && val.CloneFrom == "" {
			return fmt.Errorf("value %q must specify either source or clone_from", name)
		}
	}

	// Validate metrics reference valid values
	for _, metric := range cfg.Metrics {
		if _, exists := cfg.Values[metric.Value]; !exists {
			return fmt.Errorf("metric %q references unknown value %q", metric.Name, metric.Value)
		}
	}

	return nil
}
