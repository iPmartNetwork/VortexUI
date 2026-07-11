package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// MetricsCollectorService implements observability metrics collection
type MetricsCollectorService struct {
	metrics map[string]*domain.MetricPoint
	mu      sync.RWMutex
	log     *slog.Logger
}

// NewMetricsCollectorService creates a new metrics collector
func NewMetricsCollectorService(log *slog.Logger) *MetricsCollectorService {
	if log == nil {
		log = slog.Default()
	}
	return &MetricsCollectorService{
		metrics: make(map[string]*domain.MetricPoint),
		log:     log,
	}
}

// RecordRequestMetric records HTTP request metrics
func (m *MetricsCollectorService) RecordRequestMetric(ctx context.Context, metric *domain.RequestMetrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := metric.Method + ":" + metric.Path
	m.metrics[key] = &domain.MetricPoint{
		Name:      "http_request_duration",
		Type:      domain.MetricTypeHistogram,
		Value:     float64(metric.Latency),
		Labels:    map[string]string{"method": metric.Method, "path": metric.Path, "status": string(rune(metric.Status))},
		Timestamp: metric.Timestamp,
	}

	m.log.Debug("recorded request metric", "method", metric.Method, "path", metric.Path, "latency_ms", metric.Latency)
	return nil
}

// RecordCounter increments a counter metric
func (m *MetricsCollectorService) RecordCounter(name string, value float64, labels map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, ok := m.metrics[name]; ok {
		existing.Value += value
	} else {
		m.metrics[name] = &domain.MetricPoint{
			Name:      name,
			Type:      domain.MetricTypeCounter,
			Value:     value,
			Labels:    labels,
			Timestamp: time.Now(),
		}
	}
	return nil
}

// RecordGauge sets a gauge metric value
func (m *MetricsCollectorService) RecordGauge(name string, value float64, labels map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[name] = &domain.MetricPoint{
		Name:      name,
		Type:      domain.MetricTypeGauge,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}
	return nil
}

// RecordHistogram records a histogram measurement
func (m *MetricsCollectorService) RecordHistogram(name string, value float64, labels map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics[name] = &domain.MetricPoint{
		Name:      name,
		Type:      domain.MetricTypeHistogram,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}
	return nil
}

// GetMetrics retrieves collected metrics
func (m *MetricsCollectorService) GetMetrics(ctx context.Context, filter map[string]string) ([]*domain.MetricPoint, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []*domain.MetricPoint
	for _, metric := range m.metrics {
		result = append(result, metric)
	}
	return result, nil
}
