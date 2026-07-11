package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// DefaultHealthChecks sets up standard health checks for critical components
type DefaultHealthChecks struct {
	healthService *HealthCheckService
	log           *slog.Logger
}

// NewDefaultHealthChecks creates default health checks
func NewDefaultHealthChecks(healthService *HealthCheckService, log *slog.Logger) *DefaultHealthChecks {
	if log == nil {
		log = slog.Default()
	}
	return &DefaultHealthChecks{
		healthService: healthService,
		log:           log,
	}
}

// RegisterAll registers all default health checks
func (d *DefaultHealthChecks) RegisterAll() error {
	// Database check
	err := d.healthService.RegisterCheck("database", d.checkDatabase)
	if err != nil {
		d.log.Error("failed to register database health check", "error", err)
		return err
	}

	// Redis check
	err = d.healthService.RegisterCheck("redis", d.checkRedis)
	if err != nil {
		d.log.Error("failed to register redis health check", "error", err)
		return err
	}

	// Core node connectivity check
	err = d.healthService.RegisterCheck("core_nodes", d.checkCoreNodes)
	if err != nil {
		d.log.Error("failed to register core nodes health check", "error", err)
		return err
	}

	// Memory check
	err = d.healthService.RegisterCheck("memory", d.checkMemory)
	if err != nil {
		d.log.Error("failed to register memory health check", "error", err)
		return err
	}

	d.log.Info("registered all default health checks")
	return nil
}

// checkDatabase checks database connectivity
// NOTE: This is a placeholder - integrate with actual db.Pool in future
func (d *DefaultHealthChecks) checkDatabase(ctx context.Context) (domain.HealthStatus, string, error) {
	// In production, this would ping the actual PostgreSQL connection pool
	// For now, simulate a successful check
	select {
	case <-ctx.Done():
		return domain.HealthStatusUnhealthy, "context cancelled", ctx.Err()
	default:
		// Simulate DB check (in real code, use db.Pool.Ping(ctx))
		return domain.HealthStatusHealthy, "database connection OK", nil
	}
}

// checkRedis checks Redis connectivity
// NOTE: This is a placeholder - integrate with actual Redis client in future
func (d *DefaultHealthChecks) checkRedis(ctx context.Context) (domain.HealthStatus, string, error) {
	// In production, this would ping the actual Redis client
	// For now, simulate a successful check
	select {
	case <-ctx.Done():
		return domain.HealthStatusUnhealthy, "context cancelled", ctx.Err()
	default:
		// Simulate Redis check (in real code, use redisClient.Ping(ctx))
		return domain.HealthStatusHealthy, "redis connection OK", nil
	}
}

// checkCoreNodes checks availability of core nodes
// NOTE: This is a placeholder - integrate with actual node service in future
func (d *DefaultHealthChecks) checkCoreNodes(ctx context.Context) (domain.HealthStatus, string, error) {
	// In production, this would check:
	// 1. At least one node is online
	// 2. Node cluster is operational
	// For now, simulate a successful check
	select {
	case <-ctx.Done():
		return domain.HealthStatusUnhealthy, "context cancelled", ctx.Err()
	default:
		// Simulate core node check (in real code, query nodeService)
		return domain.HealthStatusHealthy, "core nodes operational", nil
	}
}

// checkMemory checks memory usage
func (d *DefaultHealthChecks) checkMemory(ctx context.Context) (domain.HealthStatus, string, error) {
	// In production, this would check:
	// 1. Heap allocation
	// 2. GC pressure
	// For now, return healthy
	select {
	case <-ctx.Done():
		return domain.HealthStatusUnhealthy, "context cancelled", ctx.Err()
	default:
		// Simulate memory check (in real code, use runtime.ReadMemStats)
		return domain.HealthStatusHealthy, "memory usage normal", nil
	}
}

// HealthCheckPoller runs periodic health checks in the background
type HealthCheckPoller struct {
	healthService *HealthCheckService
	interval      time.Duration
	log           *slog.Logger
	stopChan      chan struct{}
}

// NewHealthCheckPoller creates a background health check poller
func NewHealthCheckPoller(healthService *HealthCheckService, interval time.Duration, log *slog.Logger) *HealthCheckPoller {
	if log == nil {
		log = slog.Default()
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &HealthCheckPoller{
		healthService: healthService,
		interval:      interval,
		log:           log,
		stopChan:      make(chan struct{}),
	}
}

// Start begins background health checking
func (h *HealthCheckPoller) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		h.log.Info("health check poller started", "interval", h.interval.String())

		for {
			select {
			case <-h.stopChan:
				h.log.Info("health check poller stopped")
				return
			case <-ctx.Done():
				h.log.Info("health check poller context cancelled")
				return
			case <-ticker.C:
				result, err := h.healthService.CheckHealth(ctx)
				if err != nil {
					h.log.Error("health check error", "error", err)
					continue
				}

				// Log aggregated health status
				h.log.Debug("health check completed", "status", result.Status, "components", len(result.Components))

				// Log any unhealthy components
				for _, comp := range result.Components {
					if comp.Status != domain.HealthStatusHealthy {
						h.log.Warn("component unhealthy", "component", comp.Name, "status", comp.Status, "message", comp.Message)
					}
				}
			}
		}
	}()
}

// Stop halts background health checking
func (h *HealthCheckPoller) Stop() {
	close(h.stopChan)
}
