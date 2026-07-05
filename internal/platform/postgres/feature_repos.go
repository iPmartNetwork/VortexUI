package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- AnalyticsRepo ---

type AnalyticsRepo struct{ pool *pgxpool.Pool }

var _ port.AnalyticsRepository = (*AnalyticsRepo)(nil)

func (r *AnalyticsRepo) GeoBreakdown(ctx context.Context, q port.SeriesQuery) ([]domain.GeoTrafficPoint, error) {
	from := time.Unix(q.FromUnix, 0)
	to := time.Unix(q.ToUnix, 0)
	rows, err := r.pool.Query(ctx,
		`SELECT ug.country,
		        COUNT(DISTINCT tp.user_id)::bigint AS connections,
		        COALESCE(SUM(tp.up), 0)::bigint     AS bytes_up,
		        COALESCE(SUM(tp.down), 0)::bigint   AS bytes_down
		 FROM traffic_points tp
		 JOIN user_geo ug ON ug.user_id = tp.user_id
		 JOIN users u ON u.id = tp.user_id
		 WHERE tp.time >= $1 AND tp.time <= $2 AND ug.country <> ''
		   AND ($3::uuid IS NULL OR u.admin_id = $3)
		 GROUP BY ug.country
		 ORDER BY (SUM(tp.up) + SUM(tp.down)) DESC`, from, to, q.AdminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []domain.GeoTrafficPoint
	for rows.Next() {
		var p domain.GeoTrafficPoint
		if err := rows.Scan(&p.Country, &p.Connections, &p.BytesUp, &p.BytesDown); err != nil {
			return nil, err
		}
		p.Time = to
		results = append(results, p)
	}
	return results, rows.Err()
}

func (r *AnalyticsRepo) TopUsers(ctx context.Context, limit int, adminID *uuid.UUID) ([]domain.UserTrafficRank, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT u.id, u.username, u.used_traffic
		 FROM users u
		 WHERE u.used_traffic > 0
		   AND ($2::uuid IS NULL OR u.admin_id = $2)
		 ORDER BY u.used_traffic DESC
		 LIMIT $1`, limit, adminID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	var results []domain.UserTrafficRank
	for rows.Next() {
		var r domain.UserTrafficRank
		if err := rows.Scan(&r.UserID, &r.Username, &r.UsedTraffic); err != nil {
			return nil, nil
		}
		results = append(results, r)
	}
	return results, nil
}

func (r *AnalyticsRepo) PeakHours(ctx context.Context, q port.SeriesQuery) ([]domain.PeakHour, error) {
	from := time.Unix(q.FromUnix, 0)
	to := time.Unix(q.ToUnix, 0)
	rows, err := r.pool.Query(ctx,
		`SELECT EXTRACT(HOUR FROM tp.time)::int AS hour,
		        COUNT(*)::bigint AS connections,
		        COALESCE(SUM(tp.up + tp.down), 0)::bigint AS bytes_total
		 FROM traffic_points tp
		 JOIN users u ON u.id = tp.user_id
		 WHERE tp.time >= $1 AND tp.time <= $2
		   AND ($3::uuid IS NULL OR u.admin_id = $3)
		 GROUP BY hour
		 ORDER BY hour`, from, to, q.AdminID)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()
	var results []domain.PeakHour
	for rows.Next() {
		var p domain.PeakHour
		if err := rows.Scan(&p.Hour, &p.Connections, &p.BytesTotal); err != nil {
			return nil, nil
		}
		results = append(results, p)
	}
	return results, nil
}

func (r *AnalyticsRepo) TotalTraffic(ctx context.Context, q port.SeriesQuery) (int64, int64, error) {
	from := time.Unix(q.FromUnix, 0)
	to := time.Unix(q.ToUnix, 0)
	var up, down int64

	if q.AdminID == nil {
		// Try traffic_geo first (aggregated geo data).
		err := r.pool.QueryRow(ctx,
			`SELECT COALESCE(SUM(bytes_up), 0), COALESCE(SUM(bytes_down), 0)
			 FROM traffic_geo
			 WHERE time >= $1 AND time <= $2`, from, to).Scan(&up, &down)
		if err == nil && (up > 0 || down > 0) {
			return up, down, nil
		}
	}

	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(tp.up), 0), COALESCE(SUM(tp.down), 0)
		 FROM traffic_points tp
		 JOIN users u ON u.id = tp.user_id
		 WHERE tp.time >= $1 AND tp.time <= $2
		   AND ($3::uuid IS NULL OR u.admin_id = $3)`, from, to, q.AdminID).Scan(&up, &down)
	if err != nil {
		return 0, 0, nil
	}
	return up, down, nil
}


// --- UserGeoRepo ---

type UserGeoRepo struct{ pool *pgxpool.Pool }

func (r *UserGeoRepo) Upsert(ctx context.Context, userID uuid.UUID, country, ip string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_geo (user_id, country, ip, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id) DO UPDATE SET country = EXCLUDED.country, ip = EXCLUDED.ip, updated_at = now()`,
		userID, country, ip)
	return err
}


// --- QuotaPolicyRepo ---

type QuotaPolicyRepo struct{ pool *pgxpool.Pool }

var _ port.QuotaPolicyRepository = (*QuotaPolicyRepo)(nil)

func (r *QuotaPolicyRepo) Create(ctx context.Context, p *domain.QuotaPolicy) error {
	tiersJSON, err := json.Marshal(p.Tiers)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO quota_policies (id, name, tiers, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		p.ID, p.Name, tiersJSON, p.Enabled, p.CreatedAt)
	return err
}

func (r *QuotaPolicyRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.QuotaPolicy, error) {
	var p domain.QuotaPolicy
	var tiersJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, tiers, enabled, created_at
		 FROM quota_policies WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &tiersJSON, &p.Enabled, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tiersJSON, &p.Tiers); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *QuotaPolicyRepo) GetDefault(ctx context.Context) (*domain.QuotaPolicy, error) {
	var p domain.QuotaPolicy
	var tiersJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, tiers, enabled, created_at
		 FROM quota_policies WHERE enabled = TRUE
		 ORDER BY created_at ASC LIMIT 1`).
		Scan(&p.ID, &p.Name, &tiersJSON, &p.Enabled, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(tiersJSON, &p.Tiers); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *QuotaPolicyRepo) Update(ctx context.Context, p *domain.QuotaPolicy) error {
	tiersJSON, err := json.Marshal(p.Tiers)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE quota_policies SET name = $2, tiers = $3, enabled = $4
		 WHERE id = $1`,
		p.ID, p.Name, tiersJSON, p.Enabled)
	return err
}

func (r *QuotaPolicyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM quota_policies WHERE id = $1`, id)
	return err
}

func (r *QuotaPolicyRepo) List(ctx context.Context) ([]*domain.QuotaPolicy, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, tiers, enabled, created_at
		 FROM quota_policies ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.QuotaPolicy
	for rows.Next() {
		var p domain.QuotaPolicy
		var tiersJSON []byte
		if err := rows.Scan(&p.ID, &p.Name, &tiersJSON, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(tiersJSON, &p.Tiers); err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, rows.Err()
}


// --- RelayChainRepo ---

type RelayChainRepo struct{ pool *pgxpool.Pool }

var _ port.RelayChainRepository = (*RelayChainRepo)(nil)

func (r *RelayChainRepo) Create(ctx context.Context, rc *domain.RelayChain) error {
	hopsJSON, err := json.Marshal(rc.Hops)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO relay_chains (id, name, node_id, hops, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		rc.ID, rc.Name, rc.NodeID, hopsJSON, rc.Enabled, rc.CreatedAt)
	return err
}

func (r *RelayChainRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RelayChain, error) {
	var rc domain.RelayChain
	var hopsJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, node_id, hops, enabled, created_at
		 FROM relay_chains WHERE id = $1`, id).
		Scan(&rc.ID, &rc.Name, &rc.NodeID, &hopsJSON, &rc.Enabled, &rc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(hopsJSON, &rc.Hops); err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *RelayChainRepo) Update(ctx context.Context, rc *domain.RelayChain) error {
	hopsJSON, err := json.Marshal(rc.Hops)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE relay_chains SET name = $2, node_id = $3, hops = $4, enabled = $5
		 WHERE id = $1`,
		rc.ID, rc.Name, rc.NodeID, hopsJSON, rc.Enabled)
	return err
}

func (r *RelayChainRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM relay_chains WHERE id = $1`, id)
	return err
}

func (r *RelayChainRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.RelayChain, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, node_id, hops, enabled, created_at
		 FROM relay_chains WHERE node_id = $1 ORDER BY created_at`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.RelayChain
	for rows.Next() {
		var rc domain.RelayChain
		var hopsJSON []byte
		if err := rows.Scan(&rc.ID, &rc.Name, &rc.NodeID, &hopsJSON, &rc.Enabled, &rc.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(hopsJSON, &rc.Hops); err != nil {
			return nil, err
		}
		results = append(results, &rc)
	}
	return results, rows.Err()
}

func (r *RelayChainRepo) List(ctx context.Context) ([]*domain.RelayChain, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, node_id, hops, enabled, created_at
		 FROM relay_chains ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.RelayChain
	for rows.Next() {
		var rc domain.RelayChain
		var hopsJSON []byte
		if err := rows.Scan(&rc.ID, &rc.Name, &rc.NodeID, &hopsJSON, &rc.Enabled, &rc.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(hopsJSON, &rc.Hops); err != nil {
			return nil, err
		}
		results = append(results, &rc)
	}
	return results, rows.Err()
}


// --- DecoySiteRepo ---

type DecoySiteRepo struct{ pool *pgxpool.Pool }

var _ port.DecoySiteRepository = (*DecoySiteRepo)(nil)

func (r *DecoySiteRepo) Create(ctx context.Context, d *domain.DecoySite) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO decoy_sites (id, node_id, mode, target_url, static_html, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		d.ID, d.NodeID, d.Mode, d.TargetURL, d.StaticHTML, d.Enabled, d.CreatedAt)
	return err
}

func (r *DecoySiteRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.DecoySite, error) {
	var d domain.DecoySite
	err := r.pool.QueryRow(ctx,
		`SELECT id, node_id, mode, target_url, static_html, enabled, created_at
		 FROM decoy_sites WHERE id = $1`, id).
		Scan(&d.ID, &d.NodeID, &d.Mode, &d.TargetURL, &d.StaticHTML, &d.Enabled, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DecoySiteRepo) GetGlobal(ctx context.Context) (*domain.DecoySite, error) {
	var d domain.DecoySite
	err := r.pool.QueryRow(ctx,
		`SELECT id, node_id, mode, target_url, static_html, enabled, created_at
		 FROM decoy_sites WHERE node_id IS NULL LIMIT 1`).
		Scan(&d.ID, &d.NodeID, &d.Mode, &d.TargetURL, &d.StaticHTML, &d.Enabled, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DecoySiteRepo) GetByNode(ctx context.Context, nodeID uuid.UUID) (*domain.DecoySite, error) {
	var d domain.DecoySite
	err := r.pool.QueryRow(ctx,
		`SELECT id, node_id, mode, target_url, static_html, enabled, created_at
		 FROM decoy_sites WHERE node_id = $1 LIMIT 1`, nodeID).
		Scan(&d.ID, &d.NodeID, &d.Mode, &d.TargetURL, &d.StaticHTML, &d.Enabled, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *DecoySiteRepo) Update(ctx context.Context, d *domain.DecoySite) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE decoy_sites SET node_id = $2, mode = $3, target_url = $4, static_html = $5, enabled = $6
		 WHERE id = $1`,
		d.ID, d.NodeID, d.Mode, d.TargetURL, d.StaticHTML, d.Enabled)
	return err
}

func (r *DecoySiteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM decoy_sites WHERE id = $1`, id)
	return err
}

func (r *DecoySiteRepo) List(ctx context.Context) ([]*domain.DecoySite, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, node_id, mode, target_url, static_html, enabled, created_at
		 FROM decoy_sites ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.DecoySite
	for rows.Next() {
		var d domain.DecoySite
		if err := rows.Scan(&d.ID, &d.NodeID, &d.Mode, &d.TargetURL, &d.StaticHTML, &d.Enabled, &d.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}


// --- RealityScanRepo ---

type RealityScanRepo struct{ pool *pgxpool.Pool }

var _ port.RealityScanRepository = (*RealityScanRepo)(nil)

func (r *RealityScanRepo) SaveBatch(ctx context.Context, results []domain.RealityScan) error {
	if len(results) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for i := range results {
		s := &results[i]
		batch.Queue(
			`INSERT INTO reality_scans (id, node_id, sni, latency_ms, score, valid, scanned_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			s.ID, s.NodeID, s.SNI, s.LatencyMS, s.Score, s.Valid, s.ScannedAt)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close() //nolint:errcheck // batch result cleanup
	for range results {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RealityScanRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]domain.RealityScan, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, node_id, sni, latency_ms, score, valid, scanned_at
		 FROM reality_scans WHERE node_id = $1
		 ORDER BY score DESC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.RealityScan
	for rows.Next() {
		var s domain.RealityScan
		if err := rows.Scan(&s.ID, &s.NodeID, &s.SNI, &s.LatencyMS, &s.Score, &s.Valid, &s.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

func (r *RealityScanRepo) DeleteByNode(ctx context.Context, nodeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM reality_scans WHERE node_id = $1`, nodeID)
	return err
}


// --- ProbingRepo ---

type ProbingRepo struct{ pool *pgxpool.Pool }

var _ port.ProbingRepository = (*ProbingRepo)(nil)

func (r *ProbingRepo) GetPolicy(ctx context.Context) (*domain.ProbingPolicy, error) {
	var p domain.ProbingPolicy
	var whitelistJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, action, block_duration, max_probe_per_min, whitelisted_ips, honeypot_html, notify_telegram
		 FROM probing_policy WHERE id = 1`).
		Scan(&p.Enabled, &p.Action, &p.BlockDuration, &p.MaxProbePerMin, &whitelistJSON, &p.HoneypotHTML, &p.NotifyTelegram)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(whitelistJSON, &p.WhitelistedIPs); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProbingRepo) SavePolicy(ctx context.Context, p *domain.ProbingPolicy) error {
	whitelistJSON, err := json.Marshal(p.WhitelistedIPs)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO probing_policy (id, enabled, action, block_duration, max_probe_per_min, whitelisted_ips, honeypot_html, notify_telegram)
		 VALUES (1, $1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   action = EXCLUDED.action,
		   block_duration = EXCLUDED.block_duration,
		   max_probe_per_min = EXCLUDED.max_probe_per_min,
		   whitelisted_ips = EXCLUDED.whitelisted_ips,
		   honeypot_html = EXCLUDED.honeypot_html,
		   notify_telegram = EXCLUDED.notify_telegram`,
		p.Enabled, p.Action, p.BlockDuration, p.MaxProbePerMin, whitelistJSON, p.HoneypotHTML, p.NotifyTelegram)
	return err
}

func (r *ProbingRepo) SaveEvent(ctx context.Context, e *domain.ProbeEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO probe_events (id, source_ip, port, method, fingerprint, action, node_id, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ID, e.SourceIP, e.Port, e.Method, e.Fingerprint, e.Action, e.NodeID, e.CreatedAt)
	return err
}

func (r *ProbingRepo) ListEvents(ctx context.Context, limit, offset int) ([]*domain.ProbeEvent, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM probe_events`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, source_ip, port, method, fingerprint, action, node_id, created_at
		 FROM probe_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*domain.ProbeEvent
	for rows.Next() {
		var e domain.ProbeEvent
		if err := rows.Scan(&e.ID, &e.SourceIP, &e.Port, &e.Method, &e.Fingerprint, &e.Action, &e.NodeID, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, &e)
	}
	return results, total, rows.Err()
}

func (r *ProbingRepo) BlockIP(ctx context.Context, b *domain.BlockedIP) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO blocked_ips (ip, reason, blocked_at, expires_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (ip) DO UPDATE SET
		   reason = EXCLUDED.reason,
		   blocked_at = EXCLUDED.blocked_at,
		   expires_at = EXCLUDED.expires_at`,
		b.IP, b.Reason, b.BlockedAt, b.ExpiresAt)
	return err
}

func (r *ProbingRepo) UnblockIP(ctx context.Context, ip string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM blocked_ips WHERE ip = $1`, ip)
	return err
}

func (r *ProbingRepo) ListBlockedIPs(ctx context.Context) ([]domain.BlockedIP, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT ip, reason, blocked_at, expires_at FROM blocked_ips ORDER BY blocked_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.BlockedIP
	for rows.Next() {
		var b domain.BlockedIP
		if err := rows.Scan(&b.IP, &b.Reason, &b.BlockedAt, &b.ExpiresAt); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

func (r *ProbingRepo) IsBlocked(ctx context.Context, ip string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM blocked_ips WHERE ip = $1 AND expires_at > now())`, ip).
		Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}


// --- FamilyRepo ---

type FamilyRepo struct{ pool *pgxpool.Pool }

var _ port.FamilyRepository = (*FamilyRepo)(nil)

func (r *FamilyRepo) CreateGroup(ctx context.Context, g *domain.FamilyGroup) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO family_groups (id, name, owner_id, data_limit, used_traffic, max_members, member_quota, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		g.ID, g.Name, g.OwnerID, g.DataLimit, g.UsedTraffic, g.MaxMembers, g.MemberQuota, g.CreatedAt)
	return err
}

func (r *FamilyRepo) GetGroup(ctx context.Context, id uuid.UUID) (*domain.FamilyGroup, error) {
	var g domain.FamilyGroup
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, data_limit, used_traffic, max_members, member_quota, created_at
		 FROM family_groups WHERE id = $1`, id).
		Scan(&g.ID, &g.Name, &g.OwnerID, &g.DataLimit, &g.UsedTraffic, &g.MaxMembers, &g.MemberQuota, &g.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *FamilyRepo) GetGroupByOwner(ctx context.Context, ownerID uuid.UUID) (*domain.FamilyGroup, error) {
	var g domain.FamilyGroup
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, owner_id, data_limit, used_traffic, max_members, member_quota, created_at
		 FROM family_groups WHERE owner_id = $1`, ownerID).
		Scan(&g.ID, &g.Name, &g.OwnerID, &g.DataLimit, &g.UsedTraffic, &g.MaxMembers, &g.MemberQuota, &g.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *FamilyRepo) UpdateGroup(ctx context.Context, g *domain.FamilyGroup) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE family_groups SET name = $2, data_limit = $3, used_traffic = $4, max_members = $5, member_quota = $6
		 WHERE id = $1`,
		g.ID, g.Name, g.DataLimit, g.UsedTraffic, g.MaxMembers, g.MemberQuota)
	return err
}

func (r *FamilyRepo) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM family_groups WHERE id = $1`, id)
	return err
}

func (r *FamilyRepo) ListGroups(ctx context.Context, limit, offset int) ([]*domain.FamilyGroup, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM family_groups`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, name, owner_id, data_limit, used_traffic, max_members, member_quota, created_at
		 FROM family_groups ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*domain.FamilyGroup
	for rows.Next() {
		var g domain.FamilyGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.OwnerID, &g.DataLimit, &g.UsedTraffic, &g.MaxMembers, &g.MemberQuota, &g.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, &g)
	}
	return results, total, rows.Err()
}

func (r *FamilyRepo) AddMember(ctx context.Context, m *domain.FamilyMember) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO family_members (id, group_id, user_id, used_traffic, label, joined_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		m.ID, m.GroupID, m.UserID, m.UsedTraffic, m.Label, m.JoinedAt)
	return err
}

func (r *FamilyRepo) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM family_members WHERE group_id = $1 AND user_id = $2`, groupID, userID)
	return err
}

func (r *FamilyRepo) ListMembers(ctx context.Context, groupID uuid.UUID) ([]domain.FamilyMember, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, group_id, user_id, used_traffic, label, joined_at
		 FROM family_members WHERE group_id = $1 ORDER BY joined_at`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.FamilyMember
	for rows.Next() {
		var m domain.FamilyMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.UserID, &m.UsedTraffic, &m.Label, &m.JoinedAt); err != nil {
			return nil, err
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *FamilyRepo) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.FamilyMember, error) {
	var m domain.FamilyMember
	err := r.pool.QueryRow(ctx,
		`SELECT id, group_id, user_id, used_traffic, label, joined_at
		 FROM family_members WHERE group_id = $1 AND user_id = $2`, groupID, userID).
		Scan(&m.ID, &m.GroupID, &m.UserID, &m.UsedTraffic, &m.Label, &m.JoinedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *FamilyRepo) UpdateMemberTraffic(ctx context.Context, groupID, userID uuid.UUID, delta int64) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE family_members SET used_traffic = used_traffic + $3
		 WHERE group_id = $1 AND user_id = $2`, groupID, userID, delta)
	return err
}


// --- ReferralRepo ---

type ReferralRepo struct{ pool *pgxpool.Pool }

var _ port.ReferralRepository = (*ReferralRepo)(nil)

func (r *ReferralRepo) GetConfig(ctx context.Context) (*domain.ReferralConfig, error) {
	var c domain.ReferralConfig
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, reward_type, reward_amount, max_referrals, require_paid
		 FROM referral_config WHERE id = 1`).
		Scan(&c.Enabled, &c.RewardType, &c.RewardAmount, &c.MaxReferrals, &c.RequirePaid)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ReferralRepo) SaveConfig(ctx context.Context, c *domain.ReferralConfig) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO referral_config (id, enabled, reward_type, reward_amount, max_referrals, require_paid)
		 VALUES (1, $1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   reward_type = EXCLUDED.reward_type,
		   reward_amount = EXCLUDED.reward_amount,
		   max_referrals = EXCLUDED.max_referrals,
		   require_paid = EXCLUDED.require_paid`,
		c.Enabled, c.RewardType, c.RewardAmount, c.MaxReferrals, c.RequirePaid)
	return err
}

func (r *ReferralRepo) CreateCode(ctx context.Context, rc *domain.ReferralCode) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO referral_codes (id, user_id, code, uses, max_uses, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		rc.ID, rc.UserID, rc.Code, rc.Uses, rc.MaxUses, rc.CreatedAt)
	return err
}

func (r *ReferralRepo) GetCodeByCode(ctx context.Context, code string) (*domain.ReferralCode, error) {
	var rc domain.ReferralCode
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, code, uses, max_uses, created_at
		 FROM referral_codes WHERE code = $1`, code).
		Scan(&rc.ID, &rc.UserID, &rc.Code, &rc.Uses, &rc.MaxUses, &rc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReferralRepo) GetCodeByUser(ctx context.Context, userID uuid.UUID) (*domain.ReferralCode, error) {
	var rc domain.ReferralCode
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, code, uses, max_uses, created_at
		 FROM referral_codes WHERE user_id = $1`, userID).
		Scan(&rc.ID, &rc.UserID, &rc.Code, &rc.Uses, &rc.MaxUses, &rc.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

func (r *ReferralRepo) IncrementUses(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE referral_codes SET uses = uses + 1 WHERE id = $1`, id)
	return err
}

func (r *ReferralRepo) ListCodes(ctx context.Context, limit, offset int) ([]*domain.ReferralCode, int, error) {
	var total int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM referral_codes`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, code, uses, max_uses, created_at
		 FROM referral_codes ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*domain.ReferralCode
	for rows.Next() {
		var rc domain.ReferralCode
		if err := rows.Scan(&rc.ID, &rc.UserID, &rc.Code, &rc.Uses, &rc.MaxUses, &rc.CreatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, &rc)
	}
	return results, total, rows.Err()
}

func (r *ReferralRepo) SaveEvent(ctx context.Context, e *domain.ReferralEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO referral_events (id, referrer_id, referred_id, code_used, reward_type, reward_amount, reward_applied, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		e.ID, e.ReferrerID, e.ReferredID, e.CodeUsed, e.RewardType, e.RewardAmount, e.RewardApplied, e.CreatedAt)
	return err
}

func (r *ReferralRepo) ListEvents(ctx context.Context, userID *uuid.UUID, limit int) ([]*domain.ReferralEvent, error) {
	var rows pgx.Rows
	var err error
	if userID != nil {
		rows, err = r.pool.Query(ctx,
			`SELECT id, referrer_id, referred_id, code_used, reward_type, reward_amount, reward_applied, created_at
			 FROM referral_events WHERE referrer_id = $1 OR referred_id = $1
			 ORDER BY created_at DESC LIMIT $2`, *userID, limit)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, referrer_id, referred_id, code_used, reward_type, reward_amount, reward_applied, created_at
			 FROM referral_events ORDER BY created_at DESC LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.ReferralEvent
	for rows.Next() {
		var e domain.ReferralEvent
		if err := rows.Scan(&e.ID, &e.ReferrerID, &e.ReferredID, &e.CodeUsed, &e.RewardType, &e.RewardAmount, &e.RewardApplied, &e.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}


// --- DoHRepo ---

type DoHRepo struct{ pool *pgxpool.Pool }

var _ port.DoHRepository = (*DoHRepo)(nil)

func (r *DoHRepo) GetConfig(ctx context.Context) (*domain.DoHConfig, error) {
	var c domain.DoHConfig
	var upstreamJSON, blocklistJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, listen_addr, upstream_dns, block_ads, block_malware, custom_blocklist, log_queries, cache_ttl
		 FROM doh_config WHERE id = 1`).
		Scan(&c.Enabled, &c.ListenAddr, &upstreamJSON, &c.BlockAds, &c.BlockMalware, &blocklistJSON, &c.LogQueries, &c.CacheTTL)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(upstreamJSON, &c.UpstreamDNS); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(blocklistJSON, &c.CustomBlocklist); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *DoHRepo) SaveConfig(ctx context.Context, c *domain.DoHConfig) error {
	upstreamJSON, err := json.Marshal(c.UpstreamDNS)
	if err != nil {
		return err
	}
	blocklistJSON, err := json.Marshal(c.CustomBlocklist)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO doh_config (id, enabled, listen_addr, upstream_dns, block_ads, block_malware, custom_blocklist, log_queries, cache_ttl)
		 VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   listen_addr = EXCLUDED.listen_addr,
		   upstream_dns = EXCLUDED.upstream_dns,
		   block_ads = EXCLUDED.block_ads,
		   block_malware = EXCLUDED.block_malware,
		   custom_blocklist = EXCLUDED.custom_blocklist,
		   log_queries = EXCLUDED.log_queries,
		   cache_ttl = EXCLUDED.cache_ttl`,
		c.Enabled, c.ListenAddr, upstreamJSON, c.BlockAds, c.BlockMalware, blocklistJSON, c.LogQueries, c.CacheTTL)
	return err
}

func (r *DoHRepo) SaveQueryLog(ctx context.Context, log *domain.DoHQueryLog) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO doh_query_logs (domain, type, client_ip, blocked, latency_ms, timestamp)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		log.Domain, log.Type, log.ClientIP, log.Blocked, log.LatencyMS, log.Timestamp)
	return err
}

func (r *DoHRepo) ListQueryLogs(ctx context.Context, limit int) ([]domain.DoHQueryLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT domain, type, client_ip, blocked, latency_ms, timestamp
		 FROM doh_query_logs ORDER BY timestamp DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.DoHQueryLog
	for rows.Next() {
		var l domain.DoHQueryLog
		if err := rows.Scan(&l.Domain, &l.Type, &l.ClientIP, &l.Blocked, &l.LatencyMS, &l.Timestamp); err != nil {
			return nil, err
		}
		results = append(results, l)
	}
	return results, rows.Err()
}

func (r *DoHRepo) GetStats(ctx context.Context) (*port.DoHStats, error) {
	var s port.DoHStats
	err := r.pool.QueryRow(ctx,
		`SELECT
		   COUNT(*)::int,
		   COUNT(*) FILTER (WHERE blocked)::int,
		   0,
		   COALESCE(AVG(latency_ms), 0)::int
		 FROM doh_query_logs`).
		Scan(&s.TotalQueries, &s.BlockedCount, &s.CacheHits, &s.AvgLatencyMS)
	if err != nil {
		return nil, err
	}
	return &s, nil
}


// --- SNIDomainRepo ---

type SNIDomainRepo struct{ pool *pgxpool.Pool }

var _ port.SNIDomainRepository = (*SNIDomainRepo)(nil)

func (r *SNIDomainRepo) CreateDomain(ctx context.Context, d *domain.SNIDomain) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sni_domains (id, inbound_id, domain, auto_cert, cert_status, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		d.ID, d.InboundID, d.Domain, d.AutoCert, d.CertStatus, d.ExpiresAt, d.CreatedAt)
	return err
}

func (r *SNIDomainRepo) GetDomain(ctx context.Context, id uuid.UUID) (*domain.SNIDomain, error) {
	var d domain.SNIDomain
	err := r.pool.QueryRow(ctx,
		`SELECT id, inbound_id, domain, auto_cert, cert_status, expires_at, created_at
		 FROM sni_domains WHERE id = $1`, id).
		Scan(&d.ID, &d.InboundID, &d.Domain, &d.AutoCert, &d.CertStatus, &d.ExpiresAt, &d.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *SNIDomainRepo) UpdateDomain(ctx context.Context, d *domain.SNIDomain) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sni_domains SET inbound_id = $2, domain = $3, auto_cert = $4, cert_status = $5, expires_at = $6
		 WHERE id = $1`,
		d.ID, d.InboundID, d.Domain, d.AutoCert, d.CertStatus, d.ExpiresAt)
	return err
}

func (r *SNIDomainRepo) DeleteDomain(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sni_domains WHERE id = $1`, id)
	return err
}

func (r *SNIDomainRepo) ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SNIDomain, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, inbound_id, domain, auto_cert, cert_status, expires_at, created_at
		 FROM sni_domains WHERE inbound_id = $1 ORDER BY created_at`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SNIDomain
	for rows.Next() {
		var d domain.SNIDomain
		if err := rows.Scan(&d.ID, &d.InboundID, &d.Domain, &d.AutoCert, &d.CertStatus, &d.ExpiresAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}

func (r *SNIDomainRepo) ListAll(ctx context.Context) ([]*domain.SNIDomain, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, inbound_id, domain, auto_cert, cert_status, expires_at, created_at
		 FROM sni_domains ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SNIDomain
	for rows.Next() {
		var d domain.SNIDomain
		if err := rows.Scan(&d.ID, &d.InboundID, &d.Domain, &d.AutoCert, &d.CertStatus, &d.ExpiresAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	return results, rows.Err()
}

func (r *SNIDomainRepo) CreateCert(ctx context.Context, c *domain.SSLCertificate) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ssl_certificates (id, domain, wildcard, issuer, status, auto_renew, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		c.ID, c.Domain, c.Wildcard, c.Issuer, c.Status, c.AutoRenew, c.ExpiresAt, c.CreatedAt)
	return err
}

func (r *SNIDomainRepo) GetCert(ctx context.Context, id uuid.UUID) (*domain.SSLCertificate, error) {
	var c domain.SSLCertificate
	err := r.pool.QueryRow(ctx,
		`SELECT id, domain, wildcard, issuer, status, auto_renew, expires_at, created_at
		 FROM ssl_certificates WHERE id = $1`, id).
		Scan(&c.ID, &c.Domain, &c.Wildcard, &c.Issuer, &c.Status, &c.AutoRenew, &c.ExpiresAt, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SNIDomainRepo) GetCertByDomain(ctx context.Context, d string) (*domain.SSLCertificate, error) {
	var c domain.SSLCertificate
	err := r.pool.QueryRow(ctx,
		`SELECT id, domain, wildcard, issuer, status, auto_renew, expires_at, created_at
		 FROM ssl_certificates WHERE domain = $1`, d).
		Scan(&c.ID, &c.Domain, &c.Wildcard, &c.Issuer, &c.Status, &c.AutoRenew, &c.ExpiresAt, &c.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *SNIDomainRepo) UpdateCert(ctx context.Context, c *domain.SSLCertificate) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE ssl_certificates SET domain = $2, wildcard = $3, issuer = $4, status = $5, auto_renew = $6, expires_at = $7
		 WHERE id = $1`,
		c.ID, c.Domain, c.Wildcard, c.Issuer, c.Status, c.AutoRenew, c.ExpiresAt)
	return err
}

func (r *SNIDomainRepo) DeleteCert(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ssl_certificates WHERE id = $1`, id)
	return err
}

func (r *SNIDomainRepo) ListCerts(ctx context.Context) ([]*domain.SSLCertificate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, domain, wildcard, issuer, status, auto_renew, expires_at, created_at
		 FROM ssl_certificates ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SSLCertificate
	for rows.Next() {
		var c domain.SSLCertificate
		if err := rows.Scan(&c.ID, &c.Domain, &c.Wildcard, &c.Issuer, &c.Status, &c.AutoRenew, &c.ExpiresAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *SNIDomainRepo) ListExpiringSoon(ctx context.Context, withinDays int) ([]*domain.SSLCertificate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, domain, wildcard, issuer, status, auto_renew, expires_at, created_at
		 FROM ssl_certificates
		 WHERE expires_at IS NOT NULL AND expires_at <= now() + make_interval(days => $1)
		 ORDER BY expires_at`, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SSLCertificate
	for rows.Next() {
		var c domain.SSLCertificate
		if err := rows.Scan(&c.ID, &c.Domain, &c.Wildcard, &c.Issuer, &c.Status, &c.AutoRenew, &c.ExpiresAt, &c.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &c)
	}
	return results, rows.Err()
}

func (r *SNIDomainRepo) CreateRoute(ctx context.Context, route *domain.SNIRoute) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sni_routes (id, inbound_id, sni, action, target_tag, priority, enabled)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		route.ID, route.InboundID, route.SNI, route.Action, route.TargetTag, route.Priority, route.Enabled)
	return err
}

func (r *SNIDomainRepo) UpdateRoute(ctx context.Context, route *domain.SNIRoute) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sni_routes SET inbound_id = $2, sni = $3, action = $4, target_tag = $5, priority = $6, enabled = $7
		 WHERE id = $1`,
		route.ID, route.InboundID, route.SNI, route.Action, route.TargetTag, route.Priority, route.Enabled)
	return err
}

func (r *SNIDomainRepo) DeleteRoute(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sni_routes WHERE id = $1`, id)
	return err
}

func (r *SNIDomainRepo) ListRoutesByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SNIRoute, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, inbound_id, sni, action, target_tag, priority, enabled
		 FROM sni_routes WHERE inbound_id = $1 ORDER BY priority`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SNIRoute
	for rows.Next() {
		var route domain.SNIRoute
		if err := rows.Scan(&route.ID, &route.InboundID, &route.SNI, &route.Action, &route.TargetTag, &route.Priority, &route.Enabled); err != nil {
			return nil, err
		}
		results = append(results, &route)
	}
	return results, rows.Err()
}

// --- TLSTricksRepo ---

type TLSTricksRepo struct{ pool *pgxpool.Pool }

var _ port.TLSTricksRepository = (*TLSTricksRepo)(nil)

func (r *TLSTricksRepo) Create(ctx context.Context, p *domain.TLSTrickProfile) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO tls_trick_profiles (id, name, description, fragment_enabled, fragment_length, fragment_interval, fingerprint, mux_enabled, mux_protocol, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		p.ID, p.Name, p.Description, p.FragmentEnabled, p.FragmentSize, p.FragmentInterval, p.UTLSFingerprint, p.MuxEnabled, "", p.Enabled, p.CreatedAt)
	return err
}

func (r *TLSTricksRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.TLSTrickProfile, error) {
	var p domain.TLSTrickProfile
	var muxProto string
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, fragment_enabled, fragment_length, fragment_interval, fingerprint, mux_enabled, mux_protocol, enabled, created_at
		 FROM tls_trick_profiles WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.FragmentEnabled, &p.FragmentSize, &p.FragmentInterval, &p.UTLSFingerprint, &p.MuxEnabled, &muxProto, &p.Enabled, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *TLSTricksRepo) Update(ctx context.Context, p *domain.TLSTrickProfile) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE tls_trick_profiles SET name = $2, description = $3, fragment_enabled = $4, fragment_length = $5, fragment_interval = $6, fingerprint = $7, mux_enabled = $8, mux_protocol = $9, enabled = $10
		 WHERE id = $1`,
		p.ID, p.Name, p.Description, p.FragmentEnabled, p.FragmentSize, p.FragmentInterval, p.UTLSFingerprint, p.MuxEnabled, "", p.Enabled)
	return err
}

func (r *TLSTricksRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tls_trick_profiles WHERE id = $1`, id)
	return err
}

func (r *TLSTricksRepo) List(ctx context.Context) ([]*domain.TLSTrickProfile, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, fragment_enabled, fragment_length, fragment_interval, fingerprint, mux_enabled, mux_protocol, enabled, created_at
		 FROM tls_trick_profiles ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.TLSTrickProfile
	for rows.Next() {
		var p domain.TLSTrickProfile
		var muxProto string
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.FragmentEnabled, &p.FragmentSize, &p.FragmentInterval, &p.UTLSFingerprint, &p.MuxEnabled, &muxProto, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, rows.Err()
}

// --- FingerprintRepo ---

type FingerprintRepo struct{ pool *pgxpool.Pool }

var _ port.FingerprintRepository = (*FingerprintRepo)(nil)

func (r *FingerprintRepo) GetPolicy(ctx context.Context) (*domain.FingerprintPolicy, error) {
	var p domain.FingerprintPolicy
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, default_action, log_unknown
		 FROM fingerprint_policy WHERE id = 1`).
		Scan(&p.Enabled, &p.DefaultAction, &p.LogUnknown)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *FingerprintRepo) SavePolicy(ctx context.Context, p *domain.FingerprintPolicy) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fingerprint_policy (id, enabled, default_action, log_unknown)
		 VALUES (1, $1, $2, $3)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   default_action = EXCLUDED.default_action,
		   log_unknown = EXCLUDED.log_unknown`,
		p.Enabled, p.DefaultAction, p.LogUnknown)
	return err
}

func (r *FingerprintRepo) CreateRule(ctx context.Context, rule *domain.FingerprintRule) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fingerprint_rules (id, name, fingerprint, ja3_hash, action, priority, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		rule.ID, rule.Name, rule.Fingerprint, rule.JA3Hash, rule.Action, rule.Priority, rule.Enabled, rule.CreatedAt)
	return err
}

func (r *FingerprintRepo) UpdateRule(ctx context.Context, rule *domain.FingerprintRule) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE fingerprint_rules SET name = $2, fingerprint = $3, ja3_hash = $4, action = $5, priority = $6, enabled = $7
		 WHERE id = $1`,
		rule.ID, rule.Name, rule.Fingerprint, rule.JA3Hash, rule.Action, rule.Priority, rule.Enabled)
	return err
}

func (r *FingerprintRepo) DeleteRule(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM fingerprint_rules WHERE id = $1`, id)
	return err
}

func (r *FingerprintRepo) ListRules(ctx context.Context) ([]*domain.FingerprintRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, fingerprint, ja3_hash, action, priority, enabled, created_at
		 FROM fingerprint_rules ORDER BY priority`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.FingerprintRule
	for rows.Next() {
		var rule domain.FingerprintRule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Fingerprint, &rule.JA3Hash, &rule.Action, &rule.Priority, &rule.Enabled, &rule.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &rule)
	}
	return results, rows.Err()
}

func (r *FingerprintRepo) SaveEvent(ctx context.Context, e *domain.FingerprintEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO fingerprint_events (id, client_ip, fingerprint, user_agent, action, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		e.ID, e.ClientIP, e.Fingerprint, e.UserAgent, e.Action, e.CreatedAt)
	return err
}

func (r *FingerprintRepo) ListEvents(ctx context.Context, limit int) ([]*domain.FingerprintEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, client_ip, fingerprint, user_agent, action, created_at
		 FROM fingerprint_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.FingerprintEvent
	for rows.Next() {
		var e domain.FingerprintEvent
		if err := rows.Scan(&e.ID, &e.ClientIP, &e.Fingerprint, &e.UserAgent, &e.Action, &e.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// --- FederationRepo ---

type FederationRepo struct{ pool *pgxpool.Pool }

var _ port.FederationRepository = (*FederationRepo)(nil)

func (r *FederationRepo) GetConfig(ctx context.Context) (*domain.FederationConfig, error) {
	var c domain.FederationConfig
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, cluster_name, sso_enabled, sync_interval
		 FROM federation_config WHERE id = 1`).
		Scan(&c.Enabled, &c.ClusterName, &c.SSOEnabled, &c.SyncInterval)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *FederationRepo) SaveConfig(ctx context.Context, c *domain.FederationConfig) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO federation_config (id, enabled, cluster_name, sso_enabled, sync_interval)
		 VALUES (1, $1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   cluster_name = EXCLUDED.cluster_name,
		   sso_enabled = EXCLUDED.sso_enabled,
		   sync_interval = EXCLUDED.sync_interval`,
		c.Enabled, c.ClusterName, c.SSOEnabled, c.SyncInterval)
	return err
}

func (r *FederationRepo) CreatePeer(ctx context.Context, p *domain.FederationPeer) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO federation_peers (id, name, endpoint, api_key, status, sync_users, sync_nodes, last_sync, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		p.ID, p.Name, p.Endpoint, p.APIKey, p.Status, p.SyncUsers, p.SyncNodes, p.LastSync, p.CreatedAt)
	return err
}

func (r *FederationRepo) GetPeer(ctx context.Context, id uuid.UUID) (*domain.FederationPeer, error) {
	var p domain.FederationPeer
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, endpoint, api_key, status, sync_users, sync_nodes, last_sync, created_at
		 FROM federation_peers WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Endpoint, &p.APIKey, &p.Status, &p.SyncUsers, &p.SyncNodes, &p.LastSync, &p.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *FederationRepo) UpdatePeer(ctx context.Context, p *domain.FederationPeer) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE federation_peers SET name = $2, endpoint = $3, api_key = $4, status = $5, sync_users = $6, sync_nodes = $7, last_sync = $8
		 WHERE id = $1`,
		p.ID, p.Name, p.Endpoint, p.APIKey, p.Status, p.SyncUsers, p.SyncNodes, p.LastSync)
	return err
}

func (r *FederationRepo) DeletePeer(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM federation_peers WHERE id = $1`, id)
	return err
}

func (r *FederationRepo) ListPeers(ctx context.Context) ([]*domain.FederationPeer, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, endpoint, api_key, status, sync_users, sync_nodes, last_sync, created_at
		 FROM federation_peers ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.FederationPeer
	for rows.Next() {
		var p domain.FederationPeer
		if err := rows.Scan(&p.ID, &p.Name, &p.Endpoint, &p.APIKey, &p.Status, &p.SyncUsers, &p.SyncNodes, &p.LastSync, &p.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &p)
	}
	return results, rows.Err()
}

func (r *FederationRepo) SaveSyncEvent(ctx context.Context, e *domain.FederationSyncEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO federation_sync_events (id, peer_name, direction, entity_type, count, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.PeerName, e.Direction, e.EntityType, e.Count, e.Status, e.CreatedAt)
	return err
}

func (r *FederationRepo) ListSyncEvents(ctx context.Context, limit int) ([]*domain.FederationSyncEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, peer_name, direction, entity_type, count, status, created_at
		 FROM federation_sync_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.FederationSyncEvent
	for rows.Next() {
		var e domain.FederationSyncEvent
		if err := rows.Scan(&e.ID, &e.PeerName, &e.Direction, &e.EntityType, &e.Count, &e.Status, &e.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}

// --- DeepLinkRepo ---

type DeepLinkRepo struct{ pool *pgxpool.Pool }

var _ port.DeepLinkRepository = (*DeepLinkRepo)(nil)

func (r *DeepLinkRepo) GetConfig(ctx context.Context) (*domain.DeepLinkConfig, error) {
	var c domain.DeepLinkConfig
	err := r.pool.QueryRow(ctx,
		`SELECT base_url, app_scheme, include_name, qr_logo_url
		 FROM deeplink_config WHERE id = 1`).
		Scan(&c.BaseURL, &c.Scheme, &c.Enabled, &c.QRLogoURL)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Enabled = true
	return &c, nil
}

func (r *DeepLinkRepo) SaveConfig(ctx context.Context, c *domain.DeepLinkConfig) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO deeplink_config (id, base_url, app_scheme, include_name, qr_logo_url)
		 VALUES (1, $1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET
		   base_url = EXCLUDED.base_url,
		   app_scheme = EXCLUDED.app_scheme,
		   include_name = EXCLUDED.include_name,
		   qr_logo_url = EXCLUDED.qr_logo_url`,
		c.BaseURL, c.Scheme, c.Enabled, c.QRLogoURL)
	return err
}

// --- QuotaNotifyRepo ---

type QuotaNotifyRepo struct{ pool *pgxpool.Pool }

var _ port.QuotaNotifyRepository = (*QuotaNotifyRepo)(nil)

func (r *QuotaNotifyRepo) GetConfig(ctx context.Context) (*domain.QuotaNotificationConfig, error) {
	var c domain.QuotaNotificationConfig
	var enabled, notifyTelegram, notifyEmail bool
	var thresholdPct int
	var messageTemplate string
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, threshold_pct, notify_telegram, notify_email, message_template
		 FROM quota_notify_config WHERE id = 1`).
		Scan(&enabled, &thresholdPct, &notifyTelegram, &notifyEmail, &messageTemplate)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Enabled = enabled
	c.NotifyTelegram = notifyTelegram
	c.NotifyAtPercent = []int{thresholdPct}
	return &c, nil
}

func (r *QuotaNotifyRepo) SaveConfig(ctx context.Context, c *domain.QuotaNotificationConfig) error {
	thresholdPct := 80
	if len(c.NotifyAtPercent) > 0 {
		thresholdPct = c.NotifyAtPercent[0]
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO quota_notify_config (id, enabled, threshold_pct, notify_telegram, notify_email, message_template)
		 VALUES (1, $1, $2, $3, $4, $5)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   threshold_pct = EXCLUDED.threshold_pct,
		   notify_telegram = EXCLUDED.notify_telegram,
		   notify_email = EXCLUDED.notify_email,
		   message_template = EXCLUDED.message_template`,
		c.Enabled, thresholdPct, c.NotifyTelegram, c.NotifyWebhook, "")
	return err
}

func (r *QuotaNotifyRepo) SaveEvent(ctx context.Context, e *domain.QuotaNotificationEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO quota_notify_events (id, user_id, username, threshold, usage_pct, notified, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.UserID, e.Username, e.Percent, e.Percent, e.Delivered, e.SentAt)
	return err
}

func (r *QuotaNotifyRepo) ListEvents(ctx context.Context, limit int) ([]*domain.QuotaNotificationEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, username, threshold, usage_pct, notified, created_at
		 FROM quota_notify_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.QuotaNotificationEvent
	for rows.Next() {
		var e domain.QuotaNotificationEvent
		var threshold, usagePct int
		var notified bool
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &threshold, &usagePct, &notified, &e.SentAt); err != nil {
			return nil, err
		}
		e.Percent = usagePct
		e.Delivered = notified
		results = append(results, &e)
	}
	return results, rows.Err()
}


// --- SubSettingsRepo ---

type SubSettingsRepo struct{ pool *pgxpool.Pool }

func (r *SubSettingsRepo) Get(ctx context.Context) (*domain.SubSettings, error) {
	row := r.pool.QueryRow(ctx, `SELECT update_interval FROM sub_settings WHERE id = 1`)
	var s domain.SubSettings
	if err := row.Scan(&s.UpdateInterval); err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SubSettingsRepo) Save(ctx context.Context, s *domain.SubSettings) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO sub_settings (id, update_interval, updated_at)
		VALUES (1, $1, now())
		ON CONFLICT (id) DO UPDATE SET update_interval = EXCLUDED.update_interval, updated_at = now()`,
		s.UpdateInterval)
	return err
}


// --- WireGuardPeerRepo ---

type WireGuardPeerRepo struct{ pool *pgxpool.Pool }

func (r *WireGuardPeerRepo) Get(ctx context.Context, inboundID, userID uuid.UUID) (*domain.WireGuardPeer, error) {
	row := r.pool.QueryRow(ctx, `SELECT inbound_id, user_id, private_key, public_key, address FROM wireguard_peers WHERE inbound_id=$1 AND user_id=$2`, inboundID, userID)
	var p domain.WireGuardPeer
	if err := row.Scan(&p.InboundID, &p.UserID, &p.PrivateKey, &p.PublicKey, &p.Address); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *WireGuardPeerRepo) Create(ctx context.Context, p *domain.WireGuardPeer) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO wireguard_peers (inbound_id, user_id, private_key, public_key, address) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (inbound_id,user_id) DO NOTHING`, p.InboundID, p.UserID, p.PrivateKey, p.PublicKey, p.Address)
	return err
}

func (r *WireGuardPeerRepo) ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.WireGuardPeer, error) {
	rows, err := r.pool.Query(ctx, `SELECT inbound_id, user_id, private_key, public_key, address FROM wireguard_peers WHERE inbound_id=$1 ORDER BY address`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.WireGuardPeer
	for rows.Next() {
		var p domain.WireGuardPeer
		if err := rows.Scan(&p.InboundID, &p.UserID, &p.PrivateKey, &p.PublicKey, &p.Address); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	return out, rows.Err()
}


// --- CleanIPScanRepo ---

type CleanIPScanRepo struct{ pool *pgxpool.Pool }

var _ port.CleanIPScanRepository = (*CleanIPScanRepo)(nil)

func (r *CleanIPScanRepo) SaveBatch(ctx context.Context, results []*domain.CleanIPScan) error {
	if len(results) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, s := range results {
		batch.Queue(
			`INSERT INTO clean_ip_scans (id, ip, latency_ms, loss_pct, score, reachable, throughput_mbps, scanned_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			s.ID, s.IP, s.LatencyMS, s.LossPct, s.Score, s.Reachable, s.ThroughputMbps, s.ScannedAt)
	}
	br := r.pool.SendBatch(ctx, batch)
	defer br.Close() //nolint:errcheck // batch result cleanup
	for range results {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (r *CleanIPScanRepo) List(ctx context.Context) ([]*domain.CleanIPScan, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, ip, latency_ms, loss_pct, score, reachable, throughput_mbps, scanned_at
		 FROM clean_ip_scans
		 ORDER BY score DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.CleanIPScan
	for rows.Next() {
		var s domain.CleanIPScan
		if err := rows.Scan(&s.ID, &s.IP, &s.LatencyMS, &s.LossPct, &s.Score, &s.Reachable, &s.ThroughputMbps, &s.ScannedAt); err != nil {
			return nil, err
		}
		results = append(results, &s)
	}
	return results, rows.Err()
}

func (r *CleanIPScanRepo) DeleteAll(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM clean_ip_scans`)
	return err
}

func (r *CleanIPScanRepo) UpdateThroughput(ctx context.Context, id uuid.UUID, mbps float64) error {
	_, err := r.pool.Exec(ctx, `UPDATE clean_ip_scans SET throughput_mbps = $1 WHERE id = $2`, mbps, id)
	return err
}


// --- SubHostRepo ---

type SubHostRepo struct{ pool *pgxpool.Pool }

var _ port.SubHostRepository = (*SubHostRepo)(nil)

const subHostColumns = `id, inbound_id, remark, address, port, sni, host_header, path,
	alpn, fingerprint, security, allow_insecure, mux_enable, fragment, priority, enabled, created_at`

func scanSubHost(row pgx.Row) (*domain.SubHost, error) {
	var h domain.SubHost
	if err := row.Scan(
		&h.ID, &h.InboundID, &h.Remark, &h.Address, &h.Port, &h.SNI, &h.HostHeader, &h.Path,
		&h.ALPN, &h.Fingerprint, &h.Security, &h.AllowInsecure, &h.MuxEnable, &h.Fragment,
		&h.Priority, &h.Enabled, &h.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *SubHostRepo) Create(ctx context.Context, h *domain.SubHost) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO sub_hosts (id, inbound_id, remark, address, port, sni, host_header, path,
		   alpn, fingerprint, security, allow_insecure, mux_enable, fragment, priority, enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		h.ID, h.InboundID, h.Remark, h.Address, h.Port, h.SNI, h.HostHeader, h.Path,
		h.ALPN, h.Fingerprint, h.Security, h.AllowInsecure, h.MuxEnable, h.Fragment,
		h.Priority, h.Enabled, h.CreatedAt)
	return err
}

func (r *SubHostRepo) Update(ctx context.Context, h *domain.SubHost) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE sub_hosts SET inbound_id = $2, remark = $3, address = $4, port = $5, sni = $6,
		   host_header = $7, path = $8, alpn = $9, fingerprint = $10, security = $11,
		   allow_insecure = $12, mux_enable = $13, fragment = $14, priority = $15, enabled = $16
		 WHERE id = $1`,
		h.ID, h.InboundID, h.Remark, h.Address, h.Port, h.SNI, h.HostHeader, h.Path,
		h.ALPN, h.Fingerprint, h.Security, h.AllowInsecure, h.MuxEnable, h.Fragment,
		h.Priority, h.Enabled)
	return err
}

func (r *SubHostRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sub_hosts WHERE id = $1`, id)
	return err
}

func (r *SubHostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SubHost, error) {
	h, err := scanSubHost(r.pool.QueryRow(ctx,
		`SELECT `+subHostColumns+` FROM sub_hosts WHERE id = $1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return h, nil
}

func (r *SubHostRepo) ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SubHost, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+subHostColumns+`
		 FROM sub_hosts WHERE inbound_id = $1
		 ORDER BY priority ASC, created_at ASC`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SubHost
	for rows.Next() {
		h, err := scanSubHost(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, rows.Err()
}

func (r *SubHostRepo) ListByInbounds(ctx context.Context, inboundIDs []uuid.UUID) ([]*domain.SubHost, error) {
	if len(inboundIDs) == 0 {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx,
		`SELECT `+subHostColumns+`
		 FROM sub_hosts WHERE inbound_id = ANY($1)
		 ORDER BY inbound_id, priority ASC, created_at ASC`, inboundIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.SubHost
	for rows.Next() {
		h, err := scanSubHost(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, h)
	}
	return results, rows.Err()
}


// --- RoutingPackRepo ---

type RoutingPackRepo struct{ pool *pgxpool.Pool }

var _ port.RoutingPackRepository = (*RoutingPackRepo)(nil)

func (r *RoutingPackRepo) Create(ctx context.Context, p *domain.RoutingPack) error {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return err
	}
	rulesJSON, err := json.Marshal(p.Rules)
	if err != nil {
		return err
	}
	outboundsJSON, err := json.Marshal(p.Outbounds)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO routing_packs (id, name, description, category, rules, outbounds)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, p.Name, p.Description, p.Category, rulesJSON, outboundsJSON)
	return err
}

func (r *RoutingPackRepo) Update(ctx context.Context, p *domain.RoutingPack) error {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return err
	}
	rulesJSON, err := json.Marshal(p.Rules)
	if err != nil {
		return err
	}
	outboundsJSON, err := json.Marshal(p.Outbounds)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE routing_packs SET name = $2, description = $3, category = $4, rules = $5, outbounds = $6
		 WHERE id = $1`,
		id, p.Name, p.Description, p.Category, rulesJSON, outboundsJSON)
	return err
}

func (r *RoutingPackRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM routing_packs WHERE id = $1`, id)
	return err
}

func scanRoutingPack(row pgx.Row) (*domain.RoutingPack, error) {
	var (
		p             domain.RoutingPack
		id            uuid.UUID
		rulesJSON     []byte
		outboundsJSON []byte
	)
	if err := row.Scan(&id, &p.Name, &p.Description, &p.Category, &rulesJSON, &outboundsJSON); err != nil {
		return nil, err
	}
	p.ID = id.String()
	if err := json.Unmarshal(rulesJSON, &p.Rules); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(outboundsJSON, &p.Outbounds); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *RoutingPackRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoutingPack, error) {
	p, err := scanRoutingPack(r.pool.QueryRow(ctx,
		`SELECT id, name, description, category, rules, outbounds
		 FROM routing_packs WHERE id = $1`, id))
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *RoutingPackRepo) List(ctx context.Context) ([]*domain.RoutingPack, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, category, rules, outbounds
		 FROM routing_packs ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.RoutingPack
	for rows.Next() {
		p, err := scanRoutingPack(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, p)
	}
	return results, rows.Err()
}

func (r *RoutingPackRepo) GetGlobalDefault(ctx context.Context) (string, error) {
	var packID string
	err := r.pool.QueryRow(ctx,
		`SELECT pack_id FROM routing_pack_selection WHERE id = 1`).Scan(&packID)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return packID, nil
}

func (r *RoutingPackRepo) SetGlobalDefault(ctx context.Context, packID string) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO routing_pack_selection (id, pack_id) VALUES (1, $1)
		 ON CONFLICT (id) DO UPDATE SET pack_id = EXCLUDED.pack_id`, packID)
	return err
}

func (r *RoutingPackRepo) GetUserPack(ctx context.Context, userID uuid.UUID) (string, error) {
	var packID string
	err := r.pool.QueryRow(ctx,
		`SELECT routing_pack_id FROM users WHERE id = $1`, userID).Scan(&packID)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return packID, nil
}

func (r *RoutingPackRepo) SetUserPack(ctx context.Context, userID uuid.UUID, packID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET routing_pack_id = $2 WHERE id = $1`, userID, packID)
	return err
}


// --- IPLimitRepo ---

type IPLimitRepo struct{ pool *pgxpool.Pool }

var _ port.IPLimitRepository = (*IPLimitRepo)(nil)

// GetPolicy returns the singleton enforcement policy, upserting and returning
// the conservative default when no row exists yet so callers always get a
// usable policy.
func (r *IPLimitRepo) GetPolicy(ctx context.Context) (*domain.IPLimitPolicy, error) {
	var p domain.IPLimitPolicy
	var action string
	err := r.pool.QueryRow(ctx,
		`SELECT enabled, action, alert_cooldown, restore_after
		 FROM ip_limit_policy WHERE id = 1`).
		Scan(&p.Enabled, &action, &p.AlertCooldown, &p.RestoreAfter)
	if err == pgx.ErrNoRows {
		def := domain.DefaultIPLimitPolicy()
		return &def, nil
	}
	if err != nil {
		return nil, err
	}
	p.Action = domain.IPLimitAction(action)
	return &p, nil
}

func (r *IPLimitRepo) UpdatePolicy(ctx context.Context, p *domain.IPLimitPolicy) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ip_limit_policy (id, enabled, action, alert_cooldown, restore_after)
		 VALUES (1, $1, $2, $3, $4)
		 ON CONFLICT (id) DO UPDATE SET
		   enabled = EXCLUDED.enabled,
		   action = EXCLUDED.action,
		   alert_cooldown = EXCLUDED.alert_cooldown,
		   restore_after = EXCLUDED.restore_after`,
		p.Enabled, string(p.Action), p.AlertCooldown, p.RestoreAfter)
	return err
}

func (r *IPLimitRepo) InsertEvent(ctx context.Context, e *domain.IPLimitEvent) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ip_limit_events (id, user_id, username, online_ips, limit_val, action, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.UserID, e.Username, e.OnlineIPs, e.Limit, e.Action, e.CreatedAt)
	return err
}

func (r *IPLimitRepo) ListEvents(ctx context.Context, limit int) ([]*domain.IPLimitEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, username, online_ips, limit_val, action, created_at
		 FROM ip_limit_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*domain.IPLimitEvent
	for rows.Next() {
		var e domain.IPLimitEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &e.OnlineIPs, &e.Limit, &e.Action, &e.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, &e)
	}
	return results, rows.Err()
}


// --- ResellerPaymentRepo ---

// ResellerPaymentRepo manages per-reseller payment configuration.
type ResellerPaymentRepo struct{ pool *pgxpool.Pool }

// Get loads the payment config for an admin. Returns nil (not an error) when no row exists.
func (r *ResellerPaymentRepo) Get(ctx context.Context, adminID uuid.UUID) (*domain.ResellerPaymentConfig, error) {
	var cfg domain.ResellerPaymentConfig
	var cryptoJSON, methodsJSON []byte
	err := r.pool.QueryRow(ctx,
		`SELECT admin_id, card_number, card_holder, card_bank, crypto_addresses,
		        zarinpal_merchant_id, manual_instructions, enabled_methods
		 FROM reseller_payment_config WHERE admin_id = $1`, adminID).
		Scan(&cfg.AdminID, &cfg.CardNumber, &cfg.CardHolder, &cfg.CardBank,
			&cryptoJSON, &cfg.ZarinpalMerchantID, &cfg.ManualInstructions, &methodsJSON)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(cryptoJSON, &cfg.CryptoAddresses); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(methodsJSON, &cfg.EnabledMethods); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Upsert inserts or updates the payment config for an admin.
func (r *ResellerPaymentRepo) Upsert(ctx context.Context, cfg *domain.ResellerPaymentConfig) error {
	cryptoJSON, err := json.Marshal(cfg.CryptoAddresses)
	if err != nil {
		return err
	}
	methodsJSON, err := json.Marshal(cfg.EnabledMethods)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO reseller_payment_config
		   (admin_id, card_number, card_holder, card_bank, crypto_addresses,
		    zarinpal_merchant_id, manual_instructions, enabled_methods)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (admin_id) DO UPDATE SET
		   card_number = EXCLUDED.card_number,
		   card_holder = EXCLUDED.card_holder,
		   card_bank = EXCLUDED.card_bank,
		   crypto_addresses = EXCLUDED.crypto_addresses,
		   zarinpal_merchant_id = EXCLUDED.zarinpal_merchant_id,
		   manual_instructions = EXCLUDED.manual_instructions,
		   enabled_methods = EXCLUDED.enabled_methods`,
		cfg.AdminID, cfg.CardNumber, cfg.CardHolder, cfg.CardBank,
		cryptoJSON, cfg.ZarinpalMerchantID, cfg.ManualInstructions, methodsJSON)
	return err
}
