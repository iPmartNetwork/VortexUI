package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ClientTemplateRepository persists client template configurations.
type ClientTemplateRepository interface {
	Create(ctx context.Context, t *domain.ClientTemplate) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ClientTemplate, error)
	Update(ctx context.Context, t *domain.ClientTemplate) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.ClientTemplate, error)
	ListEnabled(ctx context.Context) ([]*domain.ClientTemplate, error)
}
