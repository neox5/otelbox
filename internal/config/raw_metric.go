package config

import "go.yaml.in/yaml/v4"

// RawMetricConfig with polymorphic value field
type RawMetricConfig struct {
	Name        RawMetricNameConfig `yaml:"name"`
	Type        string              `yaml:"type"`
	Description string              `yaml:"description"`
	Value       RawValueReference   `yaml:"value"`
	Attributes  map[string]string   `yaml:"attributes,omitempty"`
}

// DeepCopy creates an independent copy of the metric config
func (m RawMetricConfig) DeepCopy() RawMetricConfig {
	clone := m

	// Deep copy name config
	clone.Name = m.Name.DeepCopy()

	// Deep copy value reference
	clone.Value = m.Value.DeepCopy()

	// Deep copy attributes map
	if len(m.Attributes) > 0 {
		clone.Attributes = make(map[string]string, len(m.Attributes))
		for k, v := range m.Attributes {
			clone.Attributes[k] = v
		}
	}

	return clone
}

// FindPlaceholders implements expandable for RawMetricConfig
func (m *RawMetricConfig) FindPlaceholders() []string {
	found := make(map[string]bool)

	// Scan name fields
	for _, name := range m.Name.FindPlaceholders() {
		found[name] = true
	}

	// Scan attribute keys and values
	for key, value := range m.Attributes {
		for _, name := range extractPlaceholderNames(key) {
			found[name] = true
		}
		for _, name := range extractPlaceholderNames(value) {
			found[name] = true
		}
	}

	// Recursively scan value reference
	for _, name := range m.Value.FindPlaceholders() {
		found[name] = true
	}

	// Convert to slice
	result := make([]string, 0, len(found))
	for name := range found {
		result = append(result, name)
	}
	return result
}

// SubstitutePlaceholders implements expandable for RawMetricConfig
func (m *RawMetricConfig) SubstitutePlaceholders(iteratorValues map[string]string) {
	// Substitute in name
	m.Name.SubstitutePlaceholders(iteratorValues)

	// Substitute in attributes - both keys and values
	if len(m.Attributes) > 0 {
		newAttrs := make(map[string]string, len(m.Attributes))
		for key, value := range m.Attributes {
			newKey := substitutePlaceholders(key, iteratorValues)
			newValue := substitutePlaceholders(value, iteratorValues)
			newAttrs[newKey] = newValue
		}
		m.Attributes = newAttrs
	}

	// Recursively substitute in value reference
	m.Value.SubstitutePlaceholders(iteratorValues)
}

// RawMetricNameConfig supports both short and full forms for metric names
type RawMetricNameConfig struct {
	Simple     string
	Prometheus string
	OTEL       string
}

// DeepCopy creates an independent copy of the metric name config
func (n RawMetricNameConfig) DeepCopy() RawMetricNameConfig {
	// All fields are strings (values), no pointers to copy
	return n
}

// FindPlaceholders scans name fields for placeholders
func (n *RawMetricNameConfig) FindPlaceholders() []string {
	found := make(map[string]bool)

	// Scan all name variants
	for _, name := range extractPlaceholderNames(n.Simple) {
		found[name] = true
	}
	for _, name := range extractPlaceholderNames(n.Prometheus) {
		found[name] = true
	}
	for _, name := range extractPlaceholderNames(n.OTEL) {
		found[name] = true
	}

	// Convert to slice
	result := make([]string, 0, len(found))
	for name := range found {
		result = append(result, name)
	}
	return result
}

// SubstitutePlaceholders replaces placeholders in name fields
func (n *RawMetricNameConfig) SubstitutePlaceholders(iteratorValues map[string]string) {
	n.Simple = substitutePlaceholders(n.Simple, iteratorValues)
	n.Prometheus = substitutePlaceholders(n.Prometheus, iteratorValues)
	n.OTEL = substitutePlaceholders(n.OTEL, iteratorValues)
}

// UnmarshalYAML handles both string and object forms for metric names
func (m *RawMetricNameConfig) UnmarshalYAML(value *yaml.Node) error {
	// Try string form first (short form)
	var simple string
	if err := value.Decode(&simple); err == nil {
		m.Simple = simple
		return nil
	}

	// Try full form (object)
	type nameConfig struct {
		Prometheus string `yaml:"prometheus"`
		OTEL       string `yaml:"otel"`
	}
	var full nameConfig
	if err := value.Decode(&full); err != nil {
		return err
	}
	m.Prometheus = full.Prometheus
	m.OTEL = full.OTEL
	return nil
}

// GetPrometheusName returns the Prometheus metric name
func (m *RawMetricNameConfig) GetPrometheusName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.Prometheus
}

// GetOTELName returns the OTEL metric name
func (m *RawMetricNameConfig) GetOTELName() string {
	if m.Simple != "" {
		return m.Simple
	}
	return m.OTEL
}
