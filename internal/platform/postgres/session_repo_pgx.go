package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// SessionRepoPgx implements port.SessionRepository with pgx support
type SessionRepoPgx struct {
	pool *pgxpool.Pool
}

// NewSessionRepoPgx creates a new session repository for pgx
func NewSessionRepoPgx(pool *pgxpool.Pool) *SessionRepoPgx {
	return &SessionRepoPgx{pool: pool}
}

// CreateSession creates a new session
func (r *SessionRepoPgx) CreateSession(ctx context.Context, session *domain.AdminSession) error {
	query := `
		INSERT INTO admin_sessions (
			admin_id, token_hash, ip_address, user_agent, last_activity, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
		RETURNING id, created_at
	`

	return r.pool.QueryRow(ctx, query,
		session.AdminID,
		session.TokenHash,
		session.IPAddress,
		session.UserAgent,
		session.LastActivity,
		session.ExpiresAt,
	).Scan(&session.ID, &session.CreatedAt)
}

// GetSession retrieves a session by token hash
func (r *SessionRepoPgx) GetSession(ctx context.Context, tokenHash string) (*domain.AdminSession, error) {
	query := `
		SELECT id, admin_id, token_hash, ip_address, user_agent, last_activity,
		       expires_at, created_at, revoked_at
		FROM admin_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
		LIMIT 1
	`

	session := &domain.AdminSession{}
	var revokedAt *time.Time

	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.AdminID,
		&session.TokenHash,
		&session.IPAddress,
		&session.UserAgent,
		&session.LastActivity,
		&session.ExpiresAt,
		&session.CreatedAt,
		&revokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionExpired
		}
		return nil, err
	}

	if revokedAt != nil {
		session.RevokedAt = revokedAt
	}

	return session, nil
}

// ListSessions lists sessions with optional filters
func (r *SessionRepoPgx) ListSessions(ctx context.Context, filter *domain.SessionFilter) ([]*domain.AdminSession, int64, error) {
	whereClause := "revoked_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if filter != nil {
		if filter.AdminID != nil && *filter.AdminID != uuid.Nil {
			whereClause += " AND admin_id = $" + string(rune(argIdx))
			args = append(args, *filter.AdminID)
			argIdx++
		}
		if filter.IPAddress != nil && *filter.IPAddress != "" {
			whereClause += " AND ip_address = $" + string(rune(argIdx))
			args = append(args, *filter.IPAddress)
			argIdx++
		}
		if filter.Active != nil && *filter.Active {
			whereClause += " AND expires_at > NOW()"
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM admin_sessions WHERE " + whereClause
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	limit, offset := 100, 0
	if filter != nil {
		if filter.Limit > 0 && filter.Limit <= 1000 {
			limit = filter.Limit
		}
		offset = filter.Offset
	}

	query := `
		SELECT id, admin_id, token_hash, ip_address, user_agent, last_activity,
		       expires_at, created_at, revoked_at
		FROM admin_sessions
		WHERE ` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + string(rune(argIdx)) + ` OFFSET $` + string(rune(argIdx+1))

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	sessions := []*domain.AdminSession{}
	for rows.Next() {
		session := &domain.AdminSession{}
		var revokedAt *time.Time

		if err := rows.Scan(
			&session.ID,
			&session.AdminID,
			&session.TokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.LastActivity,
			&session.ExpiresAt,
			&session.CreatedAt,
			&revokedAt,
		); err != nil {
			return nil, 0, err
		}

		if revokedAt != nil {
			session.RevokedAt = revokedAt
		}
		sessions = append(sessions, session)
	}

	return sessions, total, rows.Err()
}

// ListAdminSessions lists all active sessions for an admin
func (r *SessionRepoPgx) ListAdminSessions(ctx context.Context, adminID uuid.UUID, active bool) ([]*domain.AdminSession, error) {
	whereClause := "admin_id = $1"
	if active {
		whereClause += " AND revoked_at IS NULL AND expires_at > NOW()"
	}

	query := `
		SELECT id, admin_id, token_hash, ip_address, user_agent, last_activity,
		       expires_at, created_at, revoked_at
		FROM admin_sessions
		WHERE ` + whereClause + `
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := []*domain.AdminSession{}
	for rows.Next() {
		session := &domain.AdminSession{}
		var revokedAt *time.Time

		if err := rows.Scan(
			&session.ID,
			&session.AdminID,
			&session.TokenHash,
			&session.IPAddress,
			&session.UserAgent,
			&session.LastActivity,
			&session.ExpiresAt,
			&session.CreatedAt,
			&revokedAt,
		); err != nil {
			return nil, err
		}

		if revokedAt != nil {
			session.RevokedAt = revokedAt
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// UpdateSession updates session last activity
func (r *SessionRepoPgx) UpdateSession(ctx context.Context, session *domain.AdminSession) error {
	query := `
		UPDATE admin_sessions
		SET last_activity = $1
		WHERE id = $2 AND revoked_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, session.LastActivity, session.ID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// RevokeSession revokes a single session
func (r *SessionRepoPgx) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE admin_sessions
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, sessionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// RevokeAdminSessions revokes all sessions for an admin
func (r *SessionRepoPgx) RevokeAdminSessions(ctx context.Context, adminID uuid.UUID) error {
	query := `
		UPDATE admin_sessions
		SET revoked_at = NOW()
		WHERE admin_id = $1 AND revoked_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, adminID)
	return err
}

// DeleteExpiredSessions deletes sessions older than expiry
func (r *SessionRepoPgx) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM admin_sessions
		WHERE expires_at < NOW() OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '7 days')
	`

	result, err := r.pool.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}


