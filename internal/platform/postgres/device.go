package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// DeviceRepo implements port.DeviceRepository on PostgreSQL.
type DeviceRepo struct {
	pool *pgxpool.Pool
}

var _ port.DeviceRepository = (*DeviceRepo)(nil)

func (r *DeviceRepo) Upsert(ctx context.Context, d *domain.Device) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO user_devices (id, user_id, hwid, os, last_seen, created_at)
		VALUES ($1, $2, $3, $4, now(), now())
		ON CONFLICT (user_id, hwid) DO UPDATE SET os = $4, last_seen = now()`,
		d.ID, d.UserID, d.HWID, d.OS)
	return err
}

func (r *DeviceRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Device, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, hwid, os, last_seen, created_at
		FROM user_devices
		WHERE user_id = $1
		ORDER BY last_seen DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	return devices, rows.Err()
}

func (r *DeviceRepo) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM user_devices WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func (r *DeviceRepo) Delete(ctx context.Context, userID uuid.UUID, hwid string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_devices WHERE user_id = $1 AND hwid = $2`, userID, hwid)
	return err
}

func (r *DeviceRepo) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_devices WHERE user_id = $1`, userID)
	return err
}

func (r *DeviceRepo) DeleteAllForUsers(ctx context.Context, userIDs []uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM user_devices WHERE user_id = ANY($1)`, userIDs)
	return err
}

func (r *DeviceRepo) Exists(ctx context.Context, userID uuid.UUID, hwid string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM user_devices WHERE user_id = $1 AND hwid = $2)`,
		userID, hwid).Scan(&exists)
	return exists, err
}

// scanDevice scans a single row into a Device.
func scanDevice(rows pgx.Rows) (*domain.Device, error) {
	var d domain.Device
	if err := rows.Scan(&d.ID, &d.UserID, &d.HWID, &d.OS, &d.LastSeen, &d.CreatedAt); err != nil {
		return nil, err
	}
	return &d, nil
}
