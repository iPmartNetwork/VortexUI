package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// BulkOperationRepository persists bulk operation history entries.
type BulkOperationRepository interface {
	Create(ctx context.Context, op *domain.BulkOperation) error
	List(ctx context.Context, adminID *uuid.UUID, limit, offset int) ([]*domain.BulkOperation, int, error)
}
