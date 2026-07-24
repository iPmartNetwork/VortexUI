package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// IPWhitelistRepository persists admin IP whitelist entries.
type IPWhitelistRepository interface {
	Create(ctx context.Context, entry *domain.AdminIPWhitelist) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context) ([]*domain.AdminIPWhitelist, error)
	ListByAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.AdminIPWhitelist, error)
}
