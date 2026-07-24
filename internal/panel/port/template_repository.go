package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// UserTemplateRepository persists user provisioning templates.
type UserTemplateRepository interface {
	Create(ctx context.Context, t *domain.UserTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.UserTemplate, error)
	GetByName(ctx context.Context, name string) (*domain.UserTemplate, error)
	Update(ctx context.Context, t *domain.UserTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.UserTemplate, error)
	ListForAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.UserTemplate, error)
}
