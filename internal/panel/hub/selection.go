package hub

import (
	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// selectFailoverTarget picks the best healthy node to absorb a failed node's
// users. It is a pure function so the policy is unit-tested in isolation.
//
// Policy: among healthy candidates (excluding the failed node), prefer the one
// with the highest configured UsageRatio (operator-assigned capacity weight),
// breaking ties by lowest current CPU so we steer load toward the least-busy
// server. Returns ok=false when no healthy candidate exists.
func selectFailoverTarget(candidates []*domain.Node, excludeID uuid.UUID) (*domain.Node, bool) {
	var best *domain.Node
	for _, n := range candidates {
		if n.ID == excludeID || !n.IsHealthy() {
			continue
		}
		if best == nil || better(n, best) {
			best = n
		}
	}
	return best, best != nil
}

func better(a, b *domain.Node) bool {
	if a.UsageRatio != b.UsageRatio {
		return a.UsageRatio > b.UsageRatio
	}
	return a.Health.CPUPercent < b.Health.CPUPercent
}
