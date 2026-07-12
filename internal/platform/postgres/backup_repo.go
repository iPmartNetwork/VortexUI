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

	for _, role := range b.Roles {
		if err := upsertRole(ctx, q, role); err != nil {
			return fmt.Errorf("restore role %s: %w", role.Name, err)
		}
	}
	for _, a := range sortAdminsParentsFirst(b.Admins) {
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
	return tx.Commit(ctx)
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
