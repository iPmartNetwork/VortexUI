package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// SubHostRepository persists Marzban-style subscription host overrides bound to
// inbounds. ListByInbound(s) return hosts ordered by priority for projection
// into subscription output.
type SubHostRepository interface {
	Create(ctx context.Context, h *domain.SubHost) error
	Update(ctx context.Context, h *domain.SubHost) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.SubHost, error)
	ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.SubHost, error)
	ListByInbounds(ctx context.Context, inboundIDs []uuid.UUID) ([]*domain.SubHost, error)
}
