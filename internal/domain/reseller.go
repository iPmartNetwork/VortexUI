package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminQuotaNotifyConfig controls reseller quota threshold alerts.
type AdminQuotaNotifyConfig struct {
	Enabled          bool   `json:"enabled"`
	NotifyTelegram   bool   `json:"notify_telegram"`
	WebhookURL       string `json:"webhook_url"`
	NotifyAtPercent  []int  `json:"notify_at_percent"`
	CooldownMinutes  int    `json:"cooldown_minutes"`
}

// AdminQuotaNotifyEvent records a fired reseller quota alert.
type AdminQuotaNotifyEvent struct {
	ID        uuid.UUID `json:"id"`
	AdminID   uuid.UUID `json:"admin_id"`
	Threshold int       `json:"threshold"`
	Metric    string    `json:"metric"` // users | traffic
	UsagePct  int       `json:"usage_pct"`
	CreatedAt time.Time `json:"created_at"`
}

// ResellerDashboard is the reseller home summary.
type ResellerDashboard struct {
	Quota          AdminQuotaUsage          `json:"quota"`
	UsersByStatus  map[string]int64         `json:"users_by_status"`
	TopUsers       []ResellerTopUser        `json:"top_users"`
	ExpiringSoon   int64                    `json:"expiring_soon"`
	NewUsers7d     int64                    `json:"new_users_7d"`
	NewUsers30d    int64                    `json:"new_users_30d"`
}

// ResellerTopUser is a high-traffic user row for the reseller dashboard.
type ResellerTopUser struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	UsedTraffic int64     `json:"used_traffic"`
	DataLimit   int64     `json:"data_limit"`
	Status      string    `json:"status"`
}
