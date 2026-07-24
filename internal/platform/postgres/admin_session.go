package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// AdminSessionRepo implements port.AdminSessionRepository on PostgreSQL.
type AdminSessionRepo struct {
	pool *pgxpool.Pool
}

var _ port.AdminSessionRepository = (*AdminSessionRepo)(nil)

func NewAdminSessionRepo(pool *pgxpool.Pool) *AdminSessionRepo {
	return &AdminSessionRepo{pool: pool}
}

func (r *AdminSessionRepo) Create(ctx context.Context, session *domain.SecuritySession) error {
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	session.LastActive = session.CreatedAt
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_sessions (id, admin_id, ip_address, user_agent, country, last_active, created_at, revoked)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		session.ID, session.AdminID, session.IPAddress, session.UserAgent,
		session.Country, session.LastActive, session.CreatedAt, session.Revoked)
	return err
}

func (r *AdminSessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SecuritySession, error) {
	var s domain.SecuritySession
	err := r.pool.QueryRow(ctx, `
		SELECT id, admin_id, ip_address, user_agent, country, last_active, created_at, revoked
		FROM admin_sessions
		WHERE id = $1`, id).Scan(
		&s.ID, &s.AdminID, &s.IPAddress, &s.UserAgent, &s.Country,
		&s.LastActive, &s.CreatedAt, &s.Revoked)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *AdminSessionRepo) ListByAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.SecuritySession, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, ip_address, user_agent, country, last_active, created_at, revoked
		FROM admin_sessions
		WHERE admin_id = $1
		ORDER BY last_active DESC`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.SecuritySession
	for rows.Next() {
		var s domain.SecuritySession
		if err := rows.Scan(&s.ID, &s.AdminID, &s.IPAddress, &s.UserAgent, &s.Country,
			&s.LastActive, &s.CreatedAt, &s.Revoked); err != nil {
			return nil, err
		}
		sessions = append(sessions, &s)
	}
	return sessions, rows.Err()
}

func (r *AdminSessionRepo) UpdateLastActive(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admin_sessions SET last_active = now() WHERE id = $1`, id)
	return err
}

func (r *AdminSessionRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admin_sessions SET revoked = true WHERE id = $1`, id)
	return err
}

func (r *AdminSessionRepo) RevokeAllForAdmin(ctx context.Context, adminID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE admin_sessions SET revoked = true WHERE admin_id = $1 AND revoked = false`, adminID)
	return err
}
