package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BalancerService manages a node's balancers and reconciles the live core after
// every change via the SyncService.
type BalancerService struct {
	repo port.BalancerRepository
	sync *SyncService
}

// NewBalancerService wires the service.
func NewBalancerService(repo port.BalancerRepository, sync *SyncService) *BalancerService {
	return &BalancerService{repo: repo, sync: sync}
}

// BalancerInput describes a balancer for create or update.
type BalancerInput struct {
	Tag           string
	Selectors     []string
	Strategy      domain.BalancerStrategy
	Observe       bool
	ProbeURL      string
	ProbeInterval string
	Enabled       *bool // nil defaults to enabled
}

// Create validates and persists a balancer, then resyncs its node.
func (s *BalancerService) Create(ctx context.Context, nodeID uuid.UUID, in BalancerInput) (*domain.Balancer, error) {
	b := balancerFromInput(uuid.New(), nodeID, in, true)
	if err := b.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, nodeID); err != nil {
		return b, errors.Join(errors.New("balancer saved but node resync failed"), err)
	}
	return b, nil
}

// Update applies changes to a balancer and resyncs its node.
func (s *BalancerService) Update(ctx context.Context, id uuid.UUID, in BalancerInput) (*domain.Balancer, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	b := balancerFromInput(existing.ID, existing.NodeID, in, existing.Enabled)
	b.Tag = existing.Tag // tag is immutable; routing rules reference it
	if err := b.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, err
	}
	if err := s.sync.Resync(ctx, b.NodeID); err != nil {
		return b, errors.Join(errors.New("balancer saved but node resync failed"), err)
	}
	return b, nil
}

// Delete removes a balancer and resyncs its node.
func (s *BalancerService) Delete(ctx context.Context, id uuid.UUID) error {
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.sync.Resync(ctx, b.NodeID)
}

// ListByNode returns a node's balancers.
func (s *BalancerService) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Balancer, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

func balancerFromInput(id, nodeID uuid.UUID, in BalancerInput, defEnabled bool) *domain.Balancer {
	return &domain.Balancer{
		ID:            id,
		NodeID:        nodeID,
		Tag:           in.Tag,
		Selectors:     in.Selectors,
		Strategy:      in.Strategy,
		Observe:       in.Observe,
		ProbeURL:      in.ProbeURL,
		ProbeInterval: in.ProbeInterval,
		Enabled:       boolOr(in.Enabled, defEnabled),
	}
}
