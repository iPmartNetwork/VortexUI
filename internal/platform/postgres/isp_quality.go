package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ISPQualityRepo implements port.ISPQualityRepository on PostgreSQL.
type ISPQualityRepo struct {
	pool *pgxpool.Pool
}

var _ port.ISPQualityRepository = (*ISPQualityRepo)(nil)

// RecordMetric stores a single ISP quality measurement for the given hour/day.
func (r *ISPQualityRepo) RecordMetric(ctx context.Context, isp string, hour, day int, score float64, date time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO isp_quality_metrics (id, isp_name, hour, day_of_week, score, sample_date)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), isp, hour, day, score, date)
	return err
}

// GetHeatmap retrieves averaged quality scores per hour/day for the given ISP
// over the last N days, returning a slice of HeatmapCell values.
func (r *ISPQualityRepo) GetHeatmap(ctx context.Context, isp string, days int) ([]domain.HeatmapCell, error) {
	since := time.Now().AddDate(0, 0, -days)

	rows, err := r.pool.Query(ctx, `
		SELECT day_of_week, hour, AVG(score) AS avg_score
		FROM isp_quality_metrics
		WHERE isp_name = $1 AND sample_date >= $2
		GROUP BY day_of_week, hour
		ORDER BY day_of_week, hour`,
		isp, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cells []domain.HeatmapCell
	for rows.Next() {
		var c domain.HeatmapCell
		if err := rows.Scan(&c.Day, &c.Hour, &c.Score); err != nil {
			return nil, err
		}
		cells = append(cells, c)
	}
	return cells, rows.Err()
}
