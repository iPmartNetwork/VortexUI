package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// ExpiryWarner periodically checks for users approaching their expiry date and
// emits a warning event (picked up by Telegram/Webhook notifiers). It fires once
// per user per warning window to avoid spamming.
type ExpiryWarner struct {
	repo     ExpiryRepo
	interval time.Duration
	warnDays int // fire warning N days before expiry
	now      func() time.Time
	log      *slog.Logger
	pub      events.Publisher
	warned   map[string]time.Time // userID -> last warned at (in-memory dedup)
}

// ExpiryRepo is the data access the expiry warner needs.
type ExpiryRepo interface {
	UsersExpiringSoon(ctx context.Context, before time.Time) ([]*domain.User, error)
}

// NewExpiryWarner wires the loop. Defaults: check every hour, warn 3 days before.
func NewExpiryWarner(repo ExpiryRepo, log *slog.Logger) *ExpiryWarner {
	if log == nil {
		log = slog.Default()
	}
	return &ExpiryWarner{
		repo:     repo,
		interval: time.Hour,
		warnDays: 3,
		now:      time.Now,
		log:      log,
		pub:      events.Nop{},
		warned:   make(map[string]time.Time),
	}
}

// SetPublisher wires the event publisher.
func (w *ExpiryWarner) SetPublisher(p events.Publisher) {
	if p != nil {
		w.pub = p
	}
}

// Run ticks until ctx is cancelled.
func (w *ExpiryWarner) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *ExpiryWarner) tick(ctx context.Context) {
	cutoff := w.now().Add(time.Duration(w.warnDays) * 24 * time.Hour)
	users, err := w.repo.UsersExpiringSoon(ctx, cutoff)
	if err != nil {
		w.log.Warn("expiry warner query failed", "err", err)
		return
	}
	now := w.now()
	for _, u := range users {
		key := u.ID.String()
		// Dedup: only warn once per 24h per user
		if last, ok := w.warned[key]; ok && now.Sub(last) < 24*time.Hour {
			continue
		}
		daysLeft := 0
		if u.ExpireAt != nil {
			daysLeft = int(time.Until(*u.ExpireAt).Hours() / 24)
		}
		w.pub.Publish(events.Event{
			Type:     events.UserExpiryWarning,
			UserID:   u.ID.String(),
			Username: u.Username,
			Message:  "User expiring soon",
			Data:     map[string]any{"days_left": daysLeft},
		})
		w.warned[key] = now
		w.log.Info("expiry warning sent", "user", u.Username, "days_left", daysLeft)
	}
	// Prune old entries
	for k, t := range w.warned {
		if now.Sub(t) > 7*24*time.Hour {
			delete(w.warned, k)
		}
	}
}
