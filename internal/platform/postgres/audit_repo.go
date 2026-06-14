package postgres

import (
	"context"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// AuditRepo persists and reads the admin action audit log.
type AuditRepo struct{ q *db.Queries }

// Insert writes one audit entry.
func (r *AuditRepo) Insert(ctx context.Context, e domain.AuditEntry) error {
	return r.q.InsertAudit(ctx, db.InsertAuditParams{
		ID:      e.ID,
		AdminID: ptrToUUID(e.AdminID),
		Method:  e.Method,
		Path:    e.Path,
		Status:  int32(e.Status),
		Ip:      e.IP,
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
			ID:       row.ID,
			Time:     row.Time.Time,
			AdminID:  uuidToPtr(row.AdminID),
			Username: row.Username,
			Method:   row.Method,
			Path:     row.Path,
			Status:   int(row.Status),
			IP:       row.Ip,
		}
	}
	return out, nil
}
