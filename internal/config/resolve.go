package config

import (
	"fmt"
	"strings"
)

// Resolver handles template and instance resolution
type Resolver struct {
	raw *RawConfig

	// Namespace tracking (all entity names)
	registeredNames map[string]string // name -> entity type

	// Resolved templates (temporary, discarded after final config built)
	templateClocks  map[string]ClockConfig
	templateSources map[string]SourceConfig
	templateValues  map[string]ValueConfig
	templateMetrics map[string]MetricConfig

	// Resolved instances (kept in final config)
	instanceClocks  map[string]ClockConfig
	instanceSources map[string]SourceConfig
	instanceValues  map[string]ValueConfig
}

// newResolver creates a new resolver
func newResolver(raw *RawConfig) *Resolver {
	return &Resolver{
		raw:             raw,
		registeredNames: make(map[string]string),
		templateClocks:  make(map[string]ClockConfig),
		templateSources: make(map[string]SourceConfig),
		templateValues:  make(map[string]ValueConfig),
		templateMetrics: make(map[string]MetricConfig),
		instanceClocks:  make(map[string]ClockConfig),
		instanceSources: make(map[string]SourceConfig),
		instanceValues:  make(map[string]ValueConfig),
	}
}

// Resolve performs hierarchical template and instance resolution and builds final config
func Resolve(raw *RawConfig) (*Config, error) {
	// Phase 0: Expand iterators (if present)
	if len(raw.Iterators) > 0 {
		registry, err := buildIteratorRegistry(raw.Iterators)
		if err != nil {
			return nil, fmt.Errorf("failed to build iterator registry: %w", err)
		}

		// Expand template clocks
		raw.Templates.Clocks, err = expandClocks(raw.Templates.Clocks, registry)
		if err != nil {
			return nil, fmt.Errorf("failed to expand template clocks: %w", err)
		}

		// Expand instance clocks
		raw.Instances.Clocks, err = expandClocks(raw.Instances.Clocks, registry)
		if err != nil {
			return nil, fmt.Errorf("failed to expand instance clocks: %w", err)
		}

		// Clear iterators - they've been consumed
		raw.Iterators = nil
	}

	r := newResolver(raw)

	// Phase 1: Resolve templates hierarchically
	if err := r.resolveTemplateClocks(); err != nil {
		return nil, err
	}
	if err := r.resolveTemplateSources(); err != nil {
		return nil, err
	}
	if err := r.resolveTemplateValues(); err != nil {
		return nil, err
	}
	if err := r.resolveTemplateMetrics(); err != nil {
		return nil, err
	}

	// Phase 2: Resolve instances hierarchically
	if err := r.resolveInstanceClocks(); err != nil {
		return nil, err
	}
	if err := r.resolveInstanceSources(); err != nil {
		return nil, err
	}
	if err := r.resolveInstanceValues(); err != nil {
		return nil, err
	}

	// Phase 3: Resolve metrics
	resolvedMetrics, err := r.resolveMetrics()
	if err != nil {
		return nil, err
	}

	// Phase 4: Resolve export config
	resolvedExport, err := resolveExport(&r.raw.Export)
	if err != nil {
		return nil, err
	}

	// Phase 5: Resolve settings config
	resolvedSettings, err := resolveSettings(&r.raw.Settings)
	if err != nil {
		return nil, err
	}

	// Phase 6: Build final config (templates discarded, instances kept)
	return &Config{
		Instances: InstanceRegistry{
			Clocks:  r.instanceClocks,
			Sources: r.instanceSources,
			Values:  r.instanceValues,
		},
		Metrics:  resolvedMetrics,
		Export:   resolvedExport,
		Settings: resolvedSettings,
	}, nil
}

// registerName validates namespace uniqueness and registers the name
func (r *Resolver) registerName(name string, entityType string) error {
	if existingType, exists := r.registeredNames[name]; exists {
		return fmt.Errorf("name %q already used by %s, cannot reuse for %s",
			name, existingType, entityType)
	}
	r.registeredNames[name] = entityType
	return nil
}

// getStringValue safely dereferences a string pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// resolveContext tracks resolution path for error messages
type resolveContext []string

func (ctx resolveContext) push(component, name string) resolveContext {
	return append(ctx, fmt.Sprintf("%s %q", component, name))
}

func (ctx resolveContext) error(msg string) error {
	if len(ctx) == 0 {
		return fmt.Errorf(msg)
	}

	var b strings.Builder
	b.WriteString(msg)
	// Print stack top-down (metric → templates → error)
	for i := len(ctx) - 1; i >= 0; i-- {
		b.WriteString("\n  in ")
		b.WriteString(ctx[i])
	}
	return fmt.Errorf(b.String())
}
