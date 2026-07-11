package domain

import (
	"time"

	"github.com/google/uuid"
)

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge    MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
)

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// MetricSeries represents a series of metric measurements
type MetricSeries struct {
	ID     uuid.UUID        `json:"id"`
	Name   string           `json:"name"`
	Type   MetricType       `json:"type"`
	Points []*MetricPoint   `json:"points"`
	Labels map[string]string `json:"labels,omitempty"`
}

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents health status of a component
type ComponentHealth struct {
	Name       string       `json:"name"`
	Status     HealthStatus `json:"status"`
	Message    string       `json:"message,omitempty"`
	LastCheck  time.Time    `json:"last_check"`
	Duration   int64        `json:"duration_ms"`
}

// HealthCheckResult represents overall health check result
type HealthCheckResult struct {
	Status     HealthStatus       `json:"status"`
	Timestamp  time.Time          `json:"timestamp"`
	Components []*ComponentHealth `json:"components"`
	Message    string             `json:"message,omitempty"`
}

// LogEntry represents a structured log entry
type LogEntry struct {
	ID        uuid.UUID              `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	TraceID   string                 `json:"trace_id,omitempty"`
	SpanID    string                 `json:"span_id,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Error     *ErrorInfo             `json:"error,omitempty"`
}

// ErrorInfo captures structured error information
type ErrorInfo struct {
	Message    string            `json:"message"`
	Code       string            `json:"code,omitempty"`
	StackTrace string            `json:"stack_trace,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

// TraceContext represents distributed tracing context
type TraceContext struct {
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
	ParentID  string `json:"parent_id,omitempty"`
	Sampled   bool   `json:"sampled"`
	Timestamp int64  `json:"timestamp"`
}

// RequestMetrics captures per-request metrics
type RequestMetrics struct {
	Method            string        `json:"method"`
	Path              string        `json:"path"`
	Status            int           `json:"status"`
	Latency           int64         `json:"latency_ms"`
	BytesIn           int64         `json:"bytes_in"`
	BytesOut          int64         `json:"bytes_out"`
	Timestamp         time.Time     `json:"timestamp"`
	TraceID           string        `json:"trace_id,omitempty"`
}
