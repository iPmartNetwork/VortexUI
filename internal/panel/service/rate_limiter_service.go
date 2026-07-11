package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// InMemoryRateLimiter implements port.RateLimiter using in-memory storage
type InMemoryRateLimiter struct {
	ruleRepo    port.RateLimitRepository
	buckets     map[string]*rateLimitBucket
	blockList   map[string]*time.Time
	mu          sync.RWMutex
	log         *slog.Logger
}

type rateLimitBucket struct {
	requestCount int
	resetTime    time.Time
	lastRequest  time.Time
	burstTokens  int
}

// NewInMemoryRateLimiter creates new rate limiter
func NewInMemoryRateLimiter(ruleRepo port.RateLimitRepository, log *slog.Logger) *InMemoryRateLimiter {
	if log == nil {
		log = slog.Default()
	}

	return &InMemoryRateLimiter{
		ruleRepo:  ruleRepo,
		buckets:   make(map[string]*rateLimitBucket),
		blockList: make(map[string]*time.Time),
		log:       log,
	}
}

// IsAllowed checks if request should be allowed
func (r *InMemoryRateLimiter) IsAllowed(ctx context.Context, clientIP, endpoint, method string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if client is blocked
	if blockedUntil, exists := r.blockList[clientIP]; exists {
		if time.Now().Before(*blockedUntil) {
			return false, nil
		}
		delete(r.blockList, clientIP)
	}

	// Get rate limit rule
	rule, err := r.ruleRepo.GetRuleByEndpoint(ctx, endpoint, method)
	if err != nil {
		// No rule means unlimited
		return true, nil
	}

	if !rule.Enabled {
		return true, nil
	}

	key := clientIP + ":" + endpoint + ":" + method
	bucket, exists := r.buckets[key]

	now := time.Now()
	if !exists {
		// Create new bucket
		bucket = &rateLimitBucket{
			requestCount: 1,
			resetTime:    now.Add(time.Minute),
			lastRequest:  now,
			burstTokens:  rule.BurstSize - 1,
		}
		r.buckets[key] = bucket
		return true, nil
	}

	// Check if bucket needs reset
	if now.After(bucket.resetTime) {
		bucket.requestCount = 1
		bucket.resetTime = now.Add(time.Minute)
		bucket.burstTokens = rule.BurstSize - 1
		bucket.lastRequest = now
		return true, nil
	}

	// Check rate limit
	if bucket.requestCount >= rule.RequestsPerMin {
		// Check burst allowance
		if bucket.burstTokens > 0 {
			bucket.burstTokens--
			bucket.requestCount++
			bucket.lastRequest = now

			// Log violation
			violation := &domain.RateLimitViolation{
				ID:           uuid.New(),
				RuleID:       rule.ID,
				ClientIP:     clientIP,
				Endpoint:     endpoint,
				RequestCount: bucket.requestCount,
				Limit:        rule.RequestsPerMin,
				Action:       "throttle",
				CreatedAt:    now,
			}
			_ = r.ruleRepo.LogViolation(ctx, violation)

			return true, nil
		}

		// Block client
		blockUntil := now.Add(5 * time.Minute)
		r.blockList[clientIP] = &blockUntil

		violation := &domain.RateLimitViolation{
			ID:           uuid.New(),
			RuleID:       rule.ID,
			ClientIP:     clientIP,
			Endpoint:     endpoint,
			RequestCount: bucket.requestCount,
			Limit:        rule.RequestsPerMin,
			Action:       "block",
			BlockedUntil: &blockUntil,
			CreatedAt:    now,
		}
		_ = r.ruleRepo.LogViolation(ctx, violation)

		return false, nil
	}

	bucket.requestCount++
	bucket.lastRequest = now
	return true, nil
}

// GetRemainingQuota returns remaining quota for client
func (r *InMemoryRateLimiter) GetRemainingQuota(ctx context.Context, clientIP, endpoint, method string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, err := r.ruleRepo.GetRuleByEndpoint(ctx, endpoint, method)
	if err != nil {
		return 0, nil
	}

	key := clientIP + ":" + endpoint + ":" + method
	bucket, exists := r.buckets[key]
	if !exists {
		return rule.RequestsPerMin, nil
	}

	remaining := rule.RequestsPerMin - bucket.requestCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// BlockClient blocks client temporarily
func (r *InMemoryRateLimiter) BlockClient(ctx context.Context, clientIP string, durationMinutes int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	blockUntil := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	r.blockList[clientIP] = &blockUntil

	return nil
}

// UnblockClient removes client block
func (r *InMemoryRateLimiter) UnblockClient(ctx context.Context, clientIP string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.blockList, clientIP)
	return nil
}

// IsClientBlocked checks if client is blocked
func (r *InMemoryRateLimiter) IsClientBlocked(ctx context.Context, clientIP string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if blockedUntil, exists := r.blockList[clientIP]; exists {
		if time.Now().Before(*blockedUntil) {
			return true, nil
		}
		delete(r.blockList, clientIP)
	}

	return false, nil
}
