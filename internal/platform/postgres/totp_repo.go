package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// TOTPRepository implements the TOTP repository using pgx
type TOTPRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewTOTPRepository creates a new TOTP repository
func NewTOTPRepository(pool *pgxpool.Pool, log *slog.Logger) *TOTPRepository {
	if log == nil {
		log = slog.Default()
	}
	return &TOTPRepository{
		pool: pool,
		log:  log,
	}
}

// SaveSecret saves a new TOTP secret
func (r *TOTPRepository) SaveSecret(ctx context.Context, secret *domain.TOTPSecret) error {
	query := `
		INSERT INTO totp_secrets (id, admin_id, secret, qr_code, verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (admin_id) DO UPDATE
		SET secret = $3, qr_code = $4, verified = $5, updated_at = $7
	`

	_, err := r.pool.Exec(ctx, query,
		secret.ID, secret.AdminID, secret.Secret, secret.QRCode, secret.Verified,
		secret.CreatedAt, secret.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save TOTP secret", "admin_id", secret.AdminID, "error", err)
		return fmt.Errorf("save TOTP secret: %w", err)
	}

	return nil
}

// GetSecret retrieves a TOTP secret by admin ID
func (r *TOTPRepository) GetSecret(ctx context.Context, adminID uuid.UUID) (*domain.TOTPSecret, error) {
	secret := &domain.TOTPSecret{}

	query := `
		SELECT id, admin_id, secret, qr_code, verified, created_at, updated_at
		FROM totp_secrets
		WHERE admin_id = $1
	`

	err := r.pool.QueryRow(ctx, query, adminID).Scan(
		&secret.ID, &secret.AdminID, &secret.Secret, &secret.QRCode,
		&secret.Verified, &secret.CreatedAt, &secret.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.NewDomainError(
				domain.ErrCodeNotFound,
				"TOTP secret not found",
				err,
			)
		}
		r.log.Error("failed to get TOTP secret", "admin_id", adminID, "error", err)
		return nil, fmt.Errorf("get TOTP secret: %w", err)
	}

	return secret, nil
}

// VerifySecret marks a TOTP secret as verified
func (r *TOTPRepository) VerifySecret(ctx context.Context, adminID uuid.UUID) error {
	query := `
		UPDATE totp_secrets
		SET verified = TRUE, updated_at = NOW()
		WHERE admin_id = $1
	`

	result, err := r.pool.Exec(ctx, query, adminID)
	if err != nil {
		r.log.Error("failed to verify TOTP secret", "admin_id", adminID, "error", err)
		return fmt.Errorf("verify TOTP secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewDomainError(
			domain.ErrCodeNotFound,
			"TOTP secret not found",
			errors.New("no rows affected"),
		)
	}

	return nil
}

// DeleteSecret removes a TOTP secret
func (r *TOTPRepository) DeleteSecret(ctx context.Context, adminID uuid.UUID) error {
	query := `DELETE FROM totp_secrets WHERE admin_id = $1`

	result, err := r.pool.Exec(ctx, query, adminID)
	if err != nil {
		r.log.Error("failed to delete TOTP secret", "admin_id", adminID, "error", err)
		return fmt.Errorf("delete TOTP secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.NewDomainError(
			domain.ErrCodeNotFound,
			"TOTP secret not found",
			errors.New("no rows affected"),
		)
	}

	return nil
}
