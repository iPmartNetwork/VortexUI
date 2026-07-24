package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// RevenueRepository persists revenue entries (income/expense) and produces
// aggregated financial reports with time-series and per-admin breakdowns.
type RevenueRepository interface {
	// Create inserts a new revenue entry.
	Create(ctx context.Context, entry *domain.RevenueEntry) error

	// Report aggregates revenue data between two timestamps. If adminID is non-nil,
	// the report is scoped to that specific admin/reseller.
	Report(ctx context.Context, adminID *uuid.UUID, from, to time.Time) (*domain.RevenueReport, error)
}
