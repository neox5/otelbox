package config

import "time"

// RawClockReference handles polymorphic clock field (instance/template/inline)
type RawClockReference struct {
	Instance string        `yaml:"instance,omitempty"`
	Template string        `yaml:"template,omitempty"`
	Type     *string       `yaml:"type,omitempty"`
	Interval time.Duration `yaml:"interval,omitempty"`
}
