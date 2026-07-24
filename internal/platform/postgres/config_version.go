package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ConfigVersionRepo implements port.ConfigVersionRepository on PostgreSQL.
type ConfigVersionRepo struct {
	pool *pgxpool.Pool
}

var _ port.ConfigVersionRepository = (*ConfigVersionRepo)(nil)

func NewConfigVersionRepo(pool *pgxpool.Pool) *ConfigVersionRepo {
	return &ConfigVersionRepo{pool: pool}
}

func (r *ConfigVersionRepo) Create(ctx context.Context, v *domain.ConfigVersion) error {
	configJSON, err := json.Marshal(v.ConfigData)
	if err != nil {
		return err
	}
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO config_versions (id, inbound_id, version, config_data, comment, admin_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		v.ID, v.InboundID, v.Version, configJSON, v.Comment, v.AdminID, v.CreatedAt)
	return err
}

func (r *ConfigVersionRepo) GetLatest(ctx context.Context, inboundID uuid.UUID) (*domain.ConfigVersion, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, inbound_id, version, config_data, comment, admin_id, created_at
		FROM config_versions
		WHERE inbound_id = $1
		ORDER BY version DESC
		LIMIT 1`, inboundID)
	return scanConfigVersion(row)
}

func (r *ConfigVersionRepo) GetByVersion(ctx context.Context, inboundID uuid.UUID, version int) (*domain.ConfigVersion, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, inbound_id, version, config_data, comment, admin_id, created_at
		FROM config_versions
		WHERE inbound_id = $1 AND version = $2`, inboundID, version)
	return scanConfigVersion(row)
}

func (r *ConfigVersionRepo) ListForInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.ConfigVersion, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, inbound_id, version, config_data, comment, admin_id, created_at
		FROM config_versions
		WHERE inbound_id = $1
		ORDER BY version DESC`, inboundID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectConfigVersions(rows)
}

func (r *ConfigVersionRepo) NextVersion(ctx context.Context, inboundID uuid.UUID) (int, error) {
	var maxVersion *int
	err := r.pool.QueryRow(ctx, `
		SELECT MAX(version) FROM config_versions WHERE inbound_id = $1`, inboundID).Scan(&maxVersion)
	if err != nil {
		return 0, err
	}
	if maxVersion == nil {
		return 1, nil
	}
	return *maxVersion + 1, nil
}

// --- helpers ---

func scanConfigVersion(row pgx.Row) (*domain.ConfigVersion, error) {
	var v domain.ConfigVersion
	var configJSON []byte
	if err := row.Scan(&v.ID, &v.InboundID, &v.Version, &configJSON, &v.Comment, &v.AdminID, &v.CreatedAt); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(configJSON, &v.ConfigData); err != nil {
		return nil, err
	}
	return &v, nil
}

func collectConfigVersions(rows pgx.Rows) ([]*domain.ConfigVersion, error) {
	var versions []*domain.ConfigVersion
	for rows.Next() {
		var v domain.ConfigVersion
		var configJSON []byte
		if err := rows.Scan(&v.ID, &v.InboundID, &v.Version, &configJSON, &v.Comment, &v.AdminID, &v.CreatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(configJSON, &v.ConfigData); err != nil {
			return nil, err
		}
		versions = append(versions, &v)
	}
	return versions, rows.Err()
}
