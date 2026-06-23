package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// ResellerIPViolationCounter counts IP-limit events for a reseller's users.
type ResellerIPViolationCounter interface {
	CountIPLimitEventsForAdminSince(ctx context.Context, adminID uuid.UUID, since time.Time) (int64, error)
}

// ResellerUserDisabler disables all users owned by a reseller.
type ResellerUserDisabler interface {
	DisableAllForAdmin(ctx context.Context, adminID uuid.UUID) (int, error)
}

// adminQuotaBreachedReader reads quota breach timestamp from storage.
type adminQuotaBreachedReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Admin, error)
}

// ResellerAutoSuspender suspends abusive or exhausted resellers after grace.
type ResellerAutoSuspender struct {
	admins   *AdminService
	users    ResellerUserDisabler
	ipEvents ResellerIPViolationCounter
	interval time.Duration
	now      func() time.Time
	log      *slog.Logger
	pub      events.Publisher
}

// NewResellerAutoSuspender wires the background checker.
func NewResellerAutoSuspender(admins *AdminService, users ResellerUserDisabler, ipEvents ResellerIPViolationCounter, log *slog.Logger) *ResellerAutoSuspender {
	if log == nil {
		log = slog.Default()
	}
	return &ResellerAutoSuspender{
		admins:   admins,
		users:    users,
		ipEvents: ipEvents,
		interval: 15 * time.Minute,
		now:      time.Now,
		log:      log,
		pub:      events.Nop{},
	}
}

func (w *ResellerAutoSuspender) SetPublisher(p events.Publisher) {
	if p != nil {
		w.pub = p
	}
}

// Run ticks until ctx is cancelled.
func (w *ResellerAutoSuspender) Run(ctx context.Context) {
	if w.admins == nil {
		return
	}
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.tick(ctx)
		}
	}
}

func (w *ResellerAutoSuspender) tick(ctx context.Context) {
	st, ok := w.admins.admins.(AdminSuspendStore)
	if !ok {
		return
	}
	candidates, err := st.ListResellerCandidates(ctx)
	if err != nil {
		w.log.Warn("list reseller candidates failed", "err", err)
		return
	}
	now := w.now()
	for _, admin := range candidates {
		if !admin.AutoSuspendEnabled {
			continue
		}
		if reason := w.checkIPViolations(ctx, admin, now); reason != "" {
			w.suspend(ctx, admin.ID, reason)
			continue
		}
		w.checkQuota(ctx, st, admin, now)
	}
}

func (w *ResellerAutoSuspender) checkIPViolations(ctx context.Context, admin *domain.Admin, now time.Time) string {
	if w.ipEvents == nil || admin.IPViolationSuspendThreshold <= 0 {
		return ""
	}
	since := now.Add(-7 * 24 * time.Hour)
	n, err := w.ipEvents.CountIPLimitEventsForAdminSince(ctx, admin.ID, since)
	if err != nil {
		return ""
	}
	if int(n) >= admin.IPViolationSuspendThreshold {
		return "ip limit violations exceeded"
	}
	return ""
}

func (w *ResellerAutoSuspender) checkQuota(ctx context.Context, st AdminSuspendStore, admin *domain.Admin, now time.Time) {
	usage, err := w.admins.quotaUsageFor(ctx, admin)
	if err != nil {
		return
	}
	exhausted := false
	if admin.UserQuota > 0 && usage.UserCount >= int64(admin.UserQuota) {
		exhausted = true
	}
	if admin.TrafficQuota > 0 {
		used := usage.TrafficAllocated
		if admin.TrafficQuotaMode == domain.TrafficQuotaConsumed {
			used = usage.TrafficUsed
		}
		if used >= admin.TrafficQuota {
			exhausted = true
		}
	}
	if !exhausted {
		_ = st.SetQuotaBreachedAt(ctx, admin.ID, nil)
		return
	}
	grace := time.Duration(admin.SuspendGraceMinutes) * time.Minute
	if grace <= 0 {
		grace = time.Hour
	}
	var breachedAt *time.Time
	if ext, ok := w.admins.admins.(adminQuotaBreachedReader); ok {
		if fresh, err := ext.GetByID(ctx, admin.ID); err == nil {
			breachedAt = fresh.QuotaBreachedAt
		}
	}
	if breachedAt == nil {
		t := now
		_ = st.SetQuotaBreachedAt(ctx, admin.ID, &t)
		return
	}
	if now.Sub(*breachedAt) >= grace {
		w.suspend(ctx, admin.ID, "quota exhausted")
	}
}

func (w *ResellerAutoSuspender) suspend(ctx context.Context, adminID uuid.UUID, reason string) {
	admin, err := w.admins.SuspendAdmin(ctx, adminID, reason)
	if err != nil {
		w.log.Warn("auto-suspend failed", "admin", adminID, "err", err)
		return
	}
	if w.users != nil {
		if n, err := w.users.DisableAllForAdmin(ctx, adminID); err != nil {
			w.log.Warn("disable users on suspend failed", "admin", adminID, "err", err)
		} else {
			w.log.Info("reseller auto-suspended", "admin", admin.Username, "users_disabled", n, "reason", reason)
		}
	}
	w.pub.Publish(events.Event{
		Type:    events.AdminQuotaWarning,
		Message: "reseller suspended: " + reason,
		Data:    map[string]any{"admin_id": adminID.String(), "admin": admin.Username, "reason": reason},
	})
}
