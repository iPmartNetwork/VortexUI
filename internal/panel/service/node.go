package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// NodeRegistrar is the slice of the hub the node service drives: bringing nodes
// under (or out of) live management. *hub.Hub satisfies it.
type NodeRegistrar interface {
	Register(ctx context.Context, node *domain.Node) error
	Deregister(nodeID uuid.UUID)
}

// NodeLogQuerier fetches recent core log lines from a node. *hub.Hub satisfies it.
type NodeLogQuerier interface {
	Logs(ctx context.Context, nodeID uuid.UUID, limit int) ([]string, error)
}

// NodeService manages the node fleet's persistent records and keeps the hub's
// live management in step with them.
type NodeService struct {
	repo      port.NodeRepository
	registrar NodeRegistrar
	logs      NodeLogQuerier
	now       func() time.Time
}

// NewNodeService wires the service.
func NewNodeService(repo port.NodeRepository, registrar NodeRegistrar) *NodeService {
	return &NodeService{repo: repo, registrar: registrar, now: time.Now}
}

// SetLogQuerier wires the source of live node logs (the hub).
func (s *NodeService) SetLogQuerier(q NodeLogQuerier) {
	if q != nil {
		s.logs = q
	}
}

// Logs fetches up to limit recent core log lines from a node. Errors when no log
// source is wired or the node is unreachable.
func (s *NodeService) Logs(ctx context.Context, id uuid.UUID, limit int) ([]string, error) {
	if s.logs == nil {
		return nil, errors.New("node logs source not configured")
	}
	return s.logs.Logs(ctx, id, limit)
}

// CreateNodeInput describes a new node.
type CreateNodeInput struct {
	Name       string
	Address    string
	Core       domain.CoreType
	UsageRatio float64
}

// Create persists a node and immediately brings it under hub management so its
// traffic/health loops start without a panel restart.
func (s *NodeService) Create(ctx context.Context, in CreateNodeInput) (*domain.Node, error) {
	if in.Name == "" || in.Address == "" {
		return nil, errors.New("name and address are required")
	}
	core := in.Core
	if core == "" {
		core = domain.CoreXray
	}
	ratio := in.UsageRatio
	if ratio == 0 {
		ratio = 1
	}
	n := &domain.Node{
		ID:         uuid.New(),
		Name:       in.Name,
		Address:    in.Address,
		Core:       core,
		Status:     domain.NodeDisconnected,
		UsageRatio: ratio,
		CreatedAt:  s.now(),
	}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	if err := s.registrar.Register(ctx, n); err != nil {
		// Persisted but not yet managed; a panel restart (or re-register) recovers.
		return n, err
	}
	return n, nil
}

// List returns all nodes.
func (s *NodeService) List(ctx context.Context) ([]*domain.Node, error) {
	return s.repo.List(ctx)
}

// Get returns one node.
func (s *NodeService) Get(ctx context.Context, id uuid.UUID) (*domain.Node, error) {
	return s.repo.GetByID(ctx, id)
}

// UpdateNodeInput is the mutable subset of a node.
type UpdateNodeInput struct {
	Name       string
	Address    string
	UsageRatio float64
}

// Update persists node changes and re-establishes hub management so an address
// or ratio change takes effect immediately (the old connection is dropped and a
// fresh one dialed against the new address).
func (s *NodeService) Update(ctx context.Context, id uuid.UUID, in UpdateNodeInput) (*domain.Node, error) {
	n, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != "" {
		n.Name = in.Name
	}
	if in.Address != "" {
		n.Address = in.Address
	}
	if in.UsageRatio > 0 {
		n.UsageRatio = in.UsageRatio
	}
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	// Re-register to pick up a possibly-changed address.
	s.registrar.Deregister(n.ID)
	if err := s.registrar.Register(ctx, n); err != nil {
		return n, err
	}
	return n, nil
}

// Delete stops managing a node and removes it (its inbounds cascade in the DB).
func (s *NodeService) Delete(ctx context.Context, id uuid.UUID) error {
	s.registrar.Deregister(id)
	return s.repo.Delete(ctx, id)
}
