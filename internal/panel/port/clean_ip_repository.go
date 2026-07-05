package port

import (
	"context"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// CleanIPScanRepository persists clean-IP scan results. A scan replaces the
// previous result set (DeleteAll + SaveBatch); List returns results scored
// best-first (score DESC).
type CleanIPScanRepository interface {
	SaveBatch(ctx context.Context, results []*domain.CleanIPScan) error
	List(ctx context.Context) ([]*domain.CleanIPScan, error)
	DeleteAll(ctx context.Context) error
	// UpdateThroughput records a real download-speed measurement (Mbps) for
	// one previously scanned IP, identified by its result ID.
	UpdateThroughput(ctx context.Context, id uuid.UUID, mbps float64) error
}
