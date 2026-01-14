package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/neox5/obsbox/internal/metric"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusExporter provides HTTP server for Prometheus metrics.
type PrometheusExporter struct {
	addr    string
	path    string
	server  *http.Server
	metrics *metric.Registry
}

// NewPrometheusExporter creates a new Prometheus HTTP exporter.
func NewPrometheusExporter(port int, path string, metrics *metric.Registry) *PrometheusExporter {
	mux := http.NewServeMux()
	addr := fmt.Sprintf(":%d", port)

	// Register Prometheus handler
	mux.Handle(path, promhttp.HandlerFor(
		metrics.PrometheusRegistry(),
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	))

	return &PrometheusExporter{
		addr:    addr,
		path:    path,
		metrics: metrics,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

// Start begins serving HTTP requests.
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
		return e.Stop()
	}
}

// Stop gracefully stops the exporter.
func (e *PrometheusExporter) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	slog.Info("shutting down prometheus exporter")
	return e.server.Shutdown(ctx)
}
