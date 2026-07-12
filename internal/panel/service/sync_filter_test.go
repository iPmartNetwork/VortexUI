package service

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestFilterProvisionableUsersDropsLimitedAndOverQuota(t *testing.T) {
	now := time.Now()
	active := &domain.User{ID: uuid.New(), Status: domain.UserStatusActive, DataLimit: 100, UsedTraffic: 50}
	limited := &domain.User{ID: uuid.New(), Status: domain.UserStatusLimited, DataLimit: 100, UsedTraffic: 100}
	overQuota := &domain.User{ID: uuid.New(), Status: domain.UserStatusActive, DataLimit: 100, UsedTraffic: 100}
	onHold := &domain.User{ID: uuid.New(), Status: domain.UserStatusOnHold, DataLimit: 100, UsedTraffic: 0}
	disabled := &domain.User{ID: uuid.New(), Status: domain.UserStatusDisabled, DataLimit: 0, UsedTraffic: 0}

	got := filterProvisionableUsers(map[string][]*domain.User{
		"in": {active, limited, overQuota, onHold, disabled},
	}, now)

	kept := got["in"]
	if len(kept) != 2 {
		t.Fatalf("kept %d users, want 2 (active + on_hold)", len(kept))
	}
	ids := map[uuid.UUID]bool{kept[0].ID: true, kept[1].ID: true}
	if !ids[active.ID] || !ids[onHold.ID] {
		t.Fatalf("unexpected survivors: %+v", kept)
	}
}
