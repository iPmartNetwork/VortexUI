package service

import (
	"context"
	"fmt"
	"time"
)

// CacheBackend is the interface for a key-value cache store (e.g. Redis).
type CacheBackend interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
}

// SubscriptionCache provides Redis-backed caching for subscription outputs.
// It invalidates on config, SNI, and node changes to ensure freshness.
type SubscriptionCache struct {
	backend CacheBackend
	ttl     time.Duration
}

// NewSubscriptionCache creates a cache with the given backend and default TTL.
func NewSubscriptionCache(backend CacheBackend, ttl time.Duration) *SubscriptionCache {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}
	return &SubscriptionCache{backend: backend, ttl: ttl}
}

// Get retrieves a cached subscription output for the given user+format.
func (c *SubscriptionCache) Get(ctx context.Context, userID, format string) ([]byte, error) {
	key := c.userKey(userID, format)
	return c.backend.Get(ctx, key)
}

// Set stores a subscription output in the cache.
func (c *SubscriptionCache) Set(ctx context.Context, userID, format string, data []byte) error {
	key := c.userKey(userID, format)
	return c.backend.Set(ctx, key, data, c.ttl)
}

// InvalidateUser removes all cached subscription outputs for a user.
func (c *SubscriptionCache) InvalidateUser(ctx context.Context, userID string) error {
	prefix := fmt.Sprintf("sub:%s:", userID)
	return c.backend.DeleteByPrefix(ctx, prefix)
}

// InvalidateNode removes all cached outputs that reference the given node.
// This is a broad invalidation — all user caches are cleared since node changes
// affect all subscriptions.
func (c *SubscriptionCache) InvalidateNode(ctx context.Context, nodeID string) error {
	// Node changes are broad — invalidate all subscription caches.
	return c.backend.DeleteByPrefix(ctx, "sub:")
}

// InvalidateAll clears the entire subscription cache.
func (c *SubscriptionCache) InvalidateAll(ctx context.Context) error {
	return c.backend.DeleteByPrefix(ctx, "sub:")
}

func (c *SubscriptionCache) userKey(userID, format string) string {
	return fmt.Sprintf("sub:%s:%s", userID, format)
}
