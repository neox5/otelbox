package config

import (
	"log/slog"
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

// LogValue implements slog.LogValuer for structured logging
func (m MetricConfig) LogValue() slog.Value {
	// Determine value name
	valueName := "inline"
	if m.Value.SourceRef != nil {
		valueName = "instance:" + *m.Value.SourceRef
	}

	attrs := []slog.Attr{
		slog.String("prometheus_name", m.PrometheusName),
		slog.String("otel_name", m.OTELName),
		slog.String("type", string(m.Type)),
		slog.String("value", valueName),
	}

	// Add attributes as nested group
	if len(m.Attributes) > 0 {
		attrPairs := make([]any, 0, len(m.Attributes)*2)
		for k, v := range m.Attributes {
			attrPairs = append(attrPairs, k, v)
		}
		attrs = append(attrs, slog.Group("attributes", attrPairs...))
	}

	return slog.GroupValue(attrs...)
}
