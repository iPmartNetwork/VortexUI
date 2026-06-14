package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// listUserRepo serves a fixed user list with pagination + bindings, enough to
// exercise BackupService.Export.
type listUserRepo struct {
	users    []*domain.User
	bindings map[uuid.UUID][]domain.Inbound
}

func (r *listUserRepo) Create(context.Context, *domain.User) error { return nil }
func (r *listUserRepo) GetByID(context.Context, uuid.UUID) (*domain.User, error) {
	return nil, domain.ErrNotFound
}
func (r *listUserRepo) GetBySubToken(context.Context, string) (*domain.User, error) {
	return nil, domain.ErrNotFound
}
func (r *listUserRepo) Update(context.Context, *domain.User) error { return nil }
func (r *listUserRepo) Delete(context.Context, uuid.UUID) error    { return nil }
func (r *listUserRepo) List(_ context.Context, f port.UserFilter) ([]*domain.User, int, error) {
	if f.Offset >= len(r.users) {
		return nil, len(r.users), nil
	}
	end := f.Offset + f.Limit
	if end > len(r.users) {
		end = len(r.users)
	}
	return r.users[f.Offset:end], len(r.users), nil
}
func (r *listUserRepo) AddUsedTraffic(context.Context, uuid.UUID, int64) error         { return nil }
func (r *listUserRepo) AddUsedTrafficBatch(context.Context, map[uuid.UUID]int64) error { return nil }
func (r *listUserRepo) SetInbounds(context.Context, uuid.UUID, []uuid.UUID) error      { return nil }
func (r *listUserRepo) InboundsFor(_ context.Context, id uuid.UUID) ([]domain.Inbound, error) {
	return r.bindings[id], nil
}

// capRestorer captures the backup handed to Restore.
type capRestorer struct{ got *domain.Backup }

func (c *capRestorer) Restore(_ context.Context, b *domain.Backup) error { c.got = b; return nil }

func TestBackupExportGathersFullConfig(t *testing.T) {
	nodeID := uuid.New()
	inID := uuid.New()
	uid := uuid.New()

	nodes := &listNodeRepo{nodes: []*domain.Node{{ID: nodeID, Name: "n1", Core: domain.CoreXray}}}
	outs := &fakeOutboundRepo{items: []*domain.Outbound{{ID: uuid.New(), NodeID: nodeID, Tag: "direct", Protocol: domain.OutFreedom, Enabled: true}}}
	routes := &fakeRoutingRepo{items: []*domain.RoutingRule{{ID: uuid.New(), NodeID: nodeID, InboundTags: []string{"in"}, OutboundTag: "direct", Enabled: true}}}
	bals := &fakeBalancerRepo{items: []*domain.Balancer{{ID: uuid.New(), NodeID: nodeID, Tag: "lb", Selectors: []string{"p"}, Strategy: domain.BalancerRandom, Enabled: true}}}
	inbounds := &fakeInboundLister{inbounds: []*domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "in", Enabled: true}}}
	users := &listUserRepo{
		users:    []*domain.User{{ID: uid, Username: "alice"}},
		bindings: map[uuid.UUID][]domain.Inbound{uid: {{ID: inID, NodeID: nodeID, Tag: "in"}}},
	}

	svc := NewBackupService(nodes, inboundRepoAdapter{inbounds}, outs, routes, bals, users, &capRestorer{})
	b, err := svc.Export(context.Background())
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if b.Version != domain.BackupVersion {
		t.Errorf("version = %d", b.Version)
	}
	if len(b.Nodes) != 1 || len(b.Inbounds) != 1 || len(b.Outbounds) != 1 || len(b.Routing) != 1 || len(b.Balancers) != 1 {
		t.Errorf("config sections incomplete: %+v", b)
	}
	if len(b.Users) != 1 || len(b.Bindings) != 1 {
		t.Errorf("users/bindings wrong: users=%d bindings=%d", len(b.Users), len(b.Bindings))
	}
	if b.Bindings[0].UserID != uid || b.Bindings[0].InboundID != inID {
		t.Errorf("binding wrong: %+v", b.Bindings[0])
	}
}

func TestBackupRestoreValidatesVersion(t *testing.T) {
	cap := &capRestorer{}
	svc := NewBackupService(
		&listNodeRepo{}, inboundRepoAdapter{&fakeInboundLister{}},
		&fakeOutboundRepo{}, &fakeRoutingRepo{}, &fakeBalancerRepo{}, &listUserRepo{}, cap,
	)
	// Wrong version is rejected before touching the restorer.
	if err := svc.Restore(context.Background(), &domain.Backup{Version: 999}); err == nil {
		t.Error("expected error for unsupported version")
	}
	if cap.got != nil {
		t.Error("restorer must not be called on a bad version")
	}
	// Correct version is delegated.
	if err := svc.Restore(context.Background(), &domain.Backup{Version: domain.BackupVersion}); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if cap.got == nil {
		t.Error("restorer was not called for a valid backup")
	}
}

// inboundRepoAdapter adapts the fakeInboundLister (ListByNode only) to the full
// port.InboundRepository the BackupService takes; only ListByNode is exercised.
type inboundRepoAdapter struct{ l *fakeInboundLister }

func (a inboundRepoAdapter) Create(context.Context, *domain.Inbound) error { return nil }
func (a inboundRepoAdapter) GetByID(context.Context, uuid.UUID) (*domain.Inbound, error) {
	return nil, domain.ErrNotFound
}
func (a inboundRepoAdapter) Update(context.Context, *domain.Inbound) error { return nil }
func (a inboundRepoAdapter) Delete(context.Context, uuid.UUID) error       { return nil }
func (a inboundRepoAdapter) ListByNode(ctx context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error) {
	return a.l.ListByNode(ctx, nodeID)
}
