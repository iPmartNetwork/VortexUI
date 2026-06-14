package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// mapNodeRepo serves nodes from a map for multi-node subscription tests.
type mapNodeRepo struct{ nodes map[uuid.UUID]*domain.Node }

func (m *mapNodeRepo) Create(context.Context, *domain.Node) error { return nil }
func (m *mapNodeRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Node, error) {
	if n, ok := m.nodes[id]; ok {
		return n, nil
	}
	return nil, domain.ErrNotFound
}
func (m *mapNodeRepo) Update(context.Context, *domain.Node) error   { return nil }
func (m *mapNodeRepo) Delete(context.Context, uuid.UUID) error      { return nil }
func (m *mapNodeRepo) List(context.Context) ([]*domain.Node, error) { return nil, nil }
func (m *mapNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error {
	return nil
}

func TestSubscriptionPrunesUnhealthyNodes(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-5 * time.Second)
	stale := now.Add(-30 * time.Minute)

	liveID, deadID := uuid.New(), uuid.New()
	user := &domain.User{ID: uuid.New(), Username: "alice", SubToken: "tok",
		Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}

	userRepo := &fakeUserRepo{
		created: user,
		inbounds: []domain.Inbound{
			{ID: uuid.New(), NodeID: liveID, Tag: "live-in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true},
			{ID: uuid.New(), NodeID: deadID, Tag: "dead-in", Protocol: domain.ProtoVLESS, Port: 443, Enabled: true},
		},
	}
	nodeRepo := &mapNodeRepo{nodes: map[uuid.UUID]*domain.Node{
		liveID: {ID: liveID, Name: "live", Address: "1.1.1.1:50051", LastSeen: &fresh, Health: domain.NodeHealth{CoreRunning: true}},
		deadID: {ID: deadID, Name: "dead", Address: "2.2.2.2:50051", LastSeen: &stale, Health: domain.NodeHealth{CoreRunning: true}},
	}}

	svc := NewSubscriptionService(userRepo, nodeRepo)
	svc.now = func() time.Time { return now }

	res, err := svc.Build(context.Background(), "tok")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if len(res.Proxies) != 1 {
		t.Fatalf("want 1 proxy (dead node pruned), got %d", len(res.Proxies))
	}
	if res.Proxies[0].Host != "1.1.1.1" {
		t.Errorf("surviving proxy host = %s, want the live node 1.1.1.1", res.Proxies[0].Host)
	}
}
