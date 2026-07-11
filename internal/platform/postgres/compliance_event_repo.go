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

// ComplianceEventRepository implements port.ComplianceEventRepository using PostgreSQL
type ComplianceEventRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewComplianceEventRepository creates new compliance event repository
func NewComplianceEventRepository(pool *pgxpool.Pool, log *slog.Logger) *ComplianceEventRepository {
	if log == nil {
		log = slog.Default()
	}
	return &ComplianceEventRepository{
		pool: pool,
		log:  log,
	}
}

// SaveEvent saves a compliance event
func (r *ComplianceEventRepository) SaveEvent(ctx context.Context, event *domain.ComplianceEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	event.UpdatedAt = time.Now()

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		r.log.Error("failed to marshal metadata", "error", err)
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO compliance_events (id, event_type, category, status, description, evidence, audit_event_id, admin_id, verified_at, expires_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			verified_at = EXCLUDED.verified_at,
			expires_at = EXCLUDED.expires_at,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.pool.Exec(ctx, query,
		event.ID,
		event.EventType,
		event.Category,
		event.Status,
		event.Description,
		event.Evidence,
		event.AuditEventID,
		event.AdminID,
		event.VerifiedAt,
		event.ExpiresAt,
		metadataJSON,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save compliance event", "error", err, "event_type", event.EventType)
		return fmt.Errorf("failed to save compliance event: %w", err)
	}

	return nil
}

// GetEvent retrieves a specific compliance event
func (r *ComplianceEventRepository) GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.ComplianceEvent, error) {
	var event domain.ComplianceEvent
	var metadataJSON []byte

	query := `
		SELECT id, event_type, category, status, description, evidence, audit_event_id, admin_id, verified_at, expires_at, metadata, created_at, updated_at
		FROM compliance_events
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.EventType,
		&event.Category,
		&event.Status,
		&event.Description,
		&event.Evidence,
		&event.AuditEventID,
		&event.AdminID,
		&event.VerifiedAt,
		&event.ExpiresAt,
		&metadataJSON,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get compliance event", "error", err, "event_id", eventID)
		return nil, fmt.Errorf("failed to get compliance event: %w", err)
	}

	event.Metadata = make(map[string]interface{})
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &event.Metadata)
	}

	return &event, nil
}

// ListEvents retrieves compliance events with filtering
func (r *ComplianceEventRepository) ListEvents(ctx context.Context, eventType *string, status *string, limit int, offset int) ([]*domain.ComplianceEvent, int64, error) {
	query := "SELECT id, event_type, category, status, description, evidence, audit_event_id, admin_id, verified_at, expires_at, metadata, created_at, updated_at FROM compliance_events WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM compliance_events WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if eventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, eventType)
		argCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.log.Error("failed to count compliance events", "error", err)
		return nil, 0, fmt.Errorf("failed to count compliance events: %w", err)
	}

	// Add pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to query compliance events", "error", err)
		return nil, 0, fmt.Errorf("failed to query compliance events: %w", err)
	}
	defer rows.Close()

	var events []*domain.ComplianceEvent
	for rows.Next() {
		var event domain.ComplianceEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Category,
			&event.Status,
			&event.Description,
			&event.Evidence,
			&event.AuditEventID,
			&event.AdminID,
			&event.VerifiedAt,
			&event.ExpiresAt,
			&metadataJSON,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan compliance event", "error", err)
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

// GetEventsByFramework retrieves events for specific compliance framework
func (r *ComplianceEventRepository) GetEventsByFramework(ctx context.Context, framework string, limit int, offset int) ([]*domain.ComplianceEvent, int64, error) {
	query := `
		SELECT id, event_type, category, status, description, evidence, audit_event_id, admin_id, verified_at, expires_at, metadata, created_at, updated_at
		FROM compliance_events
		WHERE event_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM compliance_events
		WHERE event_type = $1
	`

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, framework).Scan(&total)
	if err != nil {
		r.log.Error("failed to count compliance events by framework", "error", err)
		return nil, 0, fmt.Errorf("failed to count compliance events: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, framework, limit, offset)
	if err != nil {
		r.log.Error("failed to query compliance events by framework", "error", err)
		return nil, 0, fmt.Errorf("failed to query compliance events: %w", err)
	}
	defer rows.Close()

	var events []*domain.ComplianceEvent
	for rows.Next() {
		var event domain.ComplianceEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Category,
			&event.Status,
			&event.Description,
			&event.Evidence,
			&event.AuditEventID,
			&event.AdminID,
			&event.VerifiedAt,
			&event.ExpiresAt,
			&metadataJSON,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan compliance event", "error", err)
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

// UpdateEventStatus updates compliance event status
func (r *ComplianceEventRepository) UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string, verifiedBy *uuid.UUID) error {
	query := `
		UPDATE compliance_events
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, status, time.Now(), eventID)
	if err != nil {
		r.log.Error("failed to update compliance event status", "error", err, "event_id", eventID)
		return fmt.Errorf("failed to update compliance event status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// VerifyEvent marks event as verified/reviewed
func (r *ComplianceEventRepository) VerifyEvent(ctx context.Context, eventID uuid.UUID, verifiedBy uuid.UUID) error {
	query := `
		UPDATE compliance_events
		SET status = 'compliant', verified_at = $1, admin_id = $2, updated_at = $1
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, now, verifiedBy, eventID)
	if err != nil {
		r.log.Error("failed to verify compliance event", "error", err, "event_id", eventID)
		return fmt.Errorf("failed to verify compliance event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetExpiringEvents retrieves events expiring soon
func (r *ComplianceEventRepository) GetExpiringEvents(ctx context.Context, days int) ([]*domain.ComplianceEvent, error) {
	expiryDate := time.Now().AddDate(0, 0, days)

	query := `
		SELECT id, event_type, category, status, description, evidence, audit_event_id, admin_id, verified_at, expires_at, metadata, created_at, updated_at
		FROM compliance_events
		WHERE expires_at IS NOT NULL AND expires_at <= $1
		ORDER BY expires_at ASC
	`

	rows, err := r.pool.Query(ctx, query, expiryDate)
	if err != nil {
		r.log.Error("failed to query expiring compliance events", "error", err)
		return nil, fmt.Errorf("failed to query expiring events: %w", err)
	}
	defer rows.Close()

	var events []*domain.ComplianceEvent
	for rows.Next() {
		var event domain.ComplianceEvent
		var metadataJSON []byte

		err := rows.Scan(
			&event.ID,
			&event.EventType,
			&event.Category,
			&event.Status,
			&event.Description,
			&event.Evidence,
			&event.AuditEventID,
			&event.AdminID,
			&event.VerifiedAt,
			&event.ExpiresAt,
			&metadataJSON,
			&event.CreatedAt,
			&event.UpdatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan expiring event", "error", err)
			continue
		}

		event.Metadata = make(map[string]interface{})
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &event.Metadata)
		}

		events = append(events, &event)
	}

	return events, nil
}

// DeleteExpiredEvents deletes expired compliance events
func (r *ComplianceEventRepository) DeleteExpiredEvents(ctx context.Context) (int64, error) {
	query := "DELETE FROM compliance_events WHERE expires_at IS NOT NULL AND expires_at < $1"
	result, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		r.log.Error("failed to delete expired compliance events", "error", err)
		return 0, fmt.Errorf("failed to delete expired compliance events: %w", err)
	}

	return result.RowsAffected(), nil
}
