package config

import (
	"fmt"
	"time"
)

const (
	// Prometheus defaults
	DefaultPrometheusPort = 9090
	DefaultPrometheusPath = "/metrics"

	// OTEL defaults
	DefaultOTELReadInterval = 1 * time.Second
	DefaultOTELPushInterval = 1 * time.Second
	DefaultOTELTransport    = "grpc"
	DefaultOTELHost         = "localhost"
	DefaultOTELPortGRPC     = 4317
	DefaultOTELPortHTTP     = 4318
	DefaultServiceName      = "obsbox"
	DefaultServiceVersion   = "dev"
)

// ExportConfig defines how metrics are exposed.
type ExportConfig struct {
	Prometheus *PrometheusExportConfig
	OTEL       *OTELExportConfig
}

// Validate applies defaults and validates export configuration.
func (e *ExportConfig) Validate() error {
	// Default to Prometheus enabled if no exporters configured
	if e.Prometheus == nil && e.OTEL == nil {
		e.Prometheus = &PrometheusExportConfig{
			Enabled: true,
			Port:    DefaultPrometheusPort,
			Path:    DefaultPrometheusPath,
		}
		return nil
	}

	// Validate individual exporters
	if e.Prometheus != nil && e.Prometheus.Enabled {
		if err := e.Prometheus.Validate(); err != nil {
			return err
		}
	}

	if e.OTEL != nil && e.OTEL.Enabled {
		if err := e.OTEL.Validate(); err != nil {
			return err
		}
	}

	// Verify at least one exporter enabled
	promEnabled := e.Prometheus != nil && e.Prometheus.Enabled
	otelEnabled := e.OTEL != nil && e.OTEL.Enabled

	if !promEnabled && !otelEnabled {
		return fmt.Errorf("at least one exporter must be enabled")
	}

	// Verify only one exporter enabled (prevent read conflicts)
	if promEnabled && otelEnabled {
		return fmt.Errorf("only one exporter can be enabled at a time (prometheus or otel)")
	}

	return nil
}

// PrometheusExportConfig defines Prometheus pull endpoint settings.
type PrometheusExportConfig struct {
	Enabled bool
	Port    int
	Path    string
}

// Validate applies defaults and validates Prometheus configuration.
func (c *PrometheusExportConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// Apply defaults
	if c.Port == 0 {
		c.Port = DefaultPrometheusPort
	}
	if c.Path == "" {
		c.Path = DefaultPrometheusPath
	}

	// Validate port range
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid prometheus port: %d", c.Port)
	}

	return nil
}

// OTELExportConfig defines OTEL push settings.
type OTELExportConfig struct {
	Enabled   bool
	Transport string
	Host      string
	Port      int
	Interval  IntervalConfig
	Resource  map[string]string
	Headers   map[string]string
}

// IntervalConfig defines read and push intervals for OTEL.
type IntervalConfig struct {
	Read time.Duration
	Push time.Duration
}

// Validate applies defaults and validates OTEL configuration.
func (c *OTELExportConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// Apply transport default
	if c.Transport == "" {
		c.Transport = DefaultOTELTransport
	}

	// Validate transport
	if c.Transport != "grpc" && c.Transport != "http" {
		return fmt.Errorf("invalid transport: %s (must be grpc or http)", c.Transport)
	}

	// Apply host default
	if c.Host == "" {
		c.Host = DefaultOTELHost
	}

	// Apply port default based on transport
	if c.Port == 0 {
		if c.Transport == "grpc" {
			c.Port = DefaultOTELPortGRPC
		} else {
			c.Port = DefaultOTELPortHTTP
		}
	}

	// Apply interval defaults
	if c.Interval.Read == 0 {
		c.Interval.Read = DefaultOTELReadInterval
	}
	if c.Interval.Push == 0 {
		c.Interval.Push = DefaultOTELPushInterval
	}

	// Apply resource defaults
	if c.Resource == nil {
		c.Resource = make(map[string]string)
	}
	if _, exists := c.Resource["service.name"]; !exists {
		c.Resource["service.name"] = DefaultServiceName
	}
	if _, exists := c.Resource["service.version"]; !exists {
		c.Resource["service.version"] = DefaultServiceVersion
	}

	return nil
}

// GetEndpoint returns the full endpoint address.
func (c *OTELExportConfig) GetEndpoint() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
