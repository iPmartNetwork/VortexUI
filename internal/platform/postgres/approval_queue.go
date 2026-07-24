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

// ApprovalQueueRepo implements port.ApprovalQueueRepository on PostgreSQL.
type ApprovalQueueRepo struct {
	pool *pgxpool.Pool
}

var _ port.ApprovalQueueRepository = (*ApprovalQueueRepo)(nil)

func (r *ApprovalQueueRepo) Create(ctx context.Context, a *domain.SubscriptionApproval) error {
	dataJSON, err := json.Marshal(a.RequestData)
	if err != nil {
		return err
	}

	a.CreatedAt = time.Now()

	_, err = r.pool.Exec(ctx, `
		INSERT INTO subscription_approval_queue (id, user_id, request_data, status, admin_id, created_at, resolved_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.ID, a.UserID, dataJSON, a.Status, a.AdminID, a.CreatedAt, a.ResolvedAt)
	return err
}

func (r *ApprovalQueueRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.SubscriptionApproval, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, request_data, status, admin_id, created_at, resolved_at
		FROM subscription_approval_queue
		WHERE id = $1`, id)
	return scanApproval(row)
}

func (r *ApprovalQueueRepo) ListPending(ctx context.Context) ([]*domain.SubscriptionApproval, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, user_id, request_data, status, admin_id, created_at, resolved_at
		FROM subscription_approval_queue
		WHERE status = 'pending'
		ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectApprovals(rows)
}

func (r *ApprovalQueueRepo) Approve(ctx context.Context, id, adminID uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE subscription_approval_queue
		SET status = 'approved', admin_id = $2, resolved_at = $3
		WHERE id = $1`, id, adminID, now)
	return err
}

func (r *ApprovalQueueRepo) Reject(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := r.pool.Exec(ctx, `
		UPDATE subscription_approval_queue
		SET status = 'rejected', resolved_at = $2
		WHERE id = $1`, id, now)
	return err
}

// --- helpers ---

func scanApproval(row pgx.Row) (*domain.SubscriptionApproval, error) {
	var a domain.SubscriptionApproval
	var dataJSON []byte

	if err := row.Scan(&a.ID, &a.UserID, &dataJSON, &a.Status, &a.AdminID, &a.CreatedAt, &a.ResolvedAt); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(dataJSON, &a.RequestData); err != nil {
		return nil, err
	}
	return &a, nil
}

func collectApprovals(rows pgx.Rows) ([]*domain.SubscriptionApproval, error) {
	var approvals []*domain.SubscriptionApproval
	for rows.Next() {
		var a domain.SubscriptionApproval
		var dataJSON []byte

		if err := rows.Scan(&a.ID, &a.UserID, &dataJSON, &a.Status, &a.AdminID, &a.CreatedAt, &a.ResolvedAt); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(dataJSON, &a.RequestData); err != nil {
			return nil, err
		}
		approvals = append(approvals, &a)
	}
	return approvals, rows.Err()
}
