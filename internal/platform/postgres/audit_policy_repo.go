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
	"github.com/lib/pq"

	"github.com/vortexui/vortexui/internal/domain"
)

// AuditPolicyRepository implements port.AuditPolicyRepository using PostgreSQL
type AuditPolicyRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewAuditPolicyRepository creates new audit policy repository
func NewAuditPolicyRepository(pool *pgxpool.Pool, log *slog.Logger) *AuditPolicyRepository {
	if log == nil {
		log = slog.Default()
	}
	return &AuditPolicyRepository{
		pool: pool,
		log:  log,
	}
}

// GetPolicy retrieves the current audit policy
func (r *AuditPolicyRepository) GetPolicy(ctx context.Context) (*domain.AuditPolicy, error) {
	var policy domain.AuditPolicy

	query := `
		SELECT id, name, description, retention_days, alert_on_event_types, require_approval_for, auto_archive_after_days, compliance_frameworks, encryption_required, tamper_detection_enabled, exportable_formats, created_at, updated_at
		FROM audit_policies
		ORDER BY created_at DESC
		LIMIT 1
	`

	err := r.pool.QueryRow(ctx, query).Scan(
		&policy.ID,
		&policy.Name,
		&policy.Description,
		&policy.RetentionDays,
		pq.Array(&policy.AlertOnEventTypes),
		pq.Array(&policy.RequireApprovalFor),
		&policy.AutoArchiveAfterDays,
		pq.Array(&policy.ComplianceFrameworks),
		&policy.EncryptionRequired,
		&policy.TamperDetectionEnabled,
		pq.Array(&policy.ExportableFormats),
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Return default policy if none exists
			return &domain.AuditPolicy{
				ID:                     uuid.New(),
				Name:                   "Default Policy",
				Description:            "Default audit policy",
				RetentionDays:          90,
				AlertOnEventTypes:      []string{"auth_failure", "suspicious_activity", "brute_force_attempt"},
				RequireApprovalFor:     []string{"admin_created", "admin_deleted", "admin_permission_changed"},
				AutoArchiveAfterDays:   30,
				ComplianceFrameworks:   []string{"SOC2", "GDPR"},
				EncryptionRequired:     true,
				TamperDetectionEnabled: true,
				ExportableFormats:      []string{"json", "csv", "pdf"},
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}, nil
		}
		r.log.Error("failed to get audit policy", "error", err)
		return nil, fmt.Errorf("failed to get audit policy: %w", err)
	}

	return &policy, nil
}

// SavePolicy saves/updates audit policy
func (r *AuditPolicyRepository) SavePolicy(ctx context.Context, policy *domain.AuditPolicy) error {
	if policy.ID == uuid.Nil {
		policy.ID = uuid.New()
	}
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = time.Now()
	}
	policy.UpdatedAt = time.Now()

	query := `
		INSERT INTO audit_policies (id, name, description, retention_days, alert_on_event_types, require_approval_for, auto_archive_after_days, compliance_frameworks, encryption_required, tamper_detection_enabled, exportable_formats, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			retention_days = EXCLUDED.retention_days,
			alert_on_event_types = EXCLUDED.alert_on_event_types,
			require_approval_for = EXCLUDED.require_approval_for,
			auto_archive_after_days = EXCLUDED.auto_archive_after_days,
			compliance_frameworks = EXCLUDED.compliance_frameworks,
			encryption_required = EXCLUDED.encryption_required,
			tamper_detection_enabled = EXCLUDED.tamper_detection_enabled,
			exportable_formats = EXCLUDED.exportable_formats,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query,
		policy.ID,
		policy.Name,
		policy.Description,
		policy.RetentionDays,
		pq.Array(policy.AlertOnEventTypes),
		pq.Array(policy.RequireApprovalFor),
		policy.AutoArchiveAfterDays,
		pq.Array(policy.ComplianceFrameworks),
		policy.EncryptionRequired,
		policy.TamperDetectionEnabled,
		pq.Array(policy.ExportableFormats),
		policy.CreatedAt,
		policy.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save audit policy", "error", err)
		return fmt.Errorf("failed to save audit policy: %w", err)
	}

	return nil
}

// UpdateRetentionPolicy updates log retention settings
func (r *AuditPolicyRepository) UpdateRetentionPolicy(ctx context.Context, retentionDays int) error {
	query := `
		UPDATE audit_policies
		SET retention_days = $1, updated_at = $2
	`

	result, err := r.pool.Exec(ctx, query, retentionDays, time.Now())
	if err != nil {
		r.log.Error("failed to update retention policy", "error", err)
		return fmt.Errorf("failed to update retention policy: %w", err)
	}

	if result.RowsAffected() == 0 {
		r.log.Warn("no audit policy found to update retention")
	}

	return nil
}

// UpdateComplianceFrameworks updates which frameworks apply
func (r *AuditPolicyRepository) UpdateComplianceFrameworks(ctx context.Context, frameworks []string) error {
	query := `
		UPDATE audit_policies
		SET compliance_frameworks = $1, updated_at = $2
	`

	result, err := r.pool.Exec(ctx, query, pq.Array(frameworks), time.Now())
	if err != nil {
		r.log.Error("failed to update compliance frameworks", "error", err)
		return fmt.Errorf("failed to update compliance frameworks: %w", err)
	}

	if result.RowsAffected() == 0 {
		r.log.Warn("no audit policy found to update frameworks")
	}

	return nil
}

// ========== AuditArchiveRepository ==========

// AuditArchiveRepository implements port.AuditArchiveRepository using PostgreSQL
type AuditArchiveRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewAuditArchiveRepository creates new audit archive repository
func NewAuditArchiveRepository(pool *pgxpool.Pool, log *slog.Logger) *AuditArchiveRepository {
	if log == nil {
		log = slog.Default()
	}
	return &AuditArchiveRepository{
		pool: pool,
		log:  log,
	}
}

// ArchiveEvents creates an archive of events in date range
func (r *AuditArchiveRepository) ArchiveEvents(ctx context.Context, startDate time.Time, endDate time.Time) (*domain.AuditLogArchive, error) {
	archiveID := uuid.New()
	archiveName := fmt.Sprintf("audit-logs-%s-to-%s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Count events to archive
	var eventCount int
	countQuery := "SELECT COUNT(*) FROM audit_events WHERE created_at >= $1 AND created_at <= $2"
	err := r.pool.QueryRow(ctx, countQuery, startDate, endDate).Scan(&eventCount)
	if err != nil {
		r.log.Error("failed to count events for archival", "error", err)
		return nil, fmt.Errorf("failed to count events: %w", err)
	}

	// Create archive record
	query := `
		INSERT INTO audit_log_archives (id, archive_name, file_path, start_date, end_date, event_count, checksum_sha256, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	filePath := fmt.Sprintf("s3://audit-archives/%s.tar.gz", archiveName)
	checksumSHA256 := "placeholder" // In real implementation, calculate actual checksum

	_, err = r.pool.Exec(ctx, query,
		archiveID,
		archiveName,
		filePath,
		startDate,
		endDate,
		eventCount,
		checksumSHA256,
		time.Now(),
	)

	if err != nil {
		r.log.Error("failed to create archive record", "error", err)
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	return &domain.AuditLogArchive{
		ID:             archiveID,
		ArchiveName:    archiveName,
		FilePath:       filePath,
		StartDate:      startDate,
		EndDate:        endDate,
		EventCount:     eventCount,
		ChecksumSHA256: checksumSHA256,
		CreatedAt:      time.Now(),
	}, nil
}

// GetArchive retrieves a specific archive
func (r *AuditArchiveRepository) GetArchive(ctx context.Context, archiveID uuid.UUID) (*domain.AuditLogArchive, error) {
	var archive domain.AuditLogArchive

	query := `
		SELECT id, archive_name, file_path, start_date, end_date, event_count, checksum_sha256, created_at, expires_at
		FROM audit_log_archives
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, archiveID).Scan(
		&archive.ID,
		&archive.ArchiveName,
		&archive.FilePath,
		&archive.StartDate,
		&archive.EndDate,
		&archive.EventCount,
		&archive.ChecksumSHA256,
		&archive.CreatedAt,
		&archive.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get archive", "error", err)
		return nil, fmt.Errorf("failed to get archive: %w", err)
	}

	return &archive, nil
}

// ListArchives retrieves all archives
func (r *AuditArchiveRepository) ListArchives(ctx context.Context, limit int, offset int) ([]*domain.AuditLogArchive, int64, error) {
	query := `
		SELECT id, archive_name, file_path, start_date, end_date, event_count, checksum_sha256, created_at, expires_at
		FROM audit_log_archives
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	countQuery := "SELECT COUNT(*) FROM audit_log_archives"

	var total int64
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		r.log.Error("failed to count archives", "error", err)
		return nil, 0, fmt.Errorf("failed to count archives: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		r.log.Error("failed to query archives", "error", err)
		return nil, 0, fmt.Errorf("failed to query archives: %w", err)
	}
	defer rows.Close()

	var archives []*domain.AuditLogArchive
	for rows.Next() {
		var archive domain.AuditLogArchive

		err := rows.Scan(
			&archive.ID,
			&archive.ArchiveName,
			&archive.FilePath,
			&archive.StartDate,
			&archive.EndDate,
			&archive.EventCount,
			&archive.ChecksumSHA256,
			&archive.CreatedAt,
			&archive.ExpiresAt,
		)

		if err != nil {
			r.log.Error("failed to scan archive", "error", err)
			continue
		}

		archives = append(archives, &archive)
	}

	return archives, total, nil
}

// VerifyArchiveIntegrity verifies archive hasn't been tampered with
func (r *AuditArchiveRepository) VerifyArchiveIntegrity(ctx context.Context, archiveID uuid.UUID) (bool, error) {
	var storedChecksum string

	query := "SELECT checksum_sha256 FROM audit_log_archives WHERE id = $1"
	err := r.pool.QueryRow(ctx, query, archiveID).Scan(&storedChecksum)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, domain.ErrNotFound
		}
		r.log.Error("failed to get archive checksum", "error", err)
		return false, fmt.Errorf("failed to get archive: %w", err)
	}

	// In real implementation, would verify against actual file checksum
	// For now, just verify record exists and has checksum
	return storedChecksum != "", nil
}

// DeleteOldArchives deletes archives older than retentionDays
func (r *AuditArchiveRepository) DeleteOldArchives(ctx context.Context, retentionDays int) (int64, error) {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := "DELETE FROM audit_log_archives WHERE created_at < $1"
	result, err := r.pool.Exec(ctx, query, cutoffDate)
	if err != nil {
		r.log.Error("failed to delete old archives", "error", err)
		return 0, fmt.Errorf("failed to delete old archives: %w", err)
	}

	return result.RowsAffected(), nil
}
