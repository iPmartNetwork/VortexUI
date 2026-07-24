package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// SubscriptionAnalyticsRepository persists subscription fetch events for
// format/ISP/time analysis using TimescaleDB-friendly storage.
type SubscriptionAnalyticsRepository interface {
	// Record stores a single subscription fetch event.
	Record(ctx context.Context, userID uuid.UUID, format, isp, client string) error

	// Report aggregates subscription analytics between two timestamps, grouping
	// counts by format, ISP, and hour-of-day.
	Report(ctx context.Context, from, to time.Time) (*domain.SubAnalyticsReport, error)
}
