package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// IPWhitelistRepo implements port.IPWhitelistRepository on PostgreSQL.
type IPWhitelistRepo struct {
	pool *pgxpool.Pool
}

var _ port.IPWhitelistRepository = (*IPWhitelistRepo)(nil)

func NewIPWhitelistRepo(pool *pgxpool.Pool) *IPWhitelistRepo {
	return &IPWhitelistRepo{pool: pool}
}

func (r *IPWhitelistRepo) Create(ctx context.Context, entry *domain.AdminIPWhitelist) error {
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO admin_ip_whitelist (id, admin_id, cidr, description, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		entry.ID, entry.AdminID, entry.CIDR, entry.Description, entry.CreatedAt)
	return err
}

func (r *IPWhitelistRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM admin_ip_whitelist WHERE id = $1`, id)
	return err
}

func (r *IPWhitelistRepo) ListAll(ctx context.Context) ([]*domain.AdminIPWhitelist, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, cidr, description, created_at
		FROM admin_ip_whitelist
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AdminIPWhitelist
	for rows.Next() {
		var e domain.AdminIPWhitelist
		if err := rows.Scan(&e.ID, &e.AdminID, &e.CIDR, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}

func (r *IPWhitelistRepo) ListByAdmin(ctx context.Context, adminID uuid.UUID) ([]*domain.AdminIPWhitelist, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, cidr, description, created_at
		FROM admin_ip_whitelist
		WHERE admin_id = $1 OR admin_id IS NULL
		ORDER BY created_at DESC`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*domain.AdminIPWhitelist
	for rows.Next() {
		var e domain.AdminIPWhitelist
		if err := rows.Scan(&e.ID, &e.AdminID, &e.CIDR, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, &e)
	}
	return entries, rows.Err()
}
