package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// SessionManager defines session management interface
type SessionManager interface {
	// CreateSession creates a new session and returns token
	CreateSession(ctx context.Context, adminID uuid.UUID, ipAddress, userAgent string) (token string, err error)

	// ValidateSession validates a session token
	ValidateSession(ctx context.Context, token string) (*domain.AdminSession, error)

	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error

	// RevokeAdminSessions revokes all sessions for an admin
	RevokeAdminSessions(ctx context.Context, adminID uuid.UUID) error

	// RefreshSession updates session last activity
	RefreshSession(ctx context.Context, sessionID uuid.UUID) error

	// CleanupExpiredSessions cleans up expired sessions
	CleanupExpiredSessions(ctx context.Context) error
}

// AuditLogger defines audit logging interface
type AuditLogger interface {
	// LogAction logs an admin action
	LogAction(ctx context.Context, adminID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, oldValues, newValues interface{}) error

	// LogLogin logs admin login
	LogLogin(ctx context.Context, adminID uuid.UUID, ipAddress, userAgent string) error

	// LogLogout logs admin logout
	LogLogout(ctx context.Context, adminID uuid.UUID) error
}

// SecurityValidator defines security validation interface
type SecurityValidator interface {
	// ValidateIPWhitelist checks if IP is whitelisted for admin
	ValidateIPWhitelist(ctx context.Context, adminID uuid.UUID, ipAddress string) (bool, error)

	// IsIPWhitelistEnabled checks if IP whitelist is enabled
	IsIPWhitelistEnabled(ctx context.Context, adminID uuid.UUID) (bool, error)
}
