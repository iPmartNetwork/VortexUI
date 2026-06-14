package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// OutboundRepo implements port.OutboundRepository.
type OutboundRepo struct{ q *db.Queries }

var _ port.OutboundRepository = (*OutboundRepo)(nil)

func (r *OutboundRepo) Create(ctx context.Context, o *domain.Outbound) error {
	return r.q.CreateOutbound(ctx, db.CreateOutboundParams{
		ID:       o.ID,
		NodeID:   o.NodeID,
		Tag:      o.Tag,
		Protocol: string(o.Protocol),
		Address:  o.Address,
		Port:     int32(o.Port),
		Uuid:     o.UUID,
		Password: o.Password,
		Username: o.Username,
		Method:   o.Method,
		Flow:     o.Flow,
		Network:  o.Network,
		Security: string(o.Security),
		Sni:      o.SNI,
		Path:     o.Path,
		Host:     o.Host,
		Raw:      jsonbMap(o.Raw),
		Enabled:  o.Enabled,
	})
}

func (r *OutboundRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Outbound, error) {
	row, err := r.q.GetOutboundByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	o := outboundToDomain(row)
	return &o, nil
}

func (r *OutboundRepo) Update(ctx context.Context, o *domain.Outbound) error {
	return r.q.UpdateOutbound(ctx, db.UpdateOutboundParams{
		ID:       o.ID,
		Tag:      o.Tag,
		Protocol: string(o.Protocol),
		Address:  o.Address,
		Port:     int32(o.Port),
		Uuid:     o.UUID,
		Password: o.Password,
		Username: o.Username,
		Method:   o.Method,
		Flow:     o.Flow,
		Network:  o.Network,
		Security: string(o.Security),
		Sni:      o.SNI,
		Path:     o.Path,
		Host:     o.Host,
		Raw:      jsonbMap(o.Raw),
		Enabled:  o.Enabled,
	})
}

func (r *OutboundRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteOutbound(ctx, id)
}

func (r *OutboundRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Outbound, error) {
	rows, err := r.q.ListOutboundsByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Outbound, len(rows))
	for i := range rows {
		o := outboundToDomain(rows[i])
		out[i] = &o
	}
	return out, nil
}

func outboundToDomain(o db.Outbound) domain.Outbound {
	return domain.Outbound{
		ID:       o.ID,
		NodeID:   o.NodeID,
		Tag:      o.Tag,
		Protocol: domain.OutboundProtocol(o.Protocol),
		Address:  o.Address,
		Port:     int(o.Port),
		UUID:     o.Uuid,
		Password: o.Password,
		Username: o.Username,
		Method:   o.Method,
		Flow:     o.Flow,
		Network:  o.Network,
		Security: domain.Security(o.Security),
		SNI:      o.Sni,
		Path:     o.Path,
		Host:     o.Host,
		Raw:      mapFromJSONB(o.Raw),
		Enabled:  o.Enabled,
	}
}
