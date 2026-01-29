package config

// SourceConfig defines a fully resolved source with embedded clock
type SourceConfig struct {
	Type     string
	Clock    ClockConfig
	ClockRef *string // Instance name if clock is shared
	Min      int
	Max      int
}
