package config

import (
	"fmt"
	"strconv"
)

// Iterator provides lazy value generation for configuration expansion.
// Values are generated on-demand rather than stored in memory.
type Iterator struct {
	name      string
	generator func(index int) string // Generate value at index
	count     int                    // Total number of values
}

// NewRangeIterator creates an iterator that generates sequential integers.
// Values are generated as strings: start, start+1, ..., end (inclusive).
func NewRangeIterator(name string, start, end int) *Iterator {
	if end < start {
		return &Iterator{
			name:  name,
			count: 0,
			generator: func(index int) string {
				panic("empty range iterator")
			},
		}
	}

	return &Iterator{
		name:  name,
		count: end - start + 1,
		generator: func(index int) string {
			return strconv.Itoa(start + index)
		},
	}
}

// NewListIterator creates an iterator that cycles through explicit values.
func NewListIterator(name string, values []string) *Iterator {
	// Store values slice - acceptable memory cost for explicit lists
	valuesCopy := make([]string, len(values))
	copy(valuesCopy, values)

	return &Iterator{
		name:  name,
		count: len(valuesCopy),
		generator: func(index int) string {
			return valuesCopy[index]
		},
	}
}

// Name returns the iterator name (used in {name} placeholders).
func (it *Iterator) Name() string {
	return it.name
}

// Len returns the total number of values this iterator generates.
func (it *Iterator) Len() int {
	return it.count
}

// ValueAt generates the value at the specified index (0-based).
// Panics if index is out of range [0, Len()).
func (it *Iterator) ValueAt(index int) string {
	if index < 0 || index >= it.count {
		panic(fmt.Sprintf("iterator %q: index %d out of range [0, %d)",
			it.name, index, it.count))
	}
	return it.generator(index)
}

// AllValues materializes all values into a slice.
// Primarily useful for debugging and testing.
// Avoid using this for large ranges in production code.
func (it *Iterator) AllValues() []string {
	values := make([]string, it.count)
	for i := 0; i < it.count; i++ {
		values[i] = it.ValueAt(i)
	}
	return values
}

// IteratorRegistry manages all defined iterators.
type IteratorRegistry struct {
	iterators map[string]*Iterator
}

// NewIteratorRegistry creates an empty iterator registry.
func NewIteratorRegistry() *IteratorRegistry {
	return &IteratorRegistry{
		iterators: make(map[string]*Iterator),
	}
}

// Register adds an iterator to the registry.
// Returns error if an iterator with the same name already exists.
func (r *IteratorRegistry) Register(it *Iterator) error {
	if _, exists := r.iterators[it.Name()]; exists {
		return fmt.Errorf("iterator %q already registered", it.Name())
	}
	r.iterators[it.Name()] = it
	return nil
}

// Get retrieves an iterator by name.
// Returns (iterator, true) if found, (nil, false) otherwise.
func (r *IteratorRegistry) Get(name string) (*Iterator, bool) {
	it, exists := r.iterators[name]
	return it, exists
}

// GetIterators retrieves multiple iterators by name.
// Returns error if any iterator is not found.
func (r *IteratorRegistry) GetIterators(names []string) ([]*Iterator, error) {
	iterators := make([]*Iterator, len(names))
	for i, name := range names {
		it, exists := r.Get(name)
		if !exists {
			return nil, fmt.Errorf("iterator %q not defined", name)
		}
		iterators[i] = it
	}
	return iterators, nil
}

// CombinationGenerator generates Cartesian product combinations lazily.
// Memory usage is O(1) regardless of combination count.
type CombinationGenerator struct {
	iterators []*Iterator
	total     int
}

// NewCombinationGenerator creates a lazy combination generator.
// Combinations are generated on-demand, not stored in memory.
func NewCombinationGenerator(iterators []*Iterator) *CombinationGenerator {
	if len(iterators) == 0 {
		return &CombinationGenerator{
			iterators: iterators,
			total:     0,
		}
	}

	// Calculate total combinations (Cartesian product size)
	total := 1
	for _, it := range iterators {
		total *= it.Len()
	}

	return &CombinationGenerator{
		iterators: iterators,
		total:     total,
	}
}

// Total returns the number of combinations this generator will produce.
func (g *CombinationGenerator) Total() int {
	return g.total
}

// Generate produces the combination at the specified index.
// Index must be in range [0, Total()).
// Returns a map of iterator_name -> value for this combination.
func (g *CombinationGenerator) Generate(index int) map[string]string {
	if index < 0 || index >= g.total {
		panic(fmt.Sprintf("combination index %d out of range [0, %d)",
			index, g.total))
	}

	result := make(map[string]string, len(g.iterators))

	// Calculate which value from each iterator to use
	// Uses positional encoding: rightmost iterator cycles fastest
	repeat := 1
	for _, it := range g.iterators {
		valueIndex := (index / repeat) % it.Len()
		result[it.Name()] = it.ValueAt(valueIndex)
		repeat *= it.Len()
	}

	return result
}

// ForEach iterates through all combinations, calling fn for each.
// Iteration stops early if fn returns an error.
// Only one combination exists in memory at a time.
func (g *CombinationGenerator) ForEach(fn func(map[string]string) error) error {
	for i := 0; i < g.total; i++ {
		combo := g.Generate(i)
		if err := fn(combo); err != nil {
			return err
		}
	}
	return nil
}

// buildIteratorRegistry creates a registry from raw iterator definitions.
func buildIteratorRegistry(rawIterators []RawIterator) (*IteratorRegistry, error) {
	registry := NewIteratorRegistry()

	for _, raw := range rawIterators {
		var it *Iterator

		switch raw.Type {
		case "range":
			// Validate range parameters
			if raw.Start == nil {
				return nil, fmt.Errorf("iterator %q: start required for range type",
					raw.Name)
			}
			if raw.End == nil {
				return nil, fmt.Errorf("iterator %q: end required for range type",
					raw.Name)
			}
			it = NewRangeIterator(raw.Name, *raw.Start, *raw.End)

		case "list":
			// Validate list parameters
			if raw.Values == nil || len(raw.Values) == 0 {
				return nil, fmt.Errorf("iterator %q: values required for list type",
					raw.Name)
			}
			it = NewListIterator(raw.Name, raw.Values)

		default:
			return nil, fmt.Errorf("iterator %q: unknown type %q (must be range or list)",
				raw.Name, raw.Type)
		}

		if err := registry.Register(it); err != nil {
			return nil, err
		}
	}

	return registry, nil
}
