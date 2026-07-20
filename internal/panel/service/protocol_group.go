package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ProtocolGroupService manages protocol groups for auto-protocol switching.
type ProtocolGroupService struct {
	groups   port.ProtocolGroupRepository
	isp      port.ISPProfileRepository
	inbounds port.InboundRepository
	nodes    port.NodeRepository
	sync     *SyncService
	pub      events.Publisher
}

func NewProtocolGroupService(
	groups port.ProtocolGroupRepository,
	isp port.ISPProfileRepository,
	inbounds port.InboundRepository,
	nodes port.NodeRepository,
	sync *SyncService,
) *ProtocolGroupService {
	return &ProtocolGroupService{
		groups:   groups,
		isp:      isp,
		inbounds: inbounds,
		nodes:    nodes,
		sync:     sync,
		pub:      events.Nop{},
	}
}

func (s *ProtocolGroupService) SetPublisher(p events.Publisher) {
	if p != nil {
		s.pub = p
	}
}

// Create validates and persists a new protocol group.
func (s *ProtocolGroupService) Create(ctx context.Context, g *domain.ProtocolGroup) error {
	if g.Name == "" {
		return errors.New("name is required")
	}
	if err := s.validateProbeSettings(g); err != nil {
		return err
	}
	if err := s.validateInboundsSameNode(ctx, g.NodeID, g.InboundIDs); err != nil {
		return err
	}
	return s.groups.Create(ctx, g)
}

// Update validates and persists changes to an existing protocol group.
func (s *ProtocolGroupService) Update(ctx context.Context, g *domain.ProtocolGroup) error {
	if g.Name == "" {
		return errors.New("name is required")
	}
	if err := s.validateProbeSettings(g); err != nil {
		return err
	}
	if err := s.validateInboundsSameNode(ctx, g.NodeID, g.InboundIDs); err != nil {
		return err
	}
	if err := s.groups.Update(ctx, g); err != nil {
		return err
	}
	// Trigger node resync so subscription configs are refreshed.
	if s.sync != nil {
		_ = s.sync.Resync(ctx, g.NodeID)
	}
	return nil
}

// Delete removes a protocol group.
func (s *ProtocolGroupService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.groups.Delete(ctx, id)
}

// GetByID returns a single protocol group.
func (s *ProtocolGroupService) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProtocolGroup, error) {
	return s.groups.GetByID(ctx, id)
}

// ListByNode returns all protocol groups for a node.
func (s *ProtocolGroupService) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.ProtocolGroup, error) {
	return s.groups.ListByNode(ctx, nodeID)
}

// ReorderInbounds sets a new priority order for inbounds within a group.
func (s *ProtocolGroupService) ReorderInbounds(ctx context.Context, groupID uuid.UUID, newOrder []uuid.UUID) error {
	g, err := s.groups.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	// Validate all IDs in newOrder belong to the group.
	existing := make(map[uuid.UUID]bool, len(g.InboundIDs))
	for _, id := range g.InboundIDs {
		existing[id] = true
	}
	for _, id := range newOrder {
		if !existing[id] {
			return fmt.Errorf("inbound %s is not in group %s", id, groupID)
		}
	}
	g.InboundIDs = newOrder
	return s.groups.Update(ctx, g)
}

// GroupsForInbounds returns protocol groups containing any of the given inbound IDs.
func (s *ProtocolGroupService) GroupsForInbounds(ctx context.Context, inboundIDs []uuid.UUID) ([]*domain.ProtocolGroup, error) {
	return s.groups.GroupsForInbounds(ctx, inboundIDs)
}

// --- ISP Profile methods ---

// CreateISPProfile validates and persists a new ISP profile for a group.
func (s *ProtocolGroupService) CreateISPProfile(ctx context.Context, p *domain.ISPProfile) error {
	// Validate group exists.
	if _, err := s.groups.GetByID(ctx, p.GroupID); err != nil {
		return fmt.Errorf("group not found: %w", err)
	}
	return s.isp.Create(ctx, p)
}

// UpdateISPProfile persists changes to an existing ISP profile.
func (s *ProtocolGroupService) UpdateISPProfile(ctx context.Context, p *domain.ISPProfile) error {
	return s.isp.Update(ctx, p)
}

// DeleteISPProfile removes an ISP profile.
func (s *ProtocolGroupService) DeleteISPProfile(ctx context.Context, id uuid.UUID) error {
	return s.isp.Delete(ctx, id)
}

// ListISPProfiles returns all ISP profiles for a group.
func (s *ProtocolGroupService) ListISPProfiles(ctx context.Context, groupID uuid.UUID) ([]*domain.ISPProfile, error) {
	return s.isp.ListByGroup(ctx, groupID)
}

// MatchISPProfiles finds matching ISP profiles for a given ISP identifier and group IDs.
func (s *ProtocolGroupService) MatchISPProfiles(ctx context.Context, isp string, groupIDs []uuid.UUID) ([]*domain.ISPProfile, error) {
	return s.isp.MatchForGroups(ctx, isp, groupIDs)
}

// --- Validation helpers ---

func (s *ProtocolGroupService) validateProbeSettings(g *domain.ProtocolGroup) error {
	if g.ProbeInterval < domain.MinProbeInterval {
		return fmt.Errorf("probe_interval must be >= %d seconds", domain.MinProbeInterval)
	}
	if g.ProbeInterval > domain.MaxProbeInterval {
		return fmt.Errorf("probe_interval must be <= %d seconds", domain.MaxProbeInterval)
	}
	if g.MaxRetries < domain.MinRetries || g.MaxRetries > domain.MaxRetries {
		return fmt.Errorf("max_retries must be between %d and %d", domain.MinRetries, domain.MaxRetries)
	}
	if g.ProbeURL == "" {
		g.ProbeURL = "https://www.gstatic.com/generate_204"
	}
	if g.ProbeTimeout <= 0 {
		g.ProbeTimeout = 5
	}
	return nil
}

func (s *ProtocolGroupService) validateInboundsSameNode(ctx context.Context, nodeID uuid.UUID, inboundIDs []uuid.UUID) error {
	if len(inboundIDs) == 0 {
		return nil
	}
	for _, ibID := range inboundIDs {
		ib, err := s.inbounds.GetByID(ctx, ibID)
		if err != nil {
			return fmt.Errorf("inbound %s not found", ibID)
		}
		if ib.NodeID != nodeID {
			return fmt.Errorf("all inbounds must belong to the same node (inbound %s belongs to a different node)", ibID)
		}
	}
	return nil
}
