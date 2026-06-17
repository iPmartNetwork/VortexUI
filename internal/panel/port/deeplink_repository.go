package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// DeepLinkRepository persists deep link configuration.
type DeepLinkRepository interface {
	GetConfig(ctx context.Context) (*domain.DeepLinkConfig, error)
	SaveConfig(ctx context.Context, c *domain.DeepLinkConfig) error
}
