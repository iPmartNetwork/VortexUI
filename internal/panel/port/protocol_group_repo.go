package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ProtocolGroupRepository persists protocol groups and their inbound orderings.
type ProtocolGroupRepository interface {
	Create(ctx context.Context, g *domain.ProtocolGroup) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ProtocolGroup, error)
	Update(ctx context.Context, g *domain.ProtocolGroup) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.ProtocolGroup, error)
	// GroupsForInbounds returns all groups containing any of the given inbound IDs.
	GroupsForInbounds(ctx context.Context, inboundIDs []uuid.UUID) ([]*domain.ProtocolGroup, error)
}

// ISPProfileRepository persists ISP-specific switching preferences.
type ISPProfileRepository interface {
	Create(ctx context.Context, p *domain.ISPProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ISPProfile, error)
	Update(ctx context.Context, p *domain.ISPProfile) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ISPProfile, error)
	// MatchForGroups finds the ISP profile matching the given ISP identifier
	// within the specified group IDs.
	MatchForGroups(ctx context.Context, isp string, groupIDs []uuid.UUID) ([]*domain.ISPProfile, error)
}

// SwitchEventRepository persists and queries protocol switch events.
type SwitchEventRepository interface {
	Record(ctx context.Context, e *domain.SwitchEvent) error
	Summary(ctx context.Context, filter domain.SwitchEventFilter) (*domain.SwitchSummary, error)
}
