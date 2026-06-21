package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// IPLimitStore is the slice of the IP-limit repository ShareGuard needs to read
// the enforcement policy and persist enforcement events. *postgres.IPLimitRepo
// satisfies it.
type IPLimitStore interface {
	GetPolicy(ctx context.Context) (*domain.IPLimitPolicy, error)
	InsertEvent(ctx context.Context, e *domain.IPLimitEvent) error
}

// NodeCoreLookup resolves a node's core engine so kill_connections can choose
// the xray (true kill) vs sing-box (degrade-to-disable) branch. *postgres.NodeRepo
// satisfies it.
type NodeCoreLookup interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Node, error)
}

// ShareGuard periodically detects account sharing: a user online from more
// distinct source IPs than its device limit allows. Each violation emits a
// user.ip_limit event (so notifiers alert the admin). When auto-limit is enabled
// it additionally flips the user to "limited" and removes it from the live cores
// — the same hard action as data/expiry enforcement, but reversible by the admin.
//
// When an IP-limit policy is wired and Enabled, ShareGuard instead branches on
// the configured action (warn / disable_temporarily / kill_connections) and
// records each action to the event store. With no policy (or a disabled one) the
// legacy behavior above is preserved byte-for-byte.
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

	// IP-limit enforcement (optional; nil = legacy detection-only behavior).
	iplimit IPLimitStore
	cores   NodeCoreLookup

	// restores tracks users with a pending auto-restore so disable_temporarily
	// never double-schedules. afterFunc is injectable for deterministic tests.
	mu        sync.Mutex
	restores  map[uuid.UUID]struct{}
	afterFunc func(time.Duration, func()) *time.Timer
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
		users:     users,
		repo:      repo,
		nodes:     nodes,
		interval:  interval,
		cooldown:  15 * time.Minute,
		now:       time.Now,
		log:       log,
		pub:       events.Nop{},
		alerted:   map[uuid.UUID]time.Time{},
		restores:  map[uuid.UUID]struct{}{},
		afterFunc: time.AfterFunc,
	}
}

// SetPublisher wires the event bus so violations raise alerts.
func (g *ShareGuard) SetPublisher(p events.Publisher) {
	if p != nil {
		g.pub = p
	}
}

// SetAutoLimit toggles whether detected sharers are actually limited (true) or
// only alerted (false). This governs the legacy path used when no IP-limit
// policy is enabled.
func (g *ShareGuard) SetAutoLimit(enabled bool) { g.autoLimit = enabled }

// SetIPLimit wires the IP-limit enforcement policy source, event store, and the
// per-node core lookup. Without it (or with a disabled policy) ShareGuard keeps
// its legacy detection-only / auto-limit behavior unchanged.
func (g *ShareGuard) SetIPLimit(store IPLimitStore, cores NodeCoreLookup) {
	g.iplimit = store
	g.cores = cores
}

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
	pol := g.loadPolicy(ctx)
	now := g.now()
	for _, u := range users {
		ips, tracked, err := g.users.OnlineIPList(ctx, u.ID)
		// Fail open: a tracker error (or no tracking wired) must never lead to an
		// enforcement action against a possibly-legitimate user.
		if err != nil || !tracked {
			continue
		}
		if len(ips) <= u.DeviceLimit {
			continue
		}
		g.handleViolation(ctx, u, len(ips), now, pol)
	}
	return nil
}

// loadPolicy reads the enforcement policy for this pass. A nil result (no store
// wired, or a read error) means "no enforcement" — the guard then falls back to
// its legacy behavior and never acts on a fault.
func (g *ShareGuard) loadPolicy(ctx context.Context) *domain.IPLimitPolicy {
	if g.iplimit == nil {
		return nil
	}
	pol, err := g.iplimit.GetPolicy(ctx)
	if err != nil {
		g.log.Warn("shareguard: ip-limit policy load failed; falling back to legacy behavior", "err", err)
		return nil
	}
	return pol
}

// handleViolation dispatches a detected violation to the enforced policy path
// when one is enabled, otherwise to the unchanged legacy path.
func (g *ShareGuard) handleViolation(ctx context.Context, u *domain.User, ipCount int, now time.Time, pol *domain.IPLimitPolicy) {
	if pol == nil || !pol.Enabled {
		g.handleViolationLegacy(ctx, u, ipCount, now)
		return
	}
	g.handleViolationEnforced(ctx, u, ipCount, now, pol)
}

// handleViolationLegacy is the original behavior: alert (rate-limited per user)
// and, when auto-limit is on, flip the user to limited and deprovision it. This
// runs whenever IP-limit enforcement is disabled, so detection-only behavior is
// preserved with no regression.
func (g *ShareGuard) handleViolationLegacy(ctx context.Context, u *domain.User, ipCount int, now time.Time) {
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

// handleViolationEnforced applies the configured IP-limit policy: it dedups per
// user using the policy's AlertCooldown, branches on the action, publishes the
// alert event, and records the effective action to the event store.
func (g *ShareGuard) handleViolationEnforced(ctx context.Context, u *domain.User, ipCount int, now time.Time, pol *domain.IPLimitPolicy) {
	cooldown := g.cooldown
	if pol.AlertCooldown > 0 {
		cooldown = time.Duration(pol.AlertCooldown) * time.Second
	}
	if last, ok := g.alerted[u.ID]; ok && now.Sub(last) < cooldown {
		return
	}
	g.alerted[u.ID] = now

	effective := string(pol.Action)
	switch pol.Action {
	case domain.IPLimitActionDisable:
		g.disableTemporarily(ctx, u, pol)
	case domain.IPLimitActionKill:
		effective = g.killConnections(ctx, u, pol)
	case domain.IPLimitActionWarn:
		// Event only — no action against the user.
	default:
		// Unknown action degrades to warn so a misconfiguration never silently
		// takes a destructive action.
		effective = string(domain.IPLimitActionWarn)
	}

	g.log.Info("ip-limit enforcement", "user", u.Username, "online_ips", ipCount, "device_limit", u.DeviceLimit, "action", effective)
	g.pub.Publish(events.Event{
		Type:     events.UserIPLimit,
		UserID:   u.ID.String(),
		Username: u.Username,
		Message:  "account sharing detected",
		Data:     map[string]any{"online_ips": ipCount, "device_limit": u.DeviceLimit, "action": effective},
	})
	g.insertEvent(ctx, u, ipCount, effective)
}

// disableTemporarily flips the user to limited, deprovisions it from its nodes,
// and schedules an auto-restore after the policy's RestoreAfter window.
func (g *ShareGuard) disableTemporarily(ctx context.Context, u *domain.User, pol *domain.IPLimitPolicy) {
	u.Status = domain.UserStatusLimited
	if err := g.repo.Update(ctx, u); err != nil {
		g.log.Warn("ip-limit: limit update failed", "user", u.ID, "err", err)
		return
	}
	inbounds, err := g.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		g.log.Warn("ip-limit: inbounds lookup failed", "user", u.ID, "err", err)
		return
	}
	for _, in := range inbounds {
		_ = g.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID)
	}
	g.scheduleRestore(u, pol.RestoreAfter)
}

// scheduleRestore arms a single auto-restore for the user. If one is already
// pending it does not double-schedule.
func (g *ShareGuard) scheduleRestore(u *domain.User, afterSec int) {
	g.mu.Lock()
	if _, pending := g.restores[u.ID]; pending {
		g.mu.Unlock()
		return
	}
	g.restores[u.ID] = struct{}{}
	g.mu.Unlock()

	d := time.Duration(afterSec) * time.Second
	g.afterFunc(d, func() { g.restoreUser(context.Background(), u) })
}

// restoreUser sets the user back to active and re-provisions it on its nodes,
// undoing a temporary disable. It always clears the pending-restore marker so a
// future violation can be acted upon again.
func (g *ShareGuard) restoreUser(ctx context.Context, u *domain.User) {
	g.mu.Lock()
	delete(g.restores, u.ID)
	g.mu.Unlock()

	u.Status = domain.UserStatusActive
	if err := g.repo.Update(ctx, u); err != nil {
		g.log.Warn("ip-limit: restore update failed", "user", u.ID, "err", err)
		return
	}
	inbounds, err := g.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		g.log.Warn("ip-limit: restore inbounds lookup failed", "user", u.ID, "err", err)
		return
	}
	for _, in := range inbounds {
		_ = g.nodes.AddUser(ctx, in.NodeID, in.Tag, u)
	}
	g.log.Info("ip-limit: user restored after temporary disable", "user", u.Username)
}

// killConnections drops the user's live sessions. On xray nodes it removes then
// immediately re-adds the user on the core (xray drops active sessions on
// removal) so the still-valid account stays provisioned. If any bound node runs
// sing-box — which has no runtime online API — true kill is unavailable, so the
// whole user degrades to disable_temporarily and the degradation is logged. The
// returned string is the effective action recorded in the event.
func (g *ShareGuard) killConnections(ctx context.Context, u *domain.User, pol *domain.IPLimitPolicy) string {
	inbounds, err := g.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		g.log.Warn("ip-limit: inbounds lookup failed", "user", u.ID, "err", err)
		return string(domain.IPLimitActionKill)
	}
	for _, in := range inbounds {
		if g.coreKind(ctx, in.NodeID) == domain.CoreSingbox {
			g.log.Info("ip-limit: kill_connections unsupported on sing-box; degrading to disable_temporarily", "user", u.Username, "node", in.NodeID)
			g.disableTemporarily(ctx, u, pol)
			return string(domain.IPLimitActionDisable)
		}
	}
	for _, in := range inbounds {
		// Remove drops live sessions on xray; immediately re-add so the valid
		// account is re-provisioned (Req 5.4.2 — a transient kill is not a revoke).
		_ = g.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID)
		_ = g.nodes.AddUser(ctx, in.NodeID, in.Tag, u)
	}
	return string(domain.IPLimitActionKill)
}

// coreKind resolves a node's core engine, defaulting to xray when the lookup is
// unavailable (the xray path is non-destructive: remove+re-add keeps the user
// provisioned regardless of the real core).
func (g *ShareGuard) coreKind(ctx context.Context, nodeID uuid.UUID) domain.CoreType {
	if g.cores == nil {
		return domain.CoreXray
	}
	n, err := g.cores.GetByID(ctx, nodeID)
	if err != nil || n == nil {
		return domain.CoreXray
	}
	return n.Core
}

// insertEvent records an enforcement event for the admin audit trail.
func (g *ShareGuard) insertEvent(ctx context.Context, u *domain.User, ipCount int, action string) {
	if g.iplimit == nil {
		return
	}
	ev := &domain.IPLimitEvent{
		ID:        uuid.New(),
		UserID:    u.ID,
		Username:  u.Username,
		OnlineIPs: ipCount,
		Limit:     u.DeviceLimit,
		Action:    action,
		CreatedAt: g.now(),
	}
	if err := g.iplimit.InsertEvent(ctx, ev); err != nil {
		g.log.Warn("ip-limit: event insert failed", "user", u.ID, "err", err)
	}
}
