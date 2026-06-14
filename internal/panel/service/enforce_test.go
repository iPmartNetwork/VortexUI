package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

type fakeEnforceRepo struct {
	toLimit  []*domain.User
	inbounds map[uuid.UUID][]domain.Inbound
	updated  map[uuid.UUID]domain.UserStatus
}

func (f *fakeEnforceRepo) UsersToLimit(context.Context) ([]*domain.User, error) {
	return f.toLimit, nil
}
func (f *fakeEnforceRepo) InboundsFor(_ context.Context, id uuid.UUID) ([]domain.Inbound, error) {
	return f.inbounds[id], nil
}
func (f *fakeEnforceRepo) Update(_ context.Context, u *domain.User) error {
	if f.updated == nil {
		f.updated = map[uuid.UUID]domain.UserStatus{}
	}
	f.updated[u.ID] = u.Status
	return nil
}

func TestEnforcerTickDisablesAndDeprovisions(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	nodeID := uuid.New()

	overLimit := &domain.User{ID: uuid.New(), Username: "heavy", Status: domain.UserStatusActive, DataLimit: 100, UsedTraffic: 150}
	expired := &domain.User{ID: uuid.New(), Username: "old", Status: domain.UserStatusActive, ExpireAt: &past}

	repo := &fakeEnforceRepo{
		toLimit: []*domain.User{overLimit, expired},
		inbounds: map[uuid.UUID][]domain.Inbound{
			overLimit.ID: {{NodeID: nodeID, Tag: "in-a"}},
			expired.ID:   {{NodeID: nodeID, Tag: "in-b"}},
		},
	}
	ops := &fakeNodeOps{}
	e := NewEnforcer(repo, ops, time.Minute, nil)
	e.now = func() time.Time { return now }

	if err := e.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}

	// Each user gets the precise terminal status persisted.
	if repo.updated[overLimit.ID] != domain.UserStatusLimited {
		t.Errorf("over-limit user status = %s, want limited", repo.updated[overLimit.ID])
	}
	if repo.updated[expired.ID] != domain.UserStatusExpired {
		t.Errorf("expired user status = %s, want expired", repo.updated[expired.ID])
	}
	// Both are removed from their live cores.
	if len(ops.removed) != 2 {
		t.Errorf("expected 2 de-provision calls, got %v", ops.removed)
	}
}

func TestEnforcerTickNoopWhenNothingToLimit(t *testing.T) {
	repo := &fakeEnforceRepo{}
	ops := &fakeNodeOps{}
	e := NewEnforcer(repo, ops, time.Minute, nil)
	if err := e.Tick(context.Background()); err != nil {
		t.Fatalf("tick: %v", err)
	}
	if len(ops.removed) != 0 || len(repo.updated) != 0 {
		t.Error("no users to limit should produce no side effects")
	}
}
