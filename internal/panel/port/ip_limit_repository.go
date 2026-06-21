package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// IPLimitRepository persists the singleton IP-limit enforcement policy and the
// audit trail of enforcement events. GetPolicy returns the persisted singleton,
// falling back to the default when no row exists yet.
type IPLimitRepository interface {
	GetPolicy(ctx context.Context) (*domain.IPLimitPolicy, error)
	UpdatePolicy(ctx context.Context, p *domain.IPLimitPolicy) error
	InsertEvent(ctx context.Context, e *domain.IPLimitEvent) error
	ListEvents(ctx context.Context, limit int) ([]*domain.IPLimitEvent, error)
}
