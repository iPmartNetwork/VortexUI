package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// QueryMetricsRepository implements port.QueryMetricsRepository
type QueryMetricsRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewQueryMetricsRepository creates new query metrics repository
func NewQueryMetricsRepository(pool *pgxpool.Pool, log *slog.Logger) *QueryMetricsRepository {
	return &QueryMetricsRepository{pool: pool, log: log}
}

// LogQueryMetric records a query execution metric
func (r *QueryMetricsRepository) LogQueryMetric(ctx context.Context, metric *domain.QueryMetric) error {
	query := `
		INSERT INTO query_metrics (id, query, execution_time_ms, rows_affected, rows_examined, index_used, is_slow, slow_threshold_ms, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		metric.ID,
		metric.Query,
		metric.ExecutionTimeMs,
		metric.RowsAffected,
		metric.RowsExamined,
		metric.IndexUsed,
		metric.IsSlow,
		metric.SlowThresholdMs,
		metric.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to log query metric", "error", err)
		return fmt.Errorf("failed to log query metric: %w", err)
	}

	return nil
}

// GetQueryMetric retrieves a specific query metric
func (r *QueryMetricsRepository) GetQueryMetric(ctx context.Context, metricID uuid.UUID) (*domain.QueryMetric, error) {
	query := `
		SELECT id, query, execution_time_ms, rows_affected, rows_examined, index_used, is_slow, slow_threshold_ms, created_at
		FROM query_metrics WHERE id = $1
	`

	metric := &domain.QueryMetric{}
	err := r.pool.QueryRow(ctx, query, metricID).Scan(
		&metric.ID,
		&metric.Query,
		&metric.ExecutionTimeMs,
		&metric.RowsAffected,
		&metric.RowsExamined,
		&metric.IndexUsed,
		&metric.IsSlow,
		&metric.SlowThresholdMs,
		&metric.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get query metric", "error", err)
		return nil, fmt.Errorf("failed to get query metric: %w", err)
	}

	return metric, nil
}

// ListSlowQueries retrieves queries exceeding threshold
func (r *QueryMetricsRepository) ListSlowQueries(ctx context.Context, limit, offset int) ([]*domain.QueryMetric, error) {
	query := `
		SELECT id, query, execution_time_ms, rows_affected, rows_examined, index_used, is_slow, slow_threshold_ms, created_at
		FROM query_metrics WHERE is_slow = true
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.log.Error("failed to list slow queries", "error", err)
		return nil, fmt.Errorf("failed to list slow queries: %w", err)
	}
	defer rows.Close()

	var metrics []*domain.QueryMetric
	for rows.Next() {
		metric := &domain.QueryMetric{}
		if err := rows.Scan(
			&metric.ID,
			&metric.Query,
			&metric.ExecutionTimeMs,
			&metric.RowsAffected,
			&metric.RowsExamined,
			&metric.IndexUsed,
			&metric.IsSlow,
			&metric.SlowThresholdMs,
			&metric.CreatedAt,
		); err != nil {
			r.log.Error("failed to scan query metric", "error", err)
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetQueryMetricsByQuery retrieves metrics for specific query
func (r *QueryMetricsRepository) GetQueryMetricsByQuery(ctx context.Context, query string, limit, offset int) ([]*domain.QueryMetric, error) {
	sqlQuery := `
		SELECT id, query, execution_time_ms, rows_affected, rows_examined, index_used, is_slow, slow_threshold_ms, created_at
		FROM query_metrics WHERE query = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, sqlQuery, query, limit, offset)
	if err != nil {
		r.log.Error("failed to get query metrics", "error", err)
		return nil, fmt.Errorf("failed to get query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []*domain.QueryMetric
	for rows.Next() {
		metric := &domain.QueryMetric{}
		if err := rows.Scan(
			&metric.ID,
			&metric.Query,
			&metric.ExecutionTimeMs,
			&metric.RowsAffected,
			&metric.RowsExamined,
			&metric.IndexUsed,
			&metric.IsSlow,
			&metric.SlowThresholdMs,
			&metric.CreatedAt,
		); err != nil {
			continue
		}
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetQueryStats returns aggregated statistics
func (r *QueryMetricsRepository) GetQueryStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_queries,
			COUNT(CASE WHEN is_slow THEN 1 END) as slow_queries,
			AVG(execution_time_ms) as avg_time_ms,
			MAX(execution_time_ms) as max_time_ms,
			MIN(execution_time_ms) as min_time_ms
		FROM query_metrics
		WHERE created_at > NOW() - INTERVAL '1 hour'
	`

	var totalQueries, slowQueries int64
	var avgTime, maxTime, minTime float64

	err := r.pool.QueryRow(ctx, query).Scan(&totalQueries, &slowQueries, &avgTime, &maxTime, &minTime)
	if err != nil {
		r.log.Error("failed to get query stats", "error", err)
		return nil, fmt.Errorf("failed to get query stats: %w", err)
	}

	return map[string]interface{}{
		"total_queries": totalQueries,
		"slow_queries":  slowQueries,
		"avg_time_ms":   avgTime,
		"max_time_ms":   maxTime,
		"min_time_ms":   minTime,
	}, nil
}

// DeleteOldMetrics removes metrics older than days
func (r *QueryMetricsRepository) DeleteOldMetrics(ctx context.Context, daysOld int) error {
	query := `DELETE FROM query_metrics WHERE created_at < NOW() - INTERVAL '1 day' * $1`

	result, err := r.pool.Exec(ctx, query, daysOld)
	if err != nil {
		r.log.Error("failed to delete old metrics", "error", err)
		return fmt.Errorf("failed to delete old metrics: %w", err)
	}

	r.log.Info("deleted old query metrics", "count", result.RowsAffected())
	return nil
}

// RateLimitRepository implements port.RateLimitRepository
type RateLimitRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewRateLimitRepository creates new rate limit repository
func NewRateLimitRepository(pool *pgxpool.Pool, log *slog.Logger) *RateLimitRepository {
	return &RateLimitRepository{pool: pool, log: log}
}

// SaveRule creates or updates rate limit rule
func (r *RateLimitRepository) SaveRule(ctx context.Context, rule *domain.RateLimitRule) error {
	query := `
		INSERT INTO rate_limit_rules (id, name, endpoint, method, requests_per_min, burst_size, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (endpoint, method) DO UPDATE SET
			name = $2, requests_per_min = $5, burst_size = $6, enabled = $7, updated_at = $9
	`

	_, err := r.pool.Exec(ctx, query,
		rule.ID,
		rule.Name,
		rule.Endpoint,
		rule.Method,
		rule.RequestsPerMin,
		rule.BurstSize,
		rule.Enabled,
		rule.CreatedAt,
		rule.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save rate limit rule", "error", err)
		return fmt.Errorf("failed to save rate limit rule: %w", err)
	}

	return nil
}

// GetRule retrieves rate limit rule by ID
func (r *RateLimitRepository) GetRule(ctx context.Context, ruleID uuid.UUID) (*domain.RateLimitRule, error) {
	query := `
		SELECT id, name, endpoint, method, requests_per_min, burst_size, enabled, created_at, updated_at
		FROM rate_limit_rules WHERE id = $1
	`

	rule := &domain.RateLimitRule{}
	err := r.pool.QueryRow(ctx, query, ruleID).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Endpoint,
		&rule.Method,
		&rule.RequestsPerMin,
		&rule.BurstSize,
		&rule.Enabled,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get rate limit rule", "error", err)
		return nil, fmt.Errorf("failed to get rate limit rule: %w", err)
	}

	return rule, nil
}

// ListRules retrieves all active rate limit rules
func (r *RateLimitRepository) ListRules(ctx context.Context) ([]*domain.RateLimitRule, error) {
	query := `
		SELECT id, name, endpoint, method, requests_per_min, burst_size, enabled, created_at, updated_at
		FROM rate_limit_rules ORDER BY endpoint, method
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.log.Error("failed to list rate limit rules", "error", err)
		return nil, fmt.Errorf("failed to list rate limit rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.RateLimitRule
	for rows.Next() {
		rule := &domain.RateLimitRule{}
		if err := rows.Scan(
			&rule.ID,
			&rule.Name,
			&rule.Endpoint,
			&rule.Method,
			&rule.RequestsPerMin,
			&rule.BurstSize,
			&rule.Enabled,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			continue
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetRuleByEndpoint retrieves rule for specific endpoint
func (r *RateLimitRepository) GetRuleByEndpoint(ctx context.Context, endpoint, method string) (*domain.RateLimitRule, error) {
	query := `
		SELECT id, name, endpoint, method, requests_per_min, burst_size, enabled, created_at, updated_at
		FROM rate_limit_rules WHERE endpoint = $1 AND method = $2
	`

	rule := &domain.RateLimitRule{}
	err := r.pool.QueryRow(ctx, query, endpoint, method).Scan(
		&rule.ID,
		&rule.Name,
		&rule.Endpoint,
		&rule.Method,
		&rule.RequestsPerMin,
		&rule.BurstSize,
		&rule.Enabled,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get rate limit rule by endpoint", "error", err)
		return nil, fmt.Errorf("failed to get rate limit rule: %w", err)
	}

	return rule, nil
}

// DeleteRule removes a rate limit rule
func (r *RateLimitRepository) DeleteRule(ctx context.Context, ruleID uuid.UUID) error {
	query := `DELETE FROM rate_limit_rules WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, ruleID)
	if err != nil {
		r.log.Error("failed to delete rate limit rule", "error", err)
		return fmt.Errorf("failed to delete rate limit rule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// LogViolation records rate limit violation
func (r *RateLimitRepository) LogViolation(ctx context.Context, violation *domain.RateLimitViolation) error {
	query := `
		INSERT INTO rate_limit_violations (id, rule_id, client_ip, endpoint, request_count, rate_limit, action, blocked_until, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		violation.ID,
		violation.RuleID,
		violation.ClientIP,
		violation.Endpoint,
		violation.RequestCount,
		violation.Limit,
		violation.Action,
		violation.BlockedUntil,
		violation.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to log rate limit violation", "error", err)
		return fmt.Errorf("failed to log violation: %w", err)
	}

	return nil
}

// GetViolations retrieves recent violations for IP
func (r *RateLimitRepository) GetViolations(ctx context.Context, clientIP string, minutesBack int) ([]*domain.RateLimitViolation, error) {
	query := `
		SELECT id, rule_id, client_ip, endpoint, request_count, rate_limit, action, blocked_until, created_at
		FROM rate_limit_violations
		WHERE client_ip = $1 AND created_at > NOW() - INTERVAL '1 minute' * $2
		ORDER BY created_at DESC LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query, clientIP, minutesBack)
	if err != nil {
		r.log.Error("failed to get violations", "error", err)
		return nil, fmt.Errorf("failed to get violations: %w", err)
	}
	defer rows.Close()

	var violations []*domain.RateLimitViolation
	for rows.Next() {
		v := &domain.RateLimitViolation{}
		if err := rows.Scan(
			&v.ID,
			&v.RuleID,
			&v.ClientIP,
			&v.Endpoint,
			&v.RequestCount,
			&v.Limit,
			&v.Action,
			&v.BlockedUntil,
			&v.CreatedAt,
		); err != nil {
			continue
		}
		violations = append(violations, v)
	}

	return violations, nil
}

// IsClientBlocked checks if client is currently blocked
func (r *RateLimitRepository) IsClientBlocked(ctx context.Context, clientIP string) (bool, *time.Time, error) {
	query := `
		SELECT blocked_until FROM rate_limit_violations
		WHERE client_ip = $1 AND action = 'block' AND blocked_until > NOW()
		ORDER BY blocked_until DESC LIMIT 1
	`

	var blockedUntil *time.Time
	err := r.pool.QueryRow(ctx, query, clientIP).Scan(&blockedUntil)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to check client block status: %w", err)
	}

	return blockedUntil != nil && time.Now().Before(*blockedUntil), blockedUntil, nil
}

// PerformanceAlertRepository implements port.PerformanceAlertRepository
type PerformanceAlertRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewPerformanceAlertRepository creates new performance alert repository
func NewPerformanceAlertRepository(pool *pgxpool.Pool, log *slog.Logger) *PerformanceAlertRepository {
	return &PerformanceAlertRepository{pool: pool, log: log}
}

// SaveAlert creates performance alert
func (r *PerformanceAlertRepository) SaveAlert(ctx context.Context, alert *domain.PerformanceAlert) error {
	detailsJSON, _ := json.Marshal(alert.Details)

	query := `
		INSERT INTO performance_alerts (id, alert_type, severity, message, details, resolved, resolved_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		alert.ID,
		alert.AlertType,
		alert.Severity,
		alert.Message,
		detailsJSON,
		alert.Resolved,
		alert.ResolvedAt,
		alert.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to save performance alert", "error", err)
		return fmt.Errorf("failed to save alert: %w", err)
	}

	return nil
}

// GetAlert retrieves alert by ID
func (r *PerformanceAlertRepository) GetAlert(ctx context.Context, alertID uuid.UUID) (*domain.PerformanceAlert, error) {
	query := `
		SELECT id, alert_type, severity, message, details, resolved, resolved_at, created_at
		FROM performance_alerts WHERE id = $1
	`

	alert := &domain.PerformanceAlert{}
	var detailsJSON []byte

	err := r.pool.QueryRow(ctx, query, alertID).Scan(
		&alert.ID,
		&alert.AlertType,
		&alert.Severity,
		&alert.Message,
		&detailsJSON,
		&alert.Resolved,
		&alert.ResolvedAt,
		&alert.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get performance alert", "error", err)
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	if len(detailsJSON) > 0 {
		alert.Details = make(map[string]interface{})
		_ = json.Unmarshal(detailsJSON, &alert.Details)
	}

	return alert, nil
}

// ListActiveAlerts retrieves unresolved alerts
func (r *PerformanceAlertRepository) ListActiveAlerts(ctx context.Context) ([]*domain.PerformanceAlert, error) {
	query := `
		SELECT id, alert_type, severity, message, details, resolved, resolved_at, created_at
		FROM performance_alerts WHERE resolved = FALSE
		ORDER BY created_at DESC LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		r.log.Error("failed to list active alerts", "error", err)
		return nil, fmt.Errorf("failed to list active alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.PerformanceAlert
	for rows.Next() {
		alert := &domain.PerformanceAlert{}
		var detailsJSON []byte

		if err := rows.Scan(
			&alert.ID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Message,
			&detailsJSON,
			&alert.Resolved,
			&alert.ResolvedAt,
			&alert.CreatedAt,
		); err != nil {
			continue
		}

		if len(detailsJSON) > 0 {
			alert.Details = make(map[string]interface{})
			_ = json.Unmarshal(detailsJSON, &alert.Details)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// ListAlertsByType retrieves alerts of specific type
func (r *PerformanceAlertRepository) ListAlertsByType(ctx context.Context, alertType string) ([]*domain.PerformanceAlert, error) {
	query := `
		SELECT id, alert_type, severity, message, details, resolved, resolved_at, created_at
		FROM performance_alerts WHERE alert_type = $1
		ORDER BY created_at DESC LIMIT 100
	`

	rows, err := r.pool.Query(ctx, query, alertType)
	if err != nil {
		r.log.Error("failed to list alerts by type", "error", err)
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}
	defer rows.Close()

	var alerts []*domain.PerformanceAlert
	for rows.Next() {
		alert := &domain.PerformanceAlert{}
		var detailsJSON []byte

		if err := rows.Scan(
			&alert.ID,
			&alert.AlertType,
			&alert.Severity,
			&alert.Message,
			&detailsJSON,
			&alert.Resolved,
			&alert.ResolvedAt,
			&alert.CreatedAt,
		); err != nil {
			continue
		}

		if len(detailsJSON) > 0 {
			alert.Details = make(map[string]interface{})
			_ = json.Unmarshal(detailsJSON, &alert.Details)
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// ResolveAlert marks alert as resolved
func (r *PerformanceAlertRepository) ResolveAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `UPDATE performance_alerts SET resolved = TRUE, resolved_at = NOW() WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, alertID)
	if err != nil {
		r.log.Error("failed to resolve alert", "error", err)
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// DeleteAlert removes alert
func (r *PerformanceAlertRepository) DeleteAlert(ctx context.Context, alertID uuid.UUID) error {
	query := `DELETE FROM performance_alerts WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, alertID)
	if err != nil {
		r.log.Error("failed to delete alert", "error", err)
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
