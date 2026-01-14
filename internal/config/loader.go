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
	// Apply defaults and validate exporters
	if err := validateExport(&cfg.Export); err != nil {
		return err
	}

	// Validate simulation config exists
	if len(cfg.Simulation.Clocks) == 0 {
		return fmt.Errorf("at least one clock must be defined")
	}
	if len(cfg.Simulation.Sources) == 0 {
		return fmt.Errorf("at least one source must be defined")
	}
	if len(cfg.Simulation.Values) == 0 {
		return fmt.Errorf("at least one value must be defined")
	}

	// Validate source clock references
	for srcName, src := range cfg.Simulation.Sources {
		if _, exists := cfg.Simulation.Clocks[src.Clock]; !exists {
			return fmt.Errorf("source %q references unknown clock %q", srcName, src.Clock)
		}
	}

	// Validate value references (source or clone, not both)
	for valName, val := range cfg.Simulation.Values {
		if val.Source == "" && val.Clone == "" {
			return fmt.Errorf("value %q must specify either source or clone", valName)
		}
		if val.Source != "" && val.Clone != "" {
			return fmt.Errorf("value %q cannot specify both source and clone", valName)
		}
		if val.Source != "" {
			if _, exists := cfg.Simulation.Sources[val.Source]; !exists {
				return fmt.Errorf("value %q references unknown source %q", valName, val.Source)
			}
		}
		if val.Clone != "" {
			if _, exists := cfg.Simulation.Values[val.Clone]; !exists {
				return fmt.Errorf("value %q references unknown clone %q", valName, val.Clone)
			}
		}
	}

	// Validate metrics reference valid values
	for _, metric := range cfg.Metrics {
		if _, exists := cfg.Simulation.Values[metric.Value]; !exists {
			return fmt.Errorf("metric %q references unknown value %q", metric.Name, metric.Value)
		}
	}

	return nil
}

// validateExport applies defaults and validates export configuration.
func validateExport(export *ExportConfig) error {
	// Default to Prometheus enabled if no exporters configured
	if export.Prometheus == nil && export.OTEL == nil {
		export.Prometheus = &PrometheusExportConfig{
			Enabled: true,
			Port:    DefaultPrometheusPort,
			Path:    DefaultPrometheusPath,
		}
		return nil
	}

	// Apply Prometheus defaults
	if export.Prometheus != nil {
		if export.Prometheus.Enabled {
			if export.Prometheus.Port == 0 {
				export.Prometheus.Port = DefaultPrometheusPort
			}
			if export.Prometheus.Path == "" {
				export.Prometheus.Path = DefaultPrometheusPath
			}

			if export.Prometheus.Port <= 0 || export.Prometheus.Port > 65535 {
				return fmt.Errorf("invalid prometheus port: %d", export.Prometheus.Port)
			}
		}
	}

	// Apply OTEL defaults
	if export.OTEL != nil && export.OTEL.Enabled {
		if export.OTEL.Endpoint == "" {
			return fmt.Errorf("otel endpoint cannot be empty when enabled")
		}

		// Apply interval defaults
		if export.OTEL.Interval.Read == 0 {
			export.OTEL.Interval.Read = DefaultOTELReadInterval
		}
		if export.OTEL.Interval.Push == 0 {
			export.OTEL.Interval.Push = DefaultOTELPushInterval
		}

		// Apply resource defaults
		if export.OTEL.Resource == nil {
			export.OTEL.Resource = make(map[string]string)
		}
		if _, exists := export.OTEL.Resource["service.name"]; !exists {
			export.OTEL.Resource["service.name"] = DefaultServiceName
		}
		if _, exists := export.OTEL.Resource["service.version"]; !exists {
			export.OTEL.Resource["service.version"] = DefaultServiceVersion
		}
	}

	// Verify at least one exporter enabled
	promEnabled := export.Prometheus != nil && export.Prometheus.Enabled
	otelEnabled := export.OTEL != nil && export.OTEL.Enabled

	if !promEnabled && !otelEnabled {
		return fmt.Errorf("at least one exporter must be enabled")
	}

	return nil
}
