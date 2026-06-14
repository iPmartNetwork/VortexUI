package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

type fakeBinder struct {
	bound   []string
	unbound []string
}

func (f *fakeBinder) BindInbound(_ context.Context, userID, inboundID uuid.UUID) error {
	f.bound = append(f.bound, userID.String()+"->"+inboundID.String())
	return nil
}
func (f *fakeBinder) UnbindInbound(_ context.Context, userID, inboundID uuid.UUID) error {
	f.unbound = append(f.unbound, userID.String()+"->"+inboundID.String())
	return nil
}

func TestMigrateRehomesUsersToSameTagInbounds(t *testing.T) {
	failed := &domain.Node{ID: uuid.New(), Name: "fail"}
	target := &domain.Node{ID: uuid.New(), Name: "target"}

	// Target mirrors only the "vless-ws" inbound, not "trojan".
	targetVless := uuid.New()
	lister := &fakeInboundLister{inbounds: []*domain.Inbound{{ID: targetVless, NodeID: target.ID, Tag: "vless-ws"}}}

	u1, u2, u3 := &domain.User{ID: uuid.New()}, &domain.User{ID: uuid.New()}, &domain.User{ID: uuid.New()}
	usersBy := &fakeUsersByNoder{m: map[string][]*domain.User{
		"vless-ws": {u1, u2}, // migratable
		"trojan":   {u3},     // no matching target inbound -> skipped
	}}
	binder := &fakeBinder{}
	ops := &fakeNodeOps{}

	svc := NewMigrationService(lister, usersBy, binder, binder, ops, nil)
	if err := svc.Migrate(context.Background(), failed, target); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// u1 and u2 are bound to the target's vless-ws inbound and provisioned there.
	if len(binder.bound) != 2 {
		t.Errorf("want 2 bindings, got %v", binder.bound)
	}
	for _, b := range binder.bound {
		if want := "->" + targetVless.String(); b[len(b)-len(want):] != want {
			t.Errorf("binding %q not to target inbound", b)
		}
	}
	if len(ops.added) != 2 {
		t.Errorf("want 2 provision calls on target, got %v", ops.added)
	}
	for _, a := range ops.added {
		if want := target.ID.String() + "/vless-ws"; a != want {
			t.Errorf("provisioned %q, want %q", a, want)
		}
	}
	// u3 (trojan) must NOT be migrated — target has no trojan inbound.
	if contains3(binder.bound, u3.ID.String()) {
		t.Error("trojan user should not have been migrated")
	}
}

func TestMigrateNoTargetErrors(t *testing.T) {
	svc := NewMigrationService(&fakeInboundLister{}, &fakeUsersByNoder{}, &fakeBinder{}, &fakeBinder{}, &fakeNodeOps{}, nil)
	if err := svc.Migrate(context.Background(), &domain.Node{Name: "x"}, nil); err == nil {
		t.Fatal("expected error when no healthy target is available")
	}
}

func TestMigrateBackUndoesTemporaryPlacements(t *testing.T) {
	failed := &domain.Node{ID: uuid.New(), Name: "fail"}
	target := &domain.Node{ID: uuid.New(), Name: "target"}
	targetVless := uuid.New()
	lister := &fakeInboundLister{inbounds: []*domain.Inbound{{ID: targetVless, NodeID: target.ID, Tag: "vless-ws"}}}
	u := &domain.User{ID: uuid.New()}
	usersBy := &fakeUsersByNoder{m: map[string][]*domain.User{"vless-ws": {u}}}
	binder := &fakeBinder{}
	ops := &fakeNodeOps{}
	svc := NewMigrationService(lister, usersBy, binder, binder, ops, nil)
	ctx := context.Background()

	// Fail over: user is parked on the target.
	if err := svc.Migrate(ctx, failed, target); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Recover the failed node: the temporary placement must be undone.
	if err := svc.MigrateBack(ctx, failed); err != nil {
		t.Fatalf("migrate back: %v", err)
	}
	if len(binder.unbound) != 1 || binder.unbound[0] != u.ID.String()+"->"+targetVless.String() {
		t.Errorf("expected unbind of user from target inbound, got %v", binder.unbound)
	}
	if len(ops.removed) != 1 || ops.removed[0] != target.ID.String()+"/vless-ws" {
		t.Errorf("expected removal from target core, got %v", ops.removed)
	}

	// Idempotent: a second migrate-back (no records) is a no-op.
	if err := svc.MigrateBack(ctx, failed); err != nil {
		t.Fatalf("second migrate back: %v", err)
	}
	if len(binder.unbound) != 1 {
		t.Errorf("second migrate-back should be a no-op, unbound=%v", binder.unbound)
	}
}

func contains3(ss []string, sub string) bool {
	for _, s := range ss {
		if len(sub) > 0 && len(s) >= len(sub) && s[:len(sub)] == sub {
			return true
		}
	}
	return false
}
