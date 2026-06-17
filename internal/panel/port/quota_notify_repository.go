package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// QuotaNotifyRepository persists quota notification config and events.
type QuotaNotifyRepository interface {
	GetConfig(ctx context.Context) (*domain.QuotaNotificationConfig, error)
	SaveConfig(ctx context.Context, c *domain.QuotaNotificationConfig) error
	SaveEvent(ctx context.Context, e *domain.QuotaNotificationEvent) error
	ListEvents(ctx context.Context, limit int) ([]*domain.QuotaNotificationEvent, error)
}
