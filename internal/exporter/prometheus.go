package exporter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/neox5/obsbox/internal/config"
	"github.com/neox5/obsbox/internal/metric"
	"github.com/neox5/simv/value"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Internal metric name definitions (both formats hardcoded)
const (
	promScrapesTotalUnderscore   = "obsbox_prometheus_scrapes_total"
	promScrapesTotalDot          = "obsbox.prometheus.scrapes.total"
	promScrapeDurationUnderscore = "obsbox_prometheus_scrape_duration_seconds"
	promScrapeDurationDot        = "obsbox.prometheus.scrape.duration.seconds"
)

// PrometheusExporter provides HTTP server for Prometheus metrics.
type PrometheusExporter struct {
	addr         string
	path         string
	server       *http.Server
	promRegistry *prometheus.Registry

	// Internal metrics
	scrapesTotal   prometheus.Counter
	scrapeDuration prometheus.Histogram
}

// metricDescriptor holds metadata for a Prometheus metric.
type metricDescriptor struct {
	desc        *prometheus.Desc
	valueType   prometheus.ValueType
	value       value.Value[int]
	labelValues []string
}

// collector implements prometheus.Collector to read simv values on scrape.
type collector struct {
	descriptors []metricDescriptor
}

// NewPrometheusExporter creates a new Prometheus HTTP exporter.
func NewPrometheusExporter(
	port int,
	path string,
	metrics *metric.Registry,
	internalMetricsEnabled bool,
	namingFormat config.NamingFormat,
) *PrometheusExporter {
	promRegistry := prometheus.NewRegistry()

	// Build Prometheus-specific descriptors
	var descriptors []metricDescriptor
	for _, m := range metrics.Metrics() {
		var valueType prometheus.ValueType
		switch m.Type {
		case metric.MetricTypeCounter:
			valueType = prometheus.CounterValue
		case metric.MetricTypeGauge:
			valueType = prometheus.GaugeValue
		}

		// Extract and sort label names for consistent ordering
		var labelNames []string
		for key := range m.Attributes {
			labelNames = append(labelNames, key)
		}
		sort.Strings(labelNames)

		// Build label values in same order
		labelValues := make([]string, len(labelNames))
		for i, name := range labelNames {
			labelValues[i] = m.Attributes[name]
		}

		descriptors = append(descriptors, metricDescriptor{
			desc: prometheus.NewDesc(
				m.PrometheusName,
				m.Description,
				labelNames,
				nil, // No constant labels
			),
			valueType:   valueType,
			value:       m.Value,
			labelValues: labelValues,
		})

		slog.Info("registered prometheus metric",
			"name", m.PrometheusName,
			"type", m.Type,
			"labels", labelNames)
	}

	// Register collector
	c := &collector{descriptors: descriptors}
	promRegistry.MustRegister(c)

	// Setup HTTP server
	mux := http.NewServeMux()
	addr := fmt.Sprintf(":%d", port)

	e := &PrometheusExporter{
		addr:         addr,
		path:         path,
		promRegistry: promRegistry,
		server: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	// Register internal metrics if enabled
	if internalMetricsEnabled {
		// Select names based on format
		scrapesName := promScrapesTotalUnderscore
		durationName := promScrapeDurationUnderscore

		if namingFormat == config.NamingFormatDot {
			scrapesName = promScrapesTotalDot
			durationName = promScrapeDurationDot
		}
		// native format uses underscore for Prometheus

		e.scrapesTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Name: scrapesName,
			Help: "Total number of scrape requests",
		})

		e.scrapeDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    durationName,
			Help:    "Duration of scrape requests in seconds",
			Buckets: prometheus.DefBuckets,
		})

		promRegistry.MustRegister(e.scrapesTotal, e.scrapeDuration)

		slog.Info("registered prometheus internal metrics",
			"format", namingFormat,
			"scrapes_total", scrapesName,
			"scrape_duration", durationName)
	}

	mux.Handle(path, e.instrumentedHandler(promhttp.HandlerFor(
		promRegistry,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		},
	)))

	return e
}

// instrumentedHandler wraps the Prometheus handler with internal metrics instrumentation.
func (e *PrometheusExporter) instrumentedHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		defer func() {
			// Increment counter and observe duration atomically after completion
			if e.scrapesTotal != nil {
				e.scrapesTotal.Inc()
			}
			if e.scrapeDuration != nil {
				e.scrapeDuration.Observe(time.Since(start).Seconds())
			}
		}()

		next.ServeHTTP(w, r)
	})
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

// Describe sends metric descriptors to the channel.
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.descriptors {
		ch <- m.desc
	}
}

// Collect reads simv values and sends metrics to the channel.
// This is called on each Prometheus scrape.
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	for _, m := range c.descriptors {
		// Read value from simv (may trigger reset for reset_on_read)
		val := float64(m.value.Value())

		// Create and send metric with current value and labels
		metric, err := prometheus.NewConstMetric(
			m.desc,
			m.valueType,
			val,
			m.labelValues...,
		)
		if err != nil {
			continue
		}

		ch <- metric
	}
}
