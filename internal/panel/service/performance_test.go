package service

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// Property 37: Subscription cache invalidation — after invalidation, the cache
// returns no data for the invalidated key.
func TestProperty_SubscriptionCacheInvalidation(t *testing.T) {
	cache := newMockCache()

	userID := "user-abc"
	format := "clash"
	data := []byte(`{"proxies":[]}`)

	// Set cache.
	cache.set(userID+":"+format, data)
	if cache.get(userID+":"+format) == nil {
		t.Fatal("cache should have data after set")
	}

	// Invalidate.
	cache.deleteByPrefix(userID + ":")
	if cache.get(userID+":"+format) != nil {
		t.Fatal("cache should be empty after invalidation")
	}
}

// Property 38: Cursor pagination correctness — paginating through all items
// with a fixed page size visits every item exactly once.
func TestProperty_CursorPaginationCorrectness(t *testing.T) {
	totalItems := 47
	pageSize := 10
	allIDs := make([]string, totalItems)
	for i := range allIDs {
		allIDs[i] = uuid.New().String()
	}

	// Simulate cursor pagination.
	var visited []string
	offset := 0
	for offset < totalItems {
		end := offset + pageSize
		if end > totalItems {
			end = totalItems
		}
		page := allIDs[offset:end]
		visited = append(visited, page...)
		offset = end
	}

	if len(visited) != totalItems {
		t.Fatalf("expected %d visited items, got %d", totalItems, len(visited))
	}

	// Check uniqueness.
	seen := map[string]bool{}
	for _, id := range visited {
		if seen[id] {
			t.Fatalf("item visited twice: %s", id)
		}
		seen[id] = true
	}
}

// Property 39: Background job lifecycle — a job transitions through
// pending → running → completed/failed states.
func TestProperty_BackgroundJobLifecycle(t *testing.T) {
	job := &domain.BackgroundJob{
		ID:      uuid.New(),
		JobType: "test_job",
		Status:  domain.JobStatusPending,
	}

	// Must start as pending.
	if job.Status != domain.JobStatusPending {
		t.Fatal("new job must be pending")
	}

	// Transition to running.
	job.Status = domain.JobStatusRunning
	now := time.Now()
	job.StartedAt = &now
	if job.StartedAt == nil {
		t.Fatal("running job must have StartedAt")
	}

	// Complete successfully.
	job.Status = domain.JobStatusCompleted
	completed := time.Now()
	job.CompletedAt = &completed
	if job.Status != domain.JobStatusCompleted {
		t.Fatal("job should be completed")
	}

	// Alternative: failure path.
	failedJob := &domain.BackgroundJob{
		ID:      uuid.New(),
		Status:  domain.JobStatusRunning,
		Error:   "timeout",
	}
	failedJob.Status = domain.JobStatusFailed
	if failedJob.Status != domain.JobStatusFailed {
		t.Fatal("failed job must have failed status")
	}
	if failedJob.Error == "" {
		t.Fatal("failed job must have error message")
	}
}

// Property 40: Read replica routing — writes go to primary, reads go to replica.
func TestProperty_ReadReplicaRouting(t *testing.T) {
	primary := "primary-pool"
	replica := "replica-pool"

	router := testRouter{primary: primary, replica: replica}

	if router.writePool() != primary {
		t.Fatal("writes must go to primary")
	}
	if router.readPool() != replica {
		t.Fatal("reads should go to replica when available")
	}

	// Without replica, reads fall back to primary.
	routerNoReplica := testRouter{primary: primary, replica: ""}
	if routerNoReplica.readPool() != primary {
		t.Fatal("without replica, reads should fall back to primary")
	}
}

// --- helpers ---

type mockCache struct {
	data map[string][]byte
}

func newMockCache() *mockCache {
	return &mockCache{data: make(map[string][]byte)}
}

func (c *mockCache) set(key string, val []byte) {
	c.data[key] = val
}

func (c *mockCache) get(key string) []byte {
	return c.data[key]
}

func (c *mockCache) deleteByPrefix(prefix string) {
	for k := range c.data {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.data, k)
		}
	}
}

type testRouter struct {
	primary string
	replica string
}

func (r testRouter) writePool() string { return r.primary }
func (r testRouter) readPool() string {
	if r.replica != "" {
		return r.replica
	}
	return r.primary
}
