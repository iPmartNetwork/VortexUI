// Package redis adapts go-redis into the small surface VortexUI needs today: a
// connection and a fixed-window rate limiter. It is intentionally thin so the
// rest of the system depends on behavior (RateLimiter) rather than the driver.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client owns the redis connection.
type Client struct {
	rdb *redis.Client
}

// Open parses a redis:// URL, connects, and verifies reachability.
func Open(ctx context.Context, url string) (*Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &Client{rdb: rdb}, nil
}

// Close releases the connection.
func (c *Client) Close() error { return c.rdb.Close() }

// RateLimiter returns a fixed-window limiter backed by this connection.
func (c *Client) RateLimiter() *RateLimiter { return &RateLimiter{rdb: c.rdb} }

// RateLimiter implements a fixed-window counter: the first hit in a window sets
// the key's TTL, and each hit increments it. Simple, atomic, and good enough for
// abuse protection (login throttling) where exactness at window edges is moot.
type RateLimiter struct {
	rdb *redis.Client
}

// Allow reports whether a request under key is within limit for the window. It
// also returns how long until the window resets, for a Retry-After header.
func (r *RateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error) {
	pipe := r.rdb.TxPipeline()
	incr := pipe.Incr(ctx, key)
	// Set the TTL only when the key is newly created (NX), so the window does not
	// slide forward on every hit.
	pipe.ExpireNX(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, 0, err
	}
	count := incr.Val()
	if count > int64(limit) {
		ttl, err := r.rdb.TTL(ctx, key).Result()
		if err != nil || ttl < 0 {
			ttl = window
		}
		return false, ttl, nil
	}
	return true, 0, nil
}
