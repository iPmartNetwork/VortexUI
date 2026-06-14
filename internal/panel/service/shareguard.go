package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// ShareGuard periodically detects account sharing: a user online from more
// distinct source IPs than its device limit allows. Each violation emits a
// user.ip_limit event (so notifiers alert the admin). When auto-limit is enabled
// it additionally flips the user to "limited" and removes it from the live cores
// — the same hard action as data/expiry enforcement, but reversible by the admin.
type ShareGuard struct {
	users     *UserService
	repo      EnforceRepo // for auto-limit (status update + inbound lookup)
	nodes     NodeOps
	interval  time.Duration
	cooldown  time.Duration // minimum gap between repeat alerts for the same user
	autoLimit bool
	now       func() time.Time
	log       *slog.Logger
	pub       events.Publisher
	alerted   map[uuid.UUID]time.Time
}

// NewShareGuard wires the guard. interval of 0 defaults to two minutes.
func NewShareGuard(users *UserService, repo EnforceRepo, nodes NodeOps, interval time.Duration, log *slog.Logger) *ShareGuard {
	if interval == 0 {
		interval = 2 * time.Minute
	}
	if log == nil {
		log = slog.Default()
	}
	return &ShareGuard{
		users:    users,
		repo:     repo,
		nodes:    nodes,
		interval: interval,
		cooldown: 15 * time.Minute,
		now:      time.Now,
		log:      log,
		pub:      events.Nop{},
		alerted:  map[uuid.UUID]time.Time{},
	}
}

// SetPublisher wires the event bus so violations raise alerts.
func (g *ShareGuard) SetPublisher(p events.Publisher) {
	if p != nil {
		g.pub = p
	}
}

// SetAutoLimit toggles whether detected sharers are actually limited (true) or
// only alerted (false).
func (g *ShareGuard) SetAutoLimit(enabled bool) { g.autoLimit = enabled }

// Run ticks until ctx is cancelled.
func (g *ShareGuard) Run(ctx context.Context) {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := g.Tick(ctx); err != nil {
				g.log.Warn("shareguard tick failed", "err", err)
			}
		}
	}
}

// Tick runs one detection pass. Exposed for deterministic testing.
func (g *ShareGuard) Tick(ctx context.Context) error {
	users, err := g.users.ActiveWithDeviceLimit(ctx)
	if err != nil {
		return err
	}
	now := g.now()
	for _, u := range users {
		ips, tracked, err := g.users.OnlineIPList(ctx, u.ID)
		if err != nil || !tracked {
			continue
		}
		if len(ips) <= u.DeviceLimit {
			continue
		}
		g.handleViolation(ctx, u, len(ips), now)
	}
	return nil
}

// handleViolation alerts (rate-limited per user) and optionally limits the user.
func (g *ShareGuard) handleViolation(ctx context.Context, u *domain.User, ipCount int, now time.Time) {
	if last, ok := g.alerted[u.ID]; ok && now.Sub(last) < g.cooldown {
		return
	}
	g.alerted[u.ID] = now
	g.log.Info("account sharing detected", "user", u.Username, "online_ips", ipCount, "device_limit", u.DeviceLimit, "auto_limit", g.autoLimit)
	g.pub.Publish(events.Event{
		Type:     events.UserIPLimit,
		UserID:   u.ID.String(),
		Username: u.Username,
		Message:  "account sharing detected",
		Data:     map[string]any{"online_ips": ipCount, "device_limit": u.DeviceLimit, "limited": g.autoLimit},
	})

	if !g.autoLimit {
		return
	}
	u.Status = domain.UserStatusLimited
	if err := g.repo.Update(ctx, u); err != nil {
		g.log.Warn("shareguard: limit update failed", "user", u.ID, "err", err)
		return
	}
	inbounds, err := g.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		g.log.Warn("shareguard: inbounds lookup failed", "user", u.ID, "err", err)
		return
	}
	for _, in := range inbounds {
		_ = g.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID)
	}
}
