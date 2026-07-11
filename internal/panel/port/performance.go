package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// ========== PHASE 3B: Performance Optimization Ports ==========

// CacheService defines contract for caching operations
type CacheService interface {
	// Set stores value in cache with TTL
	Set(ctx context.Context, key string, value []byte, ttlSeconds int) error

	// Get retrieves value from cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Delete removes value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all entries from cache
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*domain.CacheStats, error)
}

// QueryMetricsRepository defines contract for query performance tracking
type QueryMetricsRepository interface {
	// LogQueryMetric records a query execution metric
	LogQueryMetric(ctx context.Context, metric *domain.QueryMetric) error

	// GetQueryMetric retrieves a specific query metric
	GetQueryMetric(ctx context.Context, metricID uuid.UUID) (*domain.QueryMetric, error)

	// ListSlowQueries retrieves queries exceeding threshold
	ListSlowQueries(ctx context.Context, limit, offset int) ([]*domain.QueryMetric, error)

	// GetQueryMetricsByQuery retrieves metrics for specific query
	GetQueryMetricsByQuery(ctx context.Context, query string, limit, offset int) ([]*domain.QueryMetric, error)

	// GetQueryStats returns aggregated statistics
	GetQueryStats(ctx context.Context) (map[string]interface{}, error)

	// DeleteOldMetrics removes metrics older than days
	DeleteOldMetrics(ctx context.Context, daysOld int) error
}

// RateLimitRepository defines contract for rate limiting storage
type RateLimitRepository interface {
	// SaveRule creates or updates rate limit rule
	SaveRule(ctx context.Context, rule *domain.RateLimitRule) error

	// GetRule retrieves rate limit rule by ID
	GetRule(ctx context.Context, ruleID uuid.UUID) (*domain.RateLimitRule, error)

	// ListRules retrieves all active rate limit rules
	ListRules(ctx context.Context) ([]*domain.RateLimitRule, error)

	// GetRuleByEndpoint retrieves rule for specific endpoint
	GetRuleByEndpoint(ctx context.Context, endpoint, method string) (*domain.RateLimitRule, error)

	// DeleteRule removes a rate limit rule
	DeleteRule(ctx context.Context, ruleID uuid.UUID) error

	// LogViolation records rate limit violation
	LogViolation(ctx context.Context, violation *domain.RateLimitViolation) error

	// GetViolations retrieves recent violations for IP
	GetViolations(ctx context.Context, clientIP string, minutesBack int) ([]*domain.RateLimitViolation, error)

	// IsClientBlocked checks if client is currently blocked
	IsClientBlocked(ctx context.Context, clientIP string) (bool, *time.Time, error)
}

// ConnectionPoolRepository defines contract for connection pool metrics
type ConnectionPoolRepository interface {
	// SaveStats records connection pool statistics
	SaveStats(ctx context.Context, stats *domain.ConnectionPoolStats) error

	// GetLatestStats retrieves most recent pool statistics
	GetLatestStats(ctx context.Context) (*domain.ConnectionPoolStats, error)

	// ListStats retrieves historical pool statistics
	ListStats(ctx context.Context, limit int) ([]*domain.ConnectionPoolStats, error)
}

// IndexUsageRepository defines contract for index usage tracking
type IndexUsageRepository interface {
	// LogIndexUsage records index usage
	LogIndexUsage(ctx context.Context, stats *domain.IndexUsageStats) error

	// GetIndexStats retrieves usage stats for index
	GetIndexStats(ctx context.Context, tableName, indexName string) (*domain.IndexUsageStats, error)

	// ListUnusedIndexes retrieves indexes not used recently
	ListUnusedIndexes(ctx context.Context, daysUnused int) ([]*domain.IndexUsageStats, error)

	// GetTableIndexes retrieves all indexes for table
	GetTableIndexes(ctx context.Context, tableName string) ([]*domain.IndexUsageStats, error)
}

// PerformanceAlertRepository defines contract for performance alerts
type PerformanceAlertRepository interface {
	// SaveAlert creates performance alert
	SaveAlert(ctx context.Context, alert *domain.PerformanceAlert) error

	// GetAlert retrieves alert by ID
	GetAlert(ctx context.Context, alertID uuid.UUID) (*domain.PerformanceAlert, error)

	// ListActiveAlerts retrieves unresolved alerts
	ListActiveAlerts(ctx context.Context) ([]*domain.PerformanceAlert, error)

	// ListAlertsByType retrieves alerts of specific type
	ListAlertsByType(ctx context.Context, alertType string) ([]*domain.PerformanceAlert, error)

	// ResolveAlert marks alert as resolved
	ResolveAlert(ctx context.Context, alertID uuid.UUID) error

	// DeleteAlert removes alert
	DeleteAlert(ctx context.Context, alertID uuid.UUID) error
}

// PerformanceMonitor defines contract for performance monitoring service
type PerformanceMonitor interface {
	// RecordQueryMetric records query performance
	RecordQueryMetric(ctx context.Context, query string, executionTimeMs float64, rowsAffected int64) error

	// CheckSlowQuery determines if query is slow
	CheckSlowQuery(ctx context.Context, executionTimeMs float64) bool

	// CheckCacheHealth evaluates cache performance
	CheckCacheHealth(ctx context.Context) error

	// CheckConnectionPool evaluates connection pool health
	CheckConnectionPool(ctx context.Context) error

	// GetHealthStatus returns overall performance health
	GetHealthStatus(ctx context.Context) (map[string]interface{}, error)

	// GeneratePerformanceReport creates performance analysis report
	GeneratePerformanceReport(ctx context.Context, hoursBack int) (map[string]interface{}, error)
}

// RateLimiter defines contract for rate limiting enforcement
type RateLimiter interface {
	// IsAllowed checks if request should be allowed
	IsAllowed(ctx context.Context, clientIP, endpoint, method string) (bool, error)

	// GetRemainingQuota returns remaining quota for client
	GetRemainingQuota(ctx context.Context, clientIP, endpoint, method string) (int, error)

	// BlockClient blocks client temporarily
	BlockClient(ctx context.Context, clientIP string, durationMinutes int) error

	// UnblockClient removes client block
	UnblockClient(ctx context.Context, clientIP string) error

	// IsClientBlocked checks if client is blocked
	IsClientBlocked(ctx context.Context, clientIP string) (bool, error)
}

// PerformancePolicyRepository defines contract for performance policy storage
type PerformancePolicyRepository interface {
	// GetPolicy retrieves current performance policy
	GetPolicy(ctx context.Context) (*domain.PerformancePolicy, error)

	// SavePolicy creates or updates performance policy
	SavePolicy(ctx context.Context, policy *domain.PerformancePolicy) error

	// GetDefaultPolicy returns system default policy
	GetDefaultPolicy(ctx context.Context) (*domain.PerformancePolicy, error)
}

// CacheEntryRepository defines contract for cache entry storage (optional, for persistent cache)
type CacheEntryRepository interface {
	// SaveEntry persists cache entry
	SaveEntry(ctx context.Context, entry *domain.CacheEntry) error

	// GetEntry retrieves cache entry
	GetEntry(ctx context.Context, key string) (*domain.CacheEntry, error)

	// DeleteEntry removes cache entry
	DeleteEntry(ctx context.Context, key string) error

	// ListExpiredEntries retrieves expired entries for cleanup
	ListExpiredEntries(ctx context.Context) ([]*domain.CacheEntry, error)

	// DeleteExpiredEntries removes all expired entries
	DeleteExpiredEntries(ctx context.Context) (int64, error)
}
