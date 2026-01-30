package config

import "fmt"

// expandSources expands source references that contain iterator placeholders.
// Returns expanded array with iterator placeholders substituted.
func expandSources(
	sources []RawSourceReference,
	registry *IteratorRegistry,
) ([]RawSourceReference, error) {
	expanded := make([]RawSourceReference, 0)

	for i, source := range sources {
		// Find all iterators referenced in this source
		usedIterators := findIteratorsInStruct(source)

		if len(usedIterators) == 0 {
			// No iterators - keep source as-is
			expanded = append(expanded, source)
			continue
		}

		// Get iterator instances
		iterators, err := registry.GetIterators(usedIterators)
		if err != nil {
			return nil, fmt.Errorf("source at index %d: %w", i, err)
		}

		// Create combination generator
		gen := NewCombinationGenerator(iterators)

		if gen.Total() == 0 {
			return nil, fmt.Errorf("source at index %d: iterator combination produces zero results", i)
		}

		// Pre-allocate space for all expanded sources
		startIdx := len(expanded)
		expanded = append(expanded, make([]RawSourceReference, gen.Total())...)

		// Generate one source per combination
		currentIdx := startIdx
		err = gen.ForEach(func(combo map[string]string) error {
			// Clone the source (struct copy)
			clone := source

			// Substitute iterator placeholders with actual values
			substituteIterators(&clone, combo)

			// Store expanded source
			expanded[currentIdx] = clone
			currentIdx++
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("source at index %d: %w", i, err)
		}
	}

	return expanded, nil
}
