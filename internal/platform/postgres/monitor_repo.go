package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MonitorRepo derives "live" connection info from recent traffic samples, which
// works regardless of the core's stats-API version. A user with a traffic point
// in the recent window is considered online.
type MonitorRepo struct{ pool *pgxpool.Pool }

// LiveUser is one currently-active user inferred from recent traffic.
type LiveUser struct {
	UserID   string
	Username string
	NodeID   string
	LastSeen time.Time
}

// RecentActive returns users with traffic in the last window, joined to their
// username. One row per (user, node) pair.
func (r *MonitorRepo) RecentActive(ctx context.Context, window time.Duration) ([]LiveUser, error) {
	since := time.Now().Add(-window)
	rows, err := r.pool.Query(ctx,
		`SELECT tp.user_id, COALESCE(u.username, ''), tp.node_id, MAX(tp.time) AS last_seen
		 FROM traffic_points tp
		 LEFT JOIN users u ON u.id = tp.user_id
		 WHERE tp.time >= $1 AND (tp.up > 0 OR tp.down > 0)
		 GROUP BY tp.user_id, u.username, tp.node_id
		 ORDER BY last_seen DESC`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []LiveUser
	for rows.Next() {
		var lu LiveUser
		if err := rows.Scan(&lu.UserID, &lu.Username, &lu.NodeID, &lu.LastSeen); err != nil {
			return nil, err
		}
		out = append(out, lu)
	}
	return out, rows.Err()
}
