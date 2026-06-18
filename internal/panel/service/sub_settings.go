package service

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// SubSettingsRepository persists subscription settings.
type SubSettingsRepository interface {
	Get(ctx context.Context) (*domain.SubSettings, error)
	Save(ctx context.Context, s *domain.SubSettings) error
}

// SubSettingsService manages panel-configurable subscription settings.
type SubSettingsService struct {
	repo SubSettingsRepository
}

// NewSubSettingsService wires the service.
func NewSubSettingsService(repo SubSettingsRepository) *SubSettingsService {
	return &SubSettingsService{repo: repo}
}

// GetConfig returns the settings, falling back to defaults when unset/unreadable.
func (s *SubSettingsService) GetConfig(ctx context.Context) (*domain.SubSettings, error) {
	c, err := s.repo.Get(ctx)
	if err != nil {
		def := domain.DefaultSubSettings()
		return &def, nil
	}
	if c.UpdateInterval <= 0 {
		c.UpdateInterval = domain.DefaultSubSettings().UpdateInterval
	}
	return c, nil
}

// UpdateConfig persists the settings, clamping invalid intervals to the default.
func (s *SubSettingsService) UpdateConfig(ctx context.Context, c *domain.SubSettings) error {
	if c.UpdateInterval <= 0 {
		c.UpdateInterval = domain.DefaultSubSettings().UpdateInterval
	}
	return s.repo.Save(ctx, c)
}

// UpdateInterval returns the configured client re-fetch interval in hours,
// returning the default on any error. Used by the subscription handler.
func (s *SubSettingsService) UpdateInterval(ctx context.Context) int {
	c, err := s.GetConfig(ctx)
	if err != nil || c.UpdateInterval <= 0 {
		return domain.DefaultSubSettings().UpdateInterval
	}
	return c.UpdateInterval
}
