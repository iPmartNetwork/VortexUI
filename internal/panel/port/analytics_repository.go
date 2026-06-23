package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// AnalyticsRepository provides aggregated analytics queries.
type AnalyticsRepository interface {
	// GeoBreakdown returns traffic aggregated by country for a time range.
	GeoBreakdown(ctx context.Context, q SeriesQuery) ([]domain.GeoTrafficPoint, error)

	// TopUsers returns the top N users by traffic usage.
	TopUsers(ctx context.Context, limit int, adminID *uuid.UUID) ([]domain.UserTrafficRank, error)

	// PeakHours returns traffic aggregated by hour-of-day for a time range.
	PeakHours(ctx context.Context, q SeriesQuery) ([]domain.PeakHour, error)

	// TotalTraffic returns total up/down bytes for a time range.
	TotalTraffic(ctx context.Context, q SeriesQuery) (int64, int64, error)
}
