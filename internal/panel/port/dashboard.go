package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// DashboardCounter exposes lightweight aggregate queries for dashboard widgets.
type DashboardCounter interface {
	CountOpenTickets(ctx context.Context) (int, error)
	CountPendingOrders(ctx context.Context, adminID *uuid.UUID, sudo bool) (int, error)
	CountUsersCreatedSince(ctx context.Context, since time.Time) (int, error)
	CountUsersCreatedBetween(ctx context.Context, from, to time.Time) (int, error)
	CountProbeEventsSince(ctx context.Context, since time.Time) (int, error)
	CountBlockedIPs(ctx context.Context) (int, error)
}
