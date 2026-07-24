package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ConfigVersionRepository persists configuration version snapshots for inbounds.
type ConfigVersionRepository interface {
	// Create stores a new config version. The version number should be pre-calculated.
	Create(ctx context.Context, v *domain.ConfigVersion) error
	// GetLatest retrieves the most recent version for an inbound.
	GetLatest(ctx context.Context, inboundID uuid.UUID) (*domain.ConfigVersion, error)
	// GetByVersion retrieves a specific version number for an inbound.
	GetByVersion(ctx context.Context, inboundID uuid.UUID, version int) (*domain.ConfigVersion, error)
	// ListForInbound returns all versions for an inbound, newest first.
	ListForInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.ConfigVersion, error)
	// NextVersion returns the next version number for the inbound (max + 1).
	NextVersion(ctx context.Context, inboundID uuid.UUID) (int, error)
}
