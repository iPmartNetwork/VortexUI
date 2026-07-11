package port

import (
	"context"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// AuditRepository defines the audit log data access interface
type AuditRepository interface {
	// SaveAudit persists an audit log entry
	SaveAudit(ctx context.Context, log *domain.AuditLog) error

	// GetAudit retrieves an audit log by ID
	GetAudit(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)

	// ListAudits retrieves audit logs with filters
	ListAudits(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int64, error)

	// ListAuditsByAdmin retrieves all audits for an admin
	ListAuditsByAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int64, error)

	// ListAuditsByResource retrieves all audits for a specific resource
	ListAuditsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*domain.AuditLog, int64, error)

	// DeleteOldAudits deletes audit logs older than retention days
	DeleteOldAudits(ctx context.Context, retentionDays int) (int64, error)

	// ExportAudits exports audit logs in specified format
	ExportAudits(ctx context.Context, filter *domain.AuditLogFilter, format domain.AuditExportFormat) (*domain.AuditLogExport, error)
}

// SessionRepository defines the session data access interface
type SessionRepository interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, session *domain.AdminSession) error

	// GetSession retrieves a session by token hash
	GetSession(ctx context.Context, tokenHash string) (*domain.AdminSession, error)

	// ListSessions lists sessions with filters
	ListSessions(ctx context.Context, filter *domain.SessionFilter) ([]*domain.AdminSession, int64, error)

	// ListAdminSessions lists all active sessions for an admin
	ListAdminSessions(ctx context.Context, adminID uuid.UUID, active bool) ([]*domain.AdminSession, error)

	// UpdateSession updates session last activity
	UpdateSession(ctx context.Context, session *domain.AdminSession) error

	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error

	// RevokeAdminSessions revokes all sessions for an admin
	RevokeAdminSessions(ctx context.Context, adminID uuid.UUID) error

	// DeleteExpiredSessions deletes expired sessions
	DeleteExpiredSessions(ctx context.Context) (int64, error)
}

// ErrorCodeRepository defines the error code data access interface
type ErrorCodeRepository interface {
	// GetErrorCode retrieves an error code
	GetErrorCode(ctx context.Context, code string) (*domain.ErrorCode, error)

	// ListErrorCodes lists all error codes
	ListErrorCodes(ctx context.Context) ([]*domain.ErrorCode, error)

	// ListErrorCodesByCategory lists error codes by category
	ListErrorCodesByCategory(ctx context.Context, category string) ([]*domain.ErrorCode, error)
}
