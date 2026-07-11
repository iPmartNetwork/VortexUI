package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// MetricsHandlers provides metrics endpoints
type MetricsHandlers struct {
	metricsService *service.MetricsCollectorService
}

// NewMetricsHandlers creates metrics endpoint handlers
func NewMetricsHandlers(metricsService *service.MetricsCollectorService) *MetricsHandlers {
	return &MetricsHandlers{
		metricsService: metricsService,
	}
}

// GetMetrics retrieves collected metrics
func (m *MetricsHandlers) GetMetrics(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse optional filters from query params
	filters := make(map[string]string)
	for key, values := range c.QueryParams() {
		if len(values) > 0 {
			filters[key] = values[0]
		}
	}

	metrics, err := m.metricsService.GetMetrics(ctx, filters)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"metrics": metrics,
		"count":   len(metrics),
	})
}

// HealthHandlers provides health check endpoints
type HealthHandlers struct {
	healthService *service.HealthCheckService
}

// NewHealthHandlers creates health check endpoint handlers
func NewHealthHandlers(healthService *service.HealthCheckService) *HealthHandlers {
	return &HealthHandlers{
		healthService: healthService,
	}
}

// GetHealth returns overall health status
func (h *HealthHandlers) GetHealth(c echo.Context) error {
	ctx := c.Request().Context()

	result, err := h.healthService.CheckHealth(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	statusCode := http.StatusOK
	if result.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, result)
}

// GetHealthComponent returns health status of a specific component
func (h *HealthHandlers) GetHealthComponent(c echo.Context) error {
	ctx := c.Request().Context()
	componentName := c.Param("component")

	component, err := h.healthService.CheckComponent(ctx, componentName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	statusCode := http.StatusOK
	if component.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	return c.JSON(statusCode, component)
}

// LogHandlers provides logging endpoints
type LogHandlers struct {
	loggerService *service.StructuredLoggerService
}

// NewLogHandlers creates logging endpoint handlers
func NewLogHandlers(loggerService *service.StructuredLoggerService) *LogHandlers {
	return &LogHandlers{
		loggerService: loggerService,
	}
}

// GetLogs retrieves recent logs
func (l *LogHandlers) GetLogs(c echo.Context) error {
	ctx := c.Request().Context()

	level := c.QueryParam("level")
	limitStr := c.QueryParam("limit")
	limit := 100

	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	logs, err := l.loggerService.GetLogs(ctx, level, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// TraceHandlers provides tracing endpoints
type TraceHandlers struct {
	traceService *service.TraceManagerService
}

// NewTraceHandlers creates tracing endpoint handlers
func NewTraceHandlers(traceService *service.TraceManagerService) *TraceHandlers {
	return &TraceHandlers{
		traceService: traceService,
	}
}

// GetTraceContext returns current trace context
func (t *TraceHandlers) GetTraceContext(c echo.Context) error {
	ctx := c.Request().Context()

	traceCtx := t.traceService.GetTraceContext(ctx)

	return c.JSON(http.StatusOK, traceCtx)
}

// PrometheusHandlers provides Prometheus metrics endpoint
type PrometheusHandlers struct {
	exporter *service.PrometheusExporter
}

// NewPrometheusHandlers creates Prometheus endpoint handlers
func NewPrometheusHandlers(exporter *service.PrometheusExporter) *PrometheusHandlers {
	return &PrometheusHandlers{
		exporter: exporter,
	}
}

// GetMetrics returns metrics in Prometheus format
func (p *PrometheusHandlers) GetMetrics(c echo.Context) error {
	output := p.exporter.Export()
	c.Response().Header().Set("Content-Type", "text/plain; version=0.0.4")
	return c.String(http.StatusOK, output)
}
