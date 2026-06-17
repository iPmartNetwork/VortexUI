package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- fakes for migration tests ---

type fakeMigrationInbounds struct {
	byNode map[uuid.UUID][]*domain.Inbound
}

func (f *fakeMigrationInbounds) Create(context.Context, *domain.Inbound) error          { return nil }
func (f *fakeMigrationInbounds) GetByID(context.Context, uuid.UUID) (*domain.Inbound, error) {
	return nil, nil
}
func (f *fakeMigrationInbounds) Update(context.Context, *domain.Inbound) error          { return nil }
func (f *fakeMigrationInbounds) Delete(context.Context, uuid.UUID) error                { return nil }
func (f *fakeMigrationInbounds) ListByNode(_ context.Context, nodeID uuid.UUID) ([]*domain.Inbound, error) {
	return f.byNode[nodeID], nil
}

type fakeMigrationUsers struct {
	users    []*domain.User
	inbounds map[uuid.UUID][]domain.Inbound // userID -> inbounds
	setBound map[uuid.UUID][]uuid.UUID      // captures SetInbounds calls
}

func (f *fakeMigrationUsers) Create(context.Context, *domain.User) error              { return nil }
func (f *fakeMigrationUsers) GetByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}
func (f *fakeMigrationUsers) GetBySubToken(context.Context, string) (*domain.User, error) {
	return nil, nil
}
func (f *fakeMigrationUsers) Update(context.Context, *domain.User) error              { return nil }
func (f *fakeMigrationUsers) Delete(context.Context, uuid.UUID) error                 { return nil }
func (f *fakeMigrationUsers) List(_ context.Context, _ port.UserFilter) ([]*domain.User, int, error) {
	return f.users, len(f.users), nil
}
func (f *fakeMigrationUsers) AddUsedTraffic(context.Context, uuid.UUID, int64) error  { return nil }
func (f *fakeMigrationUsers) AddUsedTrafficBatch(context.Context, map[uuid.UUID]int64) error {
	return nil
}
func (f *fakeMigrationUsers) SetInbounds(_ context.Context, userID uuid.UUID, ids []uuid.UUID) error {
	if f.setBound == nil {
		f.setBound = make(map[uuid.UUID][]uuid.UUID)
	}
	f.setBound[userID] = ids
	return nil
}
func (f *fakeMigrationUsers) InboundsFor(_ context.Context, userID uuid.UUID) ([]domain.Inbound, error) {
	return f.inbounds[userID], nil
}

type fakeMigrationNodes struct {
	nodes []*domain.Node
}

func (f *fakeMigrationNodes) Create(context.Context, *domain.Node) error                        { return nil }
func (f *fakeMigrationNodes) GetByID(context.Context, uuid.UUID) (*domain.Node, error)          { return nil, nil }
func (f *fakeMigrationNodes) Update(context.Context, *domain.Node) error                        { return nil }
func (f *fakeMigrationNodes) Delete(context.Context, uuid.UUID) error                           { return nil }
func (f *fakeMigrationNodes) List(context.Context) ([]*domain.Node, error)                      { return f.nodes, nil }
func (f *fakeMigrationNodes) UpdateHealth(context.Context, uuid.UUID, domain.NodeHealth) error  { return nil }

func TestMigrateRehomesUsersToSameTagInbounds(t *testing.T) {
	failed := &domain.Node{ID: uuid.New(), Name: "fail"}
	target := &domain.Node{ID: uuid.New(), Name: "target"}

	failedVless := uuid.New()
	failedTrojan := uuid.New()
	targetVless := uuid.New()

	inbounds := &fakeMigrationInbounds{byNode: map[uuid.UUID][]*domain.Inbound{
		failed.ID: {
			{ID: failedVless, NodeID: failed.ID, Tag: "vless-ws"},
			{ID: failedTrojan, NodeID: failed.ID, Tag: "trojan"},
		},
		target.ID: {
			{ID: targetVless, NodeID: target.ID, Tag: "vless-ws"},
			// No trojan on target.
		},
	}}

	u1 := &domain.User{ID: uuid.New(), Username: "alice"}
	u2 := &domain.User{ID: uuid.New(), Username: "bob"}
	u3 := &domain.User{ID: uuid.New(), Username: "charlie"}

	users := &fakeMigrationUsers{
		users: []*domain.User{u1, u2, u3},
		inbounds: map[uuid.UUID][]domain.Inbound{
			u1.ID: {{ID: failedVless, Tag: "vless-ws"}},
			u2.ID: {{ID: failedVless, Tag: "vless-ws"}},
			u3.ID: {{ID: failedTrojan, Tag: "trojan"}},
		},
	}

	ops := &fakeNodeOps{}
	nodes := &fakeMigrationNodes{}

	svc := NewMigrationService(inbounds, nodes, users, ops, nil)
	if err := svc.Migrate(context.Background(), failed, target); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// u1 and u2 (vless-ws) should be rebound to target's vless-ws inbound.
	if len(users.setBound) != 2 {
		t.Errorf("want 2 users rebound, got %d", len(users.setBound))
	}
	for _, uid := range []uuid.UUID{u1.ID, u2.ID} {
		ids, ok := users.setBound[uid]
		if !ok {
			t.Errorf("user %s not rebound", uid)
			continue
		}
		if len(ids) != 1 || ids[0] != targetVless {
			t.Errorf("user %s bound to %v, want [%s]", uid, ids, targetVless)
		}
	}

	// u1 and u2 should be provisioned on target.
	if len(ops.added) != 2 {
		t.Errorf("want 2 provision calls on target, got %d", len(ops.added))
	}

	// u3 (trojan) must NOT be migrated — target has no trojan inbound.
	if _, ok := users.setBound[u3.ID]; ok {
		t.Error("trojan user should not have been migrated")
	}
}

func TestMigrateNilTargetErrors(t *testing.T) {
	inbounds := &fakeMigrationInbounds{byNode: map[uuid.UUID][]*domain.Inbound{}}
	users := &fakeMigrationUsers{}
	nodes := &fakeMigrationNodes{}
	ops := &fakeNodeOps{}

	svc := NewMigrationService(inbounds, nodes, users, ops, nil)
	err := svc.Migrate(context.Background(), &domain.Node{ID: uuid.New(), Name: "x"}, nil)
	if err == nil {
		t.Fatal("expected error when target is nil")
	}
}
