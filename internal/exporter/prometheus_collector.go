package exporter

import (
	"log/slog"
	"sort"

	"github.com/neox5/obsbox/internal/metric"
	"github.com/neox5/simv/value"
	"github.com/prometheus/client_golang/prometheus"
)

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

// newCollector creates a collector from metric registry.
func newCollector(metrics *metric.Registry) *collector {
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

	return &collector{descriptors: descriptors}
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
