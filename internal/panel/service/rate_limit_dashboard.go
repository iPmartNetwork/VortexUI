package service

import (
	"context"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// RateLimitDashboard tracks and exposes per-endpoint rate limit consumption stats.
type RateLimitDashboard struct {
	mu       sync.RWMutex
	counters map[string]*endpointCounter
}

type endpointCounter struct {
	Endpoint  string
	Limit     int
	WindowSec int
	Hits      int
	WindowStart time.Time
}

// NewRateLimitDashboard creates a new dashboard tracker.
func NewRateLimitDashboard() *RateLimitDashboard {
	return &RateLimitDashboard{
		counters: make(map[string]*endpointCounter),
	}
}

// RegisterEndpoint configures rate limit parameters for an endpoint.
func (d *RateLimitDashboard) RegisterEndpoint(endpoint string, limit, windowSec int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.counters[endpoint] = &endpointCounter{
		Endpoint:    endpoint,
		Limit:       limit,
		WindowSec:   windowSec,
		WindowStart: time.Now(),
	}
}

// RecordHit increments the request counter for an endpoint.
func (d *RateLimitDashboard) RecordHit(endpoint string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	c, ok := d.counters[endpoint]
	if !ok {
		return
	}
	// Reset window if expired
	if time.Since(c.WindowStart) > time.Duration(c.WindowSec)*time.Second {
		c.Hits = 0
		c.WindowStart = time.Now()
	}
	c.Hits++
}

// GetStats returns rate limit info for all registered endpoints.
func (d *RateLimitDashboard) GetStats(_ context.Context) []domain.RateLimitInfo {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var stats []domain.RateLimitInfo
	now := time.Now()
	for _, c := range d.counters {
		remaining := c.Limit - c.Hits
		if remaining < 0 {
			remaining = 0
		}
		resetAt := c.WindowStart.Add(time.Duration(c.WindowSec) * time.Second)
		if now.After(resetAt) {
			remaining = c.Limit
			resetAt = now.Add(time.Duration(c.WindowSec) * time.Second)
		}
		stats = append(stats, domain.RateLimitInfo{
			Endpoint:  c.Endpoint,
			Limit:     c.Limit,
			Remaining: remaining,
			WindowSec: c.WindowSec,
			ResetAt:   resetAt.Unix(),
		})
	}
	return stats
}
