package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/neox5/obsbox/internal/app"
	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/version"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:    "obsbox",
		Usage:   "Telemetry signal generator for testing observability components",
		Version: version.String(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Value:   "config.yaml",
				Usage:   "path to configuration file",
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

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize application (handles seed initialization internally)
	application, err := app.New(cfg)
	if err != nil {
		return fmt.Errorf("initialization failed: %w", err)
	}

	slog.Info("starting obsbox", "version", version.String())

	// Setup graceful shutdown
	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start generator
	application.Generator.Start()
	defer application.Generator.Stop()

	// Start exporters
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	if application.PrometheusExporter != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := application.PrometheusExporter.Start(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("prometheus exporter: %w", err)
			}
		}()
	}

	if application.OTELExporter != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := application.OTELExporter.Start(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("otel exporter: %w", err)
			}
		}()
	}

	// Wait for shutdown or error
	select {
	case err := <-errChan:
		slog.Error("exporter error", "error", err)
		return err
	case <-shutdownCtx.Done():
		// Graceful shutdown
	}

	// Stop exporters
	if application.PrometheusExporter != nil {
		application.PrometheusExporter.Stop()
	}
	if application.OTELExporter != nil {
		application.OTELExporter.Stop()
	}

	wg.Wait()
	slog.Info("shutdown complete")
	return nil
}
