package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/neox5/otelbox/internal/app"
	"github.com/neox5/otelbox/internal/config"
	"github.com/neox5/otelbox/internal/monitor"
	"github.com/neox5/otelbox/internal/version"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "otelbox",
		Usage:   "Telemetry signal generator for testing observability components",
		Version: version.String(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "path to configuration file",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "enable debug logging",
			},
		},
		Action: serve,
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, cmd *cli.Command) error {
	configPath := cmd.String("config")
	debug := cmd.Bool("debug")

	// Configure logging level
	logLevel := slog.LevelInfo
	if debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("starting otelbox", "version", version.String(), "config", configPath)

	// Load configuration
	slog.Debug("--- Configuration Loading ---")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize application (handles seed initialization internally)
	slog.Debug("--- Generator Creation ---")
	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	// Setup graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start generator
	application.Generator.Start()
	defer application.Generator.Stop()

	// Start resource monitor
	mon := monitor.New(5*time.Second, logger)
	mon.Run(shutdownCtx)
	defer mon.Wait()

	// Start exporters
	slog.Debug("--- Exporter Initialization ---")
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	if application.PrometheusExporter != nil {
		wg.Go(func() {
			if err := application.PrometheusExporter.Start(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("prometheus exporter: %w", err)
			}
		})
	}

	if application.OTELExporter != nil {
		wg.Go(func() {
			if err := application.OTELExporter.Start(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("otel exporter: %w", err)
			}
		})
	}

	slog.Debug("--- Application Running ---")

	// Wait for shutdown or error
	select {
	case err := <-errChan:
		slog.Error("exporter error", "error", err)
		stop() // Cancel context to trigger shutdown
	case <-shutdownCtx.Done():
		// Graceful shutdown triggered
	}

	slog.Debug("--- Shutdown Initiated ---")

	// Wait for all goroutines to complete
	// The exporters' Start methods will return when shutdownCtx is cancelled
	wg.Wait()

	slog.Info("shutdown complete")
	return nil
}
