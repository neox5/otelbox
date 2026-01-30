package config

import (
	"fmt"
	"log/slog"
	"maps"
)

// resolveTemplateMetrics resolves metric templates (may reference value templates)
func (r *Resolver) resolveTemplateMetrics() error {
	// Template metrics not currently used - placeholder for future enhancement
	return nil
}

// resolveMetrics resolves final metrics from raw config
func (r *Resolver) resolveMetrics() ([]MetricConfig, error) {
	var metrics []MetricConfig

	slog.Debug("resolving metrics", "count", len(r.raw.Metrics))

	for _, raw := range r.raw.Metrics {
		promName := raw.Name.GetPrometheusName()
		ctx := resolveContext{}.push("metric", promName)

		metric, err := r.resolveMetric(&raw, ctx)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
		slog.Debug("resolved metric", "metric", metric)
	}

	return metrics, nil
}

// resolveMetric resolves a single metric with template + overrides
func (r *Resolver) resolveMetric(raw *RawMetricConfig, ctx resolveContext) (MetricConfig, error) {
	result := MetricConfig{
		PrometheusName: raw.Name.GetPrometheusName(),
		OTELName:       raw.Name.GetOTELName(),
		Type:           MetricType(raw.Type),
		Description:    raw.Description,
	}

	// Always resolve to full ValueConfig
	value, err := r.resolveValue(&raw.Value, ctx)
	if err != nil {
		return MetricConfig{}, err
	}
	result.Value = value

	// Apply attribute overrides (complete replacement if specified)
	if raw.Attributes != nil {
		result.Attributes = make(map[string]string, len(raw.Attributes))
		maps.Copy(result.Attributes, raw.Attributes)
	}

	// Validate final metric
	if err := r.validateMetric(result, ctx); err != nil {
		return MetricConfig{}, err
	}

	return result, nil
}

// validateMetric validates a resolved metric config
func (r *Resolver) validateMetric(metric MetricConfig, ctx resolveContext) error {
	// Names validated during raw syntax validation

	// Type required
	if metric.Type == "" {
		return ctx.error("type required")
	}

	// Validate type is valid
	if metric.Type != MetricTypeCounter && metric.Type != MetricTypeGauge {
		return ctx.error(fmt.Sprintf("invalid type: %s (must be counter or gauge)", metric.Type))
	}

	// Description required
	if metric.Description == "" {
		return ctx.error("description required")
	}

	// Value must be populated
	if metric.Value.Source.Type == "" {
		return ctx.error("value source required")
	}

	return nil
}

// resolveExport converts raw export config to resolved export config
func resolveExport(raw *RawExportConfig) (ExportConfig, error) {
	result := ExportConfig{}

	// Convert Prometheus config if present
	if raw.Prometheus != nil {
		result.Prometheus = &PrometheusExportConfig{
			Enabled: raw.Prometheus.Enabled,
			Port:    raw.Prometheus.Port,
			Path:    raw.Prometheus.Path,
		}
	}

	// Convert OTEL config if present
	if raw.OTEL != nil {
		result.OTEL = &OTELExportConfig{
			Enabled:   raw.OTEL.Enabled,
			Transport: raw.OTEL.Transport,
			Host:      raw.OTEL.Host,
			Port:      raw.OTEL.Port,
			Interval: IntervalConfig{
				Read: raw.OTEL.Interval.Read,
				Push: raw.OTEL.Interval.Push,
			},
			Resource: copyStringMap(raw.OTEL.Resource),
			Headers:  copyStringMap(raw.OTEL.Headers),
		}
	}

	// Validate converted config
	if err := result.Validate(); err != nil {
		return ExportConfig{}, err
	}

	return result, nil
}

// resolveSettings converts raw settings config to resolved settings config
func resolveSettings(raw *RawSettingsConfig) (SettingsConfig, error) {
	result := SettingsConfig{
		Seed: raw.Seed,
		InternalMetrics: InternalMetricsConfig{
			Enabled: raw.InternalMetrics.Enabled,
			Format:  NamingFormat(raw.InternalMetrics.Format),
		},
	}

	// Validate converted config
	if err := result.Validate(); err != nil {
		return SettingsConfig{}, err
	}

	return result, nil
}

// copyStringMap creates a copy of a string map (handles nil)
func copyStringMap(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}
	dst := make(map[string]string, len(src))
	maps.Copy(dst, src)
	return dst
}
