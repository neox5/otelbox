package config

import "fmt"

// expandClocks expands clock references that contain iterator placeholders.
// Returns expanded array with iterator placeholders substituted.
func expandClocks(
	clocks []RawClockReference,
	registry *IteratorRegistry,
) ([]RawClockReference, error) {
	expanded := make([]RawClockReference, 0)

	for i, clock := range clocks {
		// Find all iterators referenced in this clock
		usedIterators := findIteratorsInStruct(clock)

		if len(usedIterators) == 0 {
			// No iterators - keep clock as-is
			expanded = append(expanded, clock)
			continue
		}

		// Get iterator instances
		iterators, err := registry.GetIterators(usedIterators)
		if err != nil {
			return nil, fmt.Errorf("clock at index %d: %w", i, err)
		}

		// Create combination generator
		gen := NewCombinationGenerator(iterators)

		if gen.Total() == 0 {
			return nil, fmt.Errorf("clock at index %d: iterator combination produces zero results", i)
		}

		// Pre-allocate space for all expanded clocks
		startIdx := len(expanded)
		expanded = append(expanded, make([]RawClockReference, gen.Total())...)

		// Generate one clock per combination
		currentIdx := startIdx
		err = gen.ForEach(func(combo map[string]string) error {
			// Clone the clock (struct copy)
			clone := clock

			// Substitute iterator placeholders with actual values
			substituteIterators(&clone, combo)

			// Store expanded clock
			expanded[currentIdx] = clone
			currentIdx++
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("clock at index %d: %w", i, err)
		}
	}

	return expanded, nil
}
