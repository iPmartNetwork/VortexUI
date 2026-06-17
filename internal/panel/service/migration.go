package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// MigrationService handles automatic user migration when nodes become unhealthy.
type MigrationService struct {
	repo     port.MigrationRepository
	inbounds port.InboundRepository
	nodes    port.NodeRepository
	users    port.UserRepository
	nodeOps  NodeOps
	log      *slog.Logger
	now      func() time.Time
}

// NewMigrationService wires the migration service. The inbounds parameter is
// used to resolve user→inbound bindings during failover. The nodeOps parameter
// (typically the hub) re-provisions users on the target node.
func NewMigrationService(inbounds port.InboundRepository, nodes port.NodeRepository, users port.UserRepository, nodeOps NodeOps, log *slog.Logger) *MigrationService {
	return &MigrationService{inbounds: inbounds, nodes: nodes, users: users, nodeOps: nodeOps, log: log, now: time.Now}
}

// SetRepo attaches the migration repository after construction (optional; if
// nil, events are not persisted but failover still works).
func (s *MigrationService) SetRepo(repo port.MigrationRepository) {
	s.repo = repo
}

// Migrate moves all users from the failed node to the target node. It is called
// by the hub's failover callback when a node becomes unreachable.
func (s *MigrationService) Migrate(ctx context.Context, failed, target *domain.Node) error {
	if target == nil {
		return fmt.Errorf("no healthy target node available")
	}
	if s.log != nil {
		s.log.Info("failover migration triggered", "from", failed.Name, "to", target.Name)
	}
	inbounds, err := s.inbounds.ListByNode(ctx, failed.ID)
	if err != nil {
		return fmt.Errorf("list inbounds for failed node: %w", err)
	}
	// Find matching inbounds on the target node (same tag/protocol).
	targetInbounds, err := s.inbounds.ListByNode(ctx, target.ID)
	if err != nil {
		return fmt.Errorf("list inbounds for target node: %w", err)
	}
	targetByTag := make(map[string]*domain.Inbound, len(targetInbounds))
	for _, in := range targetInbounds {
		targetByTag[in.Tag] = in
	}
	var migrated int
	for _, in := range inbounds {
		tin, ok := targetByTag[in.Tag]
		if !ok {
			if s.log != nil {
				s.log.Warn("no matching inbound on target, skipping", "tag", in.Tag, "target", target.Name)
			}
			continue
		}
		// Get users bound to this inbound via the user repository.
		users, _, err := s.users.List(ctx, port.UserFilter{Limit: 10000})
		if err != nil {
			if s.log != nil {
				s.log.Error("list users for migration failed", "error", err)
			}
			continue
		}
		for _, u := range users {
			uInbounds, err := s.users.InboundsFor(ctx, u.ID)
			if err != nil {
				continue
			}
			for _, uin := range uInbounds {
				if uin.ID == in.ID {
					// Bind user to the target inbound and provision on the target.
					if err := s.users.SetInbounds(ctx, u.ID, []uuid.UUID{tin.ID}); err != nil {
						if s.log != nil {
							s.log.Error("rebind user failed", "user", u.Username, "error", err)
						}
						continue
					}
					if err := s.nodeOps.AddUser(ctx, target.ID, tin.Tag, u); err != nil {
						if s.log != nil {
							s.log.Error("provision user on target failed", "user", u.Username, "error", err)
						}
					}
					migrated++
					s.saveEvent(ctx, u, failed, target, "node_failover")
					break
				}
			}
		}
	}
	if s.log != nil {
		s.log.Info("failover migration complete", "migrated", migrated, "from", failed.Name, "to", target.Name)
	}
	return nil
}

// MigrateBack returns users to their original node when it recovers. Called by
// the hub's on-connect callback after a node that previously failed comes back.
func (s *MigrationService) MigrateBack(ctx context.Context, recovered *domain.Node) error {
	if s.repo == nil {
		return nil
	}
	events, err := s.repo.ListByNode(ctx, recovered.ID)
	if err != nil {
		return fmt.Errorf("list migration events: %w", err)
	}
	// Find events where recovered was the source (from_node_id) and migrate
	// those users back.
	recoveredInbounds, err := s.inbounds.ListByNode(ctx, recovered.ID)
	if err != nil {
		return err
	}
	recoveredByTag := make(map[string]*domain.Inbound, len(recoveredInbounds))
	for _, in := range recoveredInbounds {
		recoveredByTag[in.Tag] = in
	}
	var returned int
	for _, ev := range events {
		if ev.FromNodeID != recovered.ID || ev.Status != domain.MigrationCompleted {
			continue
		}
		if ev.UserID == uuid.Nil {
			continue
		}
		u, err := s.users.GetByID(ctx, ev.UserID)
		if err != nil {
			continue
		}
		// Re-bind user back to the recovered node's inbounds.
		uInbounds, err := s.users.InboundsFor(ctx, u.ID)
		if err != nil {
			continue
		}
		for _, uin := range uInbounds {
			if rin, ok := recoveredByTag[uin.Tag]; ok {
				if err := s.users.SetInbounds(ctx, u.ID, []uuid.UUID{rin.ID}); err == nil {
					_ = s.nodeOps.AddUser(ctx, recovered.ID, rin.Tag, u)
					returned++
				}
				break
			}
		}
	}
	if returned > 0 {
		if s.log != nil {
			s.log.Info("migrate-back complete", "returned", returned, "node", recovered.Name)
		}
	}
	return nil
}

func (s *MigrationService) saveEvent(ctx context.Context, u *domain.User, from, to *domain.Node, reason string) {
	if s.repo == nil {
		return
	}
	event := &domain.MigrationEvent{
		ID:         uuid.New(),
		UserID:     u.ID,
		Username:   u.Username,
		FromNodeID: from.ID,
		ToNodeID:   to.ID,
		Reason:     reason,
		Status:     domain.MigrationCompleted,
		CreatedAt:  s.now(),
	}
	if err := s.repo.SaveEvent(ctx, event); err != nil {
		if s.log != nil {
			s.log.Error("failed to save migration event", "error", err)
		}
	}
}

// GetPolicy returns the current migration policy.
func (s *MigrationService) GetPolicy(ctx context.Context) (*domain.MigrationPolicy, error) {
	if s.repo == nil {
		def := domain.DefaultMigrationPolicy()
		return &def, nil
	}
	p, err := s.repo.GetPolicy(ctx)
	if err != nil {
		def := domain.DefaultMigrationPolicy()
		return &def, nil
	}
	return p, nil
}

// UpdatePolicy saves a new migration policy.
func (s *MigrationService) UpdatePolicy(ctx context.Context, p *domain.MigrationPolicy) error {
	if s.repo == nil {
		return fmt.Errorf("migration repository not configured")
	}
	return s.repo.SavePolicy(ctx, p)
}

// ListEvents returns recent migration events.
func (s *MigrationService) ListEvents(ctx context.Context, limit, offset int) ([]*domain.MigrationEvent, int, error) {
	if s.repo == nil {
		return nil, 0, nil
	}
	if limit <= 0 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, limit, offset)
}

// EvaluateNode checks if a node is unhealthy and should trigger migrations.
func (s *MigrationService) EvaluateNode(_ context.Context, node *domain.Node, policy *domain.MigrationPolicy) bool {
	if !policy.Enabled {
		return false
	}
	if node.Status != domain.NodeConnected {
		return true
	}
	if !node.Health.CoreRunning {
		return true
	}
	if policy.CPUThreshold > 0 && node.Health.CPUPercent > policy.CPUThreshold {
		return true
	}
	if policy.MemThreshold > 0 && node.Health.MemPercent > policy.MemThreshold {
		return true
	}
	return false
}

// SelectTargetNode picks the healthiest available node to migrate users to.
func (s *MigrationService) SelectTargetNode(ctx context.Context, excludeID uuid.UUID) (*domain.Node, error) {
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return nil, err
	}
	var best *domain.Node
	var bestScore float64
	for _, n := range nodes {
		if n.ID == excludeID {
			continue
		}
		if n.Status != domain.NodeConnected || !n.Health.CoreRunning {
			continue
		}
		score := (100 - n.Health.CPUPercent) + (100 - n.Health.MemPercent)
		if best == nil || score > bestScore {
			best = n
			bestScore = score
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no healthy target node available")
	}
	return best, nil
}
