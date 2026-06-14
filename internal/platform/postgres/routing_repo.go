package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// RoutingRepo implements port.RoutingRepository.
type RoutingRepo struct{ q *db.Queries }

var _ port.RoutingRepository = (*RoutingRepo)(nil)

func (r *RoutingRepo) Create(ctx context.Context, rule *domain.RoutingRule) error {
	return r.q.CreateRoutingRule(ctx, db.CreateRoutingRuleParams{
		ID:          rule.ID,
		NodeID:      rule.NodeID,
		Priority:    int32(rule.Priority),
		Name:        rule.Name,
		InboundTags: jsonbStrings(rule.InboundTags),
		Domains:     jsonbStrings(rule.Domains),
		Ip:          jsonbStrings(rule.IP),
		Port:        rule.Port,
		Protocols:   jsonbStrings(rule.Protocols),
		Network:     rule.Network,
		OutboundTag: rule.OutboundTag,
		BalancerTag: rule.BalancerTag,
		Enabled:     rule.Enabled,
	})
}

func (r *RoutingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoutingRule, error) {
	row, err := r.q.GetRoutingRuleByID(ctx, id)
	if err != nil {
		return nil, mapErr(err)
	}
	rule := routingToDomain(row)
	return &rule, nil
}

func (r *RoutingRepo) Update(ctx context.Context, rule *domain.RoutingRule) error {
	return r.q.UpdateRoutingRule(ctx, db.UpdateRoutingRuleParams{
		ID:          rule.ID,
		Priority:    int32(rule.Priority),
		Name:        rule.Name,
		InboundTags: jsonbStrings(rule.InboundTags),
		Domains:     jsonbStrings(rule.Domains),
		Ip:          jsonbStrings(rule.IP),
		Port:        rule.Port,
		Protocols:   jsonbStrings(rule.Protocols),
		Network:     rule.Network,
		OutboundTag: rule.OutboundTag,
		BalancerTag: rule.BalancerTag,
		Enabled:     rule.Enabled,
	})
}

func (r *RoutingRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteRoutingRule(ctx, id)
}

func (r *RoutingRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RoutingRule, error) {
	rows, err := r.q.ListRoutingRulesByNode(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.RoutingRule, len(rows))
	for i := range rows {
		rule := routingToDomain(rows[i])
		out[i] = &rule
	}
	return out, nil
}

func routingToDomain(r db.RoutingRule) domain.RoutingRule {
	return domain.RoutingRule{
		ID:          r.ID,
		NodeID:      r.NodeID,
		Priority:    int(r.Priority),
		Name:        r.Name,
		InboundTags: stringsFromJSONB(r.InboundTags),
		Domains:     stringsFromJSONB(r.Domains),
		IP:          stringsFromJSONB(r.Ip),
		Port:        r.Port,
		Protocols:   stringsFromJSONB(r.Protocols),
		Network:     r.Network,
		OutboundTag: r.OutboundTag,
		BalancerTag: r.BalancerTag,
		Enabled:     r.Enabled,
	}
}
