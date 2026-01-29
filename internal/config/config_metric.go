package config

import (
	"regexp"
	"strings"
)

var attributeNameRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// MetricConfig defines a fully resolved metric
type MetricConfig struct {
	PrometheusName string
	OTELName       string
	Type           MetricType
	Description    string
	Value          ValueConfig
	Attributes     map[string]string
}

// MetricType defines the semantic type of a metric
type MetricType string

const (
	MetricTypeCounter MetricType = "counter"
	MetricTypeGauge   MetricType = "gauge"
)

// IsValidAttributeName checks if an attribute name follows conventions
func IsValidAttributeName(name string) bool {
	if len(name) == 0 {
		return false
	}
	if strings.HasPrefix(name, "__") {
		return false
	}
	return attributeNameRegex.MatchString(name)
}
