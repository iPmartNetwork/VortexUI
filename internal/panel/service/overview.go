package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// UserStatsRepo reports aggregate user counts. *postgres.UserRepo satisfies it
// (Stats is concrete-only; the rest of user access is on the port).
type UserStatsRepo interface {
	Stats(ctx context.Context) (domain.UserStats, error)
}

// overviewStaleAfter mirrors the subscription pruning window: a node silent
// longer than this is considered offline in the overview.
const overviewStaleAfter = 90 * time.Second

// OverviewService assembles the dashboard summary from persisted state alone
// (no live hub dependency): user aggregates from the users table and node
// connectivity derived from each node's last heartbeat, which the hub persists.
type OverviewService struct {
	users      UserStatsRepo
	nodes      port.NodeRepository
	widgets    *WidgetDeps
	staleAfter time.Duration
	now        func() time.Time
}

// NewOverviewService wires the service.
func NewOverviewService(users UserStatsRepo, nodes port.NodeRepository) *OverviewService {
	return &OverviewService{users: users, nodes: nodes, staleAfter: overviewStaleAfter, now: time.Now}
}

// SetWidgetDeps enables dashboard widgets (badges, trends, protocol breakdown).
func (s *OverviewService) SetWidgetDeps(dep WidgetDeps) {
	s.widgets = &dep
}

// Overview is the dashboard payload.
type Overview struct {
	Users   domain.UserStats     `json:"users"`
	Nodes   NodeSummary          `json:"nodes"`
	Widgets domain.DashboardWidgets `json:"widgets"`
}

// NodeSummary aggregates the fleet plus a per-node snapshot.
type NodeSummary struct {
	Total  int            `json:"total"`
	Online int            `json:"online"`
	Items  []NodeSnapshot `json:"items"`
}

// NodeSnapshot is a node's overview row: identity, derived connectivity, and its
// last reported health.
type NodeSnapshot struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Core        domain.CoreType   `json:"core"`
	Online      bool              `json:"online"`
	LastSeen    *time.Time        `json:"last_seen,omitempty"`
	Health      domain.NodeHealth `json:"health"`
	CoreVersion string            `json:"core_version,omitempty"`
	AgentVer    string            `json:"agent_version,omitempty"`
}

// Build computes the overview. Node connectivity uses the same staleness rule as
// subscriptions, so "online" here matches what clients are actually served.
func (s *OverviewService) Build(ctx context.Context, adminID *uuid.UUID, sudo bool) (*Overview, error) {
	stats, err := s.users.Stats(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := s.nodes.List(ctx)
	if err != nil {
		return nil, err
	}

	now := s.now()
	summary := NodeSummary{Total: len(nodes), Items: make([]NodeSnapshot, 0, len(nodes))}
	for _, n := range nodes {
		// Online on the dashboard means: a fresh heartbeat AND the core is
		// running — the same signal the Nodes page dot uses. We intentionally do
		// NOT use Node.Live here, because Live gives a never-polled node
		// (LastSeen==nil) the benefit of the doubt (correct for subscription
		// pruning, wrong here — it would show a node the panel has never reached
		// as "online").
		online := n.LastSeen != nil && now.Sub(*n.LastSeen) <= s.staleAfter && n.Health.CoreRunning
		if online {
			summary.Online++
		}
		summary.Items = append(summary.Items, NodeSnapshot{
			ID:          n.ID.String(),
			Name:        n.Name,
			Core:        n.Core,
			Online:      online,
			LastSeen:    n.LastSeen,
			Health:      n.Health,
			CoreVersion: n.CoreVer,
			AgentVer:    n.AgentVer,
		})
	}
	widgets := s.buildWidgets(ctx, summary, stats, adminID, sudo)
	return &Overview{Users: stats, Nodes: summary, Widgets: widgets}, nil
}
