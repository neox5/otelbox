package config

import (
	"fmt"
	"log/slog"
)

// resolveTemplateSources resolves source templates (may reference clock templates)
func (r *Resolver) resolveTemplateSources() error {
	slog.Debug("resolved template sources", "count", len(r.raw.Templates.Sources))

	for _, raw := range r.raw.Templates.Sources {
		name := raw.Name
		if err := r.registerName(name, "template source"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("source template", name)

		resolved := SourceConfig{
			Type: getStringValue(raw.Type),
		}

		// Resolve clock (inline only for templates)
		if raw.Clock != nil {
			clock, clockRef, err := r.resolveClockReference(raw.Clock, ctx)
			if err != nil {
				return err
			}
			resolved.Clock = clock
			resolved.ClockRef = clockRef
		}

		// Copy optional fields
		if raw.Min != nil {
			resolved.Min = *raw.Min
		}
		if raw.Max != nil {
			resolved.Max = *raw.Max
		}

		// Validate
		if resolved.Type == "" {
			return ctx.error("type required")
		}

		r.templateSources[name] = resolved

		slog.Debug("template source", "name", name, "source", resolved)
	}
	return nil
}

// resolveInstanceSources resolves source instances (may reference template/instance clocks)
func (r *Resolver) resolveInstanceSources() error {
	slog.Debug("resolved instance sources", "count", len(r.raw.Instances.Sources))

	for _, raw := range r.raw.Instances.Sources {
		name := raw.Name
		if err := r.registerName(name, "instance source"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("source instance", name)

		resolved := SourceConfig{
			Type: getStringValue(raw.Type),
		}

		// Resolve clock reference if present
		if raw.Clock != nil {
			clock, clockRef, err := r.resolveClockReference(raw.Clock, ctx)
			if err != nil {
				return err
			}
			resolved.Clock = clock
			resolved.ClockRef = clockRef
		}

		// Copy optional fields
		if raw.Min != nil {
			resolved.Min = *raw.Min
		}
		if raw.Max != nil {
			resolved.Max = *raw.Max
		}

		// Validate
		if resolved.Type == "" {
			return ctx.error("type required")
		}

		r.instanceSources[name] = resolved

		slog.Debug("instance source", "name", name, "source", resolved)
	}
	return nil
}

// resolveSourceReference resolves a source reference (supports instance/template/inline)
func (r *Resolver) resolveSourceReference(raw *RawSourceReference, ctx resolveContext) (SourceConfig, *string, error) {
	// Instance reference
	if raw.Instance != "" {
		instance, exists := r.instanceSources[raw.Instance]
		if !exists {
			return SourceConfig{}, nil, ctx.error(fmt.Sprintf("source instance %q not found", raw.Instance))
		}
		// No overrides allowed for instances
		if raw.Template != "" || raw.Type != nil || raw.Clock != nil || raw.Min != nil || raw.Max != nil {
			return SourceConfig{}, nil, ctx.error("cannot override instance source")
		}
		return instance, &raw.Instance, nil // Return instance ref
	}

	// Template reference (with optional overrides)
	if raw.Template != "" {
		template, exists := r.templateSources[raw.Template]
		if !exists {
			return SourceConfig{}, nil, ctx.error(fmt.Sprintf("source template %q not found", raw.Template))
		}

		// Apply overrides
		result := template
		if raw.Type != nil {
			result.Type = *raw.Type
		}
		if raw.Clock != nil {
			clock, clockRef, err := r.resolveClockReference(raw.Clock, ctx)
			if err != nil {
				return SourceConfig{}, nil, err
			}
			result.Clock = clock
			result.ClockRef = clockRef
		}
		if raw.Min != nil {
			result.Min = *raw.Min
		}
		if raw.Max != nil {
			result.Max = *raw.Max
		}
		return result, nil, nil // No instance ref for templates
	}

	// Inline definition
	if raw.Type != nil {
		result := SourceConfig{}
		result.Type = *raw.Type

		// Resolve clock if present
		if raw.Clock != nil {
			clock, clockRef, err := r.resolveClockReference(raw.Clock, ctx)
			if err != nil {
				return SourceConfig{}, nil, err
			}
			result.Clock = clock
			result.ClockRef = clockRef
		}

		// Copy optional fields
		if raw.Min != nil {
			result.Min = *raw.Min
		}
		if raw.Max != nil {
			result.Max = *raw.Max
		}

		// Validate
		if result.Type == "" {
			return SourceConfig{}, nil, ctx.error("source type required")
		}

		return result, nil, nil
	}

	return SourceConfig{}, nil, ctx.error("source must reference instance, template, or provide inline definition")
}
