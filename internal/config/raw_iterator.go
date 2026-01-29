package config

// RawIterator defines a single iterator for config expansion
type RawIterator struct {
	Name   string   `yaml:"name"`
	Type   string   `yaml:"type"` // "range" or "list"
	Start  *int     `yaml:"start,omitempty"`
	End    *int     `yaml:"end,omitempty"`
	Values []string `yaml:"values,omitempty"`
}
