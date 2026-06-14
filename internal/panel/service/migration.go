package service

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// InboundBinder attaches a user to an inbound additively (idempotent).
// *postgres.UserRepo satisfies it.
type InboundBinder interface {
	BindInbound(ctx context.Context, userID, inboundID uuid.UUID) error
}

// InboundUnbinder removes a single user→inbound binding. *postgres.UserRepo
// satisfies it.
type InboundUnbinder interface {
	UnbindInbound(ctx context.Context, userID, inboundID uuid.UUID) error
}

// migrationRecord remembers one temporary failover placement so it can be undone
// when the home node recovers.
type migrationRecord struct {
	userID        uuid.UUID
	targetNodeID  uuid.UUID
	targetInbound uuid.UUID
	targetTag     string
}

// MigrationService moves a failed node's users onto a healthy target node. It is
// the concrete action behind the hub's failover hook.
//
// Migration requires the target to mirror the failed node's inbounds by tag: a
// user on the failed node's "vless-ws" inbound is re-homed onto the target's
// "vless-ws" inbound. Users on inbounds the target lacks are skipped (logged),
// since there is nowhere equivalent to send them.
type MigrationService struct {
	inbounds InboundLister   // ListByNode
	users    UsersByNoder    // UsersByNode
	binder   InboundBinder   // BindInbound
	unbinder InboundUnbinder // UnbindInbound (for migrate-back)
	nodes    NodeOps         // AddUser / RemoveUser on the target
	log      *slog.Logger

	// migrated records temporary placements per failed node, so MigrateBack can
	// undo them on recovery. In-memory: a panel restart loses pending records
	// (the temp bindings then linger until manually cleared) — acceptable for a
	// transient failover state.
	mu       sync.Mutex
	migrated map[uuid.UUID][]migrationRecord
}

// NewMigrationService wires the service.
func NewMigrationService(inbounds InboundLister, users UsersByNoder, binder InboundBinder, unbinder InboundUnbinder, nodes NodeOps, log *slog.Logger) *MigrationService {
	if log == nil {
		log = slog.Default()
	}
	return &MigrationService{
		inbounds: inbounds, users: users, binder: binder, unbinder: unbinder,
		nodes: nodes, log: log, migrated: map[uuid.UUID][]migrationRecord{},
	}
}

// Migrate re-homes every user of the failed node onto the target's same-tag
// inbounds: it binds the user there (so subscriptions point at the healthy node)
// and provisions them live. Best-effort per user so one failure does not abort
// the rest; returns the joined errors for observability.
func (m *MigrationService) Migrate(ctx context.Context, failed, target *domain.Node) error {
	if target == nil {
		return errors.New("failover: no healthy target available")
	}
	targetInbounds, err := m.inbounds.ListByNode(ctx, target.ID)
	if err != nil {
		return err
	}
	targetByTag := make(map[string]*domain.Inbound, len(targetInbounds))
	for _, in := range targetInbounds {
		targetByTag[in.Tag] = in
	}

	usersByTag, err := m.users.UsersByNode(ctx, failed.ID)
	if err != nil {
		return err
	}

	var errs []error
	var records []migrationRecord
	for tag, users := range usersByTag {
		dst, ok := targetByTag[tag]
		if !ok {
			m.log.Warn("failover: target lacks matching inbound, users not migrated",
				"tag", tag, "target", target.Name, "count", len(users))
			continue
		}
		for _, u := range users {
			if err := m.binder.BindInbound(ctx, u.ID, dst.ID); err != nil {
				errs = append(errs, err)
				continue
			}
			if err := m.nodes.AddUser(ctx, target.ID, dst.Tag, u); err != nil {
				errs = append(errs, err)
				continue
			}
			records = append(records, migrationRecord{userID: u.ID, targetNodeID: target.ID, targetInbound: dst.ID, targetTag: dst.Tag})
		}
	}
	if len(records) > 0 {
		m.mu.Lock()
		m.migrated[failed.ID] = append(m.migrated[failed.ID], records...)
		m.mu.Unlock()
	}
	m.log.Info("failover migration complete", "failed", failed.Name, "target", target.Name, "users", len(records))
	return errors.Join(errs...)
}

// MigrateBack undoes the temporary placements made when `recovered` failed: it
// unbinds the migrated users from the target inbound and removes them from the
// target's live core. The recovered node is repopulated separately by the resync
// that the hub fires on reconnect, so this just sheds the now-redundant copies.
func (m *MigrationService) MigrateBack(ctx context.Context, recovered *domain.Node) error {
	m.mu.Lock()
	records := m.migrated[recovered.ID]
	delete(m.migrated, recovered.ID)
	m.mu.Unlock()
	if len(records) == 0 {
		return nil
	}

	var errs []error
	for _, r := range records {
		if err := m.unbinder.UnbindInbound(ctx, r.userID, r.targetInbound); err != nil {
			errs = append(errs, err)
		}
		if err := m.nodes.RemoveUser(ctx, r.targetNodeID, r.targetTag, r.userID); err != nil {
			errs = append(errs, err)
		}
	}
	m.log.Info("migrate-back complete", "recovered", recovered.Name, "users", len(records))
	return errors.Join(errs...)
}
