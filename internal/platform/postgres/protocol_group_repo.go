package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- ProtocolGroupRepo ---

// ProtocolGroupRepo implements port.ProtocolGroupRepository.
type ProtocolGroupRepo struct{ pool *pgxpool.Pool }

var _ port.ProtocolGroupRepository = (*ProtocolGroupRepo)(nil)

func (r *ProtocolGroupRepo) Create(ctx context.Context, g *domain.ProtocolGroup) error {
	if g.ID == (uuid.UUID{}) {
		g.ID = uuid.New()
	}
	ids, _ := json.Marshal(g.InboundIDs)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO protocol_groups (id, node_id, name, inbound_ids, probe_url, probe_interval, probe_timeout, max_retries, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		g.ID, g.NodeID, g.Name, ids, g.ProbeURL, g.ProbeInterval, g.ProbeTimeout, g.MaxRetries, time.Now(), time.Now())
	return err
}

func (r *ProtocolGroupRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProtocolGroup, error) {
	var g domain.ProtocolGroup
	var ids []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, node_id, name, inbound_ids, probe_url, probe_interval, probe_timeout, max_retries, created_at, updated_at
		FROM protocol_groups WHERE id = $1`, id).
		Scan(&g.ID, &g.NodeID, &g.Name, &ids, &g.ProbeURL, &g.ProbeInterval, &g.ProbeTimeout, &g.MaxRetries, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(ids, &g.InboundIDs)
	return &g, nil
}

func (r *ProtocolGroupRepo) Update(ctx context.Context, g *domain.ProtocolGroup) error {
	ids, _ := json.Marshal(g.InboundIDs)
	_, err := r.pool.Exec(ctx, `
		UPDATE protocol_groups SET name = $2, inbound_ids = $3, probe_url = $4, probe_interval = $5,
		probe_timeout = $6, max_retries = $7, updated_at = $8 WHERE id = $1`,
		g.ID, g.Name, ids, g.ProbeURL, g.ProbeInterval, g.ProbeTimeout, g.MaxRetries, time.Now())
	return err
}

func (r *ProtocolGroupRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM protocol_groups WHERE id = $1`, id)
	return err
}

func (r *ProtocolGroupRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.ProtocolGroup, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, node_id, name, inbound_ids, probe_url, probe_interval, probe_timeout, max_retries, created_at, updated_at
		FROM protocol_groups WHERE node_id = $1 ORDER BY name`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ProtocolGroup
	for rows.Next() {
		var g domain.ProtocolGroup
		var ids []byte
		if err := rows.Scan(&g.ID, &g.NodeID, &g.Name, &ids, &g.ProbeURL, &g.ProbeInterval, &g.ProbeTimeout, &g.MaxRetries, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(ids, &g.InboundIDs)
		out = append(out, &g)
	}
	return out, rows.Err()
}

func (r *ProtocolGroupRepo) GroupsForInbounds(ctx context.Context, inboundIDs []uuid.UUID) ([]*domain.ProtocolGroup, error) {
	if len(inboundIDs) == 0 {
		return nil, nil
	}
	// Use jsonb_array_elements_text to expand the JSONB array and match
	// against the provided inbound IDs. DISTINCT ensures no duplicates.
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT DISTINCT g.id, g.node_id, g.name, g.inbound_ids, g.probe_url, g.probe_interval, g.probe_timeout, g.max_retries, g.created_at, g.updated_at
		FROM protocol_groups g, jsonb_array_elements_text(g.inbound_ids) AS elem
		WHERE elem::uuid = ANY($1)`), inboundIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ProtocolGroup
	seen := map[uuid.UUID]bool{}
	for rows.Next() {
		var g domain.ProtocolGroup
		var idsBytes []byte
		if err := rows.Scan(&g.ID, &g.NodeID, &g.Name, &idsBytes, &g.ProbeURL, &g.ProbeInterval, &g.ProbeTimeout, &g.MaxRetries, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		if seen[g.ID] {
			continue
		}
		seen[g.ID] = true
		_ = json.Unmarshal(idsBytes, &g.InboundIDs)
		out = append(out, &g)
	}
	return out, rows.Err()
}

// --- ISPProfileRepo ---

// ISPProfileRepo implements port.ISPProfileRepository.
type ISPProfileRepo struct{ pool *pgxpool.Pool }

var _ port.ISPProfileRepository = (*ISPProfileRepo)(nil)

func (r *ISPProfileRepo) Create(ctx context.Context, p *domain.ISPProfile) error {
	if p.ID == (uuid.UUID{}) {
		p.ID = uuid.New()
	}
	prefs, _ := json.Marshal(p.PreferredProtocols)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO isp_profiles (id, group_id, isp_identifier, country_code, preferred_protocols, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID, p.GroupID, p.ISPIdentifier, p.CountryCode, prefs, time.Now())
	return err
}

func (r *ISPProfileRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.ISPProfile, error) {
	var p domain.ISPProfile
	var prefs []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, group_id, isp_identifier, country_code, preferred_protocols, created_at
		FROM isp_profiles WHERE id = $1`, id).
		Scan(&p.ID, &p.GroupID, &p.ISPIdentifier, &p.CountryCode, &prefs, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(prefs, &p.PreferredProtocols)
	return &p, nil
}

func (r *ISPProfileRepo) Update(ctx context.Context, p *domain.ISPProfile) error {
	prefs, _ := json.Marshal(p.PreferredProtocols)
	_, err := r.pool.Exec(ctx, `
		UPDATE isp_profiles SET isp_identifier = $2, country_code = $3, preferred_protocols = $4 WHERE id = $1`,
		p.ID, p.ISPIdentifier, p.CountryCode, prefs)
	return err
}

func (r *ISPProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM isp_profiles WHERE id = $1`, id)
	return err
}

func (r *ISPProfileRepo) ListByGroup(ctx context.Context, groupID uuid.UUID) ([]*domain.ISPProfile, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, group_id, isp_identifier, country_code, preferred_protocols, created_at
		FROM isp_profiles WHERE group_id = $1 ORDER BY isp_identifier`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ISPProfile
	for rows.Next() {
		var p domain.ISPProfile
		var prefs []byte
		if err := rows.Scan(&p.ID, &p.GroupID, &p.ISPIdentifier, &p.CountryCode, &prefs, &p.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(prefs, &p.PreferredProtocols)
		out = append(out, &p)
	}
	return out, rows.Err()
}

func (r *ISPProfileRepo) MatchForGroups(ctx context.Context, isp string, groupIDs []uuid.UUID) ([]*domain.ISPProfile, error) {
	if len(groupIDs) == 0 || isp == "" {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, group_id, isp_identifier, country_code, preferred_protocols, created_at
		FROM isp_profiles WHERE isp_identifier = $1 AND group_id = ANY($2)`, isp, groupIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.ISPProfile
	for rows.Next() {
		var p domain.ISPProfile
		var prefs []byte
		if err := rows.Scan(&p.ID, &p.GroupID, &p.ISPIdentifier, &p.CountryCode, &prefs, &p.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(prefs, &p.PreferredProtocols)
		out = append(out, &p)
	}
	return out, rows.Err()
}

// --- SwitchEventRepo ---

// SwitchEventRepo implements port.SwitchEventRepository.
type SwitchEventRepo struct{ pool *pgxpool.Pool }

var _ port.SwitchEventRepository = (*SwitchEventRepo)(nil)

func (r *SwitchEventRepo) Record(ctx context.Context, e *domain.SwitchEvent) error {
	if e.ID == (uuid.UUID{}) {
		e.ID = uuid.New()
	}
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	// group_id is nullable — pass nil when it's a zero UUID.
	var groupID *uuid.UUID
	if e.GroupID != (uuid.UUID{}) {
		groupID = &e.GroupID
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO switch_events (id, user_id, node_id, group_id, source_protocol, target_protocol, isp, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ID, e.UserID, e.NodeID, groupID, e.SourceProtocol, e.TargetProtocol, e.ISP, e.Timestamp)
	return err
}

func (r *SwitchEventRepo) Summary(ctx context.Context, filter domain.SwitchEventFilter) (*domain.SwitchSummary, error) {
	// Count total switches in the time window.
	var total int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM switch_events
		WHERE ($1::uuid IS NULL OR node_id = $1)
		AND ($2::uuid IS NULL OR user_id = $2)
		AND ($3::text = '' OR isp = $3)
		AND timestamp >= $4 AND timestamp <= $5`,
		filter.NodeID, filter.UserID, filter.ISP, filter.FromTime, filter.ToTime).Scan(&total)
	if err != nil {
		return nil, err
	}

	summary := &domain.SwitchSummary{
		TotalSwitches: total,
		ByProtocol:    make(map[string]int),
		ByNode:        make(map[string]int),
		ByISP:         make(map[string]int),
	}

	// Aggregate by target protocol.
	rows, err := r.pool.Query(ctx, `
		SELECT target_protocol, COUNT(*) FROM switch_events
		WHERE ($1::uuid IS NULL OR node_id = $1)
		AND ($2::uuid IS NULL OR user_id = $2)
		AND ($3::text = '' OR isp = $3)
		AND timestamp >= $4 AND timestamp <= $5
		GROUP BY target_protocol ORDER BY COUNT(*) DESC LIMIT 20`,
		filter.NodeID, filter.UserID, filter.ISP, filter.FromTime, filter.ToTime)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var proto string
			var count int
			if rows.Scan(&proto, &count) == nil {
				summary.ByProtocol[proto] = count
			}
		}
	}

	// Aggregate by node.
	rows2, err := r.pool.Query(ctx, `
		SELECT node_id::text, COUNT(*) FROM switch_events
		WHERE ($1::uuid IS NULL OR node_id = $1)
		AND ($2::uuid IS NULL OR user_id = $2)
		AND ($3::text = '' OR isp = $3)
		AND timestamp >= $4 AND timestamp <= $5
		GROUP BY node_id ORDER BY COUNT(*) DESC LIMIT 20`,
		filter.NodeID, filter.UserID, filter.ISP, filter.FromTime, filter.ToTime)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var nodeID string
			var count int
			if rows2.Scan(&nodeID, &count) == nil {
				summary.ByNode[nodeID] = count
			}
		}
	}

	// Aggregate by ISP.
	rows3, err := r.pool.Query(ctx, `
		SELECT isp, COUNT(*) FROM switch_events
		WHERE ($1::uuid IS NULL OR node_id = $1)
		AND ($2::uuid IS NULL OR user_id = $2)
		AND ($3::text = '' OR isp = $3)
		AND timestamp >= $4 AND timestamp <= $5
		AND isp <> ''
		GROUP BY isp ORDER BY COUNT(*) DESC LIMIT 20`,
		filter.NodeID, filter.UserID, filter.ISP, filter.FromTime, filter.ToTime)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var isp string
			var count int
			if rows3.Scan(&isp, &count) == nil {
				summary.ByISP[isp] = count
			}
		}
	}

	return summary, nil
}
