package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// BalancerRepo implements port.BalancerRepository.
type BalancerRepo struct{ q *db.Queries }

var _ port.BalancerRepository = (*BalancerRepo)(nil)

func (r *BalancerRepo) Create(ctx context.Context, b *domain.Balancer) error {
	return r.q.CreateBalancer(ctx, db.CreateBalancerParams{
		ID:            b.ID,
		NodeID:        b.NodeID,
		Tag:           b.Tag,
		Selectors:     jsonbStrings(b.Selectors),
		Strategy:      string(b.Strategy),
		Observe:       b.Observe,
		ProbeUrl:      b.ProbeURL,
		ProbeInterval: b.ProbeInterval,
		Enabled:       b.Enabled,
	})
}

func (r *BalancerRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Balancer, error) {
	row, err := r.q.GetBalancerByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	b := balancerToDomain(row)
	return &b, nil
}

func (r *BalancerRepo) Update(ctx context.Context, b *domain.Balancer) error {
	return r.q.UpdateBalancer(ctx, db.UpdateBalancerParams{
		ID:            b.ID,
		Tag:           b.Tag,
		Selectors:     jsonbStrings(b.Selectors),
		Strategy:      string(b.Strategy),
		Observe:       b.Observe,
		ProbeUrl:      b.ProbeURL,
		ProbeInterval: b.ProbeInterval,
		Enabled:       b.Enabled,
	})
}

func (r *BalancerRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteBalancer(ctx, id)
}

func (r *BalancerRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Balancer, error) {
	rows, err := r.q.ListBalancersByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Balancer, len(rows))
	for i := range rows {
		b := balancerToDomain(rows[i])
		out[i] = &b
	}
	return out, nil
}

func balancerToDomain(b db.Balancer) domain.Balancer {
	return domain.Balancer{
		ID:            b.ID,
		NodeID:        b.NodeID,
		Tag:           b.Tag,
		Selectors:     stringsFromJSONB(b.Selectors),
		Strategy:      domain.BalancerStrategy(b.Strategy),
		Observe:       b.Observe,
		ProbeURL:      b.ProbeUrl,
		ProbeInterval: b.ProbeInterval,
		Enabled:       b.Enabled,
	}
}
