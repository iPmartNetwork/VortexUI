package service

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
)

// AdminQuotaNotifyRepo persists reseller quota alert settings and history.
type AdminQuotaNotifyRepo interface {
	GetConfig(ctx context.Context) (*domain.AdminQuotaNotifyConfig, error)
	SaveConfig(ctx context.Context, cfg *domain.AdminQuotaNotifyConfig) error
	SaveEvent(ctx context.Context, e *domain.AdminQuotaNotifyEvent) error
	ListEvents(ctx context.Context, limit int) ([]*domain.AdminQuotaNotifyEvent, error)
	LastEvent(ctx context.Context, adminID uuid.UUID, metric string, threshold int) (*domain.AdminQuotaNotifyEvent, error)
}

// AdminQuotaNotifyService manages reseller quota alert configuration.
type AdminQuotaNotifyService struct {
	repo AdminQuotaNotifyRepo
}

func NewAdminQuotaNotifyService(repo AdminQuotaNotifyRepo) *AdminQuotaNotifyService {
	return &AdminQuotaNotifyService{repo: repo}
}

func (s *AdminQuotaNotifyService) GetConfig(ctx context.Context) (*domain.AdminQuotaNotifyConfig, error) {
	return s.repo.GetConfig(ctx)
}

func (s *AdminQuotaNotifyService) UpdateConfig(ctx context.Context, cfg domain.AdminQuotaNotifyConfig) error {
	if cfg.CooldownMinutes <= 0 {
		cfg.CooldownMinutes = 1440
	}
	if len(cfg.NotifyAtPercent) == 0 {
		cfg.NotifyAtPercent = []int{80, 90, 100}
	}
	return s.repo.SaveConfig(ctx, &cfg)
}

func (s *AdminQuotaNotifyService) ListEvents(ctx context.Context) ([]*domain.AdminQuotaNotifyEvent, error) {
	return s.repo.ListEvents(ctx, 100)
}

// AdminQuotaWarner checks reseller quotas and emits alert events.
type AdminQuotaWarner struct {
	admins   AdminStore
	users    AdminUserStatsReader
	notify   AdminQuotaNotifyRepo
	interval time.Duration
	now      func() time.Time
	log      *slog.Logger
	pub      events.Publisher
}

func NewAdminQuotaWarner(admins AdminStore, users AdminUserStatsReader, notify AdminQuotaNotifyRepo, log *slog.Logger) *AdminQuotaWarner {
	if log == nil {
		log = slog.Default()
	}
	return &AdminQuotaWarner{
		admins:   admins,
		users:    users,
		notify:   notify,
		interval: time.Hour,
		now:      time.Now,
		log:      log,
		pub:      events.Nop{},
	}
}

func (w *AdminQuotaWarner) SetPublisher(p events.Publisher) {
	if p != nil {
		w.pub = p
	}
}

func (w *AdminQuotaWarner) Run(ctx context.Context) {
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

func (w *AdminQuotaWarner) tick(ctx context.Context) {
	cfg, err := w.notify.GetConfig(ctx)
	if err != nil || !cfg.Enabled {
		return
	}
	admins, err := w.admins.List(ctx)
	if err != nil {
		w.log.Warn("admin quota warner: list admins failed", "err", err)
		return
	}
	cooldown := time.Duration(cfg.CooldownMinutes) * time.Minute
	now := w.now()
	for _, a := range admins {
		if a.Sudo || w.users == nil {
			continue
		}
		stats, err := w.users.StatsForAdmin(ctx, a.ID)
		if err != nil {
			continue
		}
		w.checkMetric(ctx, cfg, a, "users", stats.UserCount, int64(a.UserQuota), cooldown, now)
		if a.TrafficQuota > 0 {
			var used int64
			if a.TrafficQuotaMode == domain.TrafficQuotaConsumed {
				used = stats.TrafficUsed
			} else {
				used = stats.TrafficAllocated
			}
			w.checkMetric(ctx, cfg, a, "traffic", used, a.TrafficQuota, cooldown, now)
		}
	}
}

func (w *AdminQuotaWarner) checkMetric(ctx context.Context, cfg *domain.AdminQuotaNotifyConfig, a *domain.Admin, metric string, used, limit int64, cooldown time.Duration, now time.Time) {
	if limit <= 0 {
		return
	}
	pct := int(used * 100 / limit)
	for _, th := range cfg.NotifyAtPercent {
		if pct < th {
			continue
		}
		last, err := w.notify.LastEvent(ctx, a.ID, metric, th)
		if err == nil && now.Sub(last.CreatedAt) < cooldown {
			continue
		}
		msg := "Reseller " + a.Username + " reached " + metric + " quota " + strconv.Itoa(th) + "% (" + strconv.Itoa(pct) + "% used)"
		w.pub.Publish(events.Event{
			Type:    events.AdminQuotaWarning,
			Message: msg,
			Data: map[string]any{
				"admin_id":   a.ID.String(),
				"admin":      a.Username,
				"metric":     metric,
				"threshold":  th,
				"usage_pct":  pct,
				"webhook":    cfg.WebhookURL,
				"telegram":   cfg.NotifyTelegram,
			},
		})
		if err := w.notify.SaveEvent(ctx, &domain.AdminQuotaNotifyEvent{
			ID:        uuid.New(),
			AdminID:   a.ID,
			Threshold: th,
			Metric:    metric,
			UsagePct:  pct,
			CreatedAt: now,
		}); err != nil {
			w.log.Error("failed to save quota alert event", "admin", a.Username, "metric", metric, "err", err)
		}
		w.log.Info("reseller quota alert", "admin", a.Username, "metric", metric, "pct", pct, "threshold", th)
	}
}
