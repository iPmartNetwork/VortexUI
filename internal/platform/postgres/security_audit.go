package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SecurityAuditRepo implements port.SecurityAuditRepository on PostgreSQL.
type SecurityAuditRepo struct {
	pool *pgxpool.Pool
}

var _ port.SecurityAuditRepository = (*SecurityAuditRepo)(nil)

func NewSecurityAuditRepo(pool *pgxpool.Pool) *SecurityAuditRepo {
	return &SecurityAuditRepo{pool: pool}
}

func (r *SecurityAuditRepo) Create(ctx context.Context, entry *domain.SecurityAuditEntry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	beforeJSON, _ := json.Marshal(entry.BeforeState)
	afterJSON, _ := json.Marshal(entry.AfterState)

	_, err := r.pool.Exec(ctx, `
		INSERT INTO security_audit_log (id, admin_id, operation, resource, before_state, after_state, ip_address, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		entry.ID, entry.AdminID, entry.Operation, entry.Resource,
		beforeJSON, afterJSON, entry.IPAddress, entry.CreatedAt)
	return err
}

func (r *SecurityAuditRepo) ListByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.SecurityAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, operation, resource, before_state, after_state, ip_address, created_at
		FROM security_audit_log
		WHERE admin_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, adminID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecurityAuditEntries(rows)
}

func (r *SecurityAuditRepo) ListByOperation(ctx context.Context, operation string, limit int) ([]*domain.SecurityAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, operation, resource, before_state, after_state, ip_address, created_at
		FROM security_audit_log
		WHERE operation = $1
		ORDER BY created_at DESC
		LIMIT $2`, operation, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecurityAuditEntries(rows)
}

func (r *SecurityAuditRepo) ListRecent(ctx context.Context, limit int) ([]*domain.SecurityAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, operation, resource, before_state, after_state, ip_address, created_at
		FROM security_audit_log
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSecurityAuditEntries(rows)
}

func scanSecurityAuditEntries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.SecurityAuditEntry, error) {
	var entries []*domain.SecurityAuditEntry
	for rows.Next() {
		var e domain.SecurityAuditEntry
		var beforeJSON, afterJSON []byte
		if err := rows.Scan(&e.ID, &e.AdminID, &e.Operation, &e.Resource,
			&beforeJSON, &afterJSON, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, err
		}
		if len(beforeJSON) > 0 {
			_ = json.Unmarshal(beforeJSON, &e.BeforeState)
		}
		if len(afterJSON) > 0 {
			_ = json.Unmarshal(afterJSON, &e.AfterState)
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
