package service

import (
	"context"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ThreatDetectionService implements port.ThreatDetector using pattern matching
type ThreatDetectionService struct {
	threatRepo port.SecurityThreatRepository
	ipRepRepo  port.IPReputationRepository
	log        *slog.Logger

	// Compiled patterns for threat detection
	sqlInjectionPattern *regexp.Regexp
	xssPattern          *regexp.Regexp
	pathTraversalPattern *regexp.Regexp
}

// NewThreatDetectionService creates new threat detection service
func NewThreatDetectionService(
	threatRepo port.SecurityThreatRepository,
	ipRepRepo port.IPReputationRepository,
	log *slog.Logger,
) *ThreatDetectionService {
	if log == nil {
		log = slog.Default()
	}

	return &ThreatDetectionService{
		threatRepo:          threatRepo,
		ipRepRepo:           ipRepRepo,
		log:                 log,
		sqlInjectionPattern: regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute|script|javascript|onerror|onclick|onload|fetch|XMLHttpRequest)\s*(\(|;|--)`),
		xssPattern:          regexp.MustCompile(`(?i)(<script|javascript:|onerror|onclick|onload|<iframe|<embed|<object)`),
		pathTraversalPattern: regexp.MustCompile(`(\.\.\/|\.\.\\|%2e%2e)`),
	}
}

// DetectSQLInjection checks for SQL injection attempts
func (s *ThreatDetectionService) DetectSQLInjection(ctx context.Context, payload string) (bool, *domain.SecurityThreat) {
	if s.sqlInjectionPattern.MatchString(payload) {
		threat := &domain.SecurityThreat{
			ID:              uuid.New(),
			ThreatType:      "sql_injection",
			Severity:        "high",
			Payload:         payload[:min(len(payload), 100)], // Truncate for storage
			DetectionMethod: "regex",
			Blocked:         true,
			CreatedAt:       time.Now(),
		}
		return true, threat
	}
	return false, nil
}

// DetectXSS checks for XSS attempts
func (s *ThreatDetectionService) DetectXSS(ctx context.Context, payload string) (bool, *domain.SecurityThreat) {
	if s.xssPattern.MatchString(payload) {
		threat := &domain.SecurityThreat{
			ID:              uuid.New(),
			ThreatType:      "xss",
			Severity:        "high",
			Payload:         payload[:min(len(payload), 100)],
			DetectionMethod: "regex",
			Blocked:         true,
			CreatedAt:       time.Now(),
		}
		return true, threat
	}
	return false, nil
}

// DetectCSRF checks for CSRF indicators (simplified - token validation happens elsewhere)
func (s *ThreatDetectionService) DetectCSRF(ctx context.Context, token, expectedToken string) (bool, *domain.SecurityThreat) {
	if token != expectedToken {
		threat := &domain.SecurityThreat{
			ID:              uuid.New(),
			ThreatType:      "csrf",
			Severity:        "medium",
			DetectionMethod: "token_mismatch",
			Blocked:         true,
			CreatedAt:       time.Now(),
		}
		return true, threat
	}
	return false, nil
}

// DetectAnomalies checks for behavioral anomalies (stub implementation)
func (s *ThreatDetectionService) DetectAnomalies(ctx context.Context, userID uuid.UUID) ([]*domain.AnomalyDetection, error) {
	// Placeholder for machine learning based anomaly detection
	return make([]*domain.AnomalyDetection, 0), nil
}

// DetectDDoS checks for DDoS patterns based on request count
func (s *ThreatDetectionService) DetectDDoS(ctx context.Context, clientIP string, requestCount int) (bool, *domain.SecurityThreat) {
	// Threshold: more than 1000 requests in a minute is suspicious
	if requestCount > 1000 {
		threat := &domain.SecurityThreat{
			ID:              uuid.New(),
			ThreatType:      "ddos",
			Severity:        "critical",
			SourceIP:        clientIP,
			DetectionMethod: "rate_analysis",
			Blocked:         true,
			Metadata: map[string]interface{}{
				"request_count": requestCount,
			},
			CreatedAt: time.Now(),
		}
		return true, threat
	}
	return false, nil
}

// AnomalyDetectionService implements port.AnomalyDetector
type AnomalyDetectionService struct {
	log *slog.Logger
}

// NewAnomalyDetectionService creates new anomaly detection service
func NewAnomalyDetectionService(log *slog.Logger) *AnomalyDetectionService {
	if log == nil {
		log = slog.Default()
	}
	return &AnomalyDetectionService{log: log}
}

// AnalyzeLoginPattern detects unusual login behavior
func (a *AnomalyDetectionService) AnalyzeLoginPattern(ctx context.Context, userID uuid.UUID, loginIP, expectedIP string) (*domain.AnomalyDetection, error) {
	// If login IP differs from expected, flag as anomaly
	if loginIP != expectedIP && expectedIP != "" {
		return &domain.AnomalyDetection{
			ID:          uuid.New(),
			UserID:      userID,
			AnomalyType: "unusual_login_location",
			Severity:    "medium",
			Description: "Login from unexpected IP address",
			Context: map[string]interface{}{
				"login_ip":    loginIP,
				"expected_ip": expectedIP,
			},
			Flagged:      true,
			CreatedAt:    time.Now(),
		}, nil
	}
	return nil, nil
}

// AnalyzeBulkOperation detects suspicious bulk operations
func (a *AnomalyDetectionService) AnalyzeBulkOperation(ctx context.Context, userID uuid.UUID, operationType string, count int) (*domain.AnomalyDetection, error) {
	// Threshold: bulk delete/export of >1000 items is suspicious
	if count > 1000 && (operationType == "delete" || operationType == "export") {
		return &domain.AnomalyDetection{
			ID:          uuid.New(),
			UserID:      userID,
			AnomalyType: "bulk_operation",
			Severity:    "high",
			Description: "Suspicious bulk operation detected",
			Context: map[string]interface{}{
				"operation": operationType,
				"count":     count,
			},
			Flagged:      true,
			CreatedAt:    time.Now(),
		}, nil
	}
	return nil, nil
}

// AnalyzeGeoJump detects impossible geographic transitions
func (a *AnomalyDetectionService) AnalyzeGeoJump(ctx context.Context, userID uuid.UUID, from, to string, timeDiffMins int) (*domain.AnomalyDetection, error) {
	// If same user logs in from different countries within < 15 minutes, flag as impossible
	if from != to && timeDiffMins < 15 {
		return &domain.AnomalyDetection{
			ID:          uuid.New(),
			UserID:      userID,
			AnomalyType: "geo_jump",
			Severity:    "critical",
			Description: "Impossible geographic transition detected",
			Context: map[string]interface{}{
				"from_country": from,
				"to_country":   to,
				"time_diff_mins": timeDiffMins,
			},
			Flagged:      true,
			CreatedAt:    time.Now(),
		}, nil
	}
	return nil, nil
}

// GetAnomalyRiskScore returns user risk score (0-100)
func (a *AnomalyDetectionService) GetAnomalyRiskScore(ctx context.Context, userID uuid.UUID) (float64, error) {
	// Placeholder for ML-based risk scoring
	return 0.0, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
