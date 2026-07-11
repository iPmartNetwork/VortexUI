package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ComplianceCheckerService implements port.ComplianceChecker
type ComplianceCheckerService struct {
	complianceEventRepo port.ComplianceEventRepository
	policyRepo          port.AuditPolicyRepository
	log                 *slog.Logger
}

// NewComplianceCheckerService creates new compliance checker service
func NewComplianceCheckerService(complianceEventRepo port.ComplianceEventRepository, policyRepo port.AuditPolicyRepository, log *slog.Logger) *ComplianceCheckerService {
	if log == nil {
		log = slog.Default()
	}
	return &ComplianceCheckerService{
		complianceEventRepo: complianceEventRepo,
		policyRepo:          policyRepo,
		log:                 log,
	}
}

// CheckFramework checks compliance for specific framework
func (s *ComplianceCheckerService) CheckFramework(ctx context.Context, framework string) (map[string]interface{}, error) {
	// Get all compliance events for framework
	events, _, err := s.complianceEventRepo.GetEventsByFramework(ctx, framework, 1000, 0)
	if err != nil {
		s.log.Error("failed to get compliance events", "error", err, "framework", framework)
		return nil, err
	}

	// Calculate compliance metrics
	compliant := 0
	nonCompliant := 0
	pending := 0

	for _, event := range events {
		switch event.Status {
		case "compliant":
			compliant++
		case "non_compliant":
			nonCompliant++
		case "pending_review":
			pending++
		}
	}

	total := len(events)
	var compliancePercentage float64
	if total > 0 {
		compliancePercentage = float64(compliant) / float64(total) * 100
	}

	result := map[string]interface{}{
		"framework":              framework,
		"total_items":            total,
		"compliant":              compliant,
		"non_compliant":          nonCompliant,
		"pending_review":         pending,
		"compliance_percentage":  compliancePercentage,
		"status":                 getOverallStatus(compliant, total),
		"last_checked":           time.Now(),
	}

	return result, nil
}

// GetComplianceStatus returns overall compliance status
func (s *ComplianceCheckerService) GetComplianceStatus(ctx context.Context) (map[string]string, error) {
	policy, err := s.policyRepo.GetPolicy(ctx)
	if err != nil {
		s.log.Error("failed to get audit policy", "error", err)
		return nil, err
	}

	status := make(map[string]string)

	// Check each framework
	for _, framework := range policy.ComplianceFrameworks {
		frameworkStatus, err := s.CheckFramework(ctx, framework)
		if err != nil {
			s.log.Error("failed to check framework", "error", err, "framework", framework)
			status[framework] = "error"
			continue
		}

		// Get status from result
		if fStatus, ok := frameworkStatus["status"]; ok {
			status[framework] = fStatus.(string)
		}
	}

	return status, nil
}

// GetNonCompliantItems returns list of non-compliant items
func (s *ComplianceCheckerService) GetNonCompliantItems(ctx context.Context) ([]*domain.ComplianceEvent, error) {
	// Get all non-compliant events
	status := "non_compliant"
	events, _, err := s.complianceEventRepo.ListEvents(ctx, nil, &status, 10000, 0)
	if err != nil {
		s.log.Error("failed to get non-compliant events", "error", err)
		return nil, err
	}

	return events, nil
}

// GenerateComplianceCertificate generates proof of compliance (placeholder)
func (s *ComplianceCheckerService) GenerateComplianceCertificate(ctx context.Context, framework string) ([]byte, error) {
	// In real implementation, would generate PDF certificate
	certificateData := []byte(`
-----BEGIN COMPLIANCE CERTIFICATE-----
Framework: ` + framework + `
Generated: ` + time.Now().Format(time.RFC3339) + `
Status: VERIFIED
-----END COMPLIANCE CERTIFICATE-----
`)
	return certificateData, nil
}

// Helper function to determine overall status
func getOverallStatus(compliant, total int) string {
	if total == 0 {
		return "unknown"
	}
	percentage := float64(compliant) / float64(total) * 100
	if percentage >= 95 {
		return "compliant"
	} else if percentage >= 80 {
		return "partially_compliant"
	} else {
		return "non_compliant"
	}
}
