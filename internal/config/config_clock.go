package config

import "time"

// ClockConfig defines a fully resolved clock
type ClockConfig struct {
	Type     string
	Interval time.Duration
}
