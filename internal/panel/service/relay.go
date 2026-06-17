package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// RelayService manages CDN/relay chain definitions.
type RelayService struct {
	repo  port.RelayChainRepository
	nodes port.NodeRepository
	now   func() time.Time
}

// NewRelayService wires dependencies.
func NewRelayService(repo port.RelayChainRepository, nodes port.NodeRepository) *RelayService {
	return &RelayService{repo: repo, nodes: nodes, now: time.Now}
}

// CreateChainInput describes a new relay chain.
type CreateChainInput struct {
	Name   string
	NodeID uuid.UUID
	Hops   []domain.RelayHop
}

// CreateChain persists a new relay chain.
func (s *RelayService) CreateChain(ctx context.Context, in CreateChainInput) (*domain.RelayChain, error) {
	if in.Name == "" {
		return nil, errors.New("chain name is required")
	}
	if len(in.Hops) == 0 {
		return nil, errors.New("at least one hop is required")
	}
	// Verify node exists.
	if _, err := s.nodes.GetByID(ctx, in.NodeID); err != nil {
		return nil, errors.New("node not found")
	}
	r := &domain.RelayChain{
		ID:        uuid.New(),
		Name:      in.Name,
		NodeID:    in.NodeID,
		Hops:      in.Hops,
		Enabled:   true,
		CreatedAt: s.now(),
	}
	if err := s.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// ListChains returns all relay chains.
func (s *RelayService) ListChains(ctx context.Context) ([]*domain.RelayChain, error) {
	return s.repo.List(ctx)
}

// UpdateChain updates a relay chain.
func (s *RelayService) UpdateChain(ctx context.Context, id uuid.UUID, name string, hops []domain.RelayHop, enabled bool) (*domain.RelayChain, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		r.Name = name
	}
	if hops != nil {
		r.Hops = hops
	}
	r.Enabled = enabled
	if err := s.repo.Update(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

// DeleteChain removes a relay chain.
func (s *RelayService) DeleteChain(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
