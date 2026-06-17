package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// ProbingRepository persists probing protection state.
type ProbingRepository interface {
	GetPolicy(ctx context.Context) (*domain.ProbingPolicy, error)
	SavePolicy(ctx context.Context, p *domain.ProbingPolicy) error

	SaveEvent(ctx context.Context, e *domain.ProbeEvent) error
	ListEvents(ctx context.Context, limit, offset int) ([]*domain.ProbeEvent, int, error)

	BlockIP(ctx context.Context, b *domain.BlockedIP) error
	UnblockIP(ctx context.Context, ip string) error
	ListBlockedIPs(ctx context.Context) ([]domain.BlockedIP, error)
	IsBlocked(ctx context.Context, ip string) (bool, error)
}
