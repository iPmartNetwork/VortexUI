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
