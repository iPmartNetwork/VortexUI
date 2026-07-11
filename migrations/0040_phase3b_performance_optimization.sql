-- PHASE 3B: Performance Optimization Schema
-- Migration: 0040_phase3b_performance_optimization.sql

-- Query Metrics Table
CREATE TABLE IF NOT EXISTS query_metrics (
    id UUID PRIMARY KEY,
    query TEXT NOT NULL,
    execution_time_ms FLOAT NOT NULL,
    rows_affected BIGINT DEFAULT 0,
    rows_examined BIGINT DEFAULT 0,
    index_used VARCHAR(255),
    is_slow BOOLEAN DEFAULT FALSE,
    slow_threshold_ms FLOAT DEFAULT 100.0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_query_metrics_is_slow ON query_metrics(is_slow, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_query_metrics_created_at ON query_metrics(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_query_metrics_query ON query_metrics(query);

-- Rate Limit Rules Table
CREATE TABLE IF NOT EXISTS rate_limit_rules (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    requests_per_min INTEGER NOT NULL DEFAULT 60,
    burst_size INTEGER NOT NULL DEFAULT 10,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(endpoint, method)
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_enabled ON rate_limit_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_rate_limit_rules_endpoint ON rate_limit_rules(endpoint, method);

-- Rate Limit Violations Table
CREATE TABLE IF NOT EXISTS rate_limit_violations (
    id UUID PRIMARY KEY,
    rule_id UUID REFERENCES rate_limit_rules(id) ON DELETE SET NULL,
    client_ip INET NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INTEGER NOT NULL,
    limit INTEGER NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('throttle', 'block', 'alert')),
    blocked_until TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rate_limit_violations_client_ip ON rate_limit_violations(client_ip);
CREATE INDEX IF NOT EXISTS idx_rate_limit_violations_created_at ON rate_limit_violations(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rate_limit_violations_blocked_until ON rate_limit_violations(blocked_until);

-- Connection Pool Stats Table
CREATE TABLE IF NOT EXISTS connection_pool_stats (
    id UUID PRIMARY KEY,
    acquired_conns INTEGER DEFAULT 0,
    available_conns INTEGER DEFAULT 0,
    constructing_conns INTEGER DEFAULT 0,
    total_conns INTEGER DEFAULT 0,
    wait_count BIGINT DEFAULT 0,
    constructing_connection_wait_count BIGINT DEFAULT 0,
    max_conns INTEGER DEFAULT 20,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_connection_pool_stats_created_at ON connection_pool_stats(created_at DESC);

-- Index Usage Stats Table
CREATE TABLE IF NOT EXISTS index_usage_stats (
    id UUID PRIMARY KEY,
    table_name VARCHAR(255) NOT NULL,
    index_name VARCHAR(255) NOT NULL,
    scan_count BIGINT DEFAULT 0,
    tuple_read BIGINT DEFAULT 0,
    tuples_fetched BIGINT DEFAULT 0,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(table_name, index_name)
);

CREATE INDEX IF NOT EXISTS idx_index_usage_stats_table ON index_usage_stats(table_name);
CREATE INDEX IF NOT EXISTS idx_index_usage_stats_last_used ON index_usage_stats(last_used_at);

-- Performance Alerts Table
CREATE TABLE IF NOT EXISTS performance_alerts (
    id UUID PRIMARY KEY,
    alert_type VARCHAR(100) NOT NULL CHECK (alert_type IN ('slow_query', 'high_cache_miss', 'pool_exhaustion', 'rate_limit_violation')),
    severity VARCHAR(50) NOT NULL CHECK (severity IN ('info', 'warning', 'critical')),
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_performance_alerts_alert_type ON performance_alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_severity ON performance_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_resolved ON performance_alerts(resolved);
CREATE INDEX IF NOT EXISTS idx_performance_alerts_created_at ON performance_alerts(created_at DESC);

-- Performance Policy Table
CREATE TABLE IF NOT EXISTS performance_policies (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    slow_query_threshold_ms FLOAT DEFAULT 100.0,
    cache_enabled BOOLEAN DEFAULT TRUE,
    cache_ttl_seconds INTEGER DEFAULT 3600,
    max_cache_size_bytes BIGINT DEFAULT 104857600,  -- 100MB
    rate_limiting_enabled BOOLEAN DEFAULT TRUE,
    default_requests_per_min INTEGER DEFAULT 60,
    connection_pool_size INTEGER DEFAULT 20,
    query_timeout_ms INTEGER DEFAULT 30000,
    enable_slow_query_log BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Default Performance Policy
INSERT INTO performance_policies (
    id, name, description, slow_query_threshold_ms, cache_enabled, cache_ttl_seconds,
    max_cache_size_bytes, rate_limiting_enabled, default_requests_per_min,
    connection_pool_size, query_timeout_ms, enable_slow_query_log, created_at, updated_at
) VALUES (
    '00000000-0000-0000-0000-000000000001'::UUID,
    'Default Policy',
    'System default performance policy',
    100.0,
    TRUE,
    3600,
    104857600,
    TRUE,
    60,
    20,
    30000,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT DO NOTHING;

-- Cache Entries Table (for persistent cache, optional)
CREATE TABLE IF NOT EXISTS cache_entries (
    id UUID PRIMARY KEY,
    key VARCHAR(512) NOT NULL UNIQUE,
    value BYTEA NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    hit_count BIGINT DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_cache_entries_expires_at ON cache_entries(expires_at);
CREATE INDEX IF NOT EXISTS idx_cache_entries_key ON cache_entries(key);

-- Default Indexes for Audit Queries (Query Optimization)
CREATE INDEX IF NOT EXISTS idx_audit_events_admin_created ON audit_events(admin_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_event_created ON audit_events(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_compliance_events_status_created ON compliance_events(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_reports_created_by_created ON audit_reports(created_by, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_reports_period ON audit_reports(start_date, end_date);

-- Auto-update trigger for updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to performance tables
DROP TRIGGER IF EXISTS trigger_connection_pool_stats_updated_at ON connection_pool_stats;
CREATE TRIGGER trigger_connection_pool_stats_updated_at
BEFORE UPDATE ON connection_pool_stats
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_index_usage_stats_updated_at ON index_usage_stats;
CREATE TRIGGER trigger_index_usage_stats_updated_at
BEFORE UPDATE ON index_usage_stats
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_performance_policies_updated_at ON performance_policies;
CREATE TRIGGER trigger_performance_policies_updated_at
BEFORE UPDATE ON performance_policies
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_cache_entries_updated_at ON cache_entries;
CREATE TRIGGER trigger_cache_entries_updated_at
BEFORE UPDATE ON cache_entries
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS trigger_rate_limit_rules_updated_at ON rate_limit_rules;
CREATE TRIGGER trigger_rate_limit_rules_updated_at
BEFORE UPDATE ON rate_limit_rules
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
