package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// LoginAttemptRepository implements login attempt repository using pgx
type LoginAttemptRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewLoginAttemptRepository creates a new login attempt repository
func NewLoginAttemptRepository(pool *pgxpool.Pool, log *slog.Logger) *LoginAttemptRepository {
	if log == nil {
		log = slog.Default()
	}
	return &LoginAttemptRepository{
		pool: pool,
		log:  log,
	}
}

// RecordAttempt records a login attempt
func (r *LoginAttemptRepository) RecordAttempt(ctx context.Context, attempt *domain.LoginAttempt) error {
	query := `
		INSERT INTO login_attempts (id, admin_id, ip_address, success, reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(ctx, query,
		attempt.ID, attempt.AdminID, attempt.IPAddress, attempt.Success, attempt.Reason, attempt.CreatedAt,
	)

	if err != nil {
		r.log.Error("failed to record login attempt", "admin_id", attempt.AdminID, "error", err)
		return fmt.Errorf("record login attempt: %w", err)
	}

	return nil
}

// GetRecentAttempts retrieves recent login attempts for an admin
func (r *LoginAttemptRepository) GetRecentAttempts(ctx context.Context, adminID uuid.UUID, minutesBack int) ([]domain.LoginAttempt, error) {
	if minutesBack <= 0 {
		minutesBack = 30
	}

	query := `
		SELECT id, admin_id, ip_address, success, reason, created_at
		FROM login_attempts
		WHERE admin_id = $1 AND created_at > NOW() - INTERVAL '1 minute' * $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, adminID, minutesBack)
	if err != nil {
		r.log.Error("failed to get recent login attempts", "admin_id", adminID, "error", err)
		return nil, fmt.Errorf("get recent login attempts: %w", err)
	}
	defer rows.Close()

	attempts := []domain.LoginAttempt{}
	for rows.Next() {
		attempt := domain.LoginAttempt{}
		if err := rows.Scan(
			&attempt.ID, &attempt.AdminID, &attempt.IPAddress, &attempt.Success,
			&attempt.Reason, &attempt.CreatedAt,
		); err != nil {
			r.log.Error("failed to scan login attempt", "error", err)
			continue
		}
		attempts = append(attempts, attempt)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("error reading login attempt rows", "error", err)
		return nil, err
	}

	return attempts, nil
}

// GetFailedAttemptCount counts failed login attempts in time window
func (r *LoginAttemptRepository) GetFailedAttemptCount(ctx context.Context, adminID uuid.UUID, minutesBack int) (int64, error) {
	if minutesBack <= 0 {
		minutesBack = 30
	}

	query := `
		SELECT COUNT(*)
		FROM login_attempts
		WHERE admin_id = $1 AND success = FALSE AND created_at > NOW() - INTERVAL '1 minute' * $2
	`

	var count int64
	err := r.pool.QueryRow(ctx, query, adminID, minutesBack).Scan(&count)
	if err != nil {
		r.log.Error("failed to count failed login attempts", "admin_id", adminID, "error", err)
		return 0, fmt.Errorf("count failed login attempts: %w", err)
	}

	return count, nil
}

// ClearAttempts deletes old login attempts
func (r *LoginAttemptRepository) ClearAttempts(ctx context.Context, olderThanHours int) (int64, error) {
	if olderThanHours <= 0 {
		olderThanHours = 24
	}

	query := `
		DELETE FROM login_attempts
		WHERE created_at < NOW() - INTERVAL '1 hour' * $1
	`

	result, err := r.pool.Exec(ctx, query, olderThanHours)
	if err != nil {
		r.log.Error("failed to clear old login attempts", "error", err)
		return 0, fmt.Errorf("clear login attempts: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetAttemptsByIPInWindow retrieves all attempts from an IP in a time window
func (r *LoginAttemptRepository) GetAttemptsByIPInWindow(ctx context.Context, ipAddress string, minutesBack int) ([]domain.LoginAttempt, error) {
	if minutesBack <= 0 {
		minutesBack = 30
	}

	query := `
		SELECT id, admin_id, ip_address, success, reason, created_at
		FROM login_attempts
		WHERE ip_address = $1 AND created_at > NOW() - INTERVAL '1 minute' * $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, ipAddress, minutesBack)
	if err != nil {
		r.log.Error("failed to get attempts by IP", "ip", ipAddress, "error", err)
		return nil, fmt.Errorf("get attempts by IP: %w", err)
	}
	defer rows.Close()

	attempts := []domain.LoginAttempt{}
	for rows.Next() {
		attempt := domain.LoginAttempt{}
		if err := rows.Scan(
			&attempt.ID, &attempt.AdminID, &attempt.IPAddress, &attempt.Success,
			&attempt.Reason, &attempt.CreatedAt,
		); err != nil {
			r.log.Error("failed to scan login attempt", "error", err)
			continue
		}
		attempts = append(attempts, attempt)
	}

	if err = rows.Err(); err != nil {
		r.log.Error("error reading login attempt rows", "error", err)
		return nil, err
	}

	return attempts, nil
}

// ClearAdminAttempts clears all attempts for an admin (after password reset)
func (r *LoginAttemptRepository) ClearAdminAttempts(ctx context.Context, adminID uuid.UUID) error {
	query := `DELETE FROM login_attempts WHERE admin_id = $1`

	_, err := r.pool.Exec(ctx, query, adminID)
	if err != nil {
		r.log.Error("failed to clear admin login attempts", "admin_id", adminID, "error", err)
		return fmt.Errorf("clear admin login attempts: %w", err)
	}

	return nil
}

// StartCleanupWorker starts background cleanup of old login attempts
func (r *LoginAttemptRepository) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				r.log.Info("login attempt cleanup worker stopped")
				return
			case <-ticker.C:
				deleted, err := r.ClearAttempts(ctx, 24)
				if err != nil {
					r.log.Error("cleanup worker error", "error", err)
				} else {
					r.log.Debug("cleared old login attempts", "count", deleted)
				}
			}
		}
	}()
}
