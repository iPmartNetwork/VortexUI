package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// InboundTrafficRepo implements port.InboundTrafficRepository using raw pgx queries.
type InboundTrafficRepo struct {
	pool *pgxpool.Pool
}

var _ port.InboundTrafficRepository = (*InboundTrafficRepo)(nil)

// AddTraffic upserts today's traffic for an inbound (ON CONFLICT DO UPDATE).
func (r *InboundTrafficRepo) AddTraffic(ctx context.Context, inboundID uuid.UUID, upload, download int64) error {
	const q = `
INSERT INTO inbound_traffic_stats (inbound_id, date, upload, download)
VALUES ($1, CURRENT_DATE, $2, $3)
ON CONFLICT (inbound_id, date) DO UPDATE SET
    upload   = inbound_traffic_stats.upload   + EXCLUDED.upload,
    download = inbound_traffic_stats.download + EXCLUDED.download
`
	_, err := r.pool.Exec(ctx, q, inboundID, upload, download)
	if err != nil {
		return fmt.Errorf("inbound traffic upsert: %w", err)
	}
	return nil
}

// GetStats returns accumulated totals and daily breakdown for an inbound.
func (r *InboundTrafficRepo) GetStats(ctx context.Context, inboundID uuid.UUID, days int) (*domain.InboundTrafficStats, error) {
	if days <= 0 {
		days = 30
	}

	// Fetch daily series (most recent N days).
	const dailyQ = `
SELECT date, upload, download
FROM inbound_traffic_stats
WHERE inbound_id = $1 AND date >= (CURRENT_DATE - $2::int)
ORDER BY date DESC
LIMIT $2
`
	rows, err := r.pool.Query(ctx, dailyQ, inboundID, days)
	if err != nil {
		return nil, fmt.Errorf("inbound traffic daily: %w", err)
	}
	defer rows.Close()

	var totalUp, totalDown int64
	var daily []domain.InboundDailyPoint
	for rows.Next() {
		var d time.Time
		var up, down int64
		if err := rows.Scan(&d, &up, &down); err != nil {
			return nil, fmt.Errorf("inbound traffic scan: %w", err)
		}
		totalUp += up
		totalDown += down
		daily = append(daily, domain.InboundDailyPoint{
			Date:     d.Format("2006-01-02"),
			Upload:   up,
			Download: down,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("inbound traffic rows: %w", err)
	}

	return &domain.InboundTrafficStats{
		InboundID: inboundID,
		Upload:    totalUp,
		Download:  totalDown,
		Total:     totalUp + totalDown,
		Daily:     daily,
	}, nil
}
