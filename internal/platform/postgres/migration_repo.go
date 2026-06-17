package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// MigrationRepo implements port.MigrationRepository on PostgreSQL.
type MigrationRepo struct {
	pool *pgxpool.Pool
}

var _ port.MigrationRepository = (*MigrationRepo)(nil)

func (r *MigrationRepo) SaveEvent(ctx context.Context, e *domain.MigrationEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO migration_events (id, user_id, username, from_node_id, to_node_id, reason, status, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		e.ID, e.UserID, e.Username, e.FromNodeID, e.ToNodeID, e.Reason, e.Status, e.Error, e.CreatedAt)
	return err
}

func (r *MigrationRepo) ListEvents(ctx context.Context, limit, offset int) ([]*domain.MigrationEvent, int, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, username, from_node_id, to_node_id, reason, status, error, created_at
		FROM migration_events ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var out []*domain.MigrationEvent
	for rows.Next() {
		var e domain.MigrationEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &e.FromNodeID, &e.ToNodeID, &e.Reason, &e.Status, &e.Error, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, &e)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	_ = r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM migration_events`).Scan(&total)
	return out, total, nil
}

func (r *MigrationRepo) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.MigrationEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, username, from_node_id, to_node_id, reason, status, error, created_at
		FROM migration_events WHERE from_node_id = $1 OR to_node_id = $1
		ORDER BY created_at DESC`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.MigrationEvent
	for rows.Next() {
		var e domain.MigrationEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.Username, &e.FromNodeID, &e.ToNodeID, &e.Reason, &e.Status, &e.Error, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &e)
	}
	return out, rows.Err()
}

func (r *MigrationRepo) GetPolicy(ctx context.Context) (*domain.MigrationPolicy, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT enabled, health_check_interval, unhealthy_threshold, cpu_threshold, mem_threshold, packet_loss_max, migrate_back
		FROM migration_policy WHERE id = 1`)
	var p domain.MigrationPolicy
	err := row.Scan(&p.Enabled, &p.HealthCheckInterval, &p.UnhealthyThreshold, &p.CPUThreshold, &p.MemThreshold, &p.PacketLossMax, &p.MigrateBack)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *MigrationRepo) SavePolicy(ctx context.Context, p *domain.MigrationPolicy) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO migration_policy (id, enabled, health_check_interval, unhealthy_threshold, cpu_threshold, mem_threshold, packet_loss_max, migrate_back)
		VALUES (1, $1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			enabled = EXCLUDED.enabled,
			health_check_interval = EXCLUDED.health_check_interval,
			unhealthy_threshold = EXCLUDED.unhealthy_threshold,
			cpu_threshold = EXCLUDED.cpu_threshold,
			mem_threshold = EXCLUDED.mem_threshold,
			packet_loss_max = EXCLUDED.packet_loss_max,
			migrate_back = EXCLUDED.migrate_back`,
		p.Enabled, p.HealthCheckInterval, p.UnhealthyThreshold, p.CPUThreshold, p.MemThreshold, p.PacketLossMax, p.MigrateBack)
	return err
}
