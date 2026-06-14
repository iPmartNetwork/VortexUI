package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BackupRestorer applies a snapshot transactionally. *postgres.BackupRepo
// satisfies it; the destructive replace lives in the storage layer where a
// single transaction can span every table.
type BackupRestorer interface {
	Restore(ctx context.Context, b *domain.Backup) error
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
	restorer  BackupRestorer
	now       func() time.Time
}

// NewBackupService wires the service.
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

const backupUserPage = 500

// Export assembles a complete configuration snapshot. Per-node config is
// gathered node by node; users are paged through; bindings are derived from each
// user's inbound memberships.
func (s *BackupService) Export(ctx context.Context) (*domain.Backup, error) {
	b := &domain.Backup{Version: domain.BackupVersion, ExportedAt: s.now()}

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
	if b.Version != domain.BackupVersion {
		return fmt.Errorf("unsupported backup version %d (want %d)", b.Version, domain.BackupVersion)
	}
	return s.restorer.Restore(ctx, b)
}
