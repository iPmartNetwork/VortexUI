package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// LoginAuditRepo implements port.LoginAuditRepository on PostgreSQL.
type LoginAuditRepo struct {
	pool *pgxpool.Pool
}

var _ port.LoginAuditRepository = (*LoginAuditRepo)(nil)

func NewLoginAuditRepo(pool *pgxpool.Pool) *LoginAuditRepo {
	return &LoginAuditRepo{pool: pool}
}

func (r *LoginAuditRepo) Create(ctx context.Context, entry *domain.LoginAuditEntry) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO login_audit_log (id, admin_id, username, ip_address, user_agent, country, success, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		entry.ID, entry.AdminID, entry.Username, entry.IPAddress,
		entry.UserAgent, entry.Country, entry.Success, entry.FailureReason, entry.CreatedAt)
	return err
}

func (r *LoginAuditRepo) ListByAdmin(ctx context.Context, adminID uuid.UUID, limit int) ([]*domain.LoginAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, username, ip_address, user_agent, country, success, failure_reason, created_at
		FROM login_audit_log
		WHERE admin_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, adminID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoginAuditEntries(rows)
}

func (r *LoginAuditRepo) ListByIP(ctx context.Context, ip string, limit int) ([]*domain.LoginAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, username, ip_address, user_agent, country, success, failure_reason, created_at
		FROM login_audit_log
		WHERE ip_address = $1
		ORDER BY created_at DESC
		LIMIT $2`, ip, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoginAuditEntries(rows)
}

func (r *LoginAuditRepo) ListRecent(ctx context.Context, limit int) ([]*domain.LoginAuditEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, username, ip_address, user_agent, country, success, failure_reason, created_at
		FROM login_audit_log
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLoginAuditEntries(rows)
}

func (r *LoginAuditRepo) CountFailedSince(ctx context.Context, adminID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM login_audit_log
		WHERE admin_id = $1 AND success = false AND created_at >= $2`, adminID, since).Scan(&count)
	return count, err
}

func scanLoginAuditEntries(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]*domain.LoginAuditEntry, error) {
	var entries []*domain.LoginAuditEntry
	for rows.Next() {
		var e domain.LoginAuditEntry
		if err := rows.Scan(&e.ID, &e.AdminID, &e.Username, &e.IPAddress,
			&e.UserAgent, &e.Country, &e.Success, &e.FailureReason, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
