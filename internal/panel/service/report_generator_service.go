package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ReportGeneratorService implements port.ReportGenerator
type ReportGeneratorService struct {
	eventRepo  port.AuditEventRepository
	reportRepo port.AuditReportRepository
	log        *slog.Logger
}

// NewReportGeneratorService creates new report generator service
func NewReportGeneratorService(eventRepo port.AuditEventRepository, reportRepo port.AuditReportRepository, log *slog.Logger) *ReportGeneratorService {
	if log == nil {
		log = slog.Default()
	}
	return &ReportGeneratorService{
		eventRepo:  eventRepo,
		reportRepo: reportRepo,
		log:        log,
	}
}

// GenerateDailyReport generates a daily audit report
func (s *ReportGeneratorService) GenerateDailyReport(ctx context.Context, date time.Time, createdBy uuid.UUID) (*domain.AuditReport, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endDate := startDate.AddDate(0, 0, 1)

	// Get events for the day
	events, total, err := s.eventRepo.ListEvents(ctx, nil, nil, nil, &startDate, &endDate, 10000, 0)
	if err != nil {
		s.log.Error("failed to get daily events", "error", err, "date", date)
		return nil, err
	}

	report := &domain.AuditReport{
		ID:          uuid.New(),
		Title:       fmt.Sprintf("Daily Audit Report - %s", startDate.Format("2006-01-02")),
		Description: fmt.Sprintf("Comprehensive audit report for %s", startDate.Format("Monday, January 02, 2006")),
		ReportType:  "daily",
		Status:      "draft",
		StartDate:   startDate,
		EndDate:     endDate,
		Scope:       "system-wide",
		EventCount:  int(total),
		CreatedBy:   createdBy,
		Filters: map[string]interface{}{
			"date": startDate.Format("2006-01-02"),
		},
		Metadata: map[string]interface{}{
			"event_count": total,
			"events":      len(events),
		},
	}

	// Save report
	err = s.reportRepo.SaveReport(ctx, report)
	if err != nil {
		s.log.Error("failed to save daily report", "error", err)
		return nil, err
	}

	return report, nil
}

// GenerateCustomReport generates a report for date range
func (s *ReportGeneratorService) GenerateCustomReport(ctx context.Context, startDate time.Time, endDate time.Time, filters map[string]interface{}, createdBy uuid.UUID) (*domain.AuditReport, error) {
	// Get events within date range
	events, total, err := s.eventRepo.ListEvents(ctx, nil, nil, nil, &startDate, &endDate, 10000, 0)
	if err != nil {
		s.log.Error("failed to get custom report events", "error", err)
		return nil, err
	}

	if filters == nil {
		filters = make(map[string]interface{})
	}

	report := &domain.AuditReport{
		ID:          uuid.New(),
		Title:       fmt.Sprintf("Custom Audit Report (%s to %s)", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")),
		Description: "Custom audit report with user-specified filters",
		ReportType:  "custom",
		Status:      "draft",
		StartDate:   startDate,
		EndDate:     endDate,
		Scope:       "custom",
		EventCount:  int(total),
		Filters:     filters,
		CreatedBy:   createdBy,
		Metadata: map[string]interface{}{
			"event_count": total,
			"events":      len(events),
		},
	}

	err = s.reportRepo.SaveReport(ctx, report)
	if err != nil {
		s.log.Error("failed to save custom report", "error", err)
		return nil, err
	}

	return report, nil
}

// GenerateIncidentReport generates report for specific incident
func (s *ReportGeneratorService) GenerateIncidentReport(ctx context.Context, eventID uuid.UUID, createdBy uuid.UUID) (*domain.AuditReport, error) {
	// Get the incident event
	incident, err := s.eventRepo.GetEvent(ctx, eventID)
	if err != nil {
		s.log.Error("failed to get incident event", "error", err)
		return nil, err
	}

	// Get related events (same type, same admin, same time period)
	hourBefore := incident.CreatedAt.Add(-1 * time.Hour)
	hourAfter := incident.CreatedAt.Add(1 * time.Hour)

	eventTypeStr := string(incident.EventType)
	events, total, err := s.eventRepo.ListEvents(ctx, &incident.AdminID, &eventTypeStr, nil, &hourBefore, &hourAfter, 100, 0)
	if err != nil {
		s.log.Error("failed to get incident related events", "error", err)
		return nil, err
	}

	report := &domain.AuditReport{
		ID:          uuid.New(),
		Title:       fmt.Sprintf("Incident Report - %s", incident.Description),
		Description: fmt.Sprintf("Detailed incident analysis for event: %s", incident.EventType),
		ReportType:  "incident",
		Status:      "draft",
		StartDate:   hourBefore,
		EndDate:     hourAfter,
		Scope:       "incident-specific",
		EventCount:  int(total),
		CreatedBy:   createdBy,
		Filters: map[string]interface{}{
			"event_id": eventID.String(),
		},
		Metadata: map[string]interface{}{
			"incident_event_type": incident.EventType,
			"admin_id":            incident.AdminID.String(),
			"severity":            incident.Severity,
			"related_events":      len(events),
		},
	}

	err = s.reportRepo.SaveReport(ctx, report)
	if err != nil {
		s.log.Error("failed to save incident report", "error", err)
		return nil, err
	}

	return report, nil
}

// GenerateComplianceReport generates compliance-focused report
func (s *ReportGeneratorService) GenerateComplianceReport(ctx context.Context, framework string, period string, createdBy uuid.UUID) (*domain.AuditReport, error) {
	// Determine date range based on period
	var startDate, endDate time.Time
	now := time.Now()

	switch period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 0, 1)
	case "weekly":
		weekStart := now.AddDate(0, 0, -int(now.Weekday()))
		startDate = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
		endDate = startDate.AddDate(0, 0, 7)
	case "monthly":
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, 0)
	case "quarterly":
		quarter := (int(now.Month()) - 1) / 3
		startDate = time.Date(now.Year(), time.Month(quarter*3+1), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 3, 0)
	case "yearly":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(1, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -30)
		endDate = now
	}

	events, total, err := s.eventRepo.ListEvents(ctx, nil, nil, nil, &startDate, &endDate, 10000, 0)
	if err != nil {
		s.log.Error("failed to get compliance report events", "error", err)
		return nil, err
	}

	report := &domain.AuditReport{
		ID:          uuid.New(),
		Title:       fmt.Sprintf("%s Compliance Report - %s", framework, period),
		Description: fmt.Sprintf("Compliance verification report for %s framework (%s)", framework, period),
		ReportType:  "compliance",
		Status:      "draft",
		StartDate:   startDate,
		EndDate:     endDate,
		Scope:       "compliance",
		EventCount:  int(total),
		CreatedBy:   createdBy,
		Filters: map[string]interface{}{
			"framework": framework,
			"period":    period,
		},
		Metadata: map[string]interface{}{
			"compliance_framework": framework,
			"period":               period,
			"event_count":          total,
			"events":               len(events),
		},
	}

	err = s.reportRepo.SaveReport(ctx, report)
	if err != nil {
		s.log.Error("failed to save compliance report", "error", err)
		return nil, err
	}

	return report, nil
}

// ExportReport exports report in specified format
func (s *ReportGeneratorService) ExportReport(ctx context.Context, reportID uuid.UUID, format string) ([]byte, error) {
	report, err := s.reportRepo.GetReport(ctx, reportID)
	if err != nil {
		s.log.Error("failed to get report for export", "error", err)
		return nil, err
	}

	// In real implementation, would generate actual PDF/CSV/JSON file
	// For now, return a placeholder
	data := []byte(fmt.Sprintf("Report %s exported as %s", report.ID, format))

	// Mark report as exported
	err = s.reportRepo.ExportReport(ctx, reportID, fmt.Sprintf("s3://audit-reports/%s.%s", report.ID, format))
	if err != nil {
		s.log.Error("failed to mark report as exported", "error", err)
		// Don't fail export just because we couldn't update the path
	}

	return data, nil
}

// GetReportPreview gets preview of report before generation
func (s *ReportGeneratorService) GetReportPreview(ctx context.Context, filters map[string]interface{}) (map[string]interface{}, error) {
	preview := map[string]interface{}{
		"status":       "ready",
		"event_count":  0,
		"date_range":   "today",
		"scope":        "system-wide",
		"generated_at": time.Now().Format(time.RFC3339),
	}

	return preview, nil
}
