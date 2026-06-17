package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// DecoySiteRepository persists decoy website configurations.
type DecoySiteRepository interface {
	Create(ctx context.Context, d *domain.DecoySite) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.DecoySite, error)
	GetGlobal(ctx context.Context) (*domain.DecoySite, error)
	GetByNode(ctx context.Context, nodeID uuid.UUID) (*domain.DecoySite, error)
	Update(ctx context.Context, d *domain.DecoySite) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.DecoySite, error)
}
