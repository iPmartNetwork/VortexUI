package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// RoutingService manages a node's routing rules and reconciles the live core
// after every change via the SyncService.
type RoutingService struct {
	repo port.RoutingRepository
	sync *SyncService
}

// NewRoutingService wires the service.
func NewRoutingService(repo port.RoutingRepository, sync *SyncService) *RoutingService {
	return &RoutingService{repo: repo, sync: sync}
}

// RoutingRuleInput describes a routing rule for create or update.
type RoutingRuleInput struct {
	Priority    int
	Name        string
	InboundTags []string
	Domains     []string
	IP          []string
	Port        string
	Protocols   []string
	Network     string
	OutboundTag string
	BalancerTag string
	Enabled     *bool // nil defaults to enabled
}

// Create validates and persists a routing rule, then resyncs its node.
func (s *RoutingService) Create(ctx context.Context, nodeID uuid.UUID, in RoutingRuleInput) (*domain.RoutingRule, error) {
	rule := ruleFromInput(uuid.New(), nodeID, in, true)
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, nodeID); err != nil {
		return rule, errors.Join(errors.New("routing rule saved but node resync failed"), err)
	}
	return rule, nil
}

// Update applies changes to a routing rule and resyncs its node.
func (s *RoutingService) Update(ctx context.Context, id uuid.UUID, in RoutingRuleInput) (*domain.RoutingRule, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	rule := ruleFromInput(existing.ID, existing.NodeID, in, existing.Enabled)
	if err := rule.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, rule); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, rule.NodeID); err != nil {
		return rule, errors.Join(errors.New("routing rule saved but node resync failed"), err)
	}
	return rule, nil
}

// Delete removes a routing rule and resyncs its node.
func (s *RoutingService) Delete(ctx context.Context, id uuid.UUID) error {
	rule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.sync.Resync(ctx, rule.NodeID)
}

// ListByNode returns a node's routing rules in priority order.
func (s *RoutingService) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RoutingRule, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

func ruleFromInput(id, nodeID uuid.UUID, in RoutingRuleInput, defEnabled bool) *domain.RoutingRule {
	return &domain.RoutingRule{
		ID:          id,
		NodeID:      nodeID,
		Priority:    in.Priority,
		Name:        in.Name,
		InboundTags: in.InboundTags,
		Domains:     in.Domains,
		IP:          in.IP,
		Port:        in.Port,
		Protocols:   in.Protocols,
		Network:     in.Network,
		OutboundTag: in.OutboundTag,
		BalancerTag: in.BalancerTag,
		Enabled:     boolOr(in.Enabled, defEnabled),
	}
}
