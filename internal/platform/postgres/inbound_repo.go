package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// InboundRepo implements port.InboundRepository.
type InboundRepo struct{ q *db.Queries }

var _ port.InboundRepository = (*InboundRepo)(nil)

func (r *InboundRepo) Create(ctx context.Context, in *domain.Inbound) error {
	return r.q.CreateInbound(ctx, db.CreateInboundParams{
		ID:               in.ID,
		NodeID:           in.NodeID,
		Tag:              in.Tag,
		Protocol:         string(in.Protocol),
		Listen:           in.Listen,
		Port:             int32(in.Port),
		Network:          in.Network,
		Security:         string(in.Security),
		Sni:              jsonbStrings(in.SNI),
		Path:             in.Path,
		Host:             jsonbStrings(in.Host),
		Flow:             in.Flow,
		EvasionProfileID: ptrToUUID(in.EvasionProfileID),
		Raw:              jsonbMap(in.Raw),
		Enabled:          in.Enabled,
	})
}

func (r *InboundRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Inbound, error) {
	row, err := r.q.GetInboundByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	in := inboundToDomain(row)
	return &in, nil
}

func (r *InboundRepo) Update(ctx context.Context, in *domain.Inbound) error {
	return r.q.UpdateInbound(ctx, db.UpdateInboundParams{
		ID:               in.ID,
		Tag:              in.Tag,
		Protocol:         string(in.Protocol),
		Listen:           in.Listen,
		Port:             int32(in.Port),
		Network:          in.Network,
		Security:         string(in.Security),
		Sni:              jsonbStrings(in.SNI),
		Path:             in.Path,
		Host:             jsonbStrings(in.Host),
		Flow:             in.Flow,
		EvasionProfileID: ptrToUUID(in.EvasionProfileID),
		Raw:              jsonbMap(in.Raw),
		Enabled:          in.Enabled,
	})
}

func (r *InboundRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteInbound(ctx, id)
}

func (r *InboundRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error) {
	rows, err := r.q.ListInboundsByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Inbound, len(rows))
	for i := range rows {
		in := inboundToDomain(rows[i])
		out[i] = &in
	}
	return out, nil
}

func inboundToDomain(in db.Inbound) domain.Inbound {
	return domain.Inbound{
		ID:               in.ID,
		NodeID:           in.NodeID,
		Tag:              in.Tag,
		Protocol:         domain.Protocol(in.Protocol),
		Listen:           in.Listen,
		Port:             int(in.Port),
		Network:          in.Network,
		Security:         domain.Security(in.Security),
		SNI:              stringsFromJSONB(in.Sni),
		Path:             in.Path,
		Host:             stringsFromJSONB(in.Host),
		Flow:             in.Flow,
		EvasionProfileID: uuidToPtr(in.EvasionProfileID),
		Raw:              mapFromJSONB(in.Raw),
		Enabled:          in.Enabled,
	}
}
