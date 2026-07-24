package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserStatus is the lifecycle state of a service user (the people who consume proxies).
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // allowed to connect
	UserStatusLimited  UserStatus = "limited"  // hit data limit
	UserStatusExpired  UserStatus = "expired"  // passed expire_at
	UserStatusDisabled UserStatus = "disabled" // manually turned off
	UserStatusOnHold   UserStatus = "on_hold"  // created but timer starts on first connection
)

// ResetStrategy controls how used traffic is periodically reset back to zero.
type ResetStrategy string

const (
	ResetNone    ResetStrategy = "no_reset"
	ResetDaily   ResetStrategy = "daily"
	ResetWeekly  ResetStrategy = "weekly"
	ResetMonthly ResetStrategy = "monthly"
)

// User is the central, core-agnostic identity. A single User can be exposed
// through many inbounds/protocols at once (User-centric model). This is the
// key architectural difference from inbound-centric panels.
type User struct {
	ID       uuid.UUID  `json:"id"`
	Username string     `json:"username"`
	Status   UserStatus `json:"status"`
	Note     string     `json:"note,omitempty"`

	// AdminID is the operator who created this user (reseller ownership). Nil
	// means created by a sudo admin with no scoping.
	AdminID *uuid.UUID `json:"admin_id,omitempty"`

	// Traffic accounting (bytes). UsedTraffic is the aggregate across all nodes,
	// updated from delta reports so it is restart-safe and never double counted.
	DataLimit   int64 `json:"data_limit"`   // 0 = unlimited
	UsedTraffic int64 `json:"used_traffic"` // aggregated delta sum

	// Lifecycle
	ExpireAt      *time.Time    `json:"expire_at,omitempty"` // nil = never expires
	OnHoldExpire  *int64        `json:"on_hold_expire_duration,omitempty"`
	ResetStrategy ResetStrategy `json:"reset_strategy"`
	LastReset     *time.Time    `json:"last_reset,omitempty"`

	// Access control
	DeviceLimit  int      `json:"device_limit"` // 0 = unlimited; enforced via HWID/IP
	DeviceLock   bool     `json:"device_lock"`  // when true, only pre-registered HWIDs may connect
	AllowedHWIDs []string `json:"allowed_hwids,omitempty"`

	// Notification: user's Telegram chat ID for personal alerts (expiry, limit).
	TelegramChatID string `json:"telegram_chat_id,omitempty"`

	// Shared credentials reused across every protocol this user is bound to.
	Proxies UserCredentials `json:"proxies"`

	// SubToken is the opaque token used in the subscription URL.
	SubToken string `json:"sub_token"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCredentials holds the per-protocol secrets. They are generated once and
// reused so the same human keeps a stable identity across protocols/nodes.
type UserCredentials struct {
	VMessUUID    uuid.UUID `json:"vmess_uuid"`
	VLESSUUID    uuid.UUID `json:"vless_uuid"`
	TrojanPass   string    `json:"trojan_password"`
	ShadowsocksP string    `json:"ss_password"`
	SSMethod     string    `json:"ss_method"`
}

// IsActive reports whether the user may currently connect, deriving status from
// the live limit/expiry state rather than trusting a possibly-stale flag.
func (u *User) IsActive(now time.Time) bool {
	if u.Status == UserStatusDisabled {
		return false
	}
	if u.DataLimit > 0 && u.UsedTraffic >= u.DataLimit {
		return false
	}
	if u.ExpireAt != nil && now.After(*u.ExpireAt) {
		return false
	}
	return true
}

// DerivedStatus computes the status that should be persisted given current usage.
func (u *User) DerivedStatus(now time.Time) UserStatus {
	switch {
	case u.Status == UserStatusDisabled:
		return UserStatusDisabled
	case u.Status == UserStatusOnHold && u.UsedTraffic == 0:
		return UserStatusOnHold
	case u.DataLimit > 0 && u.UsedTraffic >= u.DataLimit:
		return UserStatusLimited
	case u.ExpireAt != nil && now.After(*u.ExpireAt):
		return UserStatusExpired
	default:
		return UserStatusActive
	}
}

// UserListItem is a user row enriched for the management table.
type UserListItem struct {
	User
	ProtocolLabel string `json:"protocol_label,omitempty"`
	DeviceCount   int    `json:"device_count"`
}

type UserStats struct {
	Total      int                `json:"total"`
	ByStatus   map[UserStatus]int `json:"by_status"`
	TotalUsed  int64              `json:"total_used"` // sum of used_traffic across all users
}
