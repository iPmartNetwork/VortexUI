package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// NodeRepo implements port.NodeRepository.
type NodeRepo struct{ q *db.Queries }

var _ port.NodeRepository = (*NodeRepo)(nil)

func (r *NodeRepo) Create(ctx context.Context, n *domain.Node) error {
	return r.q.CreateNode(ctx, db.CreateNodeParams{
		ID:         n.ID,
		Name:       n.Name,
		Address:    n.Address,
		Core:       string(n.Core),
		Status:     string(n.Status),
		UsageRatio: n.UsageRatio,
		Endpoint:   n.Endpoint,
		CreatedAt:  timeToTS(n.CreatedAt),
	})
}

func (r *NodeRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Node, error) {
	row, err := r.q.GetNodeByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	return nodeToDomain(row), nil
}

func (r *NodeRepo) Update(ctx context.Context, n *domain.Node) error {
	return r.q.UpdateNode(ctx, db.UpdateNodeParams{
		ID:         n.ID,
		Name:       n.Name,
		Address:    n.Address,
		Core:       string(n.Core),
		Status:     string(n.Status),
		UsageRatio: n.UsageRatio,
		Endpoint:   n.Endpoint,
	})
}

func (r *NodeRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteNode(ctx, id)
}

func (r *NodeRepo) List(ctx context.Context) ([]*domain.Node, error) {
	rows, err := r.q.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Node, len(rows))
	for i := range rows {
		out[i] = nodeToDomain(rows[i])
	}
	return out, nil
}

func (r *NodeRepo) UpdateHealth(ctx context.Context, id uuid.UUID, h domain.NodeHealth) error {
	return r.q.UpdateNodeHealth(ctx, db.UpdateNodeHealthParams{
		ID:           id,
		CpuPercent:   h.CPUPercent,
		MemPercent:   h.MemPercent,
		DiskPercent:  h.DiskPercent,
		CoreRunning:  h.CoreRunning,
		Connections:  int32(h.Connections),
		CoreVersion:  h.CoreVersion,
		AgentVersion: h.AgentVersion,
	})
}

func nodeToDomain(n db.Node) *domain.Node {
	return &domain.Node{
		ID:         n.ID,
		Name:       n.Name,
		Address:    n.Address,
		Core:       domain.CoreType(n.Core),
		Status:     domain.NodeStatus(n.Status),
		UsageRatio: n.UsageRatio,
		Endpoint:   n.Endpoint,
		LastSeen:   tsToPtr(n.LastSeen),
		Health: domain.NodeHealth{
			CPUPercent:  n.CpuPercent,
			MemPercent:  n.MemPercent,
			DiskPercent: n.DiskPercent,
			CoreRunning: n.CoreRunning,
			Connections: int(n.Connections),
		},
		CoreVer:   n.CoreVersion,
		AgentVer:  n.AgentVersion,
		CreatedAt: n.CreatedAt.Time,
	}
}
