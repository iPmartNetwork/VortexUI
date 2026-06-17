// Package cluster provides multi-panel high availability. Multiple panel
// instances share the same PostgreSQL database and Redis, each running the full
// stack. A leader election (via Redis or PostgreSQL advisory locks) ensures only
// one instance runs the singleton loops (enforcement, reset, expiry warning,
// share guard) at a time — the rest are hot standby and handle API/gRPC traffic.
//
// Architecture:
//   - All panels share one PostgreSQL + Redis
//   - Every panel serves API + web traffic (behind a load balancer)
//   - Node agents connect to any panel (or all via DNS round-robin)
//   - One leader runs background loops; on failure another takes over
//   - Hub state is ephemeral (rebuilt from DB on startup) — no shared memory needed
package cluster

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

// LeaderElector decides which panel instance runs the singleton background loops.
type LeaderElector interface {
	// TryAcquire attempts to become leader. Returns true if this instance won.
	TryAcquire(ctx context.Context) (bool, error)
	// Renew extends the leadership lease. Must be called periodically.
	Renew(ctx context.Context) error
	// Release gives up leadership (graceful shutdown).
	Release(ctx context.Context) error
}

// Coordinator manages the leader election loop and exposes an IsLeader() check
// that background services query before running their ticks.
type Coordinator struct {
	elector  LeaderElector
	log      *slog.Logger
	interval time.Duration
	isLeader atomic.Bool
}

// NewCoordinator builds a coordinator.
func NewCoordinator(elector LeaderElector, log *slog.Logger) *Coordinator {
	if log == nil {
		log = slog.Default()
	}
	return &Coordinator{
		elector:  elector,
		log:      log,
		interval: 5 * time.Second,
	}
}

// IsLeader reports whether this instance currently holds the leader lock.
func (c *Coordinator) IsLeader() bool {
	return c.isLeader.Load()
}

// Run starts the election loop. Blocks until ctx is cancelled.
func (c *Coordinator) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			c.release(ctx)
			return
		case <-ticker.C:
			c.tick(ctx)
		}
	}
}

func (c *Coordinator) tick(ctx context.Context) {
	if c.isLeader.Load() {
		if err := c.elector.Renew(ctx); err != nil {
			c.log.Warn("leadership renewal failed, stepping down", "err", err)
			c.isLeader.Store(false)
		}
		return
	}
	won, err := c.elector.TryAcquire(ctx)
	if err != nil {
		return
	}
	if won {
		c.log.Info("acquired leadership — running singleton loops")
		c.isLeader.Store(true)
	}
}

func (c *Coordinator) release(ctx context.Context) {
	if c.isLeader.Load() {
		_ = c.elector.Release(ctx)
		c.isLeader.Store(false)
		c.log.Info("released leadership")
	}
}

// RedisElector implements LeaderElector using a Redis key with TTL (SET NX EX).
type RedisElector struct {
	InstanceID string // unique per panel instance (e.g. hostname + PID)
	Key        string // Redis key for the lock
	TTL        time.Duration
	Client     RedisClient
}

// RedisClient is the minimal Redis interface needed for leader election.
type RedisClient interface {
	SetNX(ctx context.Context, key, value string, ttl time.Duration) (bool, error)
	Get(ctx context.Context, key string) (string, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	Del(ctx context.Context, key string) error
}

func (r *RedisElector) TryAcquire(ctx context.Context) (bool, error) {
	return r.Client.SetNX(ctx, r.Key, r.InstanceID, r.TTL)
}

func (r *RedisElector) Renew(ctx context.Context) error {
	// Only renew if we still own it
	val, err := r.Client.Get(ctx, r.Key)
	if err != nil || val != r.InstanceID {
		return err
	}
	return r.Client.Expire(ctx, r.Key, r.TTL)
}

func (r *RedisElector) Release(ctx context.Context) error {
	val, err := r.Client.Get(ctx, r.Key)
	if err != nil || val != r.InstanceID {
		return nil // someone else took it
	}
	return r.Client.Del(ctx, r.Key)
}
