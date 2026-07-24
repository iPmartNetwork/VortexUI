package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ApprovalQueueRepository persists subscription approval requests.
type ApprovalQueueRepository interface {
	Create(ctx context.Context, a *domain.SubscriptionApproval) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionApproval, error)
	ListPending(ctx context.Context) ([]*domain.SubscriptionApproval, error)
	Approve(ctx context.Context, id, adminID uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
}
