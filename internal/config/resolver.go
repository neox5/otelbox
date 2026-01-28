package config

import (
	"fmt"
	"strings"
)

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

// NewResolver creates a new resolver
func NewResolver(raw *RawConfig) *Resolver {
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

// registerName validates namespace uniqueness and registers the name
func (r *Resolver) registerName(name string, entityType string) error {
	if existingType, exists := r.registeredNames[name]; exists {
		return fmt.Errorf("name %q already used by %s, cannot reuse for %s",
			name, existingType, entityType)
	}
	r.registeredNames[name] = entityType
	return nil
}

// Resolve performs hierarchical template and instance resolution and builds final config
func (r *Resolver) Resolve() (*Config, error) {
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

	// Phase 4: Build final config (templates discarded, instances kept)
	return &Config{
		Instances: InstanceRegistry{
			Clocks:  r.instanceClocks,
			Sources: r.instanceSources,
			Values:  r.instanceValues,
		},
		Metrics:  resolvedMetrics,
		Export:   r.raw.Export,
		Settings: r.raw.Settings,
	}, nil
}

// getStringValue safely dereferences a string pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
