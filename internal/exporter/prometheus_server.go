package exporter

import (
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// createHTTPServer creates an HTTP server for Prometheus metrics.
func createHTTPServer(
	addr string,
	path string,
	promRegistry *prometheus.Registry,
	internalMetricsEnabled bool,
) *http.Server {
	mux := http.NewServeMux()

	// Create base handler
	baseHandler := promhttp.HandlerFor(
		promRegistry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)

	// Conditionally wrap with instrumentation
	var handler http.Handler
	if internalMetricsEnabled {
		handler = promhttp.InstrumentMetricHandler(promRegistry, baseHandler)
		slog.Info("enabled prometheus internal metrics",
			"metrics", []string{
				"promhttp_metric_handler_requests_total",
				"promhttp_metric_handler_requests_in_flight",
			})
	} else {
		handler = baseHandler
	}

	// Wrap with debug logging
	handler = loggingMiddleware(handler)

	mux.Handle(path, handler)

	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

// loggingMiddleware logs scrape requests when debug logging is enabled
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Debug("prometheus scrape")
		next.ServeHTTP(w, r)
	})
}
