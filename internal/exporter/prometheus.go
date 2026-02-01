package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/neox5/otelbox/internal/metric"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusExporter provides HTTP server for Prometheus metrics.
type PrometheusExporter struct {
	addr         string
	path         string
	server       *http.Server
	promRegistry *prometheus.Registry
}

// NewPrometheusExporter creates a new Prometheus HTTP exporter.
func NewPrometheusExporter(
	port int,
	path string,
	metrics *metric.Registry,
	internalMetricsEnabled bool,
) *PrometheusExporter {
	// Create registry
	promRegistry := createPrometheusRegistry(metrics)

	// Setup HTTP server
	addr := fmt.Sprintf(":%d", port)
	server := createHTTPServer(addr, path, promRegistry, internalMetricsEnabled)

	return &PrometheusExporter{
		addr:         addr,
		path:         path,
		promRegistry: promRegistry,
		server:       server,
	}
}

// Start begins serving HTTP requests.
// Blocks until context is cancelled, then shuts down gracefully.
func (e *PrometheusExporter) Start(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		slog.Info("starting prometheus exporter", "addr", e.addr, "path", e.path)
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		// Graceful shutdown
		slog.Info("shutting down prometheus exporter")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return e.server.Shutdown(shutdownCtx)
	}
}
