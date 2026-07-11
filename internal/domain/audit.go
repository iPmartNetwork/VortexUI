package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AuditEntry records one mutating admin action for accountability. It is written
// by the audit middleware for every authenticated POST/PUT/DELETE request.
type AuditEntry struct {
	ID             uuid.UUID  `json:"id"`
	Time           time.Time  `json:"time"`
	AdminID        *uuid.UUID `json:"admin_id,omitempty"`
	ImpersonatorID *uuid.UUID `json:"impersonator_id,omitempty"`
	Username       string     `json:"username"` // denormalized for display (admin may be deleted)
	Method         string     `json:"method"`
	Path           string     `json:"path"`
	Status         int        `json:"status"`
	IP             string     `json:"ip"`
}

// AdminSession represents an active admin session
type AdminSession struct {
	ID           uuid.UUID     `json:"id"`
	AdminID      uuid.UUID     `json:"admin_id"`
	TokenHash    string        `json:"-"` // Never expose hash
	IPAddress    string        `json:"ip_address"`
	UserAgent    string        `json:"user_agent"`
	LastActivity time.Time     `json:"last_activity"`
	ExpiresAt    time.Time     `json:"expires_at"`
	CreatedAt    time.Time     `json:"created_at"`
	RevokedAt    *time.Time    `json:"revoked_at,omitempty"`
}

// IsActive checks if session is still valid
func (s *AdminSession) IsActive() bool {
	if s.RevokedAt != nil {
		return false
	}
	return time.Now().Before(s.ExpiresAt)
}

// IsExpired checks if session is expired
func (s *AdminSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// SessionFilter represents filter options for sessions
type SessionFilter struct {
	AdminID    *uuid.UUID
	IPAddress  *string
	Active     *bool
	Limit      int
	Offset     int
}

// AuditLogFilter represents filter options for audit logs
type AuditLogFilter struct {
	AdminID    *uuid.UUID
	Action     *string
	ResourceID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

// AuditLog represents a detailed audit log entry
type AuditLog struct {
	ID           uuid.UUID       `json:"id"`
	AdminID      uuid.UUID       `json:"admin_id"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   *uuid.UUID      `json:"resource_id,omitempty"`
	OldValues    json.RawMessage `json:"old_values,omitempty"`
	NewValues    json.RawMessage `json:"new_values,omitempty"`
	IPAddress    string          `json:"ip_address"`
	UserAgent    string          `json:"user_agent"`
	CreatedAt    time.Time       `json:"created_at"`
	Signature    *string         `json:"signature,omitempty"`
}

// AuditExportFormat represents export format for audit logs
type AuditExportFormat string

const (
	ExportFormatJSON AuditExportFormat = "json"
	ExportFormatCSV  AuditExportFormat = "csv"
)

// AuditLogExport represents exported audit log data
type AuditLogExport struct {
	Logs       []*AuditLog   `json:"logs"`
	Total      int64         `json:"total"`
	ExportedAt time.Time     `json:"exported_at"`
	Format     string        `json:"format"`
}

// ========== PHASE 3C: Audit & Compliance Types ==========

// AuditEventType represents different types of audit events
type AuditEventType string

const (
	// Authentication events
	AuditEventAdminLogin       AuditEventType = "admin_login"
	AuditEventAdminLogout      AuditEventType = "admin_logout"
	AuditEventAuthFailure      AuditEventType = "auth_failure"
	AuditEventPasswordChanged  AuditEventType = "password_changed"
	AuditEventMFAEnabled       AuditEventType = "mfa_enabled"
	AuditEventMFADisabled      AuditEventType = "mfa_disabled"

	// Admin management
	AuditEventAdminCreated     AuditEventType = "admin_created"
	AuditEventAdminUpdated     AuditEventType = "admin_updated"
	AuditEventAdminDeleted     AuditEventType = "admin_deleted"
	AuditEventAdminPermChanged AuditEventType = "admin_permission_changed"

	// User management
	AuditEventUserCreated  AuditEventType = "user_created"
	AuditEventUserUpdated  AuditEventType = "user_updated"
	AuditEventUserDeleted  AuditEventType = "user_deleted"
	AuditEventUserBanned   AuditEventType = "user_banned"
	AuditEventUserUnbanned AuditEventType = "user_unbanned"

	// Node management
	AuditEventNodeCreated  AuditEventType = "node_created"
	AuditEventNodeUpdated  AuditEventType = "node_updated"
	AuditEventNodeDeleted  AuditEventType = "node_deleted"

	// Configuration changes
	AuditEventConfigChanged    AuditEventType = "config_changed"
	AuditEventPolicyChanged    AuditEventType = "policy_changed"
	AuditEventSecurityChanged  AuditEventType = "security_changed"

	// Security incidents
	AuditEventIPBlocked       AuditEventType = "ip_blocked"
	AuditEventBruteForceAttempt AuditEventType = "brute_force_attempt"
	AuditEventSuspiciousActivity AuditEventType = "suspicious_activity"

	// Data access
	AuditEventDataExported    AuditEventType = "data_exported"
	AuditEventDataImported    AuditEventType = "data_imported"
	AuditEventReportGenerated AuditEventType = "report_generated"
)

// AuditEventSeverity represents event severity level
type AuditEventSeverity string

const (
	AuditSeverityInfo     AuditEventSeverity = "info"
	AuditSeverityWarning  AuditEventSeverity = "warning"
	AuditSeverityError    AuditEventSeverity = "error"
	AuditSeverityCritical AuditEventSeverity = "critical"
)

// AuditEvent represents a security/admin action event
type AuditEvent struct {
	ID              uuid.UUID
	AdminID         uuid.UUID              // Actor (admin who performed action)
	EventType       AuditEventType
	Severity        AuditEventSeverity
	TargetType      string                 // What was affected (user, node, config, etc)
	TargetID        *uuid.UUID             // ID of affected entity (nullable for system events)
	Description     string                 // Human-readable description
	IPAddress       string                 // IP where action originated
	UserAgent       string                 // Browser/client info
	OldValue        string                 // Previous state (for updates, JSON)
	NewValue        string                 // New state (for updates, JSON)
	Status          string                 // success, failure, pending
	ErrorMessage    string                 // Error details if failed
	Metadata        map[string]interface{} // Additional context (stored as JSONB)
	CreatedAt       time.Time
}

// ComplianceEvent represents compliance-specific tracking
type ComplianceEvent struct {
	ID           uuid.UUID              `json:"id"`
	EventType    string                 `json:"event_type"` // SOC2, GDPR, ISO27001, etc
	Category     string                 `json:"category"`   // access_control, encryption, incident_response, etc
	Status       string                 `json:"status"`     // compliant, non_compliant, pending_review
	Description  string                 `json:"description"`
	Evidence     string                 `json:"evidence"` // Path/reference to supporting evidence
	AuditEventID *uuid.UUID             `json:"audit_event_id,omitempty"`
	AdminID      uuid.UUID              `json:"admin_id"` // Who verified/reviewed this
	VerifiedAt   *time.Time             `json:"verified_at,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"` // Compliance expires at this date
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// AuditReport represents a generated audit report
type AuditReport struct {
	ID          uuid.UUID              `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	ReportType  string                 `json:"report_type"` // daily, weekly, monthly, custom, incident
	Status      string                 `json:"status"`      // draft, pending_review, approved, rejected, exported
	StartDate   time.Time              `json:"start_date"`
	EndDate     time.Time              `json:"end_date"`
	Scope       string                 `json:"scope"` // system-wide, admin-specific, event-type-specific
	Filters     map[string]interface{} `json:"filters,omitempty"`
	EventCount  int                    `json:"event_count"`
	FilePath    *string                `json:"file_path,omitempty"`   // S3/local path to PDF/CSV export
	CreatedBy   uuid.UUID              `json:"created_by"`
	ApprovedBy  *uuid.UUID             `json:"approved_by,omitempty"`
	ApprovedAt  *time.Time             `json:"approved_at,omitempty"`
	ExportedAt  *time.Time             `json:"exported_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AuditPolicy defines audit retention and compliance rules
type AuditPolicy struct {
	ID                     uuid.UUID `json:"id"`
	Name                   string    `json:"name"`
	Description            string    `json:"description"`
	RetentionDays          int       `json:"retention_days"` // How long to keep audit logs
	AlertOnEventTypes      []string  `json:"alert_on_event_types"`
	RequireApprovalFor     []string  `json:"require_approval_for"`
	AutoArchiveAfterDays   int       `json:"auto_archive_after_days"`
	ComplianceFrameworks   []string  `json:"compliance_frameworks"` // SOC2, GDPR, ISO27001, PCI-DSS, etc
	EncryptionRequired     bool      `json:"encryption_required"`
	TamperDetectionEnabled bool      `json:"tamper_detection_enabled"`
	ExportableFormats      []string  `json:"exportable_formats"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// AuditLogArchive represents archived audit logs (immutable)
type AuditLogArchive struct {
	ID             uuid.UUID  `json:"id"`
	ArchiveName    string     `json:"archive_name"` // e.g., "audit-logs-2026-q1"
	FilePath       string     `json:"file_path"`    // S3/backup location
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	EventCount     int        `json:"event_count"`
	ChecksumSHA256 string     `json:"checksum_sha256"`
	CreatedAt      time.Time  `json:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}
