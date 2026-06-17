package service

import (
	"context"
	"fmt"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DeepLinkService manages deep link and QR subscription delivery.
type DeepLinkService struct {
	repo port.DeepLinkRepository
}

func NewDeepLinkService(repo port.DeepLinkRepository) *DeepLinkService {
	return &DeepLinkService{repo: repo}
}

func (s *DeepLinkService) GetConfig(ctx context.Context) (*domain.DeepLinkConfig, error) {
	c, err := s.repo.GetConfig(ctx)
	if err != nil {
		def := domain.DefaultDeepLinkConfig()
		return &def, nil
	}
	return c, nil
}

func (s *DeepLinkService) UpdateConfig(ctx context.Context, c *domain.DeepLinkConfig) error {
	return s.repo.SaveConfig(ctx, c)
}

// GenerateDeepLink creates a deep link URL for a subscription token.
func (s *DeepLinkService) GenerateDeepLink(ctx context.Context, subToken string) (string, error) {
	cfg, _ := s.GetConfig(ctx)
	if !cfg.Enabled {
		return "", fmt.Errorf("deep links are disabled")
	}
	// Format: vortex://import?sub=<base_url>/sub/<token>
	subURL := fmt.Sprintf("%s/sub/%s", cfg.BaseURL, subToken)
	deepLink := fmt.Sprintf("%s://import?sub=%s", cfg.Scheme, subURL)
	return deepLink, nil
}
