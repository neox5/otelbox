package config

import "log/slog"

// SourceConfig defines a fully resolved source with embedded clock
type SourceConfig struct {
	Type     string
	Clock    ClockConfig
	ClockRef *string // Instance name if clock is shared
	Min      int
	Max      int
}

// LogValue implements slog.LogValuer for structured logging
func (s SourceConfig) LogValue() slog.Value {
	clockName := "inline"
	if s.ClockRef != nil {
		clockName = "instance:" + *s.ClockRef
	}

	attrs := []slog.Attr{
		slog.String("type", s.Type),
		slog.String("clock", clockName),
		slog.Int("min", s.Min),
		slog.Int("max", s.Max),
	}
	return slog.GroupValue(attrs...)
}
