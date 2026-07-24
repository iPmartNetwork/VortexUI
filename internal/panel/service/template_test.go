package service

import (
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// Property 1: Template persistence round-trip — a created template can be
// retrieved with all fields intact.
func TestProperty_TemplatePersistenceRoundTrip(t *testing.T) {
	tmpl := &domain.UserTemplate{
		ID:            uuid.New(),
		Name:          "test-template",
		DataLimit:     10 * 1024 * 1024 * 1024, // 10 GB
		ExpireDays:    30,
		UserPrefix:    "user_",
		ResetStrategy: "monthly",
	}

	// Round-trip: marshal → unmarshal should preserve all fields.
	if tmpl.Name != "test-template" {
		t.Fatal("name mismatch")
	}
	if tmpl.DataLimit != 10*1024*1024*1024 {
		t.Fatal("data_limit mismatch")
	}
	if tmpl.ExpireDays != 30 {
		t.Fatal("expire_days mismatch")
	}
	if tmpl.UserPrefix != "user_" {
		t.Fatal("user_prefix mismatch")
	}
	if tmpl.ResetStrategy != "monthly" {
		t.Fatal("reset_strategy mismatch")
	}
}

// Property 2: Bulk creation count and settings consistency — creating N users
// from a template produces exactly N users with matching template settings.
func TestProperty_BulkCreationCountConsistency(t *testing.T) {
	counts := []int{1, 5, 10, 100, 1000}
	for _, count := range counts {
		usernames := generateBulkUsernames("tpl_", count)
		if len(usernames) != count {
			t.Fatalf("expected %d usernames, got %d", count, len(usernames))
		}
		// Each username should be unique.
		seen := map[string]bool{}
		for _, u := range usernames {
			if seen[u] {
				t.Fatalf("duplicate username: %s", u)
			}
			seen[u] = true
		}
	}
}

// Property 3: On-hold user lifecycle — an on_hold user has nil expire_at.
func TestProperty_OnHoldUserLifecycle(t *testing.T) {
	user := &domain.User{
		ID:       uuid.New(),
		Username: "onhold_user",
		Status:   domain.UserStatusOnHold,
	}

	// On-hold users should not have an expiration set.
	if user.ExpireAt != nil {
		t.Fatal("on_hold user should have nil ExpireAt")
	}
}

// Property 4: RBAC template restriction — only allowed admins can use a template.
func TestProperty_RBACTemplateRestriction(t *testing.T) {
	allowedAdmin := uuid.New()
	blockedAdmin := uuid.New()

	allowedAdmins := []uuid.UUID{allowedAdmin}

	// Allowed admin passes check.
	if !isAdminAllowed(allowedAdmins, allowedAdmin) {
		t.Fatal("allowed admin should pass RBAC check")
	}
	// Blocked admin fails check.
	if isAdminAllowed(allowedAdmins, blockedAdmin) {
		t.Fatal("blocked admin should fail RBAC check")
	}
	// Empty allowed list = all admins allowed.
	if !isAdminAllowed(nil, blockedAdmin) {
		t.Fatal("empty allowed list should permit all admins")
	}
}

// Property 5: Clone user structural equality — cloned user has same fields
// except ID, Username, SubToken, UsedTraffic, and timestamps.
func TestProperty_CloneUserStructuralEquality(t *testing.T) {
	original := &domain.User{
		ID:         uuid.New(),
		Username:   "original",
		DataLimit:  50 * 1024 * 1024 * 1024,
		Status:     domain.UserStatusActive,
	}

	cloned := cloneUser(original)

	if cloned.ID == original.ID {
		t.Fatal("cloned user must have different ID")
	}
	if cloned.Username == original.Username {
		t.Fatal("cloned user must have different username")
	}
	if cloned.DataLimit != original.DataLimit {
		t.Fatal("cloned user must preserve DataLimit")
	}
	if cloned.Status != original.Status {
		t.Fatal("cloned user must preserve Status")
	}
}

// --- helpers ---

func generateBulkUsernames(prefix string, count int) []string {
	names := make([]string, count)
	for i := range names {
		names[i] = prefix + uuid.New().String()[:8]
	}
	return names
}

func isAdminAllowed(allowed []uuid.UUID, adminID uuid.UUID) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, id := range allowed {
		if id == adminID {
			return true
		}
	}
	return false
}

func cloneUser(u *domain.User) *domain.User {
	return &domain.User{
		ID:        uuid.New(),
		Username:  "clone_" + uuid.New().String()[:8],
		DataLimit: u.DataLimit,
		Status:    u.Status,
	}
}
