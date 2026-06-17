package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// QuotaPolicyRepository persists fair-use quota policies.
type QuotaPolicyRepository interface {
	Create(ctx context.Context, p *domain.QuotaPolicy) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.QuotaPolicy, error)
	GetDefault(ctx context.Context) (*domain.QuotaPolicy, error)
	Update(ctx context.Context, p *domain.QuotaPolicy) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.QuotaPolicy, error)
}
