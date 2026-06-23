package domain

import (
	"time"

	"github.com/google/uuid"
)

// Permission is a fine-grained capability checked by the RBAC middleware.
type Permission string

const (
	PermUserRead    Permission = "user:read"
	PermUserWrite   Permission = "user:write"
	PermNodeRead    Permission = "node:read"
	PermNodeWrite   Permission = "node:write"
	PermInboundRead  Permission = "inbound:read"
	PermInboundWrite Permission = "inbound:write"
	PermAdminManage Permission = "admin:manage" // create/modify other admins
	PermSystemRead  Permission = "system:read"
)

// Role groups permissions. Sudo bypasses all checks.
type Role struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Permissions []Permission `json:"permissions"`
}

// Admin is a panel operator (distinct from a service User).
type Admin struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // bcrypt; never serialized
	Sudo         bool      `json:"sudo"`
	RoleID       *uuid.UUID `json:"role_id,omitempty"`

	// TOTP 2FA
	TOTPSecret  string `json:"-"`
	TOTPEnabled bool   `json:"totp_enabled"`

	// Quota: a non-sudo admin can be capped on how many users/traffic they sell.
	UserQuota         int              `json:"user_quota"`          // 0 = unlimited
	TrafficQuota      int64            `json:"traffic_quota"`       // 0 = unlimited
	TrafficQuotaMode  TrafficQuotaMode `json:"traffic_quota_mode"`  // allocated | consumed

	ParentAdminID *uuid.UUID `json:"parent_admin_id,omitempty"`

	// Prepaid wallet (bytes + user slots). 0 = not using wallet for that dimension.
	WalletTrafficBytes int64 `json:"wallet_traffic_bytes"`
	WalletUserCredits  int   `json:"wallet_user_credits"`

	// Outbound webhook for reseller automation.
	WebhookURL     string `json:"webhook_url,omitempty"`
	WebhookSecret  string `json:"-"`
	WebhookEnabled bool   `json:"webhook_enabled"`

	// Per-reseller abuse policies (0 = unlimited / allowed).
	PolicyMaxDataLimit    int64 `json:"policy_max_data_limit"`
	PolicyMaxExpireDays   int   `json:"policy_max_expire_days"`
	PolicyAllowBulkDelete bool  `json:"policy_allow_bulk_delete"`
	PolicyAllowBulkCreate bool  `json:"policy_allow_bulk_create"`

	// Suspension (auto or manual).
	Suspended                   bool       `json:"suspended"`
	SuspendedAt                 *time.Time `json:"suspended_at,omitempty"`
	SuspendReason               string     `json:"suspend_reason,omitempty"`
	AutoSuspendEnabled          bool       `json:"auto_suspend_enabled"`
	IPViolationSuspendThreshold int        `json:"ip_violation_suspend_threshold"`
	SuspendGraceMinutes         int        `json:"suspend_grace_minutes"`
	QuotaBreachedAt             *time.Time `json:"-"`

	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// Has reports whether the admin holds a permission, honoring the sudo bypass.
func (a *Admin) Has(p Permission, role *Role) bool {
	if a.Sudo {
		return true
	}
	if role == nil {
		return false
	}
	for _, have := range role.Permissions {
		if have == p {
			return true
		}
	}
	return false
}

// TrafficQuotaMode controls how traffic_quota is enforced for resellers.
type TrafficQuotaMode string

const (
	TrafficQuotaAllocated TrafficQuotaMode = "allocated" // sum(data_limit) pool
	TrafficQuotaConsumed  TrafficQuotaMode = "consumed"  // sum(used_traffic) cap
)
