package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// InboundLister yields a node's inbounds (port.InboundRepository satisfies it).
type InboundLister interface {
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error)
}

// UsersByNoder yields the per-inbound user map for a node (postgres.UserRepo
// satisfies it).
type UsersByNoder interface {
	UsersByNode(ctx context.Context, nodeID uuid.UUID) (map[string][]*domain.User, error)
}

// Syncer pushes a fully-assembled config to one node (*hub.Hub satisfies it).
type Syncer interface {
	Sync(ctx context.Context, nodeID uuid.UUID, cfg *core.GeneratedConfig) error
}

// SyncService rebuilds a node's complete desired configuration from the database
// and pushes it to the live core. It is invoked whenever a node's inbounds
// change; per-user edits use the lighter AddUser/RemoveUser path instead.
type SyncService struct {
	inbounds InboundLister
	users    UsersByNoder
	syncer   Syncer
}

// NewSyncService wires the service.
func NewSyncService(inbounds InboundLister, users UsersByNoder, syncer Syncer) *SyncService {
	return &SyncService{inbounds: inbounds, users: users, syncer: syncer}
}

// Resync assembles the node's enabled inbounds plus their bound users into a
// core.GeneratedConfig and pushes it. Disabled inbounds are excluded so they
// stop listening after the sync.
func (s *SyncService) Resync(ctx context.Context, nodeID uuid.UUID) error {
	inbounds, err := s.inbounds.ListByNode(ctx, nodeID)
	if err != nil {
		return err
	}
	usersByInbound, err := s.users.UsersByNode(ctx, nodeID)
	if err != nil {
		return err
	}

	cfg := &core.GeneratedConfig{
		LogLevel:       "warning",
		UsersByInbound: usersByInbound,
	}
	for _, in := range inbounds {
		if in.Enabled {
			cfg.Inbounds = append(cfg.Inbounds, *in)
		}
	}
	return s.syncer.Sync(ctx, nodeID, cfg)
}
