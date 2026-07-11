package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// AuditEventRepository handles audit event storage and retrieval
type AuditEventRepository interface {
	// LogEvent records a new audit event
	LogEvent(ctx context.Context, event *domain.AuditEvent) error

	// GetEvent retrieves a specific audit event by ID
	GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.AuditEvent, error)

	// ListEvents retrieves events with optional filtering
	ListEvents(ctx context.Context, adminID *uuid.UUID, eventType *string, severity *string, startDate *time.Time, endDate *time.Time, limit int, offset int) ([]*domain.AuditEvent, int64, error)

	// SearchEvents performs full-text search on audit events
	SearchEvents(ctx context.Context, query string, limit int, offset int) ([]*domain.AuditEvent, int64, error)

	// GetEventsByAdmin retrieves all events for a specific admin
	GetEventsByAdmin(ctx context.Context, adminID uuid.UUID, limit int, offset int) ([]*domain.AuditEvent, int64, error)

	// GetEventsByType retrieves all events of a specific type
	GetEventsByType(ctx context.Context, eventType domain.AuditEventType, limit int, offset int) ([]*domain.AuditEvent, int64, error)

	// GetEventsBySeverity retrieves all events with a specific severity
	GetEventsBySeverity(ctx context.Context, severity domain.AuditEventSeverity, limit int, offset int) ([]*domain.AuditEvent, int64, error)

	// DeleteOldEvents deletes events older than retentionDays
	DeleteOldEvents(ctx context.Context, retentionDays int) (int64, error)

	// CountEvents returns total count of audit events
	CountEvents(ctx context.Context) (int64, error)
}

// ComplianceEventRepository handles compliance event storage and retrieval
type ComplianceEventRepository interface {
	// SaveEvent saves a compliance event
	SaveEvent(ctx context.Context, event *domain.ComplianceEvent) error

	// GetEvent retrieves a specific compliance event
	GetEvent(ctx context.Context, eventID uuid.UUID) (*domain.ComplianceEvent, error)

	// ListEvents retrieves compliance events with filtering
	ListEvents(ctx context.Context, eventType *string, status *string, limit int, offset int) ([]*domain.ComplianceEvent, int64, error)

	// GetEventsByFramework retrieves events for specific compliance framework (SOC2, GDPR, etc)
	GetEventsByFramework(ctx context.Context, framework string, limit int, offset int) ([]*domain.ComplianceEvent, int64, error)

	// UpdateEventStatus updates compliance event status
	UpdateEventStatus(ctx context.Context, eventID uuid.UUID, status string, verifiedBy *uuid.UUID) error

	// VerifyEvent marks event as verified/reviewed
	VerifyEvent(ctx context.Context, eventID uuid.UUID, verifiedBy uuid.UUID) error

	// GetExpiringEvents retrieves events expiring soon
	GetExpiringEvents(ctx context.Context, days int) ([]*domain.ComplianceEvent, error)

	// DeleteExpiredEvents deletes expired compliance events
	DeleteExpiredEvents(ctx context.Context) (int64, error)
}

// AuditReportRepository handles audit report storage and retrieval
type AuditReportRepository interface {
	// SaveReport saves an audit report
	SaveReport(ctx context.Context, report *domain.AuditReport) error

	// GetReport retrieves a specific report
	GetReport(ctx context.Context, reportID uuid.UUID) (*domain.AuditReport, error)

	// ListReports retrieves reports with optional filtering
	ListReports(ctx context.Context, reportType *string, status *string, limit int, offset int) ([]*domain.AuditReport, int64, error)

	// UpdateReportStatus updates report status (draft → pending_review → approved → exported)
	UpdateReportStatus(ctx context.Context, reportID uuid.UUID, newStatus string) error

	// ApproveReport approves a pending report
	ApproveReport(ctx context.Context, reportID uuid.UUID, approvedBy uuid.UUID) error

	// ExportReport marks report as exported and stores file path
	ExportReport(ctx context.Context, reportID uuid.UUID, filePath string) error

	// GetReportsByCreator retrieves reports created by specific admin
	GetReportsByCreator(ctx context.Context, adminID uuid.UUID, limit int, offset int) ([]*domain.AuditReport, int64, error)

	// DeleteReport deletes a report (soft or hard delete)
	DeleteReport(ctx context.Context, reportID uuid.UUID) error
}

// AuditPolicyRepository handles audit policy storage
type AuditPolicyRepository interface {
	// GetPolicy retrieves the current audit policy
	GetPolicy(ctx context.Context) (*domain.AuditPolicy, error)

	// SavePolicy saves/updates audit policy
	SavePolicy(ctx context.Context, policy *domain.AuditPolicy) error

	// UpdateRetentionPolicy updates log retention settings
	UpdateRetentionPolicy(ctx context.Context, retentionDays int) error

	// UpdateComplianceFrameworks updates which frameworks apply
	UpdateComplianceFrameworks(ctx context.Context, frameworks []string) error
}

// AuditArchiveRepository handles archival of old audit logs
type AuditArchiveRepository interface {
	// ArchiveEvents creates an archive of events in date range
	ArchiveEvents(ctx context.Context, startDate time.Time, endDate time.Time) (*domain.AuditLogArchive, error)

	// GetArchive retrieves a specific archive
	GetArchive(ctx context.Context, archiveID uuid.UUID) (*domain.AuditLogArchive, error)

	// ListArchives retrieves all archives
	ListArchives(ctx context.Context, limit int, offset int) ([]*domain.AuditLogArchive, int64, error)

	// VerifyArchiveIntegrity verifies archive hasn't been tampered with
	VerifyArchiveIntegrity(ctx context.Context, archiveID uuid.UUID) (bool, error)

	// DeleteOldArchives deletes archives older than retentionDays
	DeleteOldArchives(ctx context.Context, retentionDays int) (int64, error)
}

// AuditService handles audit event logging and business logic
type AuditService interface {
	// LogAdminAction logs when admin performs action
	LogAdminAction(ctx context.Context, adminID uuid.UUID, eventType domain.AuditEventType, targetType string, targetID *uuid.UUID, description string, ipAddress string, userAgent string) error

	// LogAuthenticationEvent logs login/logout/auth failures
	LogAuthenticationEvent(ctx context.Context, adminID uuid.UUID, eventType domain.AuditEventType, ipAddress string, userAgent string, success bool, reason string) error

	// LogSecurityIncident logs security-related events (IP block, brute force, etc)
	LogSecurityIncident(ctx context.Context, eventType domain.AuditEventType, severity domain.AuditEventSeverity, adminID *uuid.UUID, description string, ipAddress string) error

	// LogDataAccess logs when data is exported/imported/accessed
	LogDataAccess(ctx context.Context, adminID uuid.UUID, eventType domain.AuditEventType, resourceType string, resourceID *uuid.UUID, description string) error

	// LogConfigurationChange logs changes to system configuration
	LogConfigurationChange(ctx context.Context, adminID uuid.UUID, resourceType string, resourceID *uuid.UUID, oldValue string, newValue string) error

	// GetSeverity determines severity level for event type
	GetSeverity(eventType domain.AuditEventType) domain.AuditEventSeverity

	// IsAlertableEvent checks if event should trigger alert
	IsAlertableEvent(eventType domain.AuditEventType) bool
}

// ReportGenerator handles audit report generation
type ReportGenerator interface {
	// GenerateDailyReport generates a daily audit report
	GenerateDailyReport(ctx context.Context, date time.Time, createdBy uuid.UUID) (*domain.AuditReport, error)

	// GenerateCustomReport generates a report for date range
	GenerateCustomReport(ctx context.Context, startDate time.Time, endDate time.Time, filters map[string]interface{}, createdBy uuid.UUID) (*domain.AuditReport, error)

	// GenerateIncidentReport generates report for specific incident
	GenerateIncidentReport(ctx context.Context, eventID uuid.UUID, createdBy uuid.UUID) (*domain.AuditReport, error)

	// GenerateComplianceReport generates compliance-focused report
	GenerateComplianceReport(ctx context.Context, framework string, period string, createdBy uuid.UUID) (*domain.AuditReport, error)

	// ExportReport exports report in specified format (json, csv, pdf)
	ExportReport(ctx context.Context, reportID uuid.UUID, format string) ([]byte, error)

	// GetReportPreview gets preview of report before generation
	GetReportPreview(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error)
}

// ComplianceChecker verifies compliance status
type ComplianceChecker interface {
	// CheckFramework checks compliance for specific framework
	CheckFramework(ctx context.Context, framework string) (map[string]interface{}, error)

	// GetComplianceStatus returns overall compliance status
	GetComplianceStatus(ctx context.Context) (map[string]string, error)

	// GetNonCompliantItems returns list of non-compliant items
	GetNonCompliantItems(ctx context.Context) ([]*domain.ComplianceEvent, error)

	// GenerateComplianceCertificate generates proof of compliance
	GenerateComplianceCertificate(ctx context.Context, framework string) ([]byte, error)
}
