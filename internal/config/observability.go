package config

import (
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// ObservabilityConfig holds observability settings
type ObservabilityConfig struct {
	MetricsEnabled bool
	HealthEnabled  bool
	LoggingEnabled bool
	TracingEnabled bool
	MaxLogs        int
	CleanupInterval time.Duration
}

// DefaultObservabilityConfig returns default settings
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		MetricsEnabled:  true,
		HealthEnabled:   true,
		LoggingEnabled:  true,
		TracingEnabled:  true,
		MaxLogs:         1000,
		CleanupInterval: 1 * time.Hour,
	}
}

// InitializeObservability sets up all observability services
func InitializeObservability(cfg *ObservabilityConfig, log *slog.Logger) (
	*service.MetricsCollectorService,
	*service.HealthCheckService,
	*service.StructuredLoggerService,
	*service.TraceManagerService,
	*service.PrometheusExporter,
) {
	if log == nil {
		log = slog.Default()
	}

	var metricsService *service.MetricsCollectorService
	var healthService *service.HealthCheckService
	var loggerService *service.StructuredLoggerService
	var traceService *service.TraceManagerService
	var prometheusExporter *service.PrometheusExporter

	if cfg.MetricsEnabled {
		metricsService = service.NewMetricsCollectorService(log)
		log.Info("initialized metrics collector service")
	}

	if cfg.HealthEnabled {
		healthService = service.NewHealthCheckService(log)
		log.Info("initialized health check service")
	}

	if cfg.LoggingEnabled {
		loggerService = service.NewStructuredLoggerService(log, cfg.MaxLogs)
		log.Info("initialized structured logger service")
	}

	if cfg.TracingEnabled {
		traceService = service.NewTraceManagerService(log)
		log.Info("initialized trace manager service")
	}

	// Prometheus exporter depends on metrics service
	if metricsService != nil {
		prometheusExporter = service.NewPrometheusExporter(metricsService, log)
		log.Info("initialized prometheus exporter")
	}

	return metricsService, healthService, loggerService, traceService, prometheusExporter
}
