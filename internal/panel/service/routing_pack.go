package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// RuleApplier persists a single routing rule for a node and reconciles the live
// core. It is satisfied by *RoutingService, letting RoutingPackService reuse the
// existing validate+resync path instead of duplicating routing-sync logic.
//
// Create's contract (see RoutingService.Create) is the seam this package relies
// on: on a resync failure it returns the saved rule together with a non-nil
// error, so a non-nil rule with a non-nil error means "saved but resync failed".
type RuleApplier interface {
	Create(ctx context.Context, nodeID uuid.UUID, in RoutingRuleInput) (*domain.RoutingRule, error)
}

// OutboundApplier persists a single outbound for a node and reconciles the live
// core. It is satisfied by *OutboundService and shares the same "non-nil result
// + non-nil error == saved but resync failed" contract as RuleApplier.
type OutboundApplier interface {
	Create(ctx context.Context, in CreateOutboundInput) (*domain.Outbound, error)
}

// RoutingPackService lists built-in and custom routing packs, manages custom
// packs, applies a pack's rules (and any outbounds they need) to a node by
// reusing RoutingService/OutboundService, and tracks pack selection (a global
// default and an optional per-user override).
type RoutingPackService struct {
	repo      port.RoutingPackRepository
	routing   RuleApplier
	outbounds OutboundApplier
}

// NewRoutingPackService wires the service. routing/outbounds are the seams used
// by ApplyToNode; they are *RoutingService and *OutboundService in production.
func NewRoutingPackService(repo port.RoutingPackRepository, routing RuleApplier, outbounds OutboundApplier) *RoutingPackService {
	return &RoutingPackService{repo: repo, routing: routing, outbounds: outbounds}
}

// RoutingPackInput carries the mutable fields of a custom routing pack.
type RoutingPackInput struct {
	Name        string
	Description string
	Category    string
	Rules       []domain.RoutingRule
	Outbounds   []domain.Outbound
}

// ListPacks returns the built-in packs merged with persisted custom packs
// (built-ins first), satisfying Requirement 3.1.1.
func (s *RoutingPackService) ListPacks(ctx context.Context) ([]domain.RoutingPack, error) {
	packs := domain.BuiltinRoutingPacks()
	custom, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range custom {
		packs = append(packs, *c)
	}
	return packs, nil
}

// GetPack resolves a pack by ID: a built-in by its name, otherwise a custom pack
// by its UUID. Returns domain.ErrNotFound when neither matches.
func (s *RoutingPackService) GetPack(ctx context.Context, id string) (*domain.RoutingPack, error) {
	for _, p := range domain.BuiltinRoutingPacks() {
		if p.ID == id {
			pack := p
			return &pack, nil
		}
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	pack, err := s.repo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}
	if pack == nil {
		return nil, domain.ErrNotFound
	}
	return pack, nil
}

// Create persists a new custom pack with a freshly generated UUID.
func (s *RoutingPackService) Create(ctx context.Context, in RoutingPackInput) (*domain.RoutingPack, error) {
	if in.Name == "" {
		return nil, errors.New("pack name is required")
	}
	pack := &domain.RoutingPack{
		ID:          uuid.New().String(),
		Name:        in.Name,
		Description: in.Description,
		Category:    in.Category,
		Builtin:     false,
		Rules:       in.Rules,
		Outbounds:   in.Outbounds,
	}
	if err := s.repo.Create(ctx, pack); err != nil {
		return nil, err
	}
	return pack, nil
}

// Update changes a custom pack. Built-in packs are immutable.
func (s *RoutingPackService) Update(ctx context.Context, id uuid.UUID, in RoutingPackInput) (*domain.RoutingPack, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	if in.Name == "" {
		return nil, errors.New("pack name is required")
	}
	existing.Name = in.Name
	existing.Description = in.Description
	existing.Category = in.Category
	existing.Rules = in.Rules
	existing.Outbounds = in.Outbounds
	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete removes a custom pack by ID.
func (s *RoutingPackService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

// ErrPackResyncFailed reports that a pack's rules/outbounds were saved but the
// node resync did not complete (Requirement 3.2.2). Callers can map it to a
// non-fatal "saved but resync failed" response.
var ErrPackResyncFailed = errors.New("routing pack saved but node resync failed")

// ApplyToNode expands a pack onto a node: it creates the pack's outbounds first
// (rules may reference their tags), then each routing rule, reusing
// OutboundService/RoutingService so validation and core resync happen exactly
// once per item through the existing path (no routing-sync duplication).
//
// A resync failure on any item is non-fatal: the rules/outbounds are persisted
// and errResyncFailed is returned so the caller can surface "saved but resync
// failed". A validation/persistence failure aborts and is returned as-is.
func (s *RoutingPackService) ApplyToNode(ctx context.Context, nodeID uuid.UUID, packID string) error {
	pack, err := s.GetPack(ctx, packID)
	if err != nil {
		return err
	}

	resyncFailed := false

	for _, o := range pack.Outbounds {
		out, err := s.outbounds.Create(ctx, outboundInputFromPack(nodeID, o))
		if err != nil {
			// A non-nil outbound with an error means it was saved but the
			// resync failed; keep going. A nil outbound is a hard failure.
			if out != nil {
				resyncFailed = true
				continue
			}
			return err
		}
	}

	for _, r := range pack.Rules {
		rule, err := s.routing.Create(ctx, nodeID, ruleInputFromPack(r))
		if err != nil {
			if rule != nil {
				resyncFailed = true
				continue
			}
			return err
		}
	}

	if resyncFailed {
		return ErrPackResyncFailed
	}
	return nil
}

// SetGlobalDefault persists the global default pack selection.
func (s *RoutingPackService) SetGlobalDefault(ctx context.Context, packID string) error {
	return s.repo.SetGlobalDefault(ctx, packID)
}

// GetGlobalDefault returns the global default pack ID ("" when none).
func (s *RoutingPackService) GetGlobalDefault(ctx context.Context) (string, error) {
	return s.repo.GetGlobalDefault(ctx)
}

// SetUserPack persists a per-subscription pack selection for a user.
func (s *RoutingPackService) SetUserPack(ctx context.Context, userID uuid.UUID, packID string) error {
	return s.repo.SetUserPack(ctx, userID, packID)
}

// GetUserPack returns a user's per-subscription pack ID ("" when none).
func (s *RoutingPackService) GetUserPack(ctx context.Context, userID uuid.UUID) (string, error) {
	return s.repo.GetUserPack(ctx, userID)
}

// ruleInputFromPack maps a pack's engine-neutral rule onto the routing-create
// input. Node-specific fields (ID/NodeID) are intentionally dropped; Create
// assigns them fresh for the target node.
func ruleInputFromPack(r domain.RoutingRule) RoutingRuleInput {
	return RoutingRuleInput{
		Priority:    r.Priority,
		Name:        r.Name,
		InboundTags: r.InboundTags,
		Domains:     r.Domains,
		IP:          r.IP,
		Port:        r.Port,
		Protocols:   r.Protocols,
		Network:     r.Network,
		OutboundTag: r.OutboundTag,
		BalancerTag: r.BalancerTag,
	}
}

// outboundInputFromPack maps a pack's outbound onto the outbound-create input
// for the target node.
func outboundInputFromPack(nodeID uuid.UUID, o domain.Outbound) CreateOutboundInput {
	return CreateOutboundInput{
		NodeID:   nodeID,
		Tag:      o.Tag,
		Protocol: o.Protocol,
		Address:  o.Address,
		Port:     o.Port,
		UUID:     o.UUID,
		Password: o.Password,
		Username: o.Username,
		Method:   o.Method,
		Flow:     o.Flow,
		Network:  o.Network,
		Security: o.Security,
		SNI:      o.SNI,
		Path:     o.Path,
		Host:     o.Host,
		Raw:      o.Raw,
	}
}
