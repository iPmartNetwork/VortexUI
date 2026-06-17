package port

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
)

// DoHRepository persists DoH configuration and query logs.
type DoHRepository interface {
	GetConfig(ctx context.Context) (*domain.DoHConfig, error)
	SaveConfig(ctx context.Context, c *domain.DoHConfig) error

	SaveQueryLog(ctx context.Context, log *domain.DoHQueryLog) error
	ListQueryLogs(ctx context.Context, limit int) ([]domain.DoHQueryLog, error)
	GetStats(ctx context.Context) (*DoHStats, error)
}

// DoHStats provides aggregated DNS resolution stats.
type DoHStats struct {
	TotalQueries int `json:"total_queries"`
	BlockedCount int `json:"blocked_count"`
	CacheHits    int `json:"cache_hits"`
	AvgLatencyMS int `json:"avg_latency_ms"`
}
