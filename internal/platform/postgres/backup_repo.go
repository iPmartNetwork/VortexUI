package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/platform/postgres/db"
)

// BackupRepo restores a configuration snapshot transactionally. Export is done
// in the service layer via the read repositories; restore lives here because it
// needs a single transaction spanning every config table.
type BackupRepo struct{ pool *pgxpool.Pool }

// Restore replaces the entire proxy configuration with the snapshot in one
// transaction: it wipes the existing config (deleting nodes and users cascades
// to inbounds/outbounds/routing/balancers/bindings) and re-inserts everything in
// foreign-key order. On any error the transaction rolls back, leaving the
// previous configuration intact.
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
	users := &UserRepo{q: q}

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
	for _, u := range b.Users {
		if err := users.Create(ctx, u); err != nil {
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
