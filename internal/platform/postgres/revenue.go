package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// RevenueRepo implements port.RevenueRepository on PostgreSQL with aggregation
// queries for income/expense reporting and per-admin breakdowns.
type RevenueRepo struct {
	pool *pgxpool.Pool
}

var _ port.RevenueRepository = (*RevenueRepo)(nil)

// Create inserts a new revenue entry.
func (r *RevenueRepo) Create(ctx context.Context, entry *domain.RevenueEntry) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO revenue_entries (id, admin_id, type, amount, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.AdminID, entry.Type, entry.Amount, entry.Description, entry.CreatedAt)
	return err
}

// Report aggregates revenue data between two timestamps. If adminID is non-nil,
// the report is scoped to that specific admin/reseller.
func (r *RevenueRepo) Report(ctx context.Context, adminID *uuid.UUID, from, to time.Time) (*domain.RevenueReport, error) {
	report := &domain.RevenueReport{}

	// Total income and expense
	var whereAdmin string
	args := []any{from, to}
	if adminID != nil {
		whereAdmin = " AND admin_id = $3"
		args = append(args, *adminID)
	}

	row := r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS total_income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS total_expense
		FROM revenue_entries
		WHERE created_at >= $1 AND created_at <= $2`+whereAdmin, args...)

	if err := row.Scan(&report.TotalIncome, &report.TotalExpense); err != nil {
		return nil, err
	}
	report.Profit = report.TotalIncome - report.TotalExpense

	// By admin breakdown
	rows, err := r.pool.Query(ctx, `
		SELECT
			admin_id,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense
		FROM revenue_entries
		WHERE created_at >= $1 AND created_at <= $2`+whereAdmin+`
		GROUP BY admin_id
		ORDER BY income DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var ar domain.AdminRevenue
		if err := rows.Scan(&ar.AdminID, &ar.Income, &ar.Expense); err != nil {
			return nil, err
		}
		report.ByAdmin = append(report.ByAdmin, ar)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Time series (daily aggregation)
	rows2, err := r.pool.Query(ctx, `
		SELECT
			TO_CHAR(created_at::DATE, 'YYYY-MM-DD') AS day,
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) AS income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) AS expense
		FROM revenue_entries
		WHERE created_at >= $1 AND created_at <= $2`+whereAdmin+`
		GROUP BY created_at::DATE
		ORDER BY day`, args...)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var dp domain.RevenueDataPoint
		if err := rows2.Scan(&dp.Date, &dp.Income, &dp.Expense); err != nil {
			return nil, err
		}
		report.TimeSeries = append(report.TimeSeries, dp)
	}
	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return report, nil
}
