package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminIPWhitelist restricts panel access for a specific admin to allowed CIDRs.
// If admin_id is nil, the rule applies globally.
type AdminIPWhitelist struct {
	ID          uuid.UUID  `json:"id"`
	AdminID     *uuid.UUID `json:"admin_id,omitempty"`
	CIDR        string     `json:"cidr"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
}

// AdminSession tracks an active login session for an admin.
type AdminSession struct {
	ID         uuid.UUID `json:"id"`
	AdminID    uuid.UUID `json:"admin_id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	Country    string    `json:"country"`
	LastActive time.Time `json:"last_active"`
	CreatedAt  time.Time `json:"created_at"`
	Revoked    bool      `json:"revoked"`
}

// LoginAuditEntry records a single authentication attempt (success or failure).
type LoginAuditEntry struct {
	ID            uuid.UUID  `json:"id"`
	AdminID       *uuid.UUID `json:"admin_id,omitempty"`
	Username      string     `json:"username"`
	IPAddress     string     `json:"ip_address"`
	UserAgent     string     `json:"user_agent"`
	Country       string     `json:"country"`
	Success       bool       `json:"success"`
	FailureReason string     `json:"failure_reason,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

// SecurityAuditEntry records a sensitive operation with before/after state.
type SecurityAuditEntry struct {
	ID          uuid.UUID      `json:"id"`
	AdminID     *uuid.UUID     `json:"admin_id,omitempty"`
	Operation   string         `json:"operation"`
	Resource    string         `json:"resource"`
	BeforeState map[string]any `json:"before_state,omitempty"`
	AfterState  map[string]any `json:"after_state,omitempty"`
	IPAddress   string         `json:"ip_address"`
	CreatedAt   time.Time      `json:"created_at"`
}

// IPBan represents a temporary or permanent IP ban entry.
type IPBan struct {
	ID        uuid.UUID  `json:"id"`
	IPAddress string     `json:"ip_address"`
	Reason    string     `json:"reason"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// ScopedAPIToken extends the existing API token with granular permission scopes.
type ScopedAPIToken struct {
	ID     uuid.UUID `json:"id"`
	Scopes []string  `json:"scopes"`
}

// Common security scopes for API tokens.
const (
	ScopeUsersRead       = "users:read"
	ScopeUsersWrite      = "users:write"
	ScopeNodesRead       = "nodes:read"
	ScopeNodesWrite      = "nodes:write"
	ScopeInboundsRead    = "inbounds:read"
	ScopeInboundsWrite   = "inbounds:write"
	ScopeAdminRead       = "admin:read"
	ScopeAdminWrite      = "admin:write"
	ScopeSettingsRead    = "settings:read"
	ScopeSettingsWrite   = "settings:write"
	ScopeSubscription    = "subscription:read"
	ScopeSecurityManage  = "security:manage"
	ScopeAll             = "*"
)

// AccountLockout tracks consecutive failed login attempts for lockout logic.
type AccountLockout struct {
	AdminID          uuid.UUID `json:"admin_id"`
	ConsecutiveFails int       `json:"consecutive_fails"`
	LockedUntil      *time.Time `json:"locked_until,omitempty"`
}
