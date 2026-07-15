package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// BackupRepo restores a configuration snapshot transactionally. Export is done
// in the service layer via the read repositories; restore lives here because it
// needs a single transaction spanning every config table.
type BackupRepo struct{ pool *pgxpool.Pool }

// ExportSupplemental reads business tables for JSON v3 backups.
func (r *BackupRepo) ExportSupplemental(ctx context.Context, includeTraffic bool) (*domain.BackupSupplemental, error) {
	tables, err := exportSupplementalTables(ctx, r.pool, includeTraffic)
	if err != nil {
		return nil, err
	}
	return &domain.BackupSupplemental{Tables: tables}, nil
}

// ExportAdminCredentials reads login secrets omitted from public Admin JSON.
func (r *BackupRepo) ExportAdminCredentials(ctx context.Context) ([]domain.BackupAdminCredential, error) {
	return exportAdminCredentials(ctx, r.pool)
}

// ExportFullArchive dumps the database and packs manifest + pg_dump into gzip tar.
func (r *BackupRepo) ExportFullArchive(ctx context.Context, databaseURL string, manifest domain.BackupManifest) ([]byte, error) {
	dump, err := DumpDatabase(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	return PackFullBackup(manifest, dump)
}

// UnpackFullArchive extracts manifest and dump from a gzip tar archive.
func (r *BackupRepo) UnpackFullArchive(data []byte) (domain.BackupManifest, []byte, error) {
	return UnpackFullBackup(data)
}


// RestoreFull replaces the entire database from a pg_dump custom-format archive.
func (r *BackupRepo) RestoreFull(ctx context.Context, databaseURL string, dump []byte) error {
	return RestoreDatabase(ctx, databaseURL, dump)
}

func (r *BackupRepo) ListAllWalletLedger(ctx context.Context) ([]domain.WalletLedgerEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, admin_id, delta_traffic, delta_users, reason, actor_admin_id, created_at
		FROM admin_wallet_ledger ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WalletLedgerEntry
	for rows.Next() {
		var e domain.WalletLedgerEntry
		if err := rows.Scan(&e.ID, &e.AdminID, &e.DeltaTraffic, &e.DeltaUsers, &e.Reason, &e.ActorAdminID, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ListOrdersForAdmin returns orders owned by or linked to a reseller's users.
func (r *BackupRepo) ListOrdersForAdmin(ctx context.Context, adminID uuid.UUID) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT o.id, o.user_id, o.admin_id, o.plan_id, o.username, o.status, o.gateway,
			o.gateway_id, o.amount, o.currency, o.created_at, o.paid_at, o.proof_image
		FROM orders o
		WHERE o.admin_id = $1 OR o.user_id IN (SELECT id FROM users WHERE admin_id = $1)
		ORDER BY o.created_at DESC`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Order
	for rows.Next() {
		var o domain.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.AdminID, &o.PlanID, &o.Username, &o.Status, &o.Gateway,
			&o.GatewayID, &o.Amount, &o.Currency, &o.CreatedAt, &o.PaidAt, &o.ProofImage); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}

// Restore replaces the entire proxy configuration with the snapshot in one
// transaction: it wipes users/nodes (cascading proxy config), re-applies operator
// records when present, then re-inserts everything in foreign-key order. On any
// error the transaction rolls back, leaving the previous configuration intact.
func (r *BackupRepo) Restore(ctx context.Context, b *domain.Backup) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after a successful Commit

	if err := wipeOrphanConfigData(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM users"); err != nil {
		return fmt.Errorf("wipe users: %w", err)
	}
	if _, err := tx.Exec(ctx, "DELETE FROM nodes"); err != nil {
		return fmt.Errorf("wipe nodes: %w", err)
	}

	q := db.New(tx)
	nodes := &NodeRepo{q: q}
	inbounds := &InboundRepo{q: q}
	outbounds := &OutboundRepo{q: q}
	routing := &RoutingRepo{q: q}
	balancers := &BalancerRepo{q: q}
	users := &UserRepo{q: q, pool: r.pool}
	admins := &AdminRepo{q: q, pool: r.pool}
	creds := credentialMap(b.AdminCredentials)

	for _, role := range b.Roles {
		if err := upsertRole(ctx, q, role); err != nil {
			return fmt.Errorf("restore role %s: %w", role.Name, err)
		}
	}
	for _, a := range sortAdminsParentsFirst(b.Admins) {
		applyAdminCredentials(a, creds)
		if err := upsertAdmin(ctx, q, admins, a); err != nil {
			return fmt.Errorf("restore admin %s: %w", a.Username, err)
		}
	}
	for _, p := range b.Plans {
		if err := upsertPlan(ctx, tx, p); err != nil {
			return fmt.Errorf("restore plan %s: %w", p.Name, err)
		}
	}

	for _, n := range b.Nodes {
		if err := nodes.Create(ctx, n); err != nil {
			return fmt.Errorf("restore node %s: %w", n.Name, err)
		}
	}
	for _, in := range b.Inbounds {
		if err := inbounds.Create(ctx, in); err != nil {
			return fmt.Errorf("restore inbound %s: %w", in.Tag, err)
		}
	}
	for _, o := range b.Outbounds {
		if err := outbounds.Create(ctx, o); err != nil {
			return fmt.Errorf("restore outbound %s: %w", o.Tag, err)
		}
	}
	for _, rule := range b.Routing {
		if err := routing.Create(ctx, rule); err != nil {
			return fmt.Errorf("restore routing rule: %w", err)
		}
	}
	for _, bl := range b.Balancers {
		if err := balancers.Create(ctx, bl); err != nil {
			return fmt.Errorf("restore balancer %s: %w", bl.Tag, err)
		}
	}

	for _, scope := range b.AdminScopes {
		if err := restoreAdminScope(ctx, q, scope); err != nil {
			return fmt.Errorf("restore admin scope for %s: %w", scope.AdminID, err)
		}
	}
	for _, branding := range b.PortalBranding {
		if err := admins.UpsertPortalBranding(ctx, branding); err != nil {
			return fmt.Errorf("restore portal branding: %w", err)
		}
	}

	for _, u := range b.Users {
		restored := *u
		if restored.AdminID != nil && !adminExists(ctx, q, *restored.AdminID) {
			if len(b.Admins) == 0 {
				restored.AdminID = nil
			} else {
				return fmt.Errorf("restore user %s: admin %s missing from backup", u.Username, restored.AdminID)
			}
		}
		if err := users.Create(ctx, &restored); err != nil {
			return fmt.Errorf("restore user %s: %w", u.Username, err)
		}
	}
	for _, bd := range b.Bindings {
		if err := q.AddUserInbound(ctx, db.AddUserInboundParams{UserID: bd.UserID, InboundID: bd.InboundID}); err != nil {
			return fmt.Errorf("restore binding: %w", err)
		}
	}
	if err := restoreSupplementalTables(ctx, tx, b.Supplemental); err != nil {
		return fmt.Errorf("restore supplemental: %w", err)
	}
	return tx.Commit(ctx)
}

// RestoreReseller applies a scoped reseller backup (users, bindings, wallet, orders).
func (r *BackupRepo) RestoreReseller(ctx context.Context, rb *domain.ResellerBackup) error {
	if rb == nil {
		return errors.New("empty reseller backup")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := db.New(tx)
	users := &UserRepo{q: q, pool: r.pool}
	admins := &AdminRepo{q: q, pool: r.pool}

	for _, u := range rb.Users {
		if u.AdminID != nil && *u.AdminID != rb.AdminID {
			return fmt.Errorf("user %s belongs to another admin", u.Username)
		}
		restored := *u
		restored.AdminID = &rb.AdminID
		if _, err := q.GetUserByID(ctx, restored.ID); err != nil {
			if !errors.Is(mapErr(err), domain.ErrNotFound) {
				return mapErr(err)
			}
			if err := users.Create(ctx, &restored); err != nil {
				return fmt.Errorf("create user %s: %w", u.Username, err)
			}
		} else if err := users.Update(ctx, &restored); err != nil {
			return fmt.Errorf("update user %s: %w", u.Username, err)
		}
	}

	for _, u := range rb.Users {
		inIDs := make([]uuid.UUID, 0)
		for _, bd := range rb.Bindings {
			if bd.UserID == u.ID {
				inIDs = append(inIDs, bd.InboundID)
			}
		}
		if err := users.SetInbounds(ctx, u.ID, inIDs); err != nil {
			return fmt.Errorf("restore bindings for %s: %w", u.Username, err)
		}
	}

	if rb.PaymentConfig != nil {
		cfg := rb.PaymentConfig
		cryptoJSON, err := json.Marshal(cfg.CryptoAddresses)
		if err != nil {
			return fmt.Errorf("payment config crypto: %w", err)
		}
		methodsJSON, err := json.Marshal(cfg.EnabledMethods)
		if err != nil {
			return fmt.Errorf("payment config methods: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO reseller_payment_config
			   (admin_id, card_number, card_holder, card_bank, crypto_addresses,
			    zarinpal_merchant_id, manual_instructions, enabled_methods)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (admin_id) DO UPDATE SET
			   card_number = EXCLUDED.card_number,
			   card_holder = EXCLUDED.card_holder,
			   card_bank = EXCLUDED.card_bank,
			   crypto_addresses = EXCLUDED.crypto_addresses,
			   zarinpal_merchant_id = EXCLUDED.zarinpal_merchant_id,
			   manual_instructions = EXCLUDED.manual_instructions,
			   enabled_methods = EXCLUDED.enabled_methods`,
			cfg.AdminID, cfg.CardNumber, cfg.CardHolder, cfg.CardBank,
			cryptoJSON, cfg.ZarinpalMerchantID, cfg.ManualInstructions, methodsJSON); err != nil {
			return fmt.Errorf("restore payment config: %w", err)
		}
	}
	if rb.PortalBranding != nil {
		if err := admins.UpsertPortalBranding(ctx, rb.PortalBranding); err != nil {
			return fmt.Errorf("restore portal branding: %w", err)
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM admin_wallet_ledger WHERE admin_id = $1`, rb.AdminID); err != nil {
		return fmt.Errorf("wipe reseller ledger: %w", err)
	}
	for _, entry := range rb.WalletLedger {
		if entry.AdminID != rb.AdminID {
			continue
		}
		if err := q.InsertWalletLedger(ctx, db.InsertWalletLedgerParams{
			ID: entry.ID, AdminID: entry.AdminID, DeltaTraffic: entry.DeltaTraffic,
			DeltaUsers: int32(entry.DeltaUsers), Reason: entry.Reason,
			ActorAdminID: ptrToUUID(entry.ActorAdminID), CreatedAt: timeToTS(entry.CreatedAt),
		}); err != nil {
			return fmt.Errorf("restore ledger entry: %w", err)
		}
	}

	for _, o := range rb.Orders {
		adminID := rb.AdminID
		o.AdminID = &adminID
		if _, err := tx.Exec(ctx, `
			INSERT INTO orders (id, user_id, admin_id, plan_id, username, status, gateway,
				gateway_id, amount, currency, created_at, paid_at, proof_image)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			ON CONFLICT (id) DO UPDATE SET
				user_id = EXCLUDED.user_id, admin_id = EXCLUDED.admin_id, status = EXCLUDED.status,
				gateway = EXCLUDED.gateway, gateway_id = EXCLUDED.gateway_id, amount = EXCLUDED.amount,
				currency = EXCLUDED.currency, paid_at = EXCLUDED.paid_at, proof_image = EXCLUDED.proof_image`,
			o.ID, o.UserID, o.AdminID, o.PlanID, o.Username, o.Status, o.Gateway,
			o.GatewayID, o.Amount, o.Currency, o.CreatedAt, o.PaidAt, o.ProofImage); err != nil {
			return fmt.Errorf("restore order %s: %w", o.ID, err)
		}
	}
	return tx.Commit(ctx)
}

func credentialMap(creds []domain.BackupAdminCredential) map[uuid.UUID]domain.BackupAdminCredential {
	out := make(map[uuid.UUID]domain.BackupAdminCredential, len(creds))
	for _, c := range creds {
		out[c.AdminID] = c
	}
	return out
}

func applyAdminCredentials(a *domain.Admin, creds map[uuid.UUID]domain.BackupAdminCredential) {
	if a == nil {
		return
	}
	c, ok := creds[a.ID]
	if !ok {
		return
	}
	if a.PasswordHash == "" && c.PasswordHash != "" {
		a.PasswordHash = c.PasswordHash
	}
	if a.TOTPSecret == "" && c.TOTPSecret != "" {
		a.TOTPSecret = c.TOTPSecret
	}
	if a.WebhookSecret == "" && c.WebhookSecret != "" {
		a.WebhookSecret = c.WebhookSecret
	}
}

func restoreAdminScope(ctx context.Context, q *db.Queries, scope domain.BackupAdminScope) error {
	if err := q.ClearAdminInbounds(ctx, scope.AdminID); err != nil {
		return err
	}
	for _, inID := range scope.InboundIDs {
		if err := q.AddAdminInbound(ctx, db.AddAdminInboundParams{AdminID: scope.AdminID, InboundID: inID}); err != nil {
			return err
		}
	}
	if err := q.ClearAdminPlans(ctx, scope.AdminID); err != nil {
		return err
	}
	for _, planID := range scope.PlanIDs {
		if err := q.AddAdminPlan(ctx, db.AddAdminPlanParams{AdminID: scope.AdminID, PlanID: planID}); err != nil {
			return err
		}
	}
	if err := q.ClearAdminNodes(ctx, scope.AdminID); err != nil {
		return err
	}
	for _, nodeID := range scope.NodeIDs {
		if err := q.AddAdminNode(ctx, db.AddAdminNodeParams{AdminID: scope.AdminID, NodeID: nodeID}); err != nil {
			return err
		}
	}
	return nil
}

func adminExists(ctx context.Context, q *db.Queries, id uuid.UUID) bool {
	_, err := q.GetAdminByID(ctx, id)
	return err == nil
}

func upsertRole(ctx context.Context, q *db.Queries, role *domain.Role) error {
	perms, err := jsonPerms(role.Permissions)
	if err != nil {
		return err
	}
	if err := q.CreateRole(ctx, db.CreateRoleParams{ID: role.ID, Name: role.Name, Permissions: perms}); err != nil {
		if !isUniqueViolation(err) {
			return err
		}
	}
	return q.UpdateRole(ctx, db.UpdateRoleParams{ID: role.ID, Name: role.Name, Permissions: perms})
}

func upsertAdmin(ctx context.Context, q *db.Queries, admins *AdminRepo, a *domain.Admin) error {
	if _, err := q.GetAdminByID(ctx, a.ID); err != nil {
		if !errors.Is(mapErr(err), domain.ErrNotFound) {
			return mapErr(err)
		}
		if err := admins.Create(ctx, a); err != nil && !isUniqueViolation(err) {
			return err
		}
	}
	if err := admins.Update(ctx, a); err != nil {
		return err
	}
	if a.Suspended {
		at := time.Now()
		if a.SuspendedAt != nil {
			at = *a.SuspendedAt
		}
		if err := admins.Suspend(ctx, a.ID, at, a.SuspendReason); err != nil {
			return err
		}
	} else {
		if err := admins.Unsuspend(ctx, a.ID); err != nil {
			return err
		}
	}
	return admins.SetQuotaBreachedAt(ctx, a.ID, a.QuotaBreachedAt)
}

func upsertPlan(ctx context.Context, tx pgx.Tx, p *domain.Plan) error {
	idsJSON, err := json.Marshal(p.InboundIDs)
	if err != nil {
		return err
	}
	if p.InboundIDs == nil {
		idsJSON = []byte("[]")
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO plans (id, name, description, data_limit, duration_days, device_limit,
			reset_strategy, inbound_ids, price_toman, price_usd, max_users, enabled, created_at, admin_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			data_limit = EXCLUDED.data_limit,
			duration_days = EXCLUDED.duration_days,
			device_limit = EXCLUDED.device_limit,
			reset_strategy = EXCLUDED.reset_strategy,
			inbound_ids = EXCLUDED.inbound_ids,
			price_toman = EXCLUDED.price_toman,
			price_usd = EXCLUDED.price_usd,
			max_users = EXCLUDED.max_users,
			enabled = EXCLUDED.enabled,
			admin_id = EXCLUDED.admin_id`,
		p.ID, p.Name, p.Description, p.DataLimit, p.Duration, p.DeviceLimit,
		p.ResetStrategy, idsJSON, p.PriceToman, p.PriceUSD, p.MaxUsers, p.Enabled, p.CreatedAt, p.AdminID)
	return err
}

func sortAdminsParentsFirst(admins []*domain.Admin) []*domain.Admin {
	if len(admins) == 0 {
		return nil
	}
	byID := make(map[uuid.UUID]*domain.Admin, len(admins))
	for _, a := range admins {
		byID[a.ID] = a
	}
	added := make(map[uuid.UUID]bool, len(admins))
	out := make([]*domain.Admin, 0, len(admins))
	for len(out) < len(admins) {
		progress := false
		for _, a := range admins {
			if added[a.ID] {
				continue
			}
			parentReady := a.ParentAdminID == nil || added[*a.ParentAdminID] || byID[*a.ParentAdminID] == nil
			if parentReady {
				out = append(out, a)
				added[a.ID] = true
				progress = true
			}
		}
		if !progress {
			break
		}
	}
	for _, a := range admins {
		if !added[a.ID] {
			out = append(out, a)
		}
	}
	return out
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func jsonPerms(perms []domain.Permission) ([]byte, error) {
	if perms == nil {
		perms = []domain.Permission{}
	}
	return json.Marshal(perms)
}
