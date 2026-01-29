package config

// ValueConfig defines a fully resolved value with embedded components.
type ValueConfig struct {
	Source     SourceConfig
	SourceRef  *string // Instance name if source is shared
	Transforms []TransformConfig
	Reset      ResetConfig
}
