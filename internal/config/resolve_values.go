package config

import (
	"fmt"
	"log/slog"
)

// resolveTemplateValues resolves value templates (may reference source templates)
func (r *Resolver) resolveTemplateValues() error {
	slog.Debug("resolved template values", "count", len(r.raw.Templates.Values))

	for _, raw := range r.raw.Templates.Values {
		name := raw.Name
		if err := r.registerName(name, "template value"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("value template", name)

		resolved := ValueConfig{}

		// Resolve source (inline only for templates)
		if raw.Source != nil {
			source, sourceRef, err := r.resolveSourceReference(raw.Source, ctx)
			if err != nil {
				return err
			}
			resolved.Source = source
			resolved.SourceRef = sourceRef
		}

		// Copy transforms and reset
		resolved.Transforms = raw.Transforms
		resolved.Reset = raw.Reset

		// Validate
		if err := r.validateValue(resolved, ctx); err != nil {
			return err
		}

		r.templateValues[name] = resolved

		slog.Debug("template value", "name", name, "value", resolved)
	}
	return nil
}

// resolveInstanceValues resolves value instances (may reference template/instance sources)
func (r *Resolver) resolveInstanceValues() error {
	slog.Debug("resolved instance values", "count", len(r.raw.Instances.Values))

	for _, raw := range r.raw.Instances.Values {
		name := raw.Name
		if err := r.registerName(name, "instance value"); err != nil {
			return err
		}

		ctx := resolveContext{}.push("value instance", name)

		resolved := ValueConfig{}

		// Resolve source reference if present
		if raw.Source != nil {
			source, sourceRef, err := r.resolveSourceReference(raw.Source, ctx)
			if err != nil {
				return err
			}
			resolved.Source = source
			resolved.SourceRef = sourceRef
		}

		// Copy transforms and reset
		resolved.Transforms = raw.Transforms
		resolved.Reset = raw.Reset

		// Validate
		if err := r.validateValue(resolved, ctx); err != nil {
			return err
		}

		r.instanceValues[name] = resolved

		slog.Debug("instance value", "name", name, "value", resolved)
	}
	return nil
}

// resolveValue resolves a value reference into fully populated ValueConfig.
// Handles three cases: instance reference, template with overrides, inline definition.
func (r *Resolver) resolveValue(raw *RawValueReference, ctx resolveContext) (ValueConfig, error) {
	// Case 1: Instance reference - return stored config
	if raw.Instance != "" {
		instance, exists := r.instanceValues[raw.Instance]
		if !exists {
			return ValueConfig{}, ctx.error(fmt.Sprintf("value instance %q not found", raw.Instance))
		}

		// No overrides allowed for instances
		if raw.Template != "" || raw.Source != nil ||
			len(raw.Transforms) > 0 || raw.Reset.Type != "" {
			return ValueConfig{}, ctx.error("cannot override instance value")
		}

		return instance, nil // Returns full config with references preserved
	}

	// Case 2: Template reference with optional overrides
	if raw.Template != "" {
		template, exists := r.templateValues[raw.Template]
		if !exists {
			return ValueConfig{}, ctx.error(fmt.Sprintf("value template %q not found", raw.Template))
		}

		// Start with template, apply overrides
		result := template

		if raw.Source != nil {
			source, sourceRef, err := r.resolveSourceReference(raw.Source, ctx)
			if err != nil {
				return ValueConfig{}, err
			}
			result.Source = source
			result.SourceRef = sourceRef // Preserve reference tracking
		}

		if len(raw.Transforms) > 0 {
			result.Transforms = raw.Transforms
		}

		if raw.Reset.Type != "" {
			result.Reset = raw.Reset
		}

		return result, nil
	}

	// Case 3: Inline definition - must have source
	if raw.Source == nil {
		return ValueConfig{}, ctx.error("value must reference instance, template, or provide inline source")
	}

	result := ValueConfig{}

	source, sourceRef, err := r.resolveSourceReference(raw.Source, ctx)
	if err != nil {
		return ValueConfig{}, err
	}
	result.Source = source
	result.SourceRef = sourceRef // Preserve reference tracking

	result.Transforms = raw.Transforms
	result.Reset = raw.Reset

	return result, nil
}

// validateValue validates a resolved value config
func (r *Resolver) validateValue(value ValueConfig, ctx resolveContext) error {
	// Source required
	if value.Source.Type == "" {
		return ctx.error("source required")
	}

	// Clock required in source
	if value.Source.Clock.Type == "" {
		return ctx.error("clock required in source")
	}

	return nil
}
