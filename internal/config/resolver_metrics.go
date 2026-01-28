package config

import "fmt"

// resolveTemplateMetrics resolves metric templates (may reference value templates)
func (r *Resolver) resolveTemplateMetrics() error {
	for name, raw := range r.raw.Templates.Metrics {
		if err := r.registerName(name, "template metric"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("metric template", name)

		resolved := MetricConfig{
			Type: MetricType(raw.Type),
		}

		// Resolve value reference if present
		if raw.Value != nil {
			value, err := r.resolveValueFromReference(raw.Value, ctx)
			if err != nil {
				return err
			}
			resolved.Value = value
		}

		// Copy attributes (can be nil)
		if raw.Attributes != nil {
			resolved.Attributes = make(map[string]string, len(raw.Attributes))
			for k, v := range raw.Attributes {
				resolved.Attributes[k] = v
			}
		}

		// Validate
		if resolved.Type == "" {
			return ctx.error("type required")
		}

		r.templateMetrics[name] = resolved
	}
	return nil
}

// resolveMetrics resolves final metrics from raw config
func (r *Resolver) resolveMetrics() ([]MetricConfig, error) {
	var metrics []MetricConfig

	for _, raw := range r.raw.Metrics {
		promName := raw.Name.GetPrometheusName()
		ctx := resolveContext{}.push("metric", promName)

		metric, err := r.resolveMetric(&raw, ctx)
		if err != nil {
			return nil, err
		}

		metrics = append(metrics, metric)
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
		for k, v := range raw.Attributes {
			result.Attributes[k] = v
		}
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
