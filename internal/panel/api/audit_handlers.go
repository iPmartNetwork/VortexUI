package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// AuditHandlers handles audit-related API endpoints
type AuditHandlers struct {
	auditService *service.AuditService
}

// NewAuditHandlers creates new audit handlers
func NewAuditHandlers(auditService *service.AuditService) *AuditHandlers {
	return &AuditHandlers{
		auditService: auditService,
	}
}

// GetAudit retrieves an audit log by ID
// @Route GET /api/audit/logs/:id
func (h *AuditHandlers) GetAudit(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &domain.ErrorResponse{
			Code:    domain.ErrCodeInvalidInput,
			Message: "Invalid audit ID",
		})
	}

	audit, err := h.auditService.GetAudit(c.Request().Context(), id)
	if err != nil {
		if err == domain.ErrNotFound {
			return c.JSON(http.StatusNotFound, &domain.ErrorResponse{
				Code:    domain.ErrCodeNotFound,
				Message: "Audit log not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to retrieve audit log",
		})
	}

	return c.JSON(http.StatusOK, audit)
}

// ListAudits lists audit logs with filters
// @Route GET /api/audit/logs
func (h *AuditHandlers) ListAudits(c echo.Context) error {
	filter := &domain.AuditLogFilter{
		Limit:  100,
		Offset: 0,
	}

	// Parse query parameters
	if limit := c.QueryParam("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			filter.Limit = l
		}
	}

	if offset := c.QueryParam("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	if action := c.QueryParam("action"); action != "" {
		filter.Action = &action
	}

	if adminIDStr := c.QueryParam("admin_id"); adminIDStr != "" {
		if adminID, err := uuid.Parse(adminIDStr); err == nil {
			filter.AdminID = &adminID
		}
	}

	if resourceIDStr := c.QueryParam("resource_id"); resourceIDStr != "" {
		if resourceID, err := uuid.Parse(resourceIDStr); err == nil {
			filter.ResourceID = &resourceID
		}
	}

	if startDate := c.QueryParam("start_date"); startDate != "" {
		if t, err := time.Parse(time.RFC3339, startDate); err == nil {
			filter.StartDate = &t
		}
	}

	if endDate := c.QueryParam("end_date"); endDate != "" {
		if t, err := time.Parse(time.RFC3339, endDate); err == nil {
			filter.EndDate = &t
		}
	}

	audits, total, err := h.auditService.ListAudits(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to list audit logs",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"logs":   audits,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// ExportAudits exports audit logs
// @Route POST /api/audit/export
func (h *AuditHandlers) ExportAudits(c echo.Context) error {
	var req struct {
		Format    string     `json:"format"`
		StartDate *time.Time `json:"start_date,omitempty"`
		EndDate   *time.Time `json:"end_date,omitempty"`
		AdminID   *uuid.UUID `json:"admin_id,omitempty"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, &domain.ErrorResponse{
			Code:    domain.ErrCodeInvalidInput,
			Message: "Invalid request body",
		})
	}

	format := domain.AuditExportFormat(req.Format)
	if format != domain.ExportFormatJSON && format != domain.ExportFormatCSV {
		format = domain.ExportFormatJSON
	}

	filter := &domain.AuditLogFilter{
		AdminID:   req.AdminID,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Limit:     10000,
	}

	export, err := h.auditService.ExportAudits(c.Request().Context(), filter, format)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to export audit logs",
		})
	}

	if format == domain.ExportFormatCSV {
		c.Response().Header().Set(echo.HeaderContentType, "text/csv")
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=audit-logs.csv")
		return c.String(http.StatusOK, convertToCSV(export))
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/json")
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename=audit-logs.json")
	return c.JSON(http.StatusOK, export)
}

// GetAuditStats gets audit statistics
// @Route GET /api/audit/stats
func (h *AuditHandlers) GetAuditStats(c echo.Context) error {
	// Get last 7 days of audits
	startDate := time.Now().AddDate(0, 0, -7)
	filter := &domain.AuditLogFilter{
		StartDate: &startDate,
		Limit:     10000,
	}

	audits, total, err := h.auditService.ListAudits(c.Request().Context(), filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to get audit statistics",
		})
	}

	// Calculate statistics
	actionCounts := make(map[string]int)
	for _, audit := range audits {
		actionCounts[audit.Action]++
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"total":         total,
		"actions":       actionCounts,
		"period_days":   7,
		"generated_at":  time.Now(),
	})
}

// convertToCSV converts audit logs to CSV format
func convertToCSV(export *domain.AuditLogExport) string {
	csv := "ID,Admin ID,Action,Resource Type,Resource ID,IP Address,User Agent,Created At\n"

	for _, log := range export.Logs {
		resourceID := ""
		if log.ResourceID != nil {
			resourceID = log.ResourceID.String()
		}

		csv += "\"" + log.ID.String() + "\""
		csv += ",\"" + log.AdminID.String() + "\""
		csv += ",\"" + log.Action + "\""
		csv += ",\"" + log.ResourceType + "\""
		csv += ",\"" + resourceID + "\""
		csv += ",\"" + log.IPAddress + "\""
		csv += ",\"" + log.UserAgent + "\""
		csv += ",\"" + log.CreatedAt.Format(time.RFC3339) + "\"\n"
	}

	return csv
}

// RegisterAuditHandlers registers audit routes
func RegisterAuditHandlers(e *echo.Echo, handler *AuditHandlers) {
	auditGroup := e.Group("/api/audit")
	auditGroup.GET("/logs", handler.ListAudits)
	auditGroup.GET("/logs/:id", handler.GetAudit)
	auditGroup.POST("/export", handler.ExportAudits)
	auditGroup.GET("/stats", handler.GetAuditStats)
}
