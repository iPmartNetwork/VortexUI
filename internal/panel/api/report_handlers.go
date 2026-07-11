package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ReportHandlers handles audit report API endpoints
type ReportHandlers struct {
	reportRepo        port.AuditReportRepository
	reportGenerator   port.ReportGenerator
	log               *slog.Logger
}

// NewReportHandlers creates new report handlers
func NewReportHandlers(reportRepo port.AuditReportRepository, reportGenerator port.ReportGenerator, log *slog.Logger) *ReportHandlers {
	if log == nil {
		log = slog.Default()
	}
	return &ReportHandlers{
		reportRepo:      reportRepo,
		reportGenerator: reportGenerator,
		log:             log,
	}
}

// GenerateDailyReport generates a daily audit report
func (h *ReportHandlers) GenerateDailyReport(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		Date string `json:"date"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	date := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			date = t
		}
	}

	report, err := h.reportGenerator.GenerateDailyReport(c.Request().Context(), date, adminID)
	if err != nil {
		h.log.Error("failed to generate daily report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate report"})
	}

	return c.JSON(http.StatusCreated, report)
}

// GenerateCustomReport generates a custom report for date range
func (h *ReportHandlers) GenerateCustomReport(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		StartDate string                 `json:"start_date"`
		EndDate   string                 `json:"end_date"`
		Filters   map[string]interface{} `json:"filters"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid start_date format"})
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid end_date format"})
	}

	if startDate.After(endDate) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "start_date must be before end_date"})
	}

	report, err := h.reportGenerator.GenerateCustomReport(c.Request().Context(), startDate, endDate, req.Filters, adminID)
	if err != nil {
		h.log.Error("failed to generate custom report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate report"})
	}

	return c.JSON(http.StatusCreated, report)
}

// GenerateComplianceReport generates compliance-focused report
func (h *ReportHandlers) GenerateComplianceReport(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		Framework string `json:"framework"`
		Period    string `json:"period"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Framework == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "framework required"})
	}

	period := "monthly"
	if req.Period != "" {
		period = req.Period
	}

	report, err := h.reportGenerator.GenerateComplianceReport(c.Request().Context(), req.Framework, period, adminID)
	if err != nil {
		h.log.Error("failed to generate compliance report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate report"})
	}

	return c.JSON(http.StatusCreated, report)
}

// GetReport retrieves a specific report
func (h *ReportHandlers) GetReport(c echo.Context) error {
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid report ID"})
	}

	report, err := h.reportRepo.GetReport(c.Request().Context(), reportID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "report not found"})
		}
		h.log.Error("failed to get report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get report"})
	}

	return c.JSON(http.StatusOK, report)
}

// ListReports retrieves reports with optional filtering
func (h *ReportHandlers) ListReports(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	reportType := c.QueryParam("report_type")
	reportTypePtr := &reportType
	if reportType == "" {
		reportTypePtr = nil
	}

	status := c.QueryParam("status")
	statusPtr := &status
	if status == "" {
		statusPtr = nil
	}

	reports, total, err := h.reportRepo.ListReports(c.Request().Context(), reportTypePtr, statusPtr, limit, offset)
	if err != nil {
		h.log.Error("failed to list reports", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list reports"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"reports": reports,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

// ApproveReport approves a pending report (admin-only)
func (h *ReportHandlers) ApproveReport(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid report ID"})
	}

	err = h.reportRepo.ApproveReport(c.Request().Context(), reportID, adminID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "report not found"})
		}
		h.log.Error("failed to approve report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to approve report"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "approved"})
}

// ExportReport exports a report in specified format
func (h *ReportHandlers) ExportReport(c echo.Context) error {
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid report ID"})
	}

	format := c.QueryParam("format")
	if format == "" {
		format = "json"
	}

	// Validate format
	validFormats := map[string]bool{"json": true, "csv": true, "pdf": true}
	if !validFormats[format] {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid format. must be json, csv, or pdf"})
	}

	data, err := h.reportGenerator.ExportReport(c.Request().Context(), reportID, format)
	if err != nil {
		h.log.Error("failed to export report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to export report"})
	}

	contentType := "application/json"
	extension := "json"

	switch format {
	case "csv":
		contentType = "text/csv"
		extension = "csv"
	case "pdf":
		contentType = "application/pdf"
		extension = "pdf"
	}

	c.Response().Header().Set("Content-Type", contentType)
	c.Response().Header().Set("Content-Disposition", "attachment; filename=report-"+reportID.String()+"."+extension)
	return c.Blob(http.StatusOK, contentType, data)
}

// DeleteReport deletes a report (admin-only)
func (h *ReportHandlers) DeleteReport(c echo.Context) error {
	reportIDStr := c.Param("id")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid report ID"})
	}

	err = h.reportRepo.DeleteReport(c.Request().Context(), reportID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "report not found"})
		}
		h.log.Error("failed to delete report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete report"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}
