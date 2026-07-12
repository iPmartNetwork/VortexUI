package service

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DoHService manages the built-in DNS-over-HTTPS server configuration.
type DoHService struct {
	repo   port.DoHRepository
	server DoHRuntime
}

// DoHRuntime starts or stops the live DoH listener.
type DoHRuntime interface {
	Reload(ctx context.Context) error
}

// NewDoHService wires the DoH service.
func NewDoHService(repo port.DoHRepository) *DoHService {
	return &DoHService{repo: repo}
}

// SetRuntime attaches the live DoH server for config reloads.
func (s *DoHService) SetRuntime(r DoHRuntime) { s.server = r }

// GetConfig returns the current DoH configuration.
func (s *DoHService) GetConfig(ctx context.Context) (*domain.DoHConfig, error) {
	c, err := s.repo.GetConfig(ctx)
	if err != nil {
		def := domain.DefaultDoHConfig()
		return &def, nil
	}
	return c, nil
}

// UpdateConfig saves the DoH configuration.
func (s *DoHService) UpdateConfig(ctx context.Context, c *domain.DoHConfig) error {
	if err := s.repo.SaveConfig(ctx, c); err != nil {
		return err
	}
	if s.server != nil {
		return s.server.Reload(ctx)
	}
	return nil
}

// GetStats returns DNS resolution statistics.
func (s *DoHService) GetStats(ctx context.Context) (*port.DoHStats, error) {
	stats, err := s.repo.GetStats(ctx)
	if err != nil || stats == nil {
		return &port.DoHStats{}, nil
	}
	return stats, nil
}

// GetQueryLogs returns recent DNS query logs.
func (s *DoHService) GetQueryLogs(ctx context.Context, limit int) ([]domain.DoHQueryLog, error) {
	if limit <= 0 {
		limit = 100
	}
	logs, err := s.repo.ListQueryLogs(ctx, limit)
	if err != nil {
		return nil, nil
	}
	return logs, nil
}
