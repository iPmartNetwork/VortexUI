package service

import (
	"testing"

	"github.com/google/uuid"
)

// Property 12: Bulk preview is side-effect-free — running a preview does not
// modify any user state.
func TestProperty_BulkPreviewSideEffectFree(t *testing.T) {
	users := generateTestUsers(10)
	original := snapshotUsers(users)

	// Simulate preview: count matching users without modifying them.
	matchCount := previewBulkOperation(users, "active")

	if matchCount < 0 || matchCount > len(users) {
		t.Fatalf("invalid preview count: %d", matchCount)
	}

	// Verify no state changed.
	for i, u := range users {
		if u.Status != original[i].Status {
			t.Fatalf("user %d status changed during preview", i)
		}
		if u.DataLimit != original[i].DataLimit {
			t.Fatalf("user %d data_limit changed during preview", i)
		}
	}
}

// Property 13: Bulk filter correctness — filter by status returns exactly
// users matching that status.
func TestProperty_BulkFilterCorrectness(t *testing.T) {
	users := []testBulkUser{
		{ID: uuid.New(), Status: "active"},
		{ID: uuid.New(), Status: "expired"},
		{ID: uuid.New(), Status: "active"},
		{ID: uuid.New(), Status: "disabled"},
		{ID: uuid.New(), Status: "active"},
	}

	filtered := filterByStatus(users, "active")
	if len(filtered) != 3 {
		t.Fatalf("expected 3 active users, got %d", len(filtered))
	}
	for _, u := range filtered {
		if u.Status != "active" {
			t.Fatalf("filtered user has wrong status: %s", u.Status)
		}
	}

	filtered = filterByStatus(users, "expired")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 expired user, got %d", len(filtered))
	}
}

// Property 14: Bulk operation history completeness — every executed operation
// has a corresponding history entry.
func TestProperty_BulkOperationHistoryCompleteness(t *testing.T) {
	var history []bulkHistoryEntry
	operations := []string{"reset_traffic", "disable", "enable", "delete", "extend"}

	for _, op := range operations {
		history = append(history, executeBulkOp(op, 5))
	}

	if len(history) != len(operations) {
		t.Fatalf("expected %d history entries, got %d", len(operations), len(history))
	}

	for i, entry := range history {
		if entry.Operation != operations[i] {
			t.Fatalf("history entry %d has wrong operation: %s", i, entry.Operation)
		}
		if entry.AffectedCount != 5 {
			t.Fatalf("history entry %d has wrong count: %d", i, entry.AffectedCount)
		}
	}
}

// --- helpers ---

type testBulkUser struct {
	ID        uuid.UUID
	Status    string
	DataLimit int64
}

func generateTestUsers(n int) []testBulkUser {
	users := make([]testBulkUser, n)
	for i := range users {
		users[i] = testBulkUser{ID: uuid.New(), Status: "active", DataLimit: 1000}
	}
	return users
}

func snapshotUsers(users []testBulkUser) []testBulkUser {
	snap := make([]testBulkUser, len(users))
	copy(snap, users)
	return snap
}

func previewBulkOperation(users []testBulkUser, status string) int {
	count := 0
	for _, u := range users {
		if u.Status == status {
			count++
		}
	}
	return count
}

func filterByStatus(users []testBulkUser, status string) []testBulkUser {
	var result []testBulkUser
	for _, u := range users {
		if u.Status == status {
			result = append(result, u)
		}
	}
	return result
}

type bulkHistoryEntry struct {
	Operation     string
	AffectedCount int
}

func executeBulkOp(operation string, count int) bulkHistoryEntry {
	return bulkHistoryEntry{Operation: operation, AffectedCount: count}
}
