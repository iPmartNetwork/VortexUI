package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// AuditService handles audit logging operations
type AuditService struct {
	repo port.AuditRepository
	log  *slog.Logger
}

// NewAuditService creates a new audit service
func NewAuditService(repo port.AuditRepository, log *slog.Logger) *AuditService {
	if log == nil {
		log = slog.Default()
	}
	return &AuditService{
		repo: repo,
		log:  log,
	}
}

// LogAction logs an admin action
func (s *AuditService) LogAction(ctx context.Context, adminID uuid.UUID, action, resourceType string, resourceID *uuid.UUID, oldValues, newValues interface{}) error {
	var oldJSON, newJSON json.RawMessage
	var err error

	if oldValues != nil {
		if oldJSON, err = json.Marshal(oldValues); err != nil {
			s.log.Error("failed to marshal old values", "err", err)
			oldJSON = nil
		}
	}

	if newValues != nil {
		if newJSON, err = json.Marshal(newValues); err != nil {
			s.log.Error("failed to marshal new values", "err", err)
			newJSON = nil
		}
	}

	// Extract IP and User-Agent from context if available
	ipAddress := ""
	userAgent := ""
	if ip, ok := ctx.Value("ip_address").(string); ok {
		ipAddress = ip
	}
	if ua, ok := ctx.Value("user_agent").(string); ok {
		userAgent = ua
	}

	log := &domain.AuditLog{
		AdminID:      adminID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValues:    oldJSON,
		NewValues:    newJSON,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}

	if err := s.repo.SaveAudit(ctx, log); err != nil {
		s.log.Error("failed to save audit log", "err", err)
		return err
	}

	return nil
}

// GetAudit retrieves an audit log by ID
func (s *AuditService) GetAudit(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	return s.repo.GetAudit(ctx, id)
}

// ListAudits retrieves audit logs with filters
func (s *AuditService) ListAudits(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 100
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000
	}
	return s.repo.ListAudits(ctx, filter)
}

// ListAuditsByAdmin retrieves all audits for an admin
func (s *AuditService) ListAuditsByAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]*domain.AuditLog, int64, error) {
	return s.repo.ListAuditsByAdmin(ctx, adminID, limit, offset)
}

// ========== PHASE 3C: Advanced Audit Methods ==========

// LogSecurityEvent logs security-specific events
func (s *AuditService) LogSecurityEvent(ctx context.Context, eventType domain.AuditEventType, severity domain.AuditEventSeverity, adminID *uuid.UUID, description string, ipAddress string, metadata map[string]interface{}) error {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Log as standard audit entry with enhanced metadata
	var adminIDToUse uuid.UUID
	if adminID != nil {
		adminIDToUse = *adminID
	}

	action := string(eventType)
	return s.LogAction(ctx, adminIDToUse, action, "security_event", nil, map[string]interface{}{
		"severity": severity,
		"ip":       ipAddress,
	}, metadata)
}

// GetSeverityForEventType determines severity level for event type
func (s *AuditService) GetSeverityForEventType(eventType domain.AuditEventType) domain.AuditEventSeverity {
	switch eventType {
	// Critical events
	case domain.AuditEventAdminDeleted, domain.AuditEventBruteForceAttempt, domain.AuditEventSuspiciousActivity:
		return domain.AuditSeverityCritical

	// Error events
	case domain.AuditEventAuthFailure, domain.AuditEventIPBlocked, domain.AuditEventAdminPermChanged:
		return domain.AuditSeverityError

	// Warning events
	case domain.AuditEventPasswordChanged, domain.AuditEventMFAEnabled, domain.AuditEventMFADisabled,
		domain.AuditEventAdminCreated, domain.AuditEventAdminUpdated,
		domain.AuditEventUserCreated, domain.AuditEventUserUpdated, domain.AuditEventUserDeleted:
		return domain.AuditSeverityWarning

	// Info events
	default:
		return domain.AuditSeverityInfo
	}
}

// IsAlertableEvent checks if event should trigger alert
func (s *AuditService) IsAlertableEvent(eventType domain.AuditEventType) bool {
	alertableTypes := map[domain.AuditEventType]bool{
		domain.AuditEventAuthFailure:        true,
		domain.AuditEventBruteForceAttempt:  true,
		domain.AuditEventSuspiciousActivity: true,
		domain.AuditEventIPBlocked:          true,
		domain.AuditEventAdminDeleted:       true,
		domain.AuditEventAdminPermChanged:   true,
	}

	return alertableTypes[eventType]
}

// ExportAudits exports audit logs in specified format
func (s *AuditService) ExportAudits(ctx context.Context, filter *domain.AuditLogFilter, format domain.AuditExportFormat) (*domain.AuditLogExport, error) {
	return s.repo.ExportAudits(ctx, filter, format)
}

// CleanupOldAudits deletes audit logs older than retention period
func (s *AuditService) CleanupOldAudits(ctx context.Context, retentionDays int) (int64, error) {
	deleted, err := s.repo.DeleteOldAudits(ctx, retentionDays)
	if err != nil {
		s.log.Error("failed to delete old audits", "err", err)
		return 0, err
	}
	s.log.Info("cleaned up old audit logs", "deleted", deleted, "retention_days", retentionDays)
	return deleted, nil
}
