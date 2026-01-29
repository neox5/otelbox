package config

import (
	"fmt"
)

// resolveTemplateClocks resolves clock templates (no dependencies)
func (r *Resolver) resolveTemplateClocks() error {
	for name, raw := range r.raw.Templates.Clocks {
		if err := r.registerName(name, "template clock"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("clock template", name)

		resolved := ClockConfig{
			Type:     getStringValue(raw.Type),
			Interval: raw.Interval,
		}

		// Validate
		if resolved.Type == "" {
			return ctx.error("type required")
		}
		if resolved.Interval == 0 {
			return ctx.error("interval required")
		}

		r.templateClocks[name] = resolved
	}
	return nil
}

// resolveInstanceClocks resolves clock instances
func (r *Resolver) resolveInstanceClocks() error {
	for name, raw := range r.raw.Instances.Clocks {
		if err := r.registerName(name, "instance clock"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("clock instance", name)

		resolved := ClockConfig{
			Type:     getStringValue(raw.Type),
			Interval: raw.Interval,
		}

		// Validate
		if resolved.Type == "" {
			return ctx.error("type required")
		}
		if resolved.Interval == 0 {
			return ctx.error("interval required")
		}

		r.instanceClocks[name] = resolved
	}
	return nil
}

// resolveClockReference resolves a clock reference (supports instance/template/inline)
func (r *Resolver) resolveClockReference(raw *RawClockReference, ctx resolveContext) (ClockConfig, *string, error) {
	// Instance reference
	if raw.Instance != "" {
		instance, exists := r.instanceClocks[raw.Instance]
		if !exists {
			return ClockConfig{}, nil, ctx.error(fmt.Sprintf("clock instance %q not found", raw.Instance))
		}
		// No overrides allowed for instances
		if raw.Template != "" || raw.Type != nil || raw.Interval != 0 {
			return ClockConfig{}, nil, ctx.error("cannot override instance clock")
		}
		return instance, &raw.Instance, nil
	}

	// Template reference (with optional overrides)
	if raw.Template != "" {
		template, exists := r.templateClocks[raw.Template]
		if !exists {
			return ClockConfig{}, nil, ctx.error(fmt.Sprintf("clock template %q not found", raw.Template))
		}

		// Apply overrides
		result := template
		if raw.Type != nil {
			result.Type = *raw.Type
		}
		if raw.Interval != 0 {
			result.Interval = raw.Interval
		}
		return result, nil, nil
	}

	// Inline definition
	if raw.Type != nil {
		resolved := ClockConfig{
			Type:     *raw.Type,
			Interval: raw.Interval,
		}

		// Validate
		if resolved.Type == "" {
			return ClockConfig{}, nil, ctx.error("clock type required")
		}
		if resolved.Interval == 0 {
			return ClockConfig{}, nil, ctx.error("clock interval required")
		}

		return resolved, nil, nil
	}

	return ClockConfig{}, nil, ctx.error("clock must reference instance, template, or provide inline definition")
}
