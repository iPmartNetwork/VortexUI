package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// EnforceRepo is the data access the enforcement loop needs. *postgres.UserRepo
// satisfies it (UsersToLimit is concrete-only; the rest are on the port).
type EnforceRepo interface {
	UsersToLimit(ctx context.Context) ([]*domain.User, error)
	InboundsFor(ctx context.Context, userID uuid.UUID) ([]domain.Inbound, error)
	Update(ctx context.Context, u *domain.User) error
}

// Enforcer periodically finds users who have hit their data cap or expiry and
// makes that real: it flips their status and removes them from the live cores.
// Without it, limits would be mere bookkeeping that never actually cuts access.
type Enforcer struct {
	repo     EnforceRepo
	nodes    NodeOps
	interval time.Duration
	now      func() time.Time
	log      *slog.Logger
	pub      events.Publisher
}

// NewEnforcer wires the loop. interval of 0 defaults to one minute.
func NewEnforcer(repo EnforceRepo, nodes NodeOps, interval time.Duration, log *slog.Logger) *Enforcer {
	if interval == 0 {
		interval = time.Minute
	}
	if log == nil {
		log = slog.Default()
	}
	return &Enforcer{repo: repo, nodes: nodes, interval: interval, now: time.Now, log: log, pub: events.Nop{}}
}

// SetPublisher wires an event publisher (so enforcement emits user.limited /
// user.expired). A nil publisher leaves the no-op default in place.
func (e *Enforcer) SetPublisher(p events.Publisher) {
	if p != nil {
		e.pub = p
	}
}

// Run ticks until ctx is cancelled, running one enforcement pass per interval.
func (e *Enforcer) Run(ctx context.Context) {
	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := e.Tick(ctx); err != nil {
				e.log.Warn("enforcement tick failed", "err", err)
			}
		}
	}
}

// Tick performs one enforcement pass. Exposed (not just Run) so it is unit
// tested deterministically without waiting on a timer.
func (e *Enforcer) Tick(ctx context.Context) error {
	users, err := e.repo.UsersToLimit(ctx)
	if err != nil {
		return err
	}
	now := e.now()
	for _, u := range users {
		// Derive the precise reason (limited vs expired) and persist it.
		u.Status = u.DerivedStatus(now)
		if err := e.repo.Update(ctx, u); err != nil {
			e.log.Warn("enforce: update status failed", "user", u.ID, "err", err)
			continue
		}
		e.deprovision(ctx, u)
		e.log.Info("enforced user limit", "user", u.Username, "status", u.Status)
		e.publishStatus(u)
	}
	return nil
}

// publishStatus emits the event matching a freshly-enforced terminal status.
func (e *Enforcer) publishStatus(u *domain.User) {
	ev := events.Event{UserID: u.ID.String(), Username: u.Username}
	switch u.Status {
	case domain.UserStatusLimited:
		ev.Type = events.UserLimited
	case domain.UserStatusExpired:
		ev.Type = events.UserExpired
	default:
		return
	}
	e.pub.Publish(ev)
}

// deprovision removes the user from every node it is bound to, best-effort so a
// single unreachable node does not stall enforcement of the rest.
func (e *Enforcer) deprovision(ctx context.Context, u *domain.User) {
	inbounds, err := e.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		e.log.Error("enforce: inbounds lookup failed during deprovision", "user", u.ID, "err", err)
		return
	}
	for _, in := range inbounds {
		if err := e.nodes.RemoveUser(ctx, in.NodeID, in.Tag, u.ID); err != nil {
			e.log.Error("enforce: failed to remove user from node", "user", u.ID, "node", in.NodeID, "err", err)
		}
	}
}
