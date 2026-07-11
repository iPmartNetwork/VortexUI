package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// MetricsCollector defines the metrics collection interface
type MetricsCollector interface {
	// RecordRequestMetric records HTTP request metrics
	RecordRequestMetric(ctx context.Context, metric *domain.RequestMetrics) error

	// RecordCounter increments a counter metric
	RecordCounter(name string, value float64, labels map[string]string) error

	// RecordGauge sets a gauge metric value
	RecordGauge(name string, value float64, labels map[string]string) error

	// RecordHistogram records a histogram measurement
	RecordHistogram(name string, value float64, labels map[string]string) error

	// GetMetrics retrieves collected metrics
	GetMetrics(ctx context.Context, filter map[string]string) ([]*domain.MetricPoint, error)
}

// HealthChecker defines the health checking interface
type HealthChecker interface {
	// CheckHealth performs a full health check
	CheckHealth(ctx context.Context) (*domain.HealthCheckResult, error)

	// CheckComponent checks the health of a specific component
	CheckComponent(ctx context.Context, name string) (*domain.ComponentHealth, error)

	// RegisterCheck registers a health check for a component
	RegisterCheck(name string, fn func(context.Context) (domain.HealthStatus, string, error)) error
}

// StructuredLogger defines structured logging interface
type StructuredLogger interface {
	// LogWithContext logs a message with context
	LogWithContext(ctx context.Context, level string, message string, attrs map[string]interface{}) error

	// LogError logs an error with structured context
	LogError(ctx context.Context, err error, message string, attrs map[string]interface{}) error

	// GetLogs retrieves recent logs with optional filtering
	GetLogs(ctx context.Context, level string, limit int) ([]*domain.LogEntry, error)
}

// TraceManager defines distributed tracing interface
type TraceManager interface {
	// StartSpan creates a new trace span
	StartSpan(ctx context.Context, spanName string, attrs map[string]interface{}) (context.Context, func() error)

	// GetTraceContext retrieves the current trace context
	GetTraceContext(ctx context.Context) *domain.TraceContext

	// InjectTraceContext injects trace context into headers
	InjectTraceContext(ctx context.Context, headers map[string]string) error

	// ExtractTraceContext extracts trace context from headers
	ExtractTraceContext(headers map[string]string) *domain.TraceContext
}
