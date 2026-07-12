package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BackupRestorer applies a snapshot transactionally. *postgres.BackupRepo
// satisfies it; the destructive replace lives in the storage layer where a
// single transaction can span every table.
type BackupRestorer interface {
	Restore(ctx context.Context, b *domain.Backup) error
}

// backupAdminSource exports operator/reseller records needed to restore user
// ownership on a new panel.
type backupAdminSource interface {
	List(ctx context.Context) ([]*domain.Admin, error)
	ListRoles(ctx context.Context) ([]*domain.Role, error)
	ListInboundIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	ListPlanIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	ListNodeIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	GetPortalBranding(ctx context.Context, adminID uuid.UUID) (*domain.PortalBranding, error)
}

// backupPlanSource exports shop plans referenced by resellers.
type backupPlanSource interface {
	ListPlans(ctx context.Context) ([]*domain.Plan, error)
}

// BackupService exports the full proxy configuration as a portable document and
// restores one back. Export composes the read repositories; restore validates
// the document and delegates the transactional replace to the storage layer.
type BackupService struct {
	nodes     port.NodeRepository
	inbounds  port.InboundRepository
	outbounds port.OutboundRepository
	routing   port.RoutingRepository
	balancers port.BalancerRepository
	users     port.UserRepository
	admins    backupAdminSource
	plans     backupPlanSource
	restorer  BackupRestorer
	now       func() time.Time
}

// NewBackupService wires the service. admins and plans may be nil; export then
// omits operator/reseller records (legacy v1-shaped documents).
func NewBackupService(
	nodes port.NodeRepository,
	inbounds port.InboundRepository,
	outbounds port.OutboundRepository,
	routing port.RoutingRepository,
	balancers port.BalancerRepository,
	users port.UserRepository,
	restorer BackupRestorer,
) *BackupService {
	return &BackupService{
		nodes: nodes, inbounds: inbounds, outbounds: outbounds,
		routing: routing, balancers: balancers, users: users,
		restorer: restorer, now: time.Now,
	}
}

// SetAdminSource wires reseller/admin export for full migration backups.
func (s *BackupService) SetAdminSource(src backupAdminSource) { s.admins = src }

// SetPlanSource wires shop plan export for full migration backups.
func (s *BackupService) SetPlanSource(src backupPlanSource) { s.plans = src }

const backupUserPage = 500

// Export assembles a complete configuration snapshot. Per-node config is
// gathered node by node; users are paged through; bindings are derived from each
// user's inbound memberships. When wired, admins/resellers and their scopes are
// included so restore on a fresh panel preserves ownership.
func (s *BackupService) Export(ctx context.Context) (*domain.Backup, error) {
	b := &domain.Backup{Version: domain.BackupVersion, ExportedAt: s.now()}

	if s.admins != nil {
		roles, err := s.admins.ListRoles(ctx)
		if err != nil {
			return nil, fmt.Errorf("list roles: %w", err)
		}
		b.Roles = roles

		admins, err := s.admins.List(ctx)
		if err != nil {
			return nil, fmt.Errorf("list admins: %w", err)
		}
		b.Admins = admins
		for _, a := range admins {
			inIDs, err := s.admins.ListInboundIDs(ctx, a.ID)
			if err != nil {
				return nil, fmt.Errorf("list admin inbounds for %s: %w", a.Username, err)
			}
			planIDs, err := s.admins.ListPlanIDs(ctx, a.ID)
			if err != nil {
				return nil, fmt.Errorf("list admin plans for %s: %w", a.Username, err)
			}
			nodeIDs, err := s.admins.ListNodeIDs(ctx, a.ID)
			if err != nil {
				return nil, fmt.Errorf("list admin nodes for %s: %w", a.Username, err)
			}
			if len(inIDs)+len(planIDs)+len(nodeIDs) > 0 {
				b.AdminScopes = append(b.AdminScopes, domain.BackupAdminScope{
					AdminID: a.ID, InboundIDs: inIDs, PlanIDs: planIDs, NodeIDs: nodeIDs,
				})
			}
			if branding, err := s.admins.GetPortalBranding(ctx, a.ID); err == nil && branding != nil {
				b.PortalBranding = append(b.PortalBranding, branding)
			}
		}
	}

	if s.plans != nil {
		plans, err := s.plans.ListPlans(ctx)
		if err != nil {
			return nil, fmt.Errorf("list plans: %w", err)
		}
		b.Plans = plans
	}

	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	b.Nodes = nodes
	for _, n := range nodes {
		ins, err := s.inbounds.ListByNode(ctx, n.ID)
		if err != nil {
			return nil, fmt.Errorf("list inbounds for %s: %w", n.Name, err)
		}
		b.Inbounds = append(b.Inbounds, ins...)

		outs, err := s.outbounds.ListByNode(ctx, n.ID)
		if err != nil {
			return nil, fmt.Errorf("list outbounds for %s: %w", n.Name, err)
		}
		b.Outbounds = append(b.Outbounds, outs...)

		rules, err := s.routing.ListByNode(ctx, n.ID)
		if err != nil {
			return nil, fmt.Errorf("list routing for %s: %w", n.Name, err)
		}
		b.Routing = append(b.Routing, rules...)

		bals, err := s.balancers.ListByNode(ctx, n.ID)
		if err != nil {
			return nil, fmt.Errorf("list balancers for %s: %w", n.Name, err)
		}
		b.Balancers = append(b.Balancers, bals...)
	}

	for offset := 0; ; offset += backupUserPage {
		page, total, err := s.users.List(ctx, port.UserFilter{Limit: backupUserPage, Offset: offset})
		if err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		b.Users = append(b.Users, page...)
		if len(page) == 0 || offset+len(page) >= total {
			break
		}
	}

	for _, u := range b.Users {
		ins, err := s.users.InboundsFor(ctx, u.ID)
		if err != nil {
			return nil, fmt.Errorf("bindings for %s: %w", u.Username, err)
		}
		for _, in := range ins {
			b.Bindings = append(b.Bindings, domain.UserProxy{UserID: u.ID, InboundID: in.ID})
		}
	}

	return b, nil
}

// Restore validates the document version and applies it, replacing the current
// configuration. Live cores are reconciled when nodes next (re)connect (the hub
// resyncs on connect); a panel restart guarantees the same.
func (s *BackupService) Restore(ctx context.Context, b *domain.Backup) error {
	if b == nil {
		return errors.New("empty backup")
	}
	if b.Version != domain.BackupVersion && b.Version != domain.BackupVersionLegacy {
		return fmt.Errorf("unsupported backup version %d (want %d or %d)", b.Version, domain.BackupVersion, domain.BackupVersionLegacy)
	}
	return s.restorer.Restore(ctx, b)
}
