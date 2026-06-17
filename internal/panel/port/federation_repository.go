package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// FederationRepository persists federation peers, config, and sync events.
type FederationRepository interface {
	GetConfig(ctx context.Context) (*domain.FederationConfig, error)
	SaveConfig(ctx context.Context, c *domain.FederationConfig) error

	CreatePeer(ctx context.Context, p *domain.FederationPeer) error
	GetPeer(ctx context.Context, id uuid.UUID) (*domain.FederationPeer, error)
	UpdatePeer(ctx context.Context, p *domain.FederationPeer) error
	DeletePeer(ctx context.Context, id uuid.UUID) error
	ListPeers(ctx context.Context) ([]*domain.FederationPeer, error)

	SaveSyncEvent(ctx context.Context, e *domain.FederationSyncEvent) error
	ListSyncEvents(ctx context.Context, limit int) ([]*domain.FederationSyncEvent, error)
}
