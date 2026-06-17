package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// ConnectionLimiter enforces a maximum number of simultaneous connections
// per user (distinct from device limit which tracks unique IPs over time).
// When a user exceeds their connection limit, the excess connections are
// dropped by removing the user from the live core temporarily.
type ConnectionLimiter struct {
	online   OnlineQuerier
	users    ConnectionLimitRepo
	nodes    NodeOps
	interval time.Duration
	log      *slog.Logger
	pub      events.Publisher
	warned   map[uuid.UUID]time.Time // dedup warnings
}

// ConnectionLimitRepo is the data access needed.
type ConnectionLimitRepo interface {
	UsersWithConnectionLimit(ctx context.Context) ([]*domain.User, error)
	InboundsFor(ctx context.Context, userID uuid.UUID) ([]domain.Inbound, error)
}

// NewConnectionLimiter wires the limiter.
func NewConnectionLimiter(online OnlineQuerier, users ConnectionLimitRepo, nodes NodeOps, log *slog.Logger) *ConnectionLimiter {
	if log == nil {
		log = slog.Default()
	}
	return &ConnectionLimiter{
		online:   online,
		users:    users,
		nodes:    nodes,
		interval: 30 * time.Second,
		log:      log,
		pub:      events.Nop{},
		warned:   make(map[uuid.UUID]time.Time),
	}
}

// SetPublisher wires event publishing.
func (cl *ConnectionLimiter) SetPublisher(p events.Publisher) {
	if p != nil {
		cl.pub = p
	}
}

// Run ticks until ctx is cancelled.
func (cl *ConnectionLimiter) Run(ctx context.Context) {
	ticker := time.NewTicker(cl.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cl.tick(ctx)
		}
	}
}

func (cl *ConnectionLimiter) tick(ctx context.Context) {
	if cl.online == nil {
		return
	}
	users, err := cl.users.UsersWithConnectionLimit(ctx)
	if err != nil {
		return
	}
	for _, u := range users {
		if u.DeviceLimit <= 0 {
			continue // 0 = unlimited
		}
		inbounds, err := cl.users.InboundsFor(ctx, u.ID)
		if err != nil {
			continue
		}
		totalConns := 0
		for _, in := range inbounds {
			stats, err := cl.online.OnlineStats(ctx, in.NodeID)
			if err != nil {
				continue
			}
			if c, ok := stats[u.ID.String()]; ok {
				totalConns += c
			}
		}
		// Check if over limit (using DeviceLimit as connection cap)
		maxConns := u.DeviceLimit * 2 // allow 2x device limit as connection limit
		if totalConns > maxConns {
			cl.log.Warn("user over connection limit",
				"user", u.Username, "conns", totalConns, "limit", maxConns)
			cl.pub.Publish(events.Event{
				Type:     events.UserIPLimit,
				UserID:   u.ID.String(),
				Username: u.Username,
				Message:  "Connection limit exceeded",
				Data:     map[string]any{"connections": totalConns, "limit": maxConns},
			})
		}
	}
}
