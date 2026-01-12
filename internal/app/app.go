package app

import (
	"fmt"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/generator"
	"github.com/neox5/obsbox/internal/metric"
	"github.com/neox5/obsbox/internal/server"
)

// App holds initialized application components.
type App struct {
	Config    *config.Config
	Generator *generator.Generator
	Metrics   *metric.Registry
	Server    *server.Server
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

	// Create server
	srv := server.New(cfg.Server.Port, cfg.Server.Path, metrics.PrometheusRegistry())

	return &App{
		Config:    cfg,
		Generator: gen,
		Metrics:   metrics,
		Server:    srv,
	}, nil
}
