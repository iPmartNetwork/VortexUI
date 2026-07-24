package domain

import (
	"time"

	"github.com/google/uuid"
)

// ClientTemplate defines a per-client-app subscription configuration template.
// When a subscription request's User-Agent matches ClientPattern (a regex), the
// template's routing rules, DNS settings, and custom outbounds are merged into
// the subscription output. Templates are evaluated in Priority order (higher first).
type ClientTemplate struct {
	ID             uuid.UUID      `json:"id"`
	Name           string         `json:"name"`
	ClientPattern  string         `json:"client_pattern"`
	RoutingRules   []any          `json:"routing_rules"`
	DNSSettings    map[string]any `json:"dns_settings"`
	CustomOutbounds []any         `json:"custom_outbounds"`
	Priority       int            `json:"priority"`
	Enabled        bool           `json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// SubscriptionApproval represents a pending subscription request that requires
// admin approval before the user is granted access.
type SubscriptionApproval struct {
	ID          uuid.UUID      `json:"id"`
	UserID      uuid.UUID      `json:"user_id"`
	RequestData map[string]any `json:"request_data"`
	Status      string         `json:"status"`
	AdminID     *uuid.UUID     `json:"admin_id,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	ResolvedAt  *time.Time     `json:"resolved_at,omitempty"`
}
