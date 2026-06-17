package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// RelayChainRepository persists CDN/relay chain definitions.
type RelayChainRepository interface {
	Create(ctx context.Context, r *domain.RelayChain) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RelayChain, error)
	Update(ctx context.Context, r *domain.RelayChain) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RelayChain, error)
	List(ctx context.Context) ([]*domain.RelayChain, error)
}
