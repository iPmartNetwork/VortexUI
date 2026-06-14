package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// DeviceTracker enforces a per-user cap on the number of distinct devices seen
// within a rolling window. Devices are tracked in a sorted set keyed by user,
// scored by last-seen time, so idle devices age out and free their slot.
type DeviceTracker struct {
	rdb *redis.Client
}

// Devices returns the device tracker for this connection.
func (c *Client) Devices() *DeviceTracker { return &DeviceTracker{rdb: c.rdb} }

// allowScript runs the whole check-and-admit atomically so concurrent requests
// from many devices cannot race past the limit:
//   1. evict members older than the window,
//   2. a device already known is refreshed and allowed,
//   3. an unknown device is admitted only if the set is below the limit.
var allowScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local cutoff = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local device = ARGV[4]
local ttl = tonumber(ARGV[5])
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)
if redis.call('ZSCORE', key, device) then
  redis.call('ZADD', key, now, device)
  redis.call('EXPIRE', key, ttl)
  return 1
end
if redis.call('ZCARD', key) < limit then
  redis.call('ZADD', key, now, device)
  redis.call('EXPIRE', key, ttl)
  return 1
end
return 0
`)

// Allow records a device access for a user and reports whether it is permitted
// under the limit. A known device is always allowed (and refreshed); a new one
// is allowed only while the active-device count is below limit.
func (d *DeviceTracker) Allow(ctx context.Context, userID, deviceID string, limit int, window time.Duration) (bool, error) {
	now := time.Now().Unix()
	cutoff := now - int64(window.Seconds())
	ttl := int(window.Seconds())
	res, err := allowScript.Run(ctx, d.rdb, []string{"devices:" + userID}, now, cutoff, limit, deviceID, ttl).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

// onlineScript evicts members older than the window, then returns the remaining
// active-device count — done in one round-trip so the count reflects the same
// eviction the limiter applies.
var onlineScript = redis.NewScript(`
local key = KEYS[1]
local cutoff = tonumber(ARGV[1])
redis.call('ZREMRANGEBYSCORE', key, '-inf', cutoff)
return redis.call('ZCARD', key)
`)

// Online reports how many distinct devices a user has been seen on within the
// window (i.e. recently active devices). It reflects subscription-fetch activity
// the tracker observes, not raw live proxy connections.
func (d *DeviceTracker) Online(ctx context.Context, userID string, window time.Duration) (int, error) {
	cutoff := time.Now().Unix() - int64(window.Seconds())
	n, err := onlineScript.Run(ctx, d.rdb, []string{"devices:" + userID}, cutoff).Int()
	if err != nil {
		return 0, err
	}
	return n, nil
}
