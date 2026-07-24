package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BulkOperationRepo implements port.BulkOperationRepository on PostgreSQL.
type BulkOperationRepo struct {
	pool *pgxpool.Pool
}

var _ port.BulkOperationRepository = (*BulkOperationRepo)(nil)

func (r *BulkOperationRepo) Create(ctx context.Context, op *domain.BulkOperation) error {
	paramsJSON, err := json.Marshal(op.Parameters)
	if err != nil {
		return err
	}
	filtersJSON, err := json.Marshal(op.Filters)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO bulk_operation_history (id, admin_id, operation_type, parameters, filters, affected_count, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		op.ID, op.AdminID, string(op.OperationType), paramsJSON, filtersJSON,
		op.AffectedCount, op.Status, op.CreatedAt)
	return err
}

func (r *BulkOperationRepo) List(ctx context.Context, adminID *uuid.UUID, limit, offset int) ([]*domain.BulkOperation, int, error) {
	if limit <= 0 {
		limit = 50
	}

	// Count total matching records.
	var total int
	if adminID != nil {
		err := r.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM bulk_operation_history WHERE admin_id = $1`, *adminID).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	} else {
		err := r.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM bulk_operation_history`).Scan(&total)
		if err != nil {
			return nil, 0, err
		}
	}

	// Fetch paginated results.
	var rows pgx.Rows
	var err error
	if adminID != nil {
		rows, err = r.pool.Query(ctx, `
			SELECT id, admin_id, operation_type, parameters, filters, affected_count, status, created_at
			FROM bulk_operation_history
			WHERE admin_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`, *adminID, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, admin_id, operation_type, parameters, filters, affected_count, status, created_at
			FROM bulk_operation_history
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`, limit, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []*domain.BulkOperation
	for rows.Next() {
		op, err := scanBulkOperation(rows)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, op)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func scanBulkOperation(rows pgx.Rows) (*domain.BulkOperation, error) {
	var op domain.BulkOperation
	var paramsJSON, filtersJSON []byte

	if err := rows.Scan(&op.ID, &op.AdminID, &op.OperationType, &paramsJSON, &filtersJSON,
		&op.AffectedCount, &op.Status, &op.CreatedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(paramsJSON, &op.Parameters); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(filtersJSON, &op.Filters); err != nil {
		return nil, err
	}
	return &op, nil
}
