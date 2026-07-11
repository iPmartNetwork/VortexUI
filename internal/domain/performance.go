package domain

import (
	"time"

	"github.com/google/uuid"
)

// ========== PHASE 3B: Performance Optimization Types ==========

// QueryMetric represents database query performance metrics
type QueryMetric struct {
	ID              uuid.UUID     `json:"id"`
	Query           string        `json:"query"`           // SQL query (hash or summary for PII)
	ExecutionTimeMs float64       `json:"execution_time_ms"`
	RowsAffected    int64         `json:"rows_affected"`
	RowsExamined    int64         `json:"rows_examined"`
	IndexUsed       string        `json:"index_used"`      // Index name if used
	IsSlow          bool          `json:"is_slow"`         // Flagged as slow query
	SlowThresholdMs float64       `json:"slow_threshold_ms"` // Threshold used
	CreatedAt       time.Time     `json:"created_at"`
}

// CacheEntry represents a cached value with metadata
type CacheEntry struct {
	ID        uuid.UUID `json:"id"`
	Key       string    `json:"key"`       // Cache key
	Value     []byte    `json:"value"`     // Serialized value
	ExpiresAt time.Time `json:"expires_at"` // TTL expiration
	HitCount  int64     `json:"hit_count"` // Number of times accessed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CacheStats represents cache performance statistics
type CacheStats struct {
	ID           uuid.UUID `json:"id"`
	CacheType    string    `json:"cache_type"`    // "session", "report", "compliance", "query"
	HitCount     int64     `json:"hit_count"`     // Successful cache hits
	MissCount    int64     `json:"miss_count"`    // Cache misses
	EvictionCount int64    `json:"eviction_count"` // Entries removed
	CurrentSize  int64     `json:"current_size"`  // Bytes in cache
	MaxSize      int64     `json:"max_size"`      // Max capacity
	HitRate      float64   `json:"hit_rate"`      // Percentage (0-100)
	UpdatedAt    time.Time `json:"updated_at"`
}

// RateLimitRule represents API rate limiting configuration
type RateLimitRule struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`           // Rule name
	Endpoint      string    `json:"endpoint"`       // API path or pattern
	Method        string    `json:"method"`         // HTTP method
	RequestsPerMin int      `json:"requests_per_min"` // Requests allowed per minute
	BurstSize     int       `json:"burst_size"`     // Burst allowance
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// RateLimitViolation represents detected rate limit violations
type RateLimitViolation struct {
	ID        uuid.UUID `json:"id"`
	RuleID    uuid.UUID `json:"rule_id"`
	ClientIP  string    `json:"client_ip"`
	Endpoint  string    `json:"endpoint"`
	RequestCount int    `json:"request_count"` // Requests made
	Limit     int       `json:"limit"`
	Action    string    `json:"action"`        // "throttle", "block", "alert"
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ConnectionPoolStats represents database connection pool metrics
type ConnectionPoolStats struct {
	ID           uuid.UUID `json:"id"`
	AcquiredConns int      `json:"acquired_conns"`    // Currently in use
	AvailableConns int     `json:"available_conns"`   // Available in pool
	ConstructingConns int  `json:"constructing_conns"` // Being created
	TotalConns   int       `json:"total_conns"`       // Total capacity
	WaitCount    int64     `json:"wait_count"`        // Times waited for conn
	ConstructingConnectionWaitCount int64 `json:"constructing_connection_wait_count"`
	MaxConns     int       `json:"max_conns"`         // Pool size limit
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PerformanceAlert represents performance issue alerts
type PerformanceAlert struct {
	ID        uuid.UUID `json:"id"`
	AlertType string    `json:"alert_type"` // "slow_query", "high_cache_miss", "pool_exhaustion", "rate_limit_violation"
	Severity  string    `json:"severity"`   // "info", "warning", "critical"
	Message   string    `json:"message"`
	Details   map[string]interface{} `json:"details"` // Alert-specific data
	Resolved  bool      `json:"resolved"`
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// IndexUsageStats represents database index usage metrics
type IndexUsageStats struct {
	ID            uuid.UUID `json:"id"`
	TableName     string    `json:"table_name"`
	IndexName     string    `json:"index_name"`
	ScanCount     int64     `json:"scan_count"`  // Times scanned
	TupleRead     int64     `json:"tuple_read"` // Rows read from index
	TuplesFetched int64     `json:"tuples_fetched"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// PerformancePolicy represents performance optimization settings
type PerformancePolicy struct {
	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	SlowQueryThresholdMs float64   `json:"slow_query_threshold_ms"` // ms
	CacheEnabled         bool      `json:"cache_enabled"`
	CacheTTLSeconds      int       `json:"cache_ttl_seconds"`
	MaxCacheSizeBytes    int64     `json:"max_cache_size_bytes"`
	RateLimitingEnabled  bool      `json:"rate_limiting_enabled"`
	DefaultRequestsPerMin int      `json:"default_requests_per_min"`
	ConnectionPoolSize   int       `json:"connection_pool_size"`
	QueryTimeoutMs       int       `json:"query_timeout_ms"`
	EnableSlowQueryLog   bool      `json:"enable_slow_query_log"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
