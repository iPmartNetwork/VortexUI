package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// IPBanRepository persists IP ban entries.
type IPBanRepository interface {
	Create(ctx context.Context, ban *domain.IPBan) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByIP(ctx context.Context, ip string) (*domain.IPBan, error)
	ListActive(ctx context.Context) ([]*domain.IPBan, error)
	DeleteExpired(ctx context.Context) (int, error)
}
