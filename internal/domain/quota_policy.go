package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuotaTier defines a speed reduction level at a usage percentage threshold.
type QuotaTier struct {
	UsagePercent int   `json:"usage_percent"` // e.g. 80, 100
	SpeedLimit   int64 `json:"speed_limit"`   // bytes/sec; 0 = full speed
	Action       string `json:"action"`        // "reduce" | "block"
}

// QuotaPolicy defines a fair-use policy with progressive speed tiers.
type QuotaPolicy struct {
	ID        uuid.UUID    `json:"id"`
	Name      string       `json:"name"`
	Tiers     []QuotaTier  `json:"tiers"`
	Enabled   bool         `json:"enabled"`
	CreatedAt time.Time    `json:"created_at"`
}

// ActiveTier returns the tier that should be enforced for a given usage ratio
// (used_traffic / data_limit). Returns nil if no tier matches.
func (p *QuotaPolicy) ActiveTier(usageRatio float64) *QuotaTier {
	if !p.Enabled || len(p.Tiers) == 0 {
		return nil
	}
	percent := int(usageRatio * 100)
	var match *QuotaTier
	for i := range p.Tiers {
		if percent >= p.Tiers[i].UsagePercent {
			if match == nil || p.Tiers[i].UsagePercent > match.UsagePercent {
				match = &p.Tiers[i]
			}
		}
	}
	return match
}
