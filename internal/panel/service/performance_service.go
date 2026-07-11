package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// InMemoryCacheService implements port.CacheService using in-memory storage
type InMemoryCacheService struct {
	data   map[string]*cacheItem
	mu     sync.RWMutex
	stats  *domain.CacheStats
	maxSize int64
}

type cacheItem struct {
	value     []byte
	expiresAt time.Time
	hitCount  int64
}

// NewInMemoryCacheService creates new in-memory cache service
func NewInMemoryCacheService(maxSizeBytes int64) *InMemoryCacheService {
	return &InMemoryCacheService{
		data:    make(map[string]*cacheItem),
		maxSize: maxSizeBytes,
		stats: &domain.CacheStats{
			ID:        uuid.New(),
			CacheType: "in-memory",
			MaxSize:   maxSizeBytes,
			UpdatedAt: time.Now(),
		},
	}
}

// Set stores value in cache with TTL
func (c *InMemoryCacheService) Set(ctx context.Context, key string, value []byte, ttlSeconds int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = &cacheItem{
		value:     value,
		expiresAt: time.Now().Add(time.Duration(ttlSeconds) * time.Second),
		hitCount:  0,
	}

	c.stats.CurrentSize = c.calculateSize()
	c.stats.UpdatedAt = time.Now()

	return nil
}

// Get retrieves value from cache
func (c *InMemoryCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		c.stats.MissCount++
		return nil, domain.ErrNotFound
	}

	// Check expiration
	if time.Now().After(item.expiresAt) {
		delete(c.data, key)
		c.stats.MissCount++
		return nil, domain.ErrNotFound
	}

	// Update hit count
	item.hitCount++
	c.stats.HitCount++
	c.updateHitRate()

	return item.value, nil
}

// Delete removes value from cache
func (c *InMemoryCacheService) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		c.stats.EvictionCount++
		c.stats.CurrentSize = c.calculateSize()
		c.stats.UpdatedAt = time.Now()
	}

	return nil
}

// Exists checks if key exists in cache
func (c *InMemoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	if time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// Clear removes all entries from cache
func (c *InMemoryCacheService) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = make(map[string]*cacheItem)
	c.stats.EvictionCount += int64(len(c.data))
	c.stats.CurrentSize = 0
	c.stats.UpdatedAt = time.Now()

	return nil
}

// GetStats returns cache statistics
func (c *InMemoryCacheService) GetStats(ctx context.Context) (*domain.CacheStats, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Clean expired entries
	now := time.Now()
	for key, item := range c.data {
		if now.After(item.expiresAt) {
			delete(c.data, key)
			c.stats.EvictionCount++
		}
	}

	c.stats.CurrentSize = c.calculateSize()
	c.updateHitRate()

	statsCopy := *c.stats
	return &statsCopy, nil
}

// calculateSize computes total size of cache in bytes
func (c *InMemoryCacheService) calculateSize() int64 {
	var totalSize int64
	for _, item := range c.data {
		totalSize += int64(len(item.value))
	}
	return totalSize
}

// updateHitRate calculates current hit rate
func (c *InMemoryCacheService) updateHitRate() {
	total := c.stats.HitCount + c.stats.MissCount
	if total > 0 {
		c.stats.HitRate = (float64(c.stats.HitCount) / float64(total)) * 100
	}
}

// PerformanceMonitorService implements port.PerformanceMonitor
type PerformanceMonitorService struct {
	metricsRepo  port.QueryMetricsRepository
	alertRepo    port.PerformanceAlertRepository
	cache        port.CacheService
	policy       *domain.PerformancePolicy
	log          *slog.Logger
}

// NewPerformanceMonitorService creates new performance monitor
func NewPerformanceMonitorService(
	metricsRepo port.QueryMetricsRepository,
	alertRepo port.PerformanceAlertRepository,
	cache port.CacheService,
	log *slog.Logger,
) *PerformanceMonitorService {
	if log == nil {
		log = slog.Default()
	}

	return &PerformanceMonitorService{
		metricsRepo: metricsRepo,
		alertRepo:   alertRepo,
		cache:       cache,
		policy: &domain.PerformancePolicy{
			SlowQueryThresholdMs: 100.0,
			CacheEnabled:         true,
			CacheTTLSeconds:      3600,
			RateLimitingEnabled:  true,
			DefaultRequestsPerMin: 60,
			ConnectionPoolSize:   20,
			QueryTimeoutMs:       30000,
		},
		log: log,
	}
}

// RecordQueryMetric records query performance
func (m *PerformanceMonitorService) RecordQueryMetric(ctx context.Context, query string, executionTimeMs float64, rowsAffected int64) error {
	metric := &domain.QueryMetric{
		ID:              uuid.New(),
		Query:           query,
		ExecutionTimeMs: executionTimeMs,
		RowsAffected:    rowsAffected,
		IsSlow:          m.CheckSlowQuery(ctx, executionTimeMs),
		SlowThresholdMs: m.policy.SlowQueryThresholdMs,
		CreatedAt:       time.Now(),
	}

	if err := m.metricsRepo.LogQueryMetric(ctx, metric); err != nil {
		m.log.Error("failed to log query metric", "error", err)
		return err
	}

	// Create alert if slow
	if metric.IsSlow && m.policy.EnableSlowQueryLog {
		alert := &domain.PerformanceAlert{
			ID:        uuid.New(),
			AlertType: "slow_query",
			Severity:  "warning",
			Message:   "Slow query detected",
			Details: map[string]interface{}{
				"query":            query,
				"execution_time_ms": executionTimeMs,
				"threshold_ms":     m.policy.SlowQueryThresholdMs,
			},
			Resolved:  false,
			CreatedAt: time.Now(),
		}
		_ = m.alertRepo.SaveAlert(ctx, alert)
	}

	return nil
}

// CheckSlowQuery determines if query is slow
func (m *PerformanceMonitorService) CheckSlowQuery(ctx context.Context, executionTimeMs float64) bool {
	return executionTimeMs > m.policy.SlowQueryThresholdMs
}

// CheckCacheHealth evaluates cache performance
func (m *PerformanceMonitorService) CheckCacheHealth(ctx context.Context) error {
	stats, err := m.cache.GetStats(ctx)
	if err != nil {
		return err
	}

	// Alert if hit rate too low
	if stats.HitRate < 50 && m.policy.CacheEnabled {
		alert := &domain.PerformanceAlert{
			ID:        uuid.New(),
			AlertType: "high_cache_miss",
			Severity:  "warning",
			Message:   "Cache hit rate below 50%",
			Details: map[string]interface{}{
				"hit_rate":   stats.HitRate,
				"hit_count":  stats.HitCount,
				"miss_count": stats.MissCount,
			},
			Resolved:  false,
			CreatedAt: time.Now(),
		}
		_ = m.alertRepo.SaveAlert(ctx, alert)
	}

	return nil
}

// CheckConnectionPool evaluates connection pool health (placeholder)
func (m *PerformanceMonitorService) CheckConnectionPool(ctx context.Context) error {
	return nil
}

// GetHealthStatus returns overall performance health
func (m *PerformanceMonitorService) GetHealthStatus(ctx context.Context) (map[string]interface{}, error) {
	cacheStats, _ := m.cache.GetStats(ctx)

	health := map[string]interface{}{
		"cache": map[string]interface{}{
			"hit_rate":      cacheStats.HitRate,
			"current_size":  cacheStats.CurrentSize,
			"max_size":      cacheStats.MaxSize,
			"hit_count":     cacheStats.HitCount,
			"miss_count":    cacheStats.MissCount,
			"eviction_count": cacheStats.EvictionCount,
		},
		"query": map[string]interface{}{
			"slow_threshold_ms": m.policy.SlowQueryThresholdMs,
		},
	}

	return health, nil
}

// GeneratePerformanceReport creates performance analysis report
func (m *PerformanceMonitorService) GeneratePerformanceReport(ctx context.Context, hoursBack int) (map[string]interface{}, error) {
	stats, _ := m.cache.GetStats(ctx)

	report := map[string]interface{}{
		"generated_at": time.Now(),
		"time_range":   map[string]interface{}{"hours_back": hoursBack},
		"cache_metrics": map[string]interface{}{
			"hit_rate":     stats.HitRate,
			"hit_count":    stats.HitCount,
			"miss_count":   stats.MissCount,
			"current_size": stats.CurrentSize,
		},
	}

	return report, nil
}
