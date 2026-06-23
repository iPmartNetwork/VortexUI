package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/notify"
)

// ResellerWebhookDispatcher delivers panel events to per-reseller webhook URLs.
type ResellerWebhookDispatcher struct {
	admins *AdminService
	log    *slog.Logger
	cache  map[uuid.UUID]domain.Admin
	mu     sync.RWMutex
	ttl    time.Duration
	last   time.Time
}

// NewResellerWebhookDispatcher builds a dispatcher. admins may be nil (no-op).
func NewResellerWebhookDispatcher(admins *AdminService, log *slog.Logger) *ResellerWebhookDispatcher {
	if log == nil {
		log = slog.Default()
	}
	return &ResellerWebhookDispatcher{
		admins: admins,
		log:    log,
		cache:  make(map[uuid.UUID]domain.Admin),
		ttl:    2 * time.Minute,
	}
}

// Run consumes events until ctx is cancelled.
func (d *ResellerWebhookDispatcher) Run(ctx context.Context, ch <-chan events.Event) {
	if d.admins == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			adminID := adminIDFromEvent(e)
			if adminID == uuid.Nil {
				continue
			}
			admin, err := d.adminWithWebhook(ctx, adminID)
			if err != nil || admin.WebhookURL == "" || !admin.WebhookEnabled {
				continue
			}
			wh := notify.NewWebhook(admin.WebhookURL, admin.WebhookSecret, d.log)
			if err := wh.Deliver(ctx, e); err != nil {
				d.log.Warn("reseller webhook failed", "admin", adminID, "type", string(e.Type), "err", err)
			}
		}
	}
}

func adminIDFromEvent(e events.Event) uuid.UUID {
	if e.Data == nil {
		return uuid.Nil
	}
	raw, ok := e.Data["admin_id"]
	if !ok || raw == nil {
		return uuid.Nil
	}
	switch v := raw.(type) {
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil
		}
		return id
	case uuid.UUID:
		return v
	default:
		return uuid.Nil
	}
}

func (d *ResellerWebhookDispatcher) adminWithWebhook(ctx context.Context, id uuid.UUID) (domain.Admin, error) {
	d.mu.RLock()
	if time.Since(d.last) < d.ttl {
		if a, ok := d.cache[id]; ok {
			d.mu.RUnlock()
			return a, nil
		}
	}
	d.mu.RUnlock()

	targets, err := d.admins.ListWebhookTargets(ctx)
	if err != nil {
		return domain.Admin{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cache = make(map[uuid.UUID]domain.Admin, len(targets))
	for _, a := range targets {
		d.cache[a.ID] = a
	}
	d.last = time.Now()
	return d.cache[id], nil
}
