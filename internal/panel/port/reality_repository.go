package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// RealityScanRepository persists Reality SNI scan results.
type RealityScanRepository interface {
	SaveBatch(ctx context.Context, results []domain.RealityScan) error
	ListByNode(ctx context.Context, nodeID uuid.UUID) ([]domain.RealityScan, error)
	DeleteByNode(ctx context.Context, nodeID uuid.UUID) error
}
