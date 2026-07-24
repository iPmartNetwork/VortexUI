package port

import (
	"context"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// ISPQualityRepository persists ISP quality metrics for heatmap visualization.
type ISPQualityRepository interface {
	// RecordMetric stores a single ISP quality measurement for a given hour/day.
	RecordMetric(ctx context.Context, isp string, hour, day int, score float64, date time.Time) error

	// GetHeatmap retrieves the heatmap grid cells for the specified ISP over the
	// last N days, returning one cell per hour/day combination with averaged scores.
	GetHeatmap(ctx context.Context, isp string, days int) ([]domain.HeatmapCell, error)
}
