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

// ComplianceHandlers handles compliance API endpoints
type ComplianceHandlers struct {
	complianceEventRepo port.ComplianceEventRepository
	complianceChecker   port.ComplianceChecker
	log                 *slog.Logger
}

// NewComplianceHandlers creates new compliance handlers
func NewComplianceHandlers(complianceEventRepo port.ComplianceEventRepository, complianceChecker port.ComplianceChecker, log *slog.Logger) *ComplianceHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &ComplianceHandlers{
		complianceEventRepo: complianceEventRepo,
		complianceChecker:   complianceChecker,
		log:                 log,
	}
}

// GetComplianceStatus retrieves overall compliance status
func (h *ComplianceHandlers) GetComplianceStatus(c echo.Context) error {
	status, err := h.complianceChecker.GetComplianceStatus(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get compliance status", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get compliance status"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"compliance_status": status,
	})
}

// CheckFramework checks compliance for specific framework
func (h *ComplianceHandlers) CheckFramework(c echo.Context) error {
	framework := c.Param("framework")
	if framework == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "framework required"})
	}

	result, err := h.complianceChecker.CheckFramework(c.Request().Context(), framework)
	if err != nil {
		h.log.Error("failed to check framework", "error", err, "framework", framework)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to check framework"})
	}

	return c.JSON(http.StatusOK, result)
}

// ListComplianceEvents retrieves compliance events with filtering
func (h *ComplianceHandlers) ListComplianceEvents(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	eventType := c.QueryParam("event_type")
	eventTypePtr := &eventType
	if eventType == "" {
		eventTypePtr = nil
	}

	status := c.QueryParam("status")
	statusPtr := &status
	if status == "" {
		statusPtr = nil
	}

	events, total, err := h.complianceEventRepo.ListEvents(c.Request().Context(), eventTypePtr, statusPtr, limit, offset)
	if err != nil {
		h.log.Error("failed to list compliance events", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list compliance events"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetComplianceEvent retrieves a specific compliance event
func (h *ComplianceHandlers) GetComplianceEvent(c echo.Context) error {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid event ID"})
	}

	event, err := h.complianceEventRepo.GetEvent(c.Request().Context(), eventID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "event not found"})
		}
		h.log.Error("failed to get compliance event", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get event"})
	}

	return c.JSON(http.StatusOK, event)
}

// GetNonCompliantItems retrieves list of non-compliant items
func (h *ComplianceHandlers) GetNonCompliantItems(c echo.Context) error {
	items, err := h.complianceChecker.GetNonCompliantItems(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get non-compliant items", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get non-compliant items"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

// VerifyComplianceEvent marks event as verified (admin-only)
func (h *ComplianceHandlers) VerifyComplianceEvent(c echo.Context) error {
	eventIDStr := c.Param("id")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid event ID"})
	}

	// Extract admin ID from context
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	err = h.complianceEventRepo.VerifyEvent(c.Request().Context(), eventID, adminID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "event not found"})
		}
		h.log.Error("failed to verify compliance event", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to verify event"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "verified"})
}

// GenerateComplianceCertificate generates compliance certificate for framework
func (h *ComplianceHandlers) GenerateComplianceCertificate(c echo.Context) error {
	framework := c.Param("framework")
	if framework == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "framework required"})
	}

	certificate, err := h.complianceChecker.GenerateComplianceCertificate(c.Request().Context(), framework)
	if err != nil {
		h.log.Error("failed to generate certificate", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate certificate"})
	}

	c.Response().Header().Set("Content-Type", "application/x-pem-file")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=compliance-certificate-"+framework+".pem")
	return c.Blob(http.StatusOK, "application/x-pem-file", certificate)
}
