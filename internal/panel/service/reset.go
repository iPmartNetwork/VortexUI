package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// ResetRepo is the data access the reset loop needs. *postgres.UserRepo
// satisfies it (UsersToReset is concrete-only; the rest are on the port).
type ResetRepo interface {
	UsersToReset(ctx context.Context) ([]*domain.User, error)
	InboundsFor(ctx context.Context, userID uuid.UUID) ([]domain.Inbound, error)
	Update(ctx context.Context, u *domain.User) error
}

// Resetter periodically zeroes the used-traffic counter for users on a reset
// schedule (daily/weekly/monthly) and re-activates any that were limited purely
// because of the quota — the complement to the Enforcer.
type Resetter struct {
	repo     ResetRepo
	nodes    NodeOps
	interval time.Duration
	now      func() time.Time
	log      *slog.Logger
}

// NewResetter wires the loop. interval of 0 defaults to one hour, which is a fine
// granularity for day/week/month windows.
func NewResetter(repo ResetRepo, nodes NodeOps, interval time.Duration, log *slog.Logger) *Resetter {
	if interval == 0 {
		interval = time.Hour
	}
	if log == nil {
		log = slog.Default()
	}
	return &Resetter{repo: repo, nodes: nodes, interval: interval, now: time.Now, log: log}
}

// Run ticks until ctx is cancelled.
func (r *Resetter) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.Tick(ctx); err != nil {
				r.log.Warn("reset tick failed", "err", err)
			}
		}
	}
}

// Tick performs one reset pass. Exposed for deterministic unit testing.
func (r *Resetter) Tick(ctx context.Context) error {
	users, err := r.repo.UsersToReset(ctx)
	if err != nil {
		return err
	}
	now := r.now()
	for _, u := range users {
		wasActive := u.IsActive(now)

		u.UsedTraffic = 0
		u.LastReset = &now
		u.Status = u.DerivedStatus(now) // may clear a quota-driven "limited"
		if err := r.repo.Update(ctx, u); err != nil {
			r.log.Warn("reset: update failed", "user", u.ID, "err", err)
			continue
		}

		// If the reset brought a previously-limited user back to active, restore
		// their access on the live cores.
		if !wasActive && u.Status == domain.UserStatusActive {
			r.reprovision(ctx, u)
		}
		r.log.Info("reset user traffic", "user", u.Username, "status", u.Status)
	}
	return nil
}

func (r *Resetter) reprovision(ctx context.Context, u *domain.User) {
	inbounds, err := r.repo.InboundsFor(ctx, u.ID)
	if err != nil {
		r.log.Warn("reset: inbounds lookup failed", "user", u.ID, "err", err)
		return
	}
	for _, in := range inbounds {
		_ = r.nodes.AddUser(ctx, in.NodeID, in.Tag, u)
	}
}
