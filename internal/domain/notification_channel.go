package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationChannel represents a configured notification destination
// (Telegram chat or webhook URL) with scope-based routing and event filtering.
type NotificationChannel struct {
	ID        uuid.UUID      `json:"id"`
	Name      string         `json:"name"`
	Type      string         `json:"type"` // "telegram" | "webhook"
	Config    map[string]any `json:"config"`
	ScopeType string         `json:"scope_type"` // "global"|"node"|"admin"|"group"
	ScopeID   string         `json:"scope_id,omitempty"`
	Events    []string       `json:"events"`
	Template  string         `json:"template,omitempty"`
	Enabled   bool           `json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// WebhookDelivery tracks an individual webhook delivery attempt with retry state.
type WebhookDelivery struct {
	ID         uuid.UUID  `json:"id"`
	ChannelID  uuid.UUID  `json:"channel_id"`
	EventType  string     `json:"event_type"`
	Payload    any        `json:"payload"`
	StatusCode int        `json:"status_code"`
	Attempts   int        `json:"attempts"`
	NextRetry  *time.Time `json:"next_retry,omitempty"`
	Delivered  bool       `json:"delivered"`
	CreatedAt  time.Time  `json:"created_at"`
}
