package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// openTestRedis connects to the redis named by VORTEX_TEST_REDIS, skipping the
// test when it is unset so `go test ./...` stays green without a redis.
func openTestRedis(t *testing.T) *Client {
	t.Helper()
	url := os.Getenv("VORTEX_TEST_REDIS")
	if url == "" {
		t.Skip("VORTEX_TEST_REDIS not set; skipping redis integration test")
	}
	c, err := Open(context.Background(), url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestDeviceTracker(t *testing.T) {
	c := openTestRedis(t)
	dt := c.Devices()
	ctx := context.Background()
	user := fmt.Sprintf("u:%d", time.Now().UnixNano())

	// Limit of 2 distinct devices.
	for _, dev := range []string{"a", "b"} {
		if ok, err := dt.Allow(ctx, user, dev, 2, time.Hour); err != nil || !ok {
			t.Fatalf("device %s should be admitted (ok=%v err=%v)", dev, ok, err)
		}
	}
	// A third distinct device is rejected.
	if ok, _ := dt.Allow(ctx, user, "c", 2, time.Hour); ok {
		t.Error("third device should be rejected")
	}
	// An already-known device is still allowed (and refreshed), not counted anew.
	if ok, _ := dt.Allow(ctx, user, "a", 2, time.Hour); !ok {
		t.Error("known device should remain allowed")
	}
}

func TestRateLimiterFixedWindow(t *testing.T) {
	c := openTestRedis(t)
	rl := c.RateLimiter()
	ctx := context.Background()
	key := fmt.Sprintf("test:rl:%d", time.Now().UnixNano())

	// First 3 hits within the window are allowed; the 4th is denied with a TTL.
	for i := 0; i < 3; i++ {
		ok, _, err := rl.Allow(ctx, key, 3, 2*time.Second)
		if err != nil || !ok {
			t.Fatalf("hit %d should be allowed (ok=%v err=%v)", i+1, ok, err)
		}
	}
	ok, retry, err := rl.Allow(ctx, key, 3, 2*time.Second)
	if err != nil {
		t.Fatalf("allow: %v", err)
	}
	if ok {
		t.Fatal("4th hit should be denied")
	}
	if retry <= 0 || retry > 2*time.Second {
		t.Errorf("retry-after = %v, want (0, 2s]", retry)
	}

	// After the window expires the counter resets and requests flow again.
	time.Sleep(2100 * time.Millisecond)
	if ok, _, err := rl.Allow(ctx, key, 3, 2*time.Second); err != nil || !ok {
		t.Fatalf("after window reset should be allowed (ok=%v err=%v)", ok, err)
	}
}
