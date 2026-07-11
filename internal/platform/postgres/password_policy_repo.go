package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// PasswordPolicyRepository implements password policy repository using pgx
type PasswordPolicyRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewPasswordPolicyRepository creates a new password policy repository
func NewPasswordPolicyRepository(pool *pgxpool.Pool, log *slog.Logger) *PasswordPolicyRepository {
	if log == nil {
		log = slog.Default()
	}
	return &PasswordPolicyRepository{
		pool: pool,
		log:  log,
	}
}

// GetPolicy retrieves the active password policy
func (r *PasswordPolicyRepository) GetPolicy(ctx context.Context) (*domain.PasswordPolicy, error) {
	policy := &domain.PasswordPolicy{}

	query := `
		SELECT id, min_length, require_uppercase, require_lowercase, require_numbers,
		       require_special_chars, expiration_days, history_count, failed_attempts_limit, lockout_duration_mins
		FROM password_policies
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query).Scan(
		&policy.ID, &policy.MinLength, &policy.RequireUppercase, &policy.RequireLowercase,
		&policy.RequireNumbers, &policy.RequireSpecialChars, &policy.ExpirationDays,
		&policy.HistoryCount, &policy.FailedAttemptsLimit, &policy.LockoutDurationMins,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return default policy if none exists
			return &domain.PasswordPolicy{
				MinLength:               12,
				RequireUppercase:        true,
				RequireLowercase:        true,
				RequireNumbers:          true,
				RequireSpecialChars:     true,
				ExpirationDays:          90,
				HistoryCount:            5,
				FailedAttemptsLimit:     5,
				LockoutDurationMins:     30,
			}, nil
		}
		r.log.Error("failed to get password policy", "error", err)
		return nil, fmt.Errorf("get password policy: %w", err)
	}

	return policy, nil
}

// UpdatePolicy updates the password policy
func (r *PasswordPolicyRepository) UpdatePolicy(ctx context.Context, policy *domain.PasswordPolicy) error {
	query := `
		UPDATE password_policies
		SET min_length = $1, require_uppercase = $2, require_lowercase = $3,
		    require_numbers = $4, require_special_chars = $5, expiration_days = $6,
		    history_count = $7, failed_attempts_limit = $8, lockout_duration_mins = $9,
		    updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		policy.MinLength, policy.RequireUppercase, policy.RequireLowercase,
		policy.RequireNumbers, policy.RequireSpecialChars, policy.ExpirationDays,
		policy.HistoryCount, policy.FailedAttemptsLimit, policy.LockoutDurationMins,
	)

	if err != nil {
		r.log.Error("failed to update password policy", "error", err)
		return fmt.Errorf("update password policy: %w", err)
	}

	return nil
}

// SaveHistory saves a password to history
func (r *PasswordPolicyRepository) SaveHistory(ctx context.Context, adminID uuid.UUID, passwordHash string) error {
	query := `
		INSERT INTO password_history (id, admin_id, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, uuid.New(), adminID, passwordHash, time.Now())
	if err != nil {
		r.log.Error("failed to save password history", "admin_id", adminID, "error", err)
		return fmt.Errorf("save password history: %w", err)
	}

	return nil
}

// GetHistory retrieves password history for an admin (most recent entries)
func (r *PasswordPolicyRepository) GetHistory(ctx context.Context, adminID uuid.UUID, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}

	query := `
		SELECT password_hash
		FROM password_history
		WHERE admin_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, adminID, limit)
	if err != nil {
		r.log.Error("failed to get password history", "admin_id", adminID, "error", err)
		return nil, fmt.Errorf("get password history: %w", err)
	}
	defer rows.Close()

	hashes := []string{}
	for rows.Next() {
		var hash string
		if err := rows.Scan(&hash); err != nil {
			r.log.Error("failed to scan password hash", "error", err)
			continue
		}
		hashes = append(hashes, hash)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("error reading password history rows", "error", err)
		return nil, err
	}

	return hashes, nil
}

// GetPasswordStatus retrieves admin password status
func (r *PasswordPolicyRepository) GetPasswordStatus(ctx context.Context, adminID uuid.UUID) (*domain.AdminPasswordStatus, error) {
	status := &domain.AdminPasswordStatus{}

	query := `
		SELECT admin_id, last_changed_at, expires_at, failed_attempts, locked_until, must_change_password
		FROM admin_password_status
		WHERE admin_id = $1
	`

	err := r.pool.QueryRow(ctx, query, adminID).Scan(
		&status.AdminID, &status.LastChangedAt, &status.ExpiresAt, &status.FailedAttempts,
		&status.LockedUntil, &status.MustChangePassword,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Initialize new status
			return &domain.AdminPasswordStatus{
				AdminID:              adminID,
				LastChangedAt:        time.Now(),
				ExpiresAt:            time.Now().AddDate(0, 0, 90),
				FailedAttempts:       0,
				LockedUntil:          time.Time{},
				MustChangePassword:   false,
			}, nil
		}
		r.log.Error("failed to get password status", "admin_id", adminID, "error", err)
		return nil, fmt.Errorf("get password status: %w", err)
	}

	return status, nil
}

// UpdatePasswordStatus updates admin password status
func (r *PasswordPolicyRepository) UpdatePasswordStatus(ctx context.Context, status *domain.AdminPasswordStatus) error {
	query := `
		INSERT INTO admin_password_status (admin_id, last_changed_at, expires_at, failed_attempts, locked_until, must_change_password, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
		ON CONFLICT (admin_id) DO UPDATE
		SET last_changed_at = $2, expires_at = $3, failed_attempts = $4, locked_until = $5, must_change_password = $6, updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		status.AdminID, status.LastChangedAt, status.ExpiresAt, status.FailedAttempts,
		status.LockedUntil, status.MustChangePassword,
	)

	if err != nil {
		r.log.Error("failed to update password status", "admin_id", status.AdminID, "error", err)
		return fmt.Errorf("update password status: %w", err)
	}

	return nil
}
