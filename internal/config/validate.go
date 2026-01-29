package config

import (
	"fmt"
)

// Validate performs syntactic validation on raw config
func Validate(raw *RawConfig) error {
	return validateRawSyntax(raw)
}

// validateRawSyntax performs basic syntactic validation on raw config
func validateRawSyntax(raw *RawConfig) error {
	// Validate at least one metric defined
	if len(raw.Metrics) == 0 {
		return fmt.Errorf("at least one metric must be defined")
	}

	// Validate metric names
	for i, metric := range raw.Metrics {
		promName := metric.Name.GetPrometheusName()
		otelName := metric.Name.GetOTELName()

		if promName == "" && otelName == "" {
			return fmt.Errorf("metric at index %d: name cannot be empty", i)
		}

		if metric.Type == "" {
			return fmt.Errorf("metric %q: type cannot be empty", promName)
		}

		if metric.Description == "" {
			return fmt.Errorf("metric %q: description cannot be empty", promName)
		}
	}

	return nil
}
