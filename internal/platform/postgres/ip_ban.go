package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// IPBanRepo implements port.IPBanRepository on PostgreSQL.
type IPBanRepo struct {
	pool *pgxpool.Pool
}

var _ port.IPBanRepository = (*IPBanRepo)(nil)

func NewIPBanRepo(pool *pgxpool.Pool) *IPBanRepo {
	return &IPBanRepo{pool: pool}
}

func (r *IPBanRepo) Create(ctx context.Context, ban *domain.IPBan) error {
	if ban.CreatedAt.IsZero() {
		ban.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO ip_ban_list (id, ip_address, reason, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (ip_address) DO UPDATE SET reason = $3, expires_at = $4`,
		ban.ID, ban.IPAddress, ban.Reason, ban.ExpiresAt, ban.CreatedAt)
	return err
}

func (r *IPBanRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM ip_ban_list WHERE id = $1`, id)
	return err
}

func (r *IPBanRepo) GetByIP(ctx context.Context, ip string) (*domain.IPBan, error) {
	var ban domain.IPBan
	err := r.pool.QueryRow(ctx, `
		SELECT id, ip_address, reason, expires_at, created_at
		FROM ip_ban_list
		WHERE ip_address = $1 AND (expires_at IS NULL OR expires_at > now())`, ip).Scan(
		&ban.ID, &ban.IPAddress, &ban.Reason, &ban.ExpiresAt, &ban.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ban, nil
}

func (r *IPBanRepo) ListActive(ctx context.Context) ([]*domain.IPBan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, ip_address, reason, expires_at, created_at
		FROM ip_ban_list
		WHERE expires_at IS NULL OR expires_at > now()
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bans []*domain.IPBan
	for rows.Next() {
		var b domain.IPBan
		if err := rows.Scan(&b.ID, &b.IPAddress, &b.Reason, &b.ExpiresAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		bans = append(bans, &b)
	}
	return bans, rows.Err()
}

func (r *IPBanRepo) DeleteExpired(ctx context.Context) (int, error) {
	tag, err := r.pool.Exec(ctx, `DELETE FROM ip_ban_list WHERE expires_at IS NOT NULL AND expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}
