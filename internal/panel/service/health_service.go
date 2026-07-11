package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// HealthCheckService implements health checking
type HealthCheckService struct {
	checks map[string]func(context.Context) (domain.HealthStatus, string, error)
	mu     sync.RWMutex
	log    *slog.Logger
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(log *slog.Logger) *HealthCheckService {
	if log == nil {
		log = slog.Default()
	}
	return &HealthCheckService{
		checks: make(map[string]func(context.Context) (domain.HealthStatus, string, error)),
		log:    log,
	}
}

// RegisterCheck registers a health check for a component
func (h *HealthCheckService) RegisterCheck(name string, fn func(context.Context) (domain.HealthStatus, string, error)) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[name] = fn
	h.log.Info("registered health check", "component", name)
	return nil
}

// CheckComponent checks the health of a specific component
func (h *HealthCheckService) CheckComponent(ctx context.Context, name string) (*domain.ComponentHealth, error) {
	h.mu.RLock()
	checkFn, ok := h.checks[name]
	h.mu.RUnlock()

	if !ok {
		return &domain.ComponentHealth{
			Name:      name,
			Status:    domain.HealthStatusUnhealthy,
			Message:   "component not found",
			LastCheck: time.Now(),
		}, nil
	}

	start := time.Now()
	status, message, err := checkFn(ctx)
	duration := time.Since(start)

	if err != nil {
		status = domain.HealthStatusUnhealthy
		if message == "" {
			message = err.Error()
		}
	}

	return &domain.ComponentHealth{
		Name:       name,
		Status:     status,
		Message:    message,
		LastCheck:  time.Now(),
		Duration:   duration.Milliseconds(),
	}, nil
}

// CheckHealth performs a full health check
func (h *HealthCheckService) CheckHealth(ctx context.Context) (*domain.HealthCheckResult, error) {
	h.mu.RLock()
	checkNames := make([]string, 0, len(h.checks))
	for name := range h.checks {
		checkNames = append(checkNames, name)
	}
	h.mu.RUnlock()

	components := make([]*domain.ComponentHealth, len(checkNames))
	overallStatus := domain.HealthStatusHealthy

	for i, name := range checkNames {
		component, _ := h.CheckComponent(ctx, name)
		components[i] = component

		if component.Status == domain.HealthStatusUnhealthy {
			overallStatus = domain.HealthStatusUnhealthy
		} else if component.Status == domain.HealthStatusDegraded && overallStatus == domain.HealthStatusHealthy {
			overallStatus = domain.HealthStatusDegraded
		}
	}

	return &domain.HealthCheckResult{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Components: components,
		Message:    "health check completed",
	}, nil
}
