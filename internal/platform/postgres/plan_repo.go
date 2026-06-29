package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// PlanRepo implements service.PlanRepository on PostgreSQL. It persists
// subscription plans and purchase orders.
type PlanRepo struct{ pool *pgxpool.Pool }

const planCols = `id, name, description, data_limit, duration_days, device_limit,
	reset_strategy, inbound_ids, price_toman, price_usd, max_users, enabled, created_at, admin_id`

const orderCols = `id, user_id, admin_id, plan_id, username, status, gateway,
	gateway_id, amount, currency, created_at, paid_at, proof_image`

func marshalUUIDs(ids []uuid.UUID) ([]byte, error) {
	if ids == nil {
		ids = []uuid.UUID{}
	}
	return json.Marshal(ids)
}

func scanPlan(row pgx.Row) (*domain.Plan, error) {
	var p domain.Plan
	var idsJSON []byte
	var adminID pgtype.UUID
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.DataLimit, &p.Duration,
		&p.DeviceLimit, &p.ResetStrategy, &idsJSON, &p.PriceToman, &p.PriceUSD,
		&p.MaxUsers, &p.Enabled, &p.CreatedAt, &adminID); err != nil {
		return nil, err
	}
	if len(idsJSON) > 0 {
		_ = json.Unmarshal(idsJSON, &p.InboundIDs)
	}
	if adminID.Valid {
		a := uuid.UUID(adminID.Bytes)
		p.AdminID = &a
	}
	return &p, nil
}

func scanOrder(row pgx.Row) (*domain.Order, error) {
	var o domain.Order
	var userID, adminID pgtype.UUID
	var paidAt pgtype.Timestamptz
	if err := row.Scan(&o.ID, &userID, &adminID, &o.PlanID, &o.Username, &o.Status,
		&o.Gateway, &o.GatewayID, &o.Amount, &o.Currency, &o.CreatedAt, &paidAt, &o.ProofImage); err != nil {
		return nil, err
	}
	if userID.Valid {
		u := uuid.UUID(userID.Bytes)
		o.UserID = &u
	}
	if adminID.Valid {
		a := uuid.UUID(adminID.Bytes)
		o.AdminID = &a
	}
	if paidAt.Valid {
		t := paidAt.Time
		o.PaidAt = &t
	}
	return &o, nil
}

func (r *PlanRepo) CreatePlan(ctx context.Context, p *domain.Plan) error {
	idsJSON, err := marshalUUIDs(p.InboundIDs)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO plans (id, name, description, data_limit, duration_days, device_limit,
			reset_strategy, inbound_ids, price_toman, price_usd, max_users, enabled, created_at, admin_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		p.ID, p.Name, p.Description, p.DataLimit, p.Duration, p.DeviceLimit,
		p.ResetStrategy, idsJSON, p.PriceToman, p.PriceUSD, p.MaxUsers, p.Enabled, p.CreatedAt, p.AdminID)
	return err
}

func (r *PlanRepo) GetPlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+planCols+` FROM plans WHERE id = $1`, id)
	return scanPlan(row)
}

func (r *PlanRepo) ListPlans(ctx context.Context) ([]*domain.Plan, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+planCols+` FROM plans ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Plan
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PlanRepo) UpdatePlan(ctx context.Context, p *domain.Plan) error {
	idsJSON, err := marshalUUIDs(p.InboundIDs)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE plans SET name=$2, description=$3, data_limit=$4, duration_days=$5,
			device_limit=$6, reset_strategy=$7, inbound_ids=$8, price_toman=$9,
			price_usd=$10, max_users=$11, enabled=$12
		WHERE id=$1`,
		p.ID, p.Name, p.Description, p.DataLimit, p.Duration, p.DeviceLimit,
		p.ResetStrategy, idsJSON, p.PriceToman, p.PriceUSD, p.MaxUsers, p.Enabled)
	return err
}

func (r *PlanRepo) DeletePlan(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM plans WHERE id=$1`, id)
	return err
}

func (r *PlanRepo) CreateOrder(ctx context.Context, o *domain.Order) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO orders (id, user_id, admin_id, plan_id, username, status, gateway,
			gateway_id, amount, currency, created_at, paid_at, proof_image)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		o.ID, o.UserID, o.AdminID, o.PlanID, o.Username, o.Status, o.Gateway,
		o.GatewayID, o.Amount, o.Currency, o.CreatedAt, o.PaidAt, o.ProofImage)
	return err
}

func (r *PlanRepo) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+orderCols+` FROM orders WHERE id=$1`, id)
	return scanOrder(row)
}

func (r *PlanRepo) UpdateOrder(ctx context.Context, o *domain.Order) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE orders SET user_id=$2, admin_id=$3, status=$4, gateway=$5, gateway_id=$6,
			amount=$7, currency=$8, paid_at=$9, proof_image=$10
		WHERE id=$1`,
		o.ID, o.UserID, o.AdminID, o.Status, o.Gateway, o.GatewayID, o.Amount, o.Currency, o.PaidAt, o.ProofImage)
	return err
}

func (r *PlanRepo) ListOrders(ctx context.Context, userID *uuid.UUID, limit int) ([]*domain.Order, error) {
	if limit <= 0 {
		limit = 100
	}
	var rows pgx.Rows
	var err error
	if userID != nil {
		rows, err = r.pool.Query(ctx, `SELECT `+orderCols+` FROM orders WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2`, *userID, limit)
	} else {
		rows, err = r.pool.Query(ctx, `SELECT `+orderCols+` FROM orders ORDER BY created_at DESC LIMIT $1`, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
