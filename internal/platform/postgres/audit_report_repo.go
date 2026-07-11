package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// AuditReportRepository implements port.AuditReportRepository using PostgreSQL
type AuditReportRepository struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

// NewAuditReportRepository creates new audit report repository
func NewAuditReportRepository(pool *pgxpool.Pool, log *slog.Logger) *AuditReportRepository {
	if log == nil {
		log = slog.Default()
	}
	return &AuditReportRepository{
		pool: pool,
		log:  log,
	}
}

// SaveReport saves an audit report
func (r *AuditReportRepository) SaveReport(ctx context.Context, report *domain.AuditReport) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	if report.CreatedAt.IsZero() {
		report.CreatedAt = time.Now()
	}
	report.UpdatedAt = time.Now()

	filtersJSON, err := json.Marshal(report.Filters)
	if err != nil {
		r.log.Error("failed to marshal filters", "error", err)
		filtersJSON = []byte("{}")
	}

	metadataJSON, err := json.Marshal(report.Metadata)
	if err != nil {
		r.log.Error("failed to marshal metadata", "error", err)
		metadataJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_reports (id, title, description, report_type, status, start_date, end_date, scope, filters, event_count, file_path, created_by, approved_by, approved_at, exported_at, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			status = EXCLUDED.status,
			event_count = EXCLUDED.event_count,
			file_path = EXCLUDED.file_path,
			approved_by = EXCLUDED.approved_by,
			approved_at = EXCLUDED.approved_at,
			exported_at = EXCLUDED.exported_at,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.pool.Exec(ctx, query,
		report.ID,
		report.Title,
		report.Description,
		report.ReportType,
		report.Status,
		report.StartDate,
		report.EndDate,
		report.Scope,
		filtersJSON,
		report.EventCount,
		report.FilePath,
		report.CreatedBy,
		report.ApprovedBy,
		report.ApprovedAt,
		report.ExportedAt,
		metadataJSON,
		report.CreatedAt,
		report.UpdatedAt,
	)

	if err != nil {
		r.log.Error("failed to save audit report", "error", err, "report_id", report.ID)
		return fmt.Errorf("failed to save audit report: %w", err)
	}

	return nil
}

// GetReport retrieves a specific report
func (r *AuditReportRepository) GetReport(ctx context.Context, reportID uuid.UUID) (*domain.AuditReport, error) {
	var report domain.AuditReport
	var filtersJSON, metadataJSON []byte

	query := `
		SELECT id, title, description, report_type, status, start_date, end_date, scope, filters, event_count, file_path, created_by, approved_by, approved_at, exported_at, metadata, created_at, updated_at
		FROM audit_reports
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, reportID).Scan(
		&report.ID,
		&report.Title,
		&report.Description,
		&report.ReportType,
		&report.Status,
		&report.StartDate,
		&report.EndDate,
		&report.Scope,
		&filtersJSON,
		&report.EventCount,
		&report.FilePath,
		&report.CreatedBy,
		&report.ApprovedBy,
		&report.ApprovedAt,
		&report.ExportedAt,
		&metadataJSON,
		&report.CreatedAt,
		&report.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		r.log.Error("failed to get audit report", "error", err, "report_id", reportID)
		return nil, fmt.Errorf("failed to get audit report: %w", err)
	}

	report.Filters = make(map[string]interface{})
	if len(filtersJSON) > 0 {
		_ = json.Unmarshal(filtersJSON, &report.Filters)
	}

	report.Metadata = make(map[string]interface{})
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &report.Metadata)
	}

	return &report, nil
}

// ListReports retrieves reports with optional filtering
func (r *AuditReportRepository) ListReports(ctx context.Context, reportType *string, status *string, limit int, offset int) ([]*domain.AuditReport, int64, error) {
	query := "SELECT id, title, description, report_type, status, start_date, end_date, scope, filters, event_count, file_path, created_by, approved_by, approved_at, exported_at, metadata, created_at, updated_at FROM audit_reports WHERE 1=1"
	countQuery := "SELECT COUNT(*) FROM audit_reports WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if reportType != nil {
		query += fmt.Sprintf(" AND report_type = $%d", argCount)
		countQuery += fmt.Sprintf(" AND report_type = $%d", argCount)
		args = append(args, reportType)
		argCount++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		countQuery += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
		argCount++
	}

	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		r.log.Error("failed to count audit reports", "error", err)
		return nil, 0, fmt.Errorf("failed to count audit reports: %w", err)
	}

	// Add pagination
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
		args = append(args, limit, offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to query audit reports", "error", err)
		return nil, 0, fmt.Errorf("failed to query audit reports: %w", err)
	}
	defer rows.Close()

	var reports []*domain.AuditReport
	for rows.Next() {
		var report domain.AuditReport
		var filtersJSON, metadataJSON []byte

		err := rows.Scan(
			&report.ID,
			&report.Title,
			&report.Description,
			&report.ReportType,
			&report.Status,
			&report.StartDate,
			&report.EndDate,
			&report.Scope,
			&filtersJSON,
			&report.EventCount,
			&report.FilePath,
			&report.CreatedBy,
			&report.ApprovedBy,
			&report.ApprovedAt,
			&report.ExportedAt,
			&metadataJSON,
			&report.CreatedAt,
			&report.UpdatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan audit report", "error", err)
			continue
		}

		report.Filters = make(map[string]interface{})
		if len(filtersJSON) > 0 {
			_ = json.Unmarshal(filtersJSON, &report.Filters)
		}

		report.Metadata = make(map[string]interface{})
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &report.Metadata)
		}

		reports = append(reports, &report)
	}

	return reports, total, nil
}

// UpdateReportStatus updates report status
func (r *AuditReportRepository) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, newStatus string) error {
	query := `
		UPDATE audit_reports
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, newStatus, time.Now(), reportID)
	if err != nil {
		r.log.Error("failed to update report status", "error", err, "report_id", reportID)
		return fmt.Errorf("failed to update report status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ApproveReport approves a pending report
func (r *AuditReportRepository) ApproveReport(ctx context.Context, reportID uuid.UUID, approvedBy uuid.UUID) error {
	now := time.Now()

	query := `
		UPDATE audit_reports
		SET status = 'approved', approved_by = $1, approved_at = $2, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, approvedBy, now, reportID)
	if err != nil {
		r.log.Error("failed to approve report", "error", err, "report_id", reportID)
		return fmt.Errorf("failed to approve report: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// ExportReport marks report as exported and stores file path
func (r *AuditReportRepository) ExportReport(ctx context.Context, reportID uuid.UUID, filePath string) error {
	now := time.Now()

	query := `
		UPDATE audit_reports
		SET status = 'exported', file_path = $1, exported_at = $2, updated_at = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, filePath, now, reportID)
	if err != nil {
		r.log.Error("failed to export report", "error", err, "report_id", reportID)
		return fmt.Errorf("failed to export report: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// GetReportsByCreator retrieves reports created by specific admin
func (r *AuditReportRepository) GetReportsByCreator(ctx context.Context, adminID uuid.UUID, limit int, offset int) ([]*domain.AuditReport, int64, error) {
	query := `
		SELECT id, title, description, report_type, status, start_date, end_date, scope, filters, event_count, file_path, created_by, approved_by, approved_at, exported_at, metadata, created_at, updated_at
		FROM audit_reports
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*)
		FROM audit_reports
		WHERE created_by = $1
	`

	var total int64
	err := r.pool.QueryRow(ctx, countQuery, adminID).Scan(&total)
	if err != nil {
		r.log.Error("failed to count reports by creator", "error", err)
		return nil, 0, fmt.Errorf("failed to count reports: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, adminID, limit, offset)
	if err != nil {
		r.log.Error("failed to query reports by creator", "error", err)
		return nil, 0, fmt.Errorf("failed to query reports: %w", err)
	}
	defer rows.Close()

	var reports []*domain.AuditReport
	for rows.Next() {
		var report domain.AuditReport
		var filtersJSON, metadataJSON []byte

		err := rows.Scan(
			&report.ID,
			&report.Title,
			&report.Description,
			&report.ReportType,
			&report.Status,
			&report.StartDate,
			&report.EndDate,
			&report.Scope,
			&filtersJSON,
			&report.EventCount,
			&report.FilePath,
			&report.CreatedBy,
			&report.ApprovedBy,
			&report.ApprovedAt,
			&report.ExportedAt,
			&metadataJSON,
			&report.CreatedAt,
			&report.UpdatedAt,
		)

		if err != nil {
			r.log.Error("failed to scan report", "error", err)
			continue
		}

		report.Filters = make(map[string]interface{})
		if len(filtersJSON) > 0 {
			_ = json.Unmarshal(filtersJSON, &report.Filters)
		}

		report.Metadata = make(map[string]interface{})
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &report.Metadata)
		}

		reports = append(reports, &report)
	}

	return reports, total, nil
}

// DeleteReport deletes a report (soft or hard delete)
func (r *AuditReportRepository) DeleteReport(ctx context.Context, reportID uuid.UUID) error {
	query := "DELETE FROM audit_reports WHERE id = $1"

	result, err := r.pool.Exec(ctx, query, reportID)
	if err != nil {
		r.log.Error("failed to delete report", "error", err, "report_id", reportID)
		return fmt.Errorf("failed to delete report: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
