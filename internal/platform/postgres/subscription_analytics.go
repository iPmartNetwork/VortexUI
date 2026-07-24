package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SubAnalyticsRepo implements port.SubscriptionAnalyticsRepository on PostgreSQL
// with TimescaleDB-friendly queries for efficient time-series aggregation.
type SubAnalyticsRepo struct {
	pool *pgxpool.Pool
}

var _ port.SubscriptionAnalyticsRepository = (*SubAnalyticsRepo)(nil)

// Record stores a single subscription fetch event.
func (r *SubAnalyticsRepo) Record(ctx context.Context, userID uuid.UUID, format, isp, client string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscription_analytics (id, user_id, format_type, isp_name, client_app, fetched_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), userID, format, isp, client, time.Now())
	return err
}

// Report aggregates subscription analytics between two timestamps, grouping
// counts by format, ISP, and hour-of-day.
func (r *SubAnalyticsRepo) Report(ctx context.Context, from, to time.Time) (*domain.SubAnalyticsReport, error) {
	report := &domain.SubAnalyticsReport{}

	// By format
	rows, err := r.pool.Query(ctx, `
		SELECT format_type, COUNT(*) AS cnt
		FROM subscription_analytics
		WHERE fetched_at >= $1 AND fetched_at <= $2
		GROUP BY format_type
		ORDER BY cnt DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var fc domain.FormatCount
		if err := rows.Scan(&fc.Format, &fc.Count); err != nil {
			return nil, err
		}
		report.ByFormat = append(report.ByFormat, fc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// By ISP
	rows2, err := r.pool.Query(ctx, `
		SELECT isp_name, COUNT(*) AS cnt
		FROM subscription_analytics
		WHERE fetched_at >= $1 AND fetched_at <= $2 AND isp_name <> ''
		GROUP BY isp_name
		ORDER BY cnt DESC`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var ic domain.ISPCount
		if err := rows2.Scan(&ic.ISP, &ic.Count); err != nil {
			return nil, err
		}
		report.ByISP = append(report.ByISP, ic)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	// By hour of day
	rows3, err := r.pool.Query(ctx, `
		SELECT EXTRACT(HOUR FROM fetched_at)::INT AS h, COUNT(*) AS cnt
		FROM subscription_analytics
		WHERE fetched_at >= $1 AND fetched_at <= $2
		GROUP BY h
		ORDER BY h`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var tc domain.TimeCount
		if err := rows3.Scan(&tc.Hour, &tc.Count); err != nil {
			return nil, err
		}
		report.ByTime = append(report.ByTime, tc)
	}
	if err := rows3.Err(); err != nil {
		return nil, err
	}

	return report, nil
}
