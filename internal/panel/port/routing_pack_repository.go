package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// RoutingPackRepository persists admin-defined custom routing packs and the
// pack selection (global default + per-user). Built-in packs are not stored
// here; the service merges them in from domain.BuiltinRoutingPacks.
type RoutingPackRepository interface {
	// Custom pack CRUD. The pack's ID is a UUID in string form.
	Create(ctx context.Context, p *domain.RoutingPack) error
	Update(ctx context.Context, p *domain.RoutingPack) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RoutingPack, error)
	List(ctx context.Context) ([]*domain.RoutingPack, error)

	// Global default selection (singleton). Empty string means none selected.
	GetGlobalDefault(ctx context.Context) (string, error)
	SetGlobalDefault(ctx context.Context, packID string) error

	// Per-user (per-subscription) selection, stored on the user row.
	GetUserPack(ctx context.Context, userID uuid.UUID) (string, error)
	SetUserPack(ctx context.Context, userID uuid.UUID, packID string) error
}
