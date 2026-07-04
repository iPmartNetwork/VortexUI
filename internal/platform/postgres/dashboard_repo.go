package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DashboardRepo implements port.DashboardCounter.
type DashboardRepo struct{ pool *pgxpool.Pool }

var _ port.DashboardCounter = (*DashboardRepo)(nil)

func (s *Store) Dashboard() *DashboardRepo { return &DashboardRepo{pool: s.pool} }

func (r *DashboardRepo) CountOpenTickets(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tickets WHERE status IN ('open', 'answered')`,
	).Scan(&n)
	return n, err
}

func (r *DashboardRepo) CountPendingOrders(ctx context.Context, adminID *uuid.UUID, sudo bool) (int, error) {
	var n int
	var err error
	if sudo || adminID == nil {
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM orders WHERE status = $1`, domain.OrderPending,
		).Scan(&n)
	} else {
		err = r.pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM orders WHERE status = $1 AND admin_id = $2`,
			domain.OrderPending, *adminID,
		).Scan(&n)
	}
	return n, err
}

func (r *DashboardRepo) CountUsersCreatedSince(ctx context.Context, since time.Time) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at >= $1`, since,
	).Scan(&n)
	return n, err
}

func (r *DashboardRepo) CountUsersCreatedBetween(ctx context.Context, from, to time.Time) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2`, from, to,
	).Scan(&n)
	return n, err
}

func (r *DashboardRepo) CountProbeEventsSince(ctx context.Context, since time.Time) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM probe_events WHERE created_at >= $1`, since,
	).Scan(&n)
	if err == pgx.ErrNoRows {
		return 0, nil
	}
	return n, err
}

func (r *DashboardRepo) CountBlockedIPs(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM blocked_ips`).Scan(&n)
	return n, err
}

func (r *DashboardRepo) UsersCountByNode(ctx context.Context) (map[uuid.UUID]int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT i.node_id, COUNT(DISTINCT ui.user_id)::int
		FROM user_inbounds ui
		JOIN inbounds i ON i.id = ui.inbound_id
		JOIN users u ON u.id = ui.user_id
		WHERE u.status = $1
		GROUP BY i.node_id`,
		domain.UserStatusActive,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[uuid.UUID]int{}
	for rows.Next() {
		var nodeID uuid.UUID
		var count int
		if err := rows.Scan(&nodeID, &count); err != nil {
			return nil, err
		}
		out[nodeID] = count
	}
	return out, rows.Err()
}

func (r *DashboardRepo) TopUsersOverview(ctx context.Context, limit int, adminID *uuid.UUID, sudo bool) ([]domain.OverviewUserRow, error) {
	if limit <= 0 {
		limit = 5
	}
	const base = `
		SELECT u.id, u.username, u.status, u.used_traffic, u.data_limit, u.expire_at,
		       COALESCE(i.protocol, ''), COALESCE(i.network, ''), COALESCE(i.security, '')
		FROM users u
		LEFT JOIN LATERAL (
			SELECT ib.protocol, ib.network, ib.security
			FROM user_inbounds ui
			JOIN inbounds ib ON ib.id = ui.inbound_id
			WHERE ui.user_id = u.id
			ORDER BY ib.tag
			LIMIT 1
		) i ON true
		WHERE u.status = $1`
	var rows pgx.Rows
	var err error
	if sudo || adminID == nil {
		rows, err = r.pool.Query(ctx, base+` ORDER BY u.used_traffic DESC LIMIT $2`, domain.UserStatusActive, limit)
	} else {
		rows, err = r.pool.Query(ctx, base+` AND u.admin_id = $2 ORDER BY u.used_traffic DESC LIMIT $3`, domain.UserStatusActive, *adminID, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.OverviewUserRow, 0, limit)
	for rows.Next() {
		var row domain.OverviewUserRow
		var id uuid.UUID
		var protocol, network, security string
		if err := rows.Scan(&id, &row.Username, &row.Status, &row.UsedTraffic, &row.DataLimit, &row.ExpireAt, &protocol, &network, &security); err != nil {
			return nil, err
		}
		row.ID = id.String()
		row.ProtocolLabel = ProtocolLabel(protocol, network, security)
		out = append(out, row)
	}
	return out, rows.Err()
}
