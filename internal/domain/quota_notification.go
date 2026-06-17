package domain

import (
	"time"

	"github.com/google/uuid"
)

// QuotaNotificationConfig stores notification settings for smart quota tiers.
type QuotaNotificationConfig struct {
	Enabled         bool    `json:"enabled"`
	NotifyTelegram  bool    `json:"notify_telegram"`
	NotifyWebhook   bool    `json:"notify_webhook"`
	WebhookURL      string  `json:"webhook_url,omitempty"`
	NotifyAtPercent []int   `json:"notify_at_percent"` // e.g. [50, 80, 90, 100]
	CooldownMinutes int    `json:"cooldown_minutes"`   // min gap between repeat notifications
}

// DefaultQuotaNotificationConfig returns sensible defaults.
func DefaultQuotaNotificationConfig() QuotaNotificationConfig {
	return QuotaNotificationConfig{
		Enabled:         false,
		NotifyTelegram:  true,
		NotifyAtPercent: []int{50, 80, 90, 100},
		CooldownMinutes: 60,
	}
}

// QuotaNotificationEvent records a sent notification.
type QuotaNotificationEvent struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	Username   string    `json:"username"`
	Percent    int       `json:"percent"`
	Channel    string    `json:"channel"` // "telegram" | "webhook"
	Delivered  bool      `json:"delivered"`
	Error      string    `json:"error,omitempty"`
	SentAt     time.Time `json:"sent_at"`
}
