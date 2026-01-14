package app

import (
	"fmt"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/exporter"
	"github.com/neox5/obsbox/internal/generator"
	"github.com/neox5/obsbox/internal/metric"
)

// App holds initialized application components.
type App struct {
	Config             *config.Config
	Generator          *generator.Generator
	Metrics            *metric.Registry
	PrometheusExporter *exporter.PrometheusExporter
	OTELExporter       *exporter.OTELExporter
}

// New initializes the application from a configuration file.
func New(configPath string) (*App, error) {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Create generator
	gen, err := generator.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	// Create metrics
	metrics, err := metric.New(cfg, gen)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics: %w", err)
	}

	var promExporter *exporter.PrometheusExporter
	var otelExporter *exporter.OTELExporter

	// Create Prometheus exporter if enabled
	if cfg.Export.Prometheus != nil && cfg.Export.Prometheus.Enabled {
		promExporter = exporter.NewPrometheusExporter(
			cfg.Export.Prometheus.Port,
			cfg.Export.Prometheus.Path,
			metrics,
		)
	}

	// Create OTEL exporter if enabled
	if cfg.Export.OTEL != nil && cfg.Export.OTEL.Enabled {
		otelExporter, err = exporter.NewOTELExporter(
			cfg.Export.OTEL,
			metrics,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTEL exporter: %w", err)
		}
	}

	return &App{
		Config:             cfg,
		Generator:          gen,
		Metrics:            metrics,
		PrometheusExporter: promExporter,
		OTELExporter:       otelExporter,
	}, nil
}
