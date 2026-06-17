package service

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DoHService manages the built-in DNS-over-HTTPS server configuration.
type DoHService struct {
	repo port.DoHRepository
}

// NewDoHService wires the DoH service.
func NewDoHService(repo port.DoHRepository) *DoHService {
	return &DoHService{repo: repo}
}

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
	return s.repo.SaveConfig(ctx, c)
}

// GetStats returns DNS resolution statistics.
func (s *DoHService) GetStats(ctx context.Context) (*port.DoHStats, error) {
	return s.repo.GetStats(ctx)
}

// GetQueryLogs returns recent DNS query logs.
func (s *DoHService) GetQueryLogs(ctx context.Context, limit int) ([]domain.DoHQueryLog, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.repo.ListQueryLogs(ctx, limit)
}
