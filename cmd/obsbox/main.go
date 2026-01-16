package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/neox5/obsbox/internal/app"
	"github.com/neox5/obsbox/internal/version"
)

func main() {
	// Handle --version flag
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("obsbox", version.String())
		os.Exit(0)
	}

	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	// Initialize application
	application, err := app.New(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialization failed: %v\n", err)
		os.Exit(1)
	}

	slog.Info("starting obsbox", "version", version.String())

	// Setup graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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
			if err := application.PrometheusExporter.Start(ctx); err != nil {
				errChan <- fmt.Errorf("prometheus exporter: %w", err)
			}
		}()
	}

	if application.OTELExporter != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := application.OTELExporter.Start(ctx); err != nil {
				errChan <- fmt.Errorf("otel exporter: %w", err)
			}
		}()
	}

	// Wait for shutdown or error
	select {
	case err := <-errChan:
		slog.Error("exporter error", "error", err)
		os.Exit(1)
	case <-ctx.Done():
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
}
