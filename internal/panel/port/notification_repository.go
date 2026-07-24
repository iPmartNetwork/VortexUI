package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// NotificationChannelRepository persists notification channel configurations.
type NotificationChannelRepository interface {
	Create(ctx context.Context, ch *domain.NotificationChannel) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.NotificationChannel, error)
	Update(ctx context.Context, ch *domain.NotificationChannel) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.NotificationChannel, error)
	ListByScope(ctx context.Context, scopeType, scopeID string) ([]*domain.NotificationChannel, error)
}

// WebhookDeliveryRepository persists webhook delivery attempts and retry state.
type WebhookDeliveryRepository interface {
	Create(ctx context.Context, d *domain.WebhookDelivery) error
	ListPending(ctx context.Context) ([]*domain.WebhookDelivery, error)
	MarkDelivered(ctx context.Context, id uuid.UUID) error
	IncrementAttempt(ctx context.Context, id uuid.UUID, nextRetry *time.Time, statusCode int) error
}
