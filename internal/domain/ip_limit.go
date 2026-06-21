package domain

import (
	"time"

	"github.com/google/uuid"
)

// IPLimitAction is the enforcement action taken when a user exceeds its device
// (online-IP) limit. It builds on the existing ShareGuard detection.
type IPLimitAction string

const (
	// IPLimitActionWarn emits an event only — the current detection-only
	// behavior, taking no action against the user.
	IPLimitActionWarn IPLimitAction = "warn"
	// IPLimitActionDisable sets the user to "limited" and deprovisions it from
	// its nodes, auto-restoring after RestoreAfter seconds.
	IPLimitActionDisable IPLimitAction = "disable_temporarily"
	// IPLimitActionKill drops the user's live connections on xray nodes by
	// removing and immediately re-adding it on the core. On sing-box nodes (no
	// runtime online API) it degrades to IPLimitActionDisable.
	IPLimitActionKill IPLimitAction = "kill_connections"
)

// IPLimitPolicy is the singleton enforcement configuration. When Enabled is
// false, ShareGuard keeps its prior detection-only behavior unchanged.
type IPLimitPolicy struct {
	Enabled       bool          `json:"enabled"`
	Action        IPLimitAction `json:"action"`
	AlertCooldown int           `json:"alert_cooldown"` // seconds; per-user alert/action dedup
	RestoreAfter  int           `json:"restore_after"`  // seconds; auto-undo disable_temporarily
}

// DefaultIPLimitPolicy is the conservative default: disabled (detection only),
// warn action, 15-minute cooldown and restore window.
func DefaultIPLimitPolicy() IPLimitPolicy {
	return IPLimitPolicy{
		Enabled:       false,
		Action:        IPLimitActionWarn,
		AlertCooldown: 900,
		RestoreAfter:  900,
	}
}

// IPLimitEvent is the audit record of an enforcement action taken (or, for the
// warn action, of a detected violation).
type IPLimitEvent struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	OnlineIPs int       `json:"online_ips"`
	Limit     int       `json:"limit"`
	Action    string    `json:"action"`
	CreatedAt time.Time `json:"created_at"`
}
