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

// AuditEventRepository implements port.AuditEventRepository using PostgreSQL
type AuditEventRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewAuditEventRepository creates new audit event repository
func NewAuditEventRepository(pool *pgxpool.Pool, log *slog.Logger) *AuditEventRepository {
	if log == nil {
		log = slog.Default()
	}
	return &AuditEventRepository{
		pool: pool,
		log:  log,
	}
}

// LogEvent records a new audit event
func (r *AuditEventRepository) LogEvent(ctx context.Context, event *domain.AuditEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		r.log.Error("failed to marshal metadata", "error", err)
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_events (id, admin_id, event_type, severity, target_type, target_id, description, ip_address, user_agent, old_value, new_value, status, error_message, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	_, err = r.pool.Exec(ctx, query,
		event.ID,
		event.AdminID,
		event.EventType,
		event.Severity,
		event.TargetType,
		event.TargetID,
		event.Description,
		event.IPAddress,
		event.UserAgent,
		event.OldValue,
		event.NewValue,
		event.Status,
		event.ErrorMessage,
		metadataJSON,
		event.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to log audit event", "error", err, "event_type", event.EventType)
		return fmt.Errorf("failed to log audit event: %w", err)
	}

	return nil
}

// GetEvent retrieves a specific audit event by ID
func (r *AuditEventRepository) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.AuditEvent, error) {
	var event domain.AuditEvent
	var metadataJSON []byte

	query := `
		SELECT id, admin_id, event_type, severity, target_type, target_id, description, ip_address, user_agent, old_value, new_value, status, error_message, metadata, created_at
		FROM audit_events
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.AdminID,
		&event.EventType,
		&event.Severity,
		&event.TargetType,
		&event.TargetID,
		&event.Description,
		&event.IPAddress,
		&event.UserAgent,
		&event.OldValue,
		&event.NewValue,
		&event.Status,
		&event.ErrorMessage,
		&metadataJSON,
		&event.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get audit event", "error", err, "event_id", eventID)
		return nil, fmt.Errorf("failed to get audit event: %w", err)
	}

	event.Metadata = make(map[string]interface{})
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &event.Metadata)
	}

	return &event, nil
}

// ListEvents retrieves events with optional filtering
func (r *AuditEventRepository) ListEvents(ctx context.Context, adminID *uuid.UUID, eventType *string, severity *string, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]*domain.AuditEvent, int64, error) {
	query := "SELECT id, admin_id, event_type, severity, target_type, target_id, description, ip_address, user_agent, old_value, new_value, status, error_message, metadata, created_at FROM audit_events WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM audit_events WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if adminID != nil {
		query += fmt.Sprintf(" AND admin_id = $%d", argCount)
		countQuery += fmt.Sprintf(" AND admin_id = $%d", argCount)
		args = append(args, adminID)
		argCount++
	}

	if eventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, eventType)
		argCount++
	}

	if severity != nil {
		query += fmt.Sprintf(" AND severity = $%d", argCount)
		countQuery += fmt.Sprintf(" AND severity = $%d", argCount)
		args = append(args, severity)
		argCount++
	}

	if startDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startDate)
		argCount++
	}

	if endDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		countQuery += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endDate)
		argCount++
	}

	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.log.Error("failed to count audit events", "error", err)
		return nil, 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	// Add pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to query audit events", "error", err)
		return nil, 0, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []*domain.AuditEvent
	for rows.Next() {
		var event domain.AuditEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AdminID,
			&event.EventType,
			&event.Severity,
			&event.TargetType,
			&event.TargetID,
			&event.Description,
			&event.IPAddress,
			&event.UserAgent,
			&event.OldValue,
			&event.NewValue,
			&event.Status,
			&event.ErrorMessage,
			&metadataJSON,
			&event.CreatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan audit event", "error", err)
			continue
		}

		event.Metadata = make(map[string]interface{})
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &event.Metadata)
		}

		events = append(events, &event)
	}

	return events, total, nil
}

// SearchEvents performs full-text search on audit events
func (r *AuditEventRepository) SearchEvents(ctx context.Context, query string, limit int, offset int) ([]*domain.AuditEvent, int64, error) {
	searchQuery := `
		SELECT id, admin_id, event_type, severity, target_type, target_id, description, ip_address, user_agent, old_value, new_value, status, error_message, metadata, created_at
		FROM audit_events
		WHERE to_tsvector('english', description) @@ plainto_tsquery('english', $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM audit_events
		WHERE to_tsvector('english', description) @@ plainto_tsquery('english', $1)
	`

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, query).Scan(&total)
	if err != nil {
		r.log.Error("failed to count search results", "error", err)
		return nil, 0, fmt.Errorf("failed to count search results: %w", err)
	}

	rows, err := r.pool.Query(ctx, searchQuery, query, limit, offset)
	if err != nil {
		r.log.Error("failed to search audit events", "error", err)
		return nil, 0, fmt.Errorf("failed to search audit events: %w", err)
	}
	defer rows.Close()

	var events []*domain.AuditEvent
	for rows.Next() {
		var event domain.AuditEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.AdminID,
			&event.EventType,
			&event.Severity,
			&event.TargetType,
			&event.TargetID,
			&event.Description,
			&event.IPAddress,
			&event.UserAgent,
			&event.OldValue,
			&event.NewValue,
			&event.Status,
			&event.ErrorMessage,
			&metadataJSON,
			&event.CreatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan search result", "error", err)
			continue
		}

		event.Metadata = make(map[string]interface{})
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &event.Metadata)
		}

		events = append(events, &event)
	}

	return events, total, nil
}

// GetEventsByAdmin retrieves all events for a specific admin
func (r *AuditEventRepository) GetEventsByAdmin(ctx context.Context, adminID uuid.UUID, limit int, offset int) ([]*domain.AuditEvent, int64, error) {
	return r.ListEvents(ctx, &adminID, nil, nil, nil, nil, limit, offset)
}

// GetEventsByType retrieves all events of a specific type
func (r *AuditEventRepository) GetEventsByType(ctx context.Context, eventType domain.AuditEventType, limit int, offset int) ([]*domain.AuditEvent, int64, error) {
	typeStr := string(eventType)
	return r.ListEvents(ctx, nil, &typeStr, nil, nil, nil, limit, offset)
}

// GetEventsBySeverity retrieves all events with a specific severity
func (r *AuditEventRepository) GetEventsBySeverity(ctx context.Context, severity domain.AuditEventSeverity, limit int, offset int) ([]*domain.AuditEvent, int64, error) {
	sevStr := string(severity)
	return r.ListEvents(ctx, nil, nil, &sevStr, nil, nil, limit, offset)
}

// DeleteOldEvents deletes events older than retentionDays
func (r *AuditEventRepository) DeleteOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := "DELETE FROM audit_events WHERE created_at < $1"
	result, err := r.pool.Exec(ctx, query, cutoffDate)
	if err != nil {
		r.log.Error("failed to delete old audit events", "error", err, "retention_days", retentionDays)
		return 0, fmt.Errorf("failed to delete old audit events: %w", err)
	}

	return result.RowsAffected(), nil
}

// CountEvents returns total count of audit events
func (r *AuditEventRepository) CountEvents(ctx context.Context) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM audit_events"

	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		r.log.Error("failed to count audit events", "error", err)
		return 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	return count, nil
}
