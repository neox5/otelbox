package config

import (
	"fmt"
)

// Load reads and resolves a YAML configuration file
func Load(path string) (*Config, error) {
	raw, err := Parse(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	cfg, err := Resolve(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config: %w", err)
	}

	return cfg, nil
}
