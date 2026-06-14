package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// --- fakes ---

type fakeInboundLister struct{ inbounds []*domain.Inbound }

func (f *fakeInboundLister) ListByNode(context.Context, uuid.UUID) ([]*domain.Inbound, error) {
	return f.inbounds, nil
}

type fakeUsersByNoder struct{ m map[string][]*domain.User }

func (f *fakeUsersByNoder) UsersByNode(context.Context, uuid.UUID) (map[string][]*domain.User, error) {
	return f.m, nil
}

type fakeSyncer struct {
	nodeID uuid.UUID
	cfg    *core.GeneratedConfig
}

func (f *fakeSyncer) Sync(_ context.Context, nodeID uuid.UUID, cfg *core.GeneratedConfig) error {
	f.nodeID = nodeID
	f.cfg = cfg
	return nil
}

type fakeNodeRepo struct {
	created  *domain.Node
	deleted  bool
}

func (f *fakeNodeRepo) Create(_ context.Context, n *domain.Node) error { f.created = n; return nil }
func (f *fakeNodeRepo) GetByID(context.Context, uuid.UUID) (*domain.Node, error) {
	return f.created, nil
}
func (f *fakeNodeRepo) Update(context.Context, *domain.Node) error { return nil }
func (f *fakeNodeRepo) Delete(context.Context, uuid.UUID) error    { f.deleted = true; return nil }
func (f *fakeNodeRepo) List(context.Context) ([]*domain.Node, error) {
	return []*domain.Node{f.created}, nil
}
func (f *fakeNodeRepo) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error { return nil }

type fakeRegistrar struct {
	registered   *domain.Node
	deregistered uuid.UUID
}

func (f *fakeRegistrar) Register(_ context.Context, n *domain.Node) error { f.registered = n; return nil }
func (f *fakeRegistrar) Deregister(id uuid.UUID)                          { f.deregistered = id }

// --- tests ---

func TestSyncResyncAssemblesEnabledInboundsAndUsers(t *testing.T) {
	nodeID := uuid.New()
	lister := &fakeInboundLister{inbounds: []*domain.Inbound{
		{Tag: "on", NodeID: nodeID, Protocol: domain.ProtoVLESS, Port: 443, Enabled: true},
		{Tag: "off", NodeID: nodeID, Protocol: domain.ProtoVMess, Port: 80, Enabled: false},
	}}
	usersBy := &fakeUsersByNoder{m: map[string][]*domain.User{"on": {{ID: uuid.New()}}}}
	syncer := &fakeSyncer{}

	svc := NewSyncService(lister, usersBy, syncer)
	if err := svc.Resync(context.Background(), nodeID); err != nil {
		t.Fatalf("resync: %v", err)
	}

	if syncer.nodeID != nodeID {
		t.Errorf("synced node = %s, want %s", syncer.nodeID, nodeID)
	}
	// Only the enabled inbound is pushed.
	if len(syncer.cfg.Inbounds) != 1 || syncer.cfg.Inbounds[0].Tag != "on" {
		t.Errorf("config inbounds = %+v, want only [on]", syncer.cfg.Inbounds)
	}
	if len(syncer.cfg.UsersByInbound["on"]) != 1 {
		t.Errorf("expected 1 user on inbound 'on', got %d", len(syncer.cfg.UsersByInbound["on"]))
	}
}

func TestNodeCreateRegistersWithHub(t *testing.T) {
	repo := &fakeNodeRepo{}
	reg := &fakeRegistrar{}
	svc := NewNodeService(repo, reg)

	n, err := svc.Create(context.Background(), CreateNodeInput{Name: "de1", Address: "1.2.3.4:50051"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if repo.created == nil {
		t.Fatal("node not persisted")
	}
	if reg.registered == nil || reg.registered.ID != n.ID {
		t.Error("node not registered with hub")
	}
	// Defaults applied.
	if n.Core != domain.CoreXray || n.UsageRatio != 1 {
		t.Errorf("defaults not applied: core=%s ratio=%v", n.Core, n.UsageRatio)
	}

	if err := svc.Delete(context.Background(), n.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !repo.deleted || reg.deregistered != n.ID {
		t.Error("delete did not deregister + remove")
	}
}

func TestNodeUpdateReRegisters(t *testing.T) {
	id := uuid.New()
	repo := &fakeNodeRepo{created: &domain.Node{ID: id, Name: "old", Address: "1.1.1.1:50051", Core: domain.CoreXray, UsageRatio: 1}}
	reg := &fakeRegistrar{}
	svc := NewNodeService(repo, reg)

	n, err := svc.Update(context.Background(), id, UpdateNodeInput{Name: "new", Address: "2.2.2.2:50051"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if n.Name != "new" || n.Address != "2.2.2.2:50051" {
		t.Errorf("fields not applied: %+v", n)
	}
	// Address change must drop and re-dial the connection.
	if reg.deregistered != id || reg.registered == nil || reg.registered.Address != "2.2.2.2:50051" {
		t.Errorf("update did not re-register with new address: dereg=%s reg=%+v", reg.deregistered, reg.registered)
	}
}
