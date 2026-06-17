package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// TLSTricksRepository persists TLS trick profiles.
type TLSTricksRepository interface {
	Create(ctx context.Context, p *domain.TLSTrickProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TLSTrickProfile, error)
	Update(ctx context.Context, p *domain.TLSTrickProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.TLSTrickProfile, error)
}
