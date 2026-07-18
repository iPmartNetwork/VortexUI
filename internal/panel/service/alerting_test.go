package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- test doubles ---

type stubNodeRepo struct {
	nodes []*domain.Node
}

func (r *stubNodeRepo) Create(_ context.Context, _ *domain.Node) error                       { return nil }
func (r *stubNodeRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.Node, error)         { return nil, nil }
func (r *stubNodeRepo) Update(_ context.Context, _ *domain.Node) error                       { return nil }
func (r *stubNodeRepo) Delete(_ context.Context, _ uuid.UUID) error                          { return nil }
func (r *stubNodeRepo) List(_ context.Context) ([]*domain.Node, error)                       { return r.nodes, nil }
func (r *stubNodeRepo) UpdateHealth(_ context.Context, _ uuid.UUID, _ domain.NodeHealth) error { return nil }

type stubUserRepo struct {
	users []*domain.User
}

func (r *stubUserRepo) Create(_ context.Context, _ *domain.User) error                 { return nil }
func (r *stubUserRepo) GetByID(_ context.Context, _ uuid.UUID) (*domain.User, error)   { return nil, nil }
func (r *stubUserRepo) GetBySubToken(_ context.Context, _ string) (*domain.User, error) { return nil, nil }
func (r *stubUserRepo) Update(_ context.Context, _ *domain.User) error                 { return nil }
func (r *stubUserRepo) Delete(_ context.Context, _ uuid.UUID) error                    { return nil }
func (r *stubUserRepo) List(_ context.Context, _ port.UserFilter) ([]*domain.User, int, error) {
	return r.users, len(r.users), nil
}
func (r *stubUserRepo) AddUsedTraffic(_ context.Context, _ uuid.UUID, _ int64) error         { return nil }
func (r *stubUserRepo) AddUsedTrafficBatch(_ context.Context, _ map[uuid.UUID]int64) error   { return nil }
func (r *stubUserRepo) SetInbounds(_ context.Context, _ uuid.UUID, _ []uuid.UUID) error      { return nil }
func (r *stubUserRepo) InboundsFor(_ context.Context, _ uuid.UUID) ([]domain.Inbound, error) { return nil, nil }
func (r *stubUserRepo) PrimaryInboundProtocols(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]string, error) {
	return nil, nil
}
func (r *stubUserRepo) StatsForAdmin(_ context.Context, _ uuid.UUID) (domain.AdminUserStats, error) {
	return domain.AdminUserStats{}, nil
}

type collectingPublisher struct {
	events []events.Event
}

func (p *collectingPublisher) Publish(e events.Event) {
	p.events = append(p.events, e)
}

// --- tests ---

func TestAlertingService_NodeCPUAlert(t *testing.T) {
	nodeID := uuid.New()
	nodes := &stubNodeRepo{nodes: []*domain.Node{
		{ID: nodeID, Name: "test-node", Health: domain.NodeHealth{CPUPercent: 95, CoreRunning: true}},
	}}
	users := &stubUserRepo{}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if pub.events[0].Type != events.NodeAlert {
		t.Errorf("expected type %s, got %s", events.NodeAlert, pub.events[0].Type)
	}
	if pub.events[0].NodeID != nodeID.String() {
		t.Errorf("expected node_id %s, got %s", nodeID.String(), pub.events[0].NodeID)
	}
}

func TestAlertingService_NodeMemoryAlert(t *testing.T) {
	nodeID := uuid.New()
	nodes := &stubNodeRepo{nodes: []*domain.Node{
		{ID: nodeID, Name: "mem-node", Health: domain.NodeHealth{MemPercent: 92, CoreRunning: true}},
	}}
	users := &stubUserRepo{}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if pub.events[0].Type != events.NodeAlert {
		t.Errorf("expected type %s, got %s", events.NodeAlert, pub.events[0].Type)
	}
}

func TestAlertingService_NodeOfflineAlert(t *testing.T) {
	nodeID := uuid.New()
	tenMinAgo := time.Now().Add(-10 * time.Minute)
	nodes := &stubNodeRepo{nodes: []*domain.Node{
		{ID: nodeID, Name: "offline-node", LastSeen: &tenMinAgo, Health: domain.NodeHealth{CoreRunning: true}},
	}}
	users := &stubUserRepo{}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if pub.events[0].Type != events.NodeDown {
		t.Errorf("expected type %s, got %s", events.NodeDown, pub.events[0].Type)
	}
}

func TestAlertingService_UserQuotaAlert(t *testing.T) {
	userID := uuid.New()
	nodes := &stubNodeRepo{}
	users := &stubUserRepo{users: []*domain.User{
		{ID: userID, Username: "heavyuser", DataLimit: 1000, UsedTraffic: 850},
	}}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(pub.events))
	}
	if pub.events[0].Type != events.UserQuotaWarn {
		t.Errorf("expected type %s, got %s", events.UserQuotaWarn, pub.events[0].Type)
	}
	if pub.events[0].Username != "heavyuser" {
		t.Errorf("expected username heavyuser, got %s", pub.events[0].Username)
	}
}

func TestAlertingService_NoAlertBelowThreshold(t *testing.T) {
	nodes := &stubNodeRepo{nodes: []*domain.Node{
		{ID: uuid.New(), Name: "healthy", Health: domain.NodeHealth{CPUPercent: 50, MemPercent: 60, CoreRunning: true}},
	}}
	users := &stubUserRepo{users: []*domain.User{
		{ID: uuid.New(), Username: "lightuser", DataLimit: 1000, UsedTraffic: 100},
	}}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 0 {
		t.Fatalf("expected 0 events, got %d: %v", len(pub.events), pub.events)
	}
}

func TestAlertingService_CooldownPreventsRepeat(t *testing.T) {
	nodeID := uuid.New()
	nodes := &stubNodeRepo{nodes: []*domain.Node{
		{ID: nodeID, Name: "hot-node", Health: domain.NodeHealth{CPUPercent: 95, CoreRunning: true}},
	}}
	users := &stubUserRepo{}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	// First evaluation — fires
	svc.evaluateAll(context.Background())
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event on first eval, got %d", len(pub.events))
	}

	// Second evaluation — cooldown prevents re-fire
	svc.evaluateAll(context.Background())
	if len(pub.events) != 1 {
		t.Fatalf("expected still 1 event after second eval (cooldown), got %d", len(pub.events))
	}
}

func TestAlertingService_UnlimitedUserSkipped(t *testing.T) {
	nodes := &stubNodeRepo{}
	users := &stubUserRepo{users: []*domain.User{
		{ID: uuid.New(), Username: "unlimited", DataLimit: 0, UsedTraffic: 999999},
	}}
	pub := &collectingPublisher{}

	svc := NewAlertingService(nodes, users, nil, nil)
	svc.pub = pub

	svc.evaluateAll(context.Background())

	if len(pub.events) != 0 {
		t.Fatalf("expected 0 events for unlimited user, got %d", len(pub.events))
	}
}
