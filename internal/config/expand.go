package config

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"
)

// iteratorPattern matches {iterator_name} placeholders in strings
var iteratorPattern = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// Expander orchestrates iterator expansion across all configuration types.
type Expander struct {
	registry *IteratorRegistry
}

// NewExpander creates an expander from iterator definitions.
func NewExpander(iterators []RawIterator) (*Expander, error) {
	if len(iterators) == 0 {
		return &Expander{registry: nil}, nil
	}

	registry, err := buildIteratorRegistry(iterators)
	if err != nil {
		return nil, fmt.Errorf("failed to build iterator registry: %w", err)
	}

	slog.Debug("iterator registry built", "count", len(iterators))
	for _, it := range registry.iterators {
		slog.Debug("registered iterator", "name", it.Name(), "count", it.Len())
	}

	return &Expander{registry: registry}, nil
}

// expandable defines operations needed for generic expansion with pointer receiver support
type expandable[T any, PT interface {
	*T
	FindPlaceholders() []string
	SubstitutePlaceholders(map[string]string)
}] interface {
	DeepCopy() T
}

// expand is the generic expansion implementation using two-type-parameter pattern
func expand[T expandable[T, PT], PT interface {
	*T
	FindPlaceholders() []string
	SubstitutePlaceholders(map[string]string)
}](items []T, registry *IteratorRegistry, entityType string) ([]T, error) {
	if registry == nil {
		return items, nil
	}

	expanded := make([]T, 0)

	for i, item := range items {
		placeholders := PT(&item).FindPlaceholders()

		if len(placeholders) == 0 {
			expanded = append(expanded, item)
			continue
		}

		iterators, err := registry.GetIterators(placeholders)
		if err != nil {
			return nil, fmt.Errorf("%s at index %d: %w", entityType, i, err)
		}

		gen := NewCombinationGenerator(iterators)

		if gen.Total() == 0 {
			return nil, fmt.Errorf("%s at index %d: iterator combination produces zero results", entityType, i)
		}

		err = gen.ForEach(func(iteratorValues map[string]string) error {
			clone := item.DeepCopy()
			PT(&clone).SubstitutePlaceholders(iteratorValues)
			expanded = append(expanded, clone)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("%s at index %d: %w", entityType, i, err)
		}
	}

	return expanded, nil
}

// ExpandClocks expands clock references containing iterator placeholders.
func (e *Expander) ExpandClocks(clocks []RawClockReference) ([]RawClockReference, error) {
	return expand(clocks, e.registry, "clock")
}

// ExpandSources expands source references containing iterator placeholders.
func (e *Expander) ExpandSources(sources []RawSourceReference) ([]RawSourceReference, error) {
	return expand(sources, e.registry, "source")
}

// ExpandValues expands value references containing iterator placeholders.
func (e *Expander) ExpandValues(values []RawValueReference) ([]RawValueReference, error) {
	return expand(values, e.registry, "value")
}

// ExpandMetrics expands metric configs containing iterator placeholders.
func (e *Expander) ExpandMetrics(metrics []RawMetricConfig) ([]RawMetricConfig, error) {
	return expand(metrics, e.registry, "metric")
}

// Expand performs iterator expansion on raw configuration.
// Mutates raw config in place by replacing arrays with expanded versions.
func Expand(raw *RawConfig) error {
	expander, err := NewExpander(raw.Iterators)
	if err != nil {
		return err
	}

	// Expand template clocks
	raw.Templates.Clocks, err = expander.ExpandClocks(raw.Templates.Clocks)
	if err != nil {
		return fmt.Errorf("failed to expand template clocks: %w", err)
	}

	// Expand instance clocks
	raw.Instances.Clocks, err = expander.ExpandClocks(raw.Instances.Clocks)
	if err != nil {
		return fmt.Errorf("failed to expand instance clocks: %w", err)
	}

	// Expand template sources
	raw.Templates.Sources, err = expander.ExpandSources(raw.Templates.Sources)
	if err != nil {
		return fmt.Errorf("failed to expand template sources: %w", err)
	}

	// Expand instance sources
	raw.Instances.Sources, err = expander.ExpandSources(raw.Instances.Sources)
	if err != nil {
		return fmt.Errorf("failed to expand instance sources: %w", err)
	}

	// Expand template values
	raw.Templates.Values, err = expander.ExpandValues(raw.Templates.Values)
	if err != nil {
		return fmt.Errorf("failed to expand template values: %w", err)
	}

	// Expand instance values
	raw.Instances.Values, err = expander.ExpandValues(raw.Instances.Values)
	if err != nil {
		return fmt.Errorf("failed to expand instance values: %w", err)
	}

	// Expand metrics
	raw.Metrics, err = expander.ExpandMetrics(raw.Metrics)
	if err != nil {
		return fmt.Errorf("failed to expand metrics: %w", err)
	}

	// Clear consumed iterators
	raw.Iterators = nil

	return nil
}

// substitutePlaceholders replaces {name} patterns in a string with values
func substitutePlaceholders(s string, iteratorValues map[string]string) string {
	result := s
	for name, value := range iteratorValues {
		placeholder := "{" + name + "}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// extractPlaceholderNames extracts placeholder names from {name} patterns in a string
func extractPlaceholderNames(s string) []string {
	matches := iteratorPattern.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	names := make([]string, len(matches))
	for i, match := range matches {
		names[i] = match[1] // Capture group 1 contains the iterator name
	}
	return names
}
