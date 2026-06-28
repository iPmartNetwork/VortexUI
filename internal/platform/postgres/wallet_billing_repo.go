package postgres

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
)

// WalletBillingRepo persists wallet packages, deposits, and billing settings.
type WalletBillingRepo struct{ pool *pgxpool.Pool }

const depositCols = `d.id, d.admin_id, a.username, d.package_id, p.name, d.method, d.status,
	d.amount, d.currency, d.traffic_bytes, d.user_credits, d.gateway_id, d.tx_id, d.crypto_coin,
	d.proof_image, d.reseller_note, d.admin_note, d.reviewer_id, d.reviewed_at, d.paid_at, d.created_at`

func (r *WalletBillingRepo) scanDeposit(row pgx.Row) (*domain.WalletDeposit, error) {
	var d domain.WalletDeposit
	var pkgID, reviewerID pgtype.UUID
	var pkgName pgtype.Text
	var reviewedAt, paidAt pgtype.Timestamptz
	if err := row.Scan(
		&d.ID, &d.AdminID, &d.AdminUsername, &pkgID, &pkgName, &d.Method, &d.Status,
		&d.Amount, &d.Currency, &d.TrafficBytes, &d.UserCredits, &d.GatewayID, &d.TxID, &d.CryptoCoin,
		&d.ProofImage, &d.ResellerNote, &d.AdminNote, &reviewerID, &reviewedAt, &paidAt, &d.CreatedAt,
	); err != nil {
		return nil, err
	}
	if pkgID.Valid {
		id := uuid.UUID(pkgID.Bytes)
		d.PackageID = &id
	}
	if pkgName.Valid {
		d.PackageName = pkgName.String
	}
	if reviewerID.Valid {
		id := uuid.UUID(reviewerID.Bytes)
		d.ReviewerID = &id
	}
	if reviewedAt.Valid {
		t := reviewedAt.Time
		d.ReviewedAt = &t
	}
	if paidAt.Valid {
		t := paidAt.Time
		d.PaidAt = &t
	}
	return &d, nil
}

func (r *WalletBillingRepo) CreatePackage(ctx context.Context, p *domain.WalletPackage) error {
	methods, err := json.Marshal(p.Methods)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		INSERT INTO wallet_packages (id, name, description, traffic_bytes, user_credits,
			price_amount, currency, methods, enabled, sort_order, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		p.ID, p.Name, p.Description, p.TrafficBytes, p.UserCredits,
		p.PriceAmount, p.Currency, methods, p.Enabled, p.SortOrder, p.CreatedAt)
	return err
}

func (r *WalletBillingRepo) UpdatePackage(ctx context.Context, p *domain.WalletPackage) error {
	methods, err := json.Marshal(p.Methods)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE wallet_packages SET name=$2, description=$3, traffic_bytes=$4, user_credits=$5,
			price_amount=$6, currency=$7, methods=$8, enabled=$9, sort_order=$10
		WHERE id=$1`,
		p.ID, p.Name, p.Description, p.TrafficBytes, p.UserCredits,
		p.PriceAmount, p.Currency, methods, p.Enabled, p.SortOrder)
	return err
}

func (r *WalletBillingRepo) DeletePackage(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM wallet_packages WHERE id=$1`, id)
	return err
}

func (r *WalletBillingRepo) GetPackage(ctx context.Context, id uuid.UUID) (*domain.WalletPackage, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, name, description, traffic_bytes, user_credits, price_amount, currency,
			methods, enabled, sort_order, created_at
		FROM wallet_packages WHERE id=$1`, id)
	return scanPackage(row)
}

func scanPackage(row pgx.Row) (*domain.WalletPackage, error) {
	var p domain.WalletPackage
	var methodsJSON []byte
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.TrafficBytes, &p.UserCredits,
		&p.PriceAmount, &p.Currency, &methodsJSON, &p.Enabled, &p.SortOrder, &p.CreatedAt); err != nil {
		return nil, err
	}
	if len(methodsJSON) > 0 {
		_ = json.Unmarshal(methodsJSON, &p.Methods)
	}
	return &p, nil
}

func (r *WalletBillingRepo) ListPackages(ctx context.Context, enabledOnly bool) ([]*domain.WalletPackage, error) {
	q := `SELECT id, name, description, traffic_bytes, user_credits, price_amount, currency,
		methods, enabled, sort_order, created_at FROM wallet_packages`
	if enabledOnly {
		q += ` WHERE enabled = TRUE`
	}
	q += ` ORDER BY sort_order ASC, created_at DESC`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.WalletPackage
	for rows.Next() {
		p, err := scanPackage(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *WalletBillingRepo) GetBillingSettings(ctx context.Context) (*domain.BillingSettings, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT card_number, card_holder, card_bank, crypto_addresses, manual_instructions
		FROM billing_settings WHERE id=1`)
	var s domain.BillingSettings
	var cryptoJSON []byte
	if err := row.Scan(&s.CardNumber, &s.CardHolder, &s.CardBank, &cryptoJSON, &s.ManualInstructions); err != nil {
		return nil, err
	}
	s.CryptoAddresses = map[string]string{}
	if len(cryptoJSON) > 0 {
		_ = json.Unmarshal(cryptoJSON, &s.CryptoAddresses)
	}
	return &s, nil
}

func (r *WalletBillingRepo) UpdateBillingSettings(ctx context.Context, s *domain.BillingSettings) error {
	cryptoJSON, err := json.Marshal(s.CryptoAddresses)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `
		UPDATE billing_settings SET card_number=$1, card_holder=$2, card_bank=$3,
			crypto_addresses=$4, manual_instructions=$5 WHERE id=1`,
		s.CardNumber, s.CardHolder, s.CardBank, cryptoJSON, s.ManualInstructions)
	return err
}

func (r *WalletBillingRepo) CreateDeposit(ctx context.Context, d *domain.WalletDeposit) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO wallet_deposits (id, admin_id, package_id, method, status, amount, currency,
			traffic_bytes, user_credits, gateway_id, tx_id, crypto_coin, proof_image, reseller_note, admin_note,
			reviewer_id, reviewed_at, paid_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		d.ID, d.AdminID, d.PackageID, d.Method, d.Status, d.Amount, d.Currency,
		d.TrafficBytes, d.UserCredits, d.GatewayID, d.TxID, d.CryptoCoin, d.ProofImage, d.ResellerNote, d.AdminNote,
		d.ReviewerID, d.ReviewedAt, d.PaidAt, d.CreatedAt)
	return err
}

func (r *WalletBillingRepo) UpdateDeposit(ctx context.Context, d *domain.WalletDeposit) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE wallet_deposits SET status=$2, gateway_id=$3, tx_id=$4, crypto_coin=$5, proof_image=$6,
			reseller_note=$7, admin_note=$8, reviewer_id=$9, reviewed_at=$10, paid_at=$11
		WHERE id=$1`,
		d.ID, d.Status, d.GatewayID, d.TxID, d.CryptoCoin, d.ProofImage, d.ResellerNote, d.AdminNote,
		d.ReviewerID, d.ReviewedAt, d.PaidAt)
	return err
}

func (r *WalletBillingRepo) GetDeposit(ctx context.Context, id uuid.UUID) (*domain.WalletDeposit, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+depositCols+`
		FROM wallet_deposits d
		JOIN admins a ON a.id = d.admin_id
		LEFT JOIN wallet_packages p ON p.id = d.package_id
		WHERE d.id=$1`, id)
	return r.scanDeposit(row)
}

func (r *WalletBillingRepo) GetDepositByGatewayID(ctx context.Context, gatewayID string) (*domain.WalletDeposit, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT `+depositCols+`
		FROM wallet_deposits d
		JOIN admins a ON a.id = d.admin_id
		LEFT JOIN wallet_packages p ON p.id = d.package_id
		WHERE d.gateway_id=$1`, gatewayID)
	return r.scanDeposit(row)
}

func (r *WalletBillingRepo) ListDeposits(ctx context.Context, adminID *uuid.UUID, status *domain.WalletDepositStatus, limit int) ([]*domain.WalletDeposit, error) {
	if limit <= 0 {
		limit = 100
	}
	q := `SELECT ` + depositCols + `
		FROM wallet_deposits d
		JOIN admins a ON a.id = d.admin_id
		LEFT JOIN wallet_packages p ON p.id = d.package_id WHERE 1=1`
	args := []any{}
	n := 1
	if adminID != nil {
		q += ` AND d.admin_id = $` + strconv.Itoa(n)
		args = append(args, *adminID)
		n++
	}
	if status != nil {
		q += ` AND d.status = $` + strconv.Itoa(n)
		args = append(args, string(*status))
		n++
	}
	q += ` ORDER BY d.created_at DESC LIMIT $` + strconv.Itoa(n)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.WalletDeposit
	for rows.Next() {
		d, err := r.scanDeposit(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
