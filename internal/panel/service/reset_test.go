package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

type fakeResetRepo struct {
	toReset  []*domain.User
	inbounds map[uuid.UUID][]domain.Inbound
	updated  map[uuid.UUID]*domain.User
}

func (f *fakeResetRepo) UsersToReset(context.Context) ([]*domain.User, error) { return f.toReset, nil }
func (f *fakeResetRepo) InboundsFor(_ context.Context, id uuid.UUID) ([]domain.Inbound, error) {
	return f.inbounds[id], nil
}
func (f *fakeResetRepo) Update(_ context.Context, u *domain.User) error {
	if f.updated == nil {
		f.updated = map[uuid.UUID]*domain.User{}
	}
	cp := *u
	f.updated[u.ID] = &cp
	return nil
}

func TestResetterZeroesTrafficAndReactivatesLimited(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	future := now.Add(48 * time.Hour)
	nodeID := uuid.New()

	// Limited purely by quota (not expired): reset must reactivate + re-provision.
	limited := &domain.User{ID: uuid.New(), Username: "capped", Status: domain.UserStatusLimited,
		DataLimit: 100, UsedTraffic: 100, ExpireAt: &future, ResetStrategy: domain.ResetDaily}
	// Active heavy user: reset just zeroes the counter, no provisioning change.
	active := &domain.User{ID: uuid.New(), Username: "normal", Status: domain.UserStatusActive,
		DataLimit: 100, UsedTraffic: 60, ResetStrategy: domain.ResetMonthly}

	repo := &fakeResetRepo{
		toReset:  []*domain.User{limited, active},
		inbounds: map[uuid.UUID][]domain.Inbound{limited.ID: {{NodeID: nodeID, Tag: "in-a"}}},
	}
	ops := &fakeNodeOps{}
	r := NewResetter(repo, ops, time.Hour, nil)
	r.now = func() time.Time { return now }

	if err := r.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}

	// Both counters zeroed and last_reset stamped.
	for _, id := range []uuid.UUID{limited.ID, active.ID} {
		u := repo.updated[id]
		if u.UsedTraffic != 0 || u.LastReset == nil {
			t.Errorf("user %s not reset: used=%d lastReset=%v", u.Username, u.UsedTraffic, u.LastReset)
		}
	}
	// The previously-limited user is reactivated and re-provisioned.
	if repo.updated[limited.ID].Status != domain.UserStatusActive {
		t.Errorf("limited user status = %s, want active", repo.updated[limited.ID].Status)
	}
	if len(ops.added) != 1 || ops.added[0] != nodeID.String()+"/in-a" {
		t.Errorf("limited user not re-provisioned: %v", ops.added)
	}
	// The already-active user triggers no provisioning.
	if repo.updated[active.ID].Status != domain.UserStatusActive {
		t.Errorf("active user status changed unexpectedly: %s", repo.updated[active.ID].Status)
	}
}

func TestResetterDoesNotReactivateExpired(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)

	// Expired AND limited: resetting traffic must NOT bring an expired user back.
	u := &domain.User{ID: uuid.New(), Username: "gone", Status: domain.UserStatusLimited,
		DataLimit: 100, UsedTraffic: 100, ExpireAt: &past, ResetStrategy: domain.ResetDaily}
	repo := &fakeResetRepo{toReset: []*domain.User{u}}
	ops := &fakeNodeOps{}
	r := NewResetter(repo, ops, time.Hour, nil)
	r.now = func() time.Time { return now }

	if err := r.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if repo.updated[u.ID].Status != domain.UserStatusExpired {
		t.Errorf("status = %s, want expired (reset must not revive an expired user)", repo.updated[u.ID].Status)
	}
	if len(ops.added) != 0 {
		t.Errorf("expired user must not be re-provisioned: %v", ops.added)
	}
}
