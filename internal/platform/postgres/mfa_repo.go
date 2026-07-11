package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"

	"github.com/vortexui/vortexui/internal/domain"
)

// MFARepository implements the MFA repository using pgx
type MFARepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewMFARepository creates a new MFA repository
func NewMFARepository(pool *pgxpool.Pool, log *slog.Logger) *MFARepository {
	if log == nil {
		log = slog.Default()
	}
	return &MFARepository{
		pool: pool,
		log:  log,
	}
}

// SaveConfig saves a new MFA configuration
func (r *MFARepository) SaveConfig(ctx context.Context, config *domain.MFAConfig) error {
	query := `
		INSERT INTO mfa_configs (id, admin_id, totp_enabled, email_enabled, backup_codes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (admin_id) DO UPDATE
		SET totp_enabled = $3, email_enabled = $4, backup_codes = $5, updated_at = $7
	`

	backupCodesArray := pq.Array(config.BackupCodes)

	_, err := r.pool.Exec(ctx, query,
		config.ID, config.AdminID, config.TOTPEnabled, config.EmailEnabled,
		backupCodesArray, config.CreatedAt, config.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save MFA config", "admin_id", config.AdminID, "error", err)
		return fmt.Errorf("save MFA config: %w", err)
	}

	return nil
}

// GetConfig retrieves MFA configuration by admin ID
func (r *MFARepository) GetConfig(ctx context.Context, adminID uuid.UUID) (*domain.MFAConfig, error) {
	config := &domain.MFAConfig{}

	query := `
		SELECT id, admin_id, totp_enabled, email_enabled, backup_codes, created_at, updated_at
		FROM mfa_configs
		WHERE admin_id = $1
	`

	var backupCodesArray pq.StringArray

	err := r.pool.QueryRow(ctx, query, adminID).Scan(
		&config.ID, &config.AdminID, &config.TOTPEnabled, &config.EmailEnabled,
		&backupCodesArray, &config.CreatedAt, &config.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewDomainError(
				domain.ErrCodeNotFound,
				"MFA config not found",
				err,
			)
		}
		r.log.Error("failed to get MFA config", "admin_id", adminID, "error", err)
		return nil, fmt.Errorf("get MFA config: %w", err)
	}

	config.BackupCodes = []string(backupCodesArray)

	return config, nil
}

// UpdateConfig updates an existing MFA configuration
func (r *MFARepository) UpdateConfig(ctx context.Context, config *domain.MFAConfig) error {
	query := `
		UPDATE mfa_configs
		SET totp_enabled = $1, email_enabled = $2, backup_codes = $3, updated_at = $4
		WHERE admin_id = $5
	`

	backupCodesArray := pq.Array(config.BackupCodes)

	result, err := r.pool.Exec(ctx, query,
		config.TOTPEnabled, config.EmailEnabled, backupCodesArray, config.UpdatedAt, config.AdminID,
	)

	if err != nil {
		r.log.Error("failed to update MFA config", "admin_id", config.AdminID, "error", err)
		return fmt.Errorf("update MFA config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewDomainError(
			domain.ErrCodeNotFound,
			"MFA config not found",
			errors.New("no rows affected"),
		)
	}

	return nil
}
