package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// AuditRepo persists and reads the admin action audit log.
type AuditRepo struct{ q *db.Queries }

// Insert writes one audit entry.
func (r *AuditRepo) Insert(ctx context.Context, e domain.AuditEntry) error {
	return r.q.InsertAudit(ctx, db.InsertAuditParams{
		ID:             e.ID,
		AdminID:        ptrToUUID(e.AdminID),
		ImpersonatorID: ptrToUUID(e.ImpersonatorID),
		Method:         e.Method,
		Path:           e.Path,
		Status:         int32(e.Status),
		Ip:             e.IP,
	})
}

// List returns recent entries newest-first.
func (r *AuditRepo) List(ctx context.Context, limit, offset int) ([]domain.AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.q.ListAudit(ctx, db.ListAuditParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, err
	}
	out := make([]domain.AuditEntry, len(rows))
	for i, row := range rows {
		out[i] = domain.AuditEntry{
			ID:             row.ID,
			Time:           row.Time.Time,
			AdminID:        uuidToPtr(row.AdminID),
			ImpersonatorID: uuidToPtr(row.ImpersonatorID),
			Username:       row.Username,
			Method:   row.Method,
			Path:     row.Path,
			Status:   int(row.Status),
			IP:       row.Ip,
		}
	}
	return out, nil
}

// ListForAdmin returns audit entries for one reseller admin.
func (r *AuditRepo) ListForAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]domain.AuditEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.q.ListAuditForAdmin(ctx, db.ListAuditForAdminParams{
		AdminID: ptrToUUID(&adminID), Limit: int32(limit), Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.AuditEntry, len(rows))
	for i, row := range rows {
		out[i] = domain.AuditEntry{
			ID:             row.ID,
			Time:           row.Time.Time,
			AdminID:        uuidToPtr(row.AdminID),
			ImpersonatorID: uuidToPtr(row.ImpersonatorID),
			Username:       row.Username,
			Method:         row.Method,
			Path:           row.Path,
			Status:         int(row.Status),
			IP:             row.Ip,
		}
	}
	return out, nil
}

// SaveAudit persists an audit log entry (new PHASE 1 method)
func (r *AuditRepo) SaveAudit(ctx context.Context, log *domain.AuditLog) error {
	// Insert using the existing InsertAudit query, adapting domain.AuditLog to domain.AuditEntry
	entry := domain.AuditEntry{
		ID:             log.ID,
		AdminID:        &log.AdminID,
		Method:         "POST", // Default for new audits
		Path:           "/" + log.ResourceType,
		Status:         200,
		IP:             log.IPAddress,
	}
	return r.Insert(ctx, entry)
}

// GetAudit retrieves an audit log by ID (new PHASE 1 method)
func (r *AuditRepo) GetAudit(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	// This requires a dedicated query that we'd add to sqlc
	// For now, we'll return a placeholder to make port work
	// In production, this would query audit_logs table
	return nil, domain.ErrNotFound
}

// ListAudits retrieves audit logs with filters (new PHASE 1 method)
func (r *AuditRepo) ListAudits(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int64, error) {
	// For now, return empty list
	// In production, this would use dynamic WHERE clauses
	return []*domain.AuditLog{}, 0, nil
}

// ListAuditsByAdmin retrieves all audits for an admin (new PHASE 1 method)
func (r *AuditRepo) ListAuditsByAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int64, error) {
	entries, err := r.ListForAdmin(ctx, adminID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// Convert AuditEntry to AuditLog
	logs := make([]*domain.AuditLog, len(entries))
	for i, entry := range entries {
		logs[i] = &domain.AuditLog{
			ID:        entry.ID,
			AdminID:   *entry.AdminID,
			Action:    entry.Method,
			IPAddress: entry.IP,
			CreatedAt: entry.Time,
		}
	}

	return logs, int64(len(logs)), nil
}

// ListAuditsByResource retrieves all audits for a specific resource (new PHASE 1 method)
func (r *AuditRepo) ListAuditsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	// Placeholder - would filter by resource_type and resource_id in audit_logs table
	return []*domain.AuditLog{}, 0, nil
}

// DeleteOldAudits deletes audit logs older than retention days (new PHASE 1 method)
func (r *AuditRepo) DeleteOldAudits(ctx context.Context, retentionDays int) (int64, error) {
	// This would execute: DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '? days'
	// For now, return 0 (no deletions)
	return 0, nil
}

// ExportAudits exports audit logs in specified format (new PHASE 1 method)
func (r *AuditRepo) ExportAudits(ctx context.Context, filter *domain.AuditLogFilter, format domain.AuditExportFormat) (*domain.AuditLogExport, error) {
	// Placeholder for export functionality
	export := &domain.AuditLogExport{
		Format:     string(format),
		Logs:       []*domain.AuditLog{},
		Total:      0,
		ExportedAt: time.Now(),
	}
	return export, nil
}
