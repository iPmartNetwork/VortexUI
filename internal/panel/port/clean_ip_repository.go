package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// CleanIPScanRepository persists clean-IP scan results. A scan replaces the
// previous result set (DeleteAll + SaveBatch); List returns results scored
// best-first (score DESC).
type CleanIPScanRepository interface {
	SaveBatch(ctx context.Context, results []*domain.CleanIPScan) error
	List(ctx context.Context) ([]*domain.CleanIPScan, error)
	DeleteAll(ctx context.Context) error
}
