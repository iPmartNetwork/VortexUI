package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// MigrationRepository persists migration events and policy.
type MigrationRepository interface {
	SaveEvent(ctx context.Context, e *domain.MigrationEvent) error
	ListEvents(ctx context.Context, limit, offset int) ([]*domain.MigrationEvent, int, error)
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.MigrationEvent, error)

	GetPolicy(ctx context.Context) (*domain.MigrationPolicy, error)
	SavePolicy(ctx context.Context, p *domain.MigrationPolicy) error
}
