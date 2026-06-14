package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// TrafficRepo implements port.TrafficRepository against the traffic_points
// hypertable.
type TrafficRepo struct{ q *db.Queries }

var _ port.TrafficRepository = (*TrafficRepo)(nil)

// WriteBatch inserts samples via pgx CopyFrom for throughput.
func (r *TrafficRepo) WriteBatch(ctx context.Context, points []domain.TrafficPoint) error {
	if len(points) == 0 {
		return nil
	}
	rows := make([]db.WriteTrafficPointsParams, len(points))
	for i, p := range points {
		rows[i] = db.WriteTrafficPointsParams{
			Time:   timeToTS(p.Time),
			UserID: p.UserID,
			NodeID: p.NodeID,
			Up:     p.Up,
			Down:   p.Down,
		}
	}
	_, err := r.q.WriteTrafficPoints(ctx, rows)
	return err
}

func (r *TrafficRepo) UsageSeries(ctx context.Context, userID uuid.UUID, q port.SeriesQuery) ([]domain.TrafficPoint, error) {
	bucket, err := parseBucket(q.Bucket)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.UsageSeries(ctx, db.UsageSeriesParams{
		Bucket: bucket,
		UserID: userID,
		FromTs: timeToTS(unixToTime(q.FromUnix)),
		ToTs:   timeToTS(unixToTime(q.ToUnix)),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.TrafficPoint, len(rows))
	for i, row := range rows {
		out[i] = domain.TrafficPoint{
			Time:   row.Bucket.Time,
			UserID: userID,
			Up:     row.Up,
			Down:   row.Down,
		}
	}
	return out, nil
}
