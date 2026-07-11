package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// SecurityHandlers handles security hardening API endpoints
type SecurityHandlers struct {
	threatRepo port.SecurityThreatRepository
	policyRepo port.SecurityPolicyRepository
	threatDetector port.ThreatDetector
	anomalyDetector port.AnomalyDetector
	log *slog.Logger
}

// NewSecurityHandlers creates new security handlers
func NewSecurityHandlers(
	threatRepo port.SecurityThreatRepository,
	policyRepo port.SecurityPolicyRepository,
	threatDetector port.ThreatDetector,
	anomalyDetector port.AnomalyDetector,
	log *slog.Logger,
) *SecurityHandlers {
	return &SecurityHandlers{
		threatRepo: threatRepo,
		policyRepo: policyRepo,
		threatDetector: threatDetector,
		anomalyDetector: anomalyDetector,
		log: log,
	}
}

// GetSecurityThreats returns recent security threats
// GET /security/threats?limit=50&offset=0&type=sql_injection
func (h *SecurityHandlers) GetSecurityThreats(c echo.Context) error {
	limit := 50
	offset := 0
	threatType := c.QueryParam("type")

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	threats, err := h.threatRepo.ListThreats(c.Request().Context(), threatType, limit, offset)
	if err != nil {
		h.log.Error("failed to list threats", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get threats"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":   len(threats),
		"limit":   limit,
		"offset":  offset,
		"threats": threats,
	})
}

// GetSecurityPolicy returns current security policy
// GET /security/policy
func (h *SecurityHandlers) GetSecurityPolicy(c echo.Context) error {
	policy, err := h.policyRepo.GetPolicy(c.Request().Context())
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Return default policy if not found
			policy, _ = h.policyRepo.GetDefaultPolicy(c.Request().Context())
		}
		if policy == nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get policy"})
		}
	}

	return c.JSON(http.StatusOK, policy)
}

// UpdateSecurityPolicy updates security policy
// PUT /security/policy
func (h *SecurityHandlers) UpdateSecurityPolicy(c echo.Context) error {
	policy := &domain.SecurityPolicy{}
	if err := c.Bind(policy); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	policy.ID = uuid.New()
	if err := h.policyRepo.SavePolicy(c.Request().Context(), policy); err != nil {
		h.log.Error("failed to save policy", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update policy"})
	}

	return c.JSON(http.StatusOK, policy)
}

// GetBlockedThreats returns threats that were blocked
// GET /security/threats/blocked?limit=50
func (h *SecurityHandlers) GetBlockedThreats(c echo.Context) error {
	limit := 50
	offset := 0

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	threats, err := h.threatRepo.GetBlockedThreats(c.Request().Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to get blocked threats", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get threats"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":   len(threats),
		"threats": threats,
	})
}

// GetThreatCount returns total threat count
// GET /security/threats/count
func (h *SecurityHandlers) GetThreatCount(c echo.Context) error {
	count, err := h.threatRepo.CountThreats(c.Request().Context())
	if err != nil {
		h.log.Error("failed to count threats", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get count"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count": count,
	})
}

// GetSecurityScore returns system security score
// GET /security/score
func (h *SecurityHandlers) GetSecurityScore(c echo.Context) error {
	// Placeholder for comprehensive security scoring
	// In real implementation, would aggregate multiple metrics
	score := map[string]interface{}{
		"overall_score": 85.0,
		"components": map[string]interface{}{
			"threat_detection":  90,
			"policy_compliance": 85,
			"encryption":        80,
			"access_control":    88,
		},
		"timestamp": c.Request().Context(),
	}

	return c.JSON(http.StatusOK, score)
}

// ValidateCompliance checks if system meets policy requirements
// GET /security/compliance/validate
func (h *SecurityHandlers) ValidateCompliance(c echo.Context) error {
	compliantMap, issues, err := h.policyRepo.ValidateCompliance(c.Request().Context())
	if err != nil {
		h.log.Error("failed to validate compliance", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to validate"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"compliant": compliantMap,
		"issues":    issues,
	})
}
