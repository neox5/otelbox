package generator

import (
	"fmt"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/simv/clock"
	"github.com/neox5/simv/source"
	"github.com/neox5/simv/transform"
	"github.com/neox5/simv/value"
)

// Generator manages simv components and value generation.
type Generator struct {
	clock  clock.Clock
	values map[string]value.Value[int]
}

// New creates a generator from configuration.
func New(cfg *config.Config) (*Generator, error) {
	// Create clock
	clk := clock.NewPeriodicClock(cfg.Clock.Interval)

	// Create sources
	sources := make(map[string]source.Publisher[int])
	for name, srcCfg := range cfg.Sources {
		switch srcCfg.Type {
		case "random_int":
			sources[name] = source.NewRandomIntSource(clk, srcCfg.Min, srcCfg.Max)
		default:
			return nil, fmt.Errorf("unknown source type: %s", srcCfg.Type)
		}
	}

	// Create values
	values := make(map[string]value.Value[int])

	// First pass: create base values from sources
	for name, valCfg := range cfg.Values {
		if valCfg.Source != "" {
			src, exists := sources[valCfg.Source]
			if !exists {
				return nil, fmt.Errorf("source %q not found for value %q", valCfg.Source, name)
			}

			// Apply transforms
			var transforms []transform.Transformation[int]
			for _, tfName := range valCfg.Transforms {
				switch tfName {
				case "accumulate":
					transforms = append(transforms, transform.NewAccumulate[int]())
				default:
					return nil, fmt.Errorf("unknown transform: %s", tfName)
				}
			}

			values[name] = value.New(src, transforms...)
		}
	}

	// Second pass: create derived values (clones, wraps)
	for name, valCfg := range cfg.Values {
		if valCfg.CloneFrom != "" {
			baseVal, exists := values[valCfg.CloneFrom]
			if !exists {
				return nil, fmt.Errorf("clone_from %q not found for value %q", valCfg.CloneFrom, name)
			}

			cloned := baseVal.Clone()

			// Apply wrapping
			if valCfg.Wrap == "reset_on_read" {
				values[name] = value.NewResetOnRead(cloned, valCfg.ResetValue)
			} else if valCfg.Wrap != "" {
				return nil, fmt.Errorf("unknown wrap type: %s", valCfg.Wrap)
			} else {
				values[name] = cloned
			}
		}
	}

	return &Generator{
		clock:  clk,
		values: values,
	}, nil
}

// Start begins value generation.
func (g *Generator) Start() {
	g.clock.Start()
}

// Stop halts value generation.
func (g *Generator) Stop() {
	g.clock.Stop()
}

// GetValue returns a named value.
func (g *Generator) GetValue(name string) (value.Value[int], bool) {
	val, exists := g.values[name]
	return val, exists
}
