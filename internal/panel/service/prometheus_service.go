package service

import (
	"bytes"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// PrometheusExporter exports metrics in Prometheus format
type PrometheusExporter struct {
	metricsService *MetricsCollectorService
	log            *slog.Logger
	mu             sync.RWMutex
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(metricsService *MetricsCollectorService, log *slog.Logger) *PrometheusExporter {
	if log == nil {
		log = slog.Default()
	}
	return &PrometheusExporter{
		metricsService: metricsService,
		log:            log,
	}
}

// Export generates Prometheus format metrics output
func (p *PrometheusExporter) Export() string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var buf bytes.Buffer

	// Write header comments
	buf.WriteString("# HELP vortex_http_request_duration HTTP request duration in milliseconds\n")
	buf.WriteString("# TYPE vortex_http_request_duration histogram\n")

	// Get all metrics
	metrics, err := p.metricsService.GetMetrics(nil, nil)
	if err != nil {
		p.log.Error("failed to get metrics for prometheus export", "error", err)
		return buf.String()
	}

	// Group metrics by type and name
	counters := make(map[string]float64)
	gauges := make(map[string]float64)
	histograms := make(map[string][]float64)

	for _, metric := range metrics {
		switch metric.Type {
		case "counter":
			counters[metric.Name] = metric.Value
		case "gauge":
			gauges[metric.Name] = metric.Value
		case "histogram":
			histograms[metric.Name] = append(histograms[metric.Name], metric.Value)
		}
	}

	// Write counters
	if len(counters) > 0 {
		buf.WriteString("\n# Counters\n")
		for name, value := range counters {
			buf.WriteString(fmt.Sprintf("vortex_%s %f\n", name, value))
		}
	}

	// Write gauges
	if len(gauges) > 0 {
		buf.WriteString("\n# Gauges\n")
		for name, value := range gauges {
			buf.WriteString(fmt.Sprintf("vortex_%s %f\n", name, value))
		}
	}

	// Write histograms with buckets
	if len(histograms) > 0 {
		buf.WriteString("\n# Histograms\n")
		for name, values := range histograms {
			// Calculate histogram statistics
			min, max, sum := calculateHistogramStats(values)

			// Prometheus histogram buckets (in milliseconds)
			buckets := []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000, 5000}
			for _, bucket := range buckets {
				count := 0
				for _, v := range values {
					if v <= bucket {
						count++
					}
				}
				buf.WriteString(fmt.Sprintf("vortex_%s_bucket{le=\"%f\"} %d\n", name, bucket, count))
			}
			buf.WriteString(fmt.Sprintf("vortex_%s_bucket{le=\"+Inf\"} %d\n", name, len(values)))
			buf.WriteString(fmt.Sprintf("vortex_%s_sum %f\n", name, sum))
			buf.WriteString(fmt.Sprintf("vortex_%s_count %d\n", name, len(values)))
			buf.WriteString(fmt.Sprintf("vortex_%s_min %f\n", name, min))
			buf.WriteString(fmt.Sprintf("vortex_%s_max %f\n", name, max))
		}
	}

	// Write timestamp
	buf.WriteString(fmt.Sprintf("\n# Exported at %s\n", time.Now().UTC().Format(time.RFC3339)))

	return buf.String()
}

// calculateHistogramStats calculates min, max, sum from histogram values
func calculateHistogramStats(values []float64) (float64, float64, float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}

	min := values[0]
	max := values[0]
	sum := 0.0

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	return min, max, sum
}

// RequestMetricsAggregator aggregates HTTP metrics by endpoint
type RequestMetricsAggregator struct {
	metricsService *MetricsCollectorService
	log            *slog.Logger
	mu             sync.RWMutex
}

// NewRequestMetricsAggregator creates a new metrics aggregator
func NewRequestMetricsAggregator(metricsService *MetricsCollectorService, log *slog.Logger) *RequestMetricsAggregator {
	if log == nil {
		log = slog.Default()
	}
	return &RequestMetricsAggregator{
		metricsService: metricsService,
		log:            log,
	}
}

// AggregateMetrics returns aggregated metrics by method/path
func (r *RequestMetricsAggregator) AggregateMetrics() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics, err := r.metricsService.GetMetrics(nil, nil)
	if err != nil {
		r.log.Error("failed to aggregate metrics", "error", err)
		return make(map[string]interface{})
	}

	endpoints := make(map[string]map[string]interface{})

	for _, metric := range metrics {
		if metric.Labels == nil {
			continue
		}

		method := metric.Labels["method"]
		path := metric.Labels["path"]
		if method == "" || path == "" {
			continue
		}

		key := fmt.Sprintf("%s %s", method, path)

		if _, exists := endpoints[key]; !exists {
			endpoints[key] = map[string]interface{}{
				"method":     method,
				"path":       path,
				"count":      0,
				"total_time": 0,
				"min_time":   float64(^uint64(0)),
				"max_time":   0,
			}
		}

		ep := endpoints[key]
		ep["count"] = ep["count"].(int) + 1
		ep["total_time"] = ep["total_time"].(float64) + metric.Value

		if metric.Value < ep["min_time"].(float64) {
			ep["min_time"] = metric.Value
		}
		if metric.Value > ep["max_time"].(float64) {
			ep["max_time"] = metric.Value
		}
	}

	// Convert to response format
	result := make(map[string]interface{})
	for key, ep := range endpoints {
		count := ep["count"].(int)
		if count > 0 {
			total := ep["total_time"].(float64)
			ep["avg_time"] = total / float64(count)
		}
		result[key] = ep
	}

	return result
}
