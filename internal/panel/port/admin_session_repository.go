package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// AdminSessionRepository persists admin login sessions.
type AdminSessionRepository interface {
	Create(ctx context.Context, session *domain.AdminSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AdminSession, error)
	ListByAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.AdminSession, error)
	UpdateLastActive(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForAdmin(ctx context.Context, adminID uuid.UUID) error
}
