package config

import (
	"log/slog"
	"strings"
)

// ValueConfig defines a fully resolved value with embedded components.
type ValueConfig struct {
	Source     SourceConfig
	SourceRef  *string // Instance name if source is shared
	Transforms []TransformConfig
	Reset      ResetConfig
}

// LogValue implements slog.LogValuer for structured logging
func (v ValueConfig) LogValue() slog.Value {
	sourceName := "inline"
	if v.SourceRef != nil {
		sourceName = "instance:" + *v.SourceRef
	}

	attrs := []slog.Attr{
		slog.String("source", sourceName),
		slog.Int("transforms", len(v.Transforms)),
	}

	// Add reset info if configured
	if v.Reset.Type != "" {
		resetDesc := v.Reset.Type
		if v.Reset.Value != 0 {
			resetDesc += ":" + strings.TrimSpace(string(rune(v.Reset.Value)))
		}
		attrs = append(attrs, slog.String("reset", resetDesc))
	}

	return slog.GroupValue(attrs...)
}
