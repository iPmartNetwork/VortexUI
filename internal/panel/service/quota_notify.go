package service

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// QuotaNotifyService manages smart quota notification logic.
type QuotaNotifyService struct {
	repo port.QuotaNotifyRepository
}

func NewQuotaNotifyService(repo port.QuotaNotifyRepository) *QuotaNotifyService {
	return &QuotaNotifyService{repo: repo}
}

func (s *QuotaNotifyService) GetConfig(ctx context.Context) (*domain.QuotaNotificationConfig, error) {
	c, err := s.repo.GetConfig(ctx)
	if err != nil {
		def := domain.DefaultQuotaNotificationConfig()
		return &def, nil
	}
	return c, nil
}

func (s *QuotaNotifyService) UpdateConfig(ctx context.Context, c *domain.QuotaNotificationConfig) error {
	return s.repo.SaveConfig(ctx, c)
}

func (s *QuotaNotifyService) ListEvents(ctx context.Context, limit int) ([]*domain.QuotaNotificationEvent, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit)
}
