package config

import (
	"log/slog"
	"time"
)

// ClockConfig defines a fully resolved clock
type ClockConfig struct {
	Type     string
	Interval time.Duration
}

// LogValue implements slog.LogValuer for structured logging
func (c ClockConfig) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("type", c.Type),
		slog.Duration("interval", c.Interval),
	}
	return slog.GroupValue(attrs...)
}
