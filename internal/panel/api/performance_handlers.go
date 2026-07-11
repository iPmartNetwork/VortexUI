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

// PerformanceHandlers handles performance monitoring API endpoints
type PerformanceHandlers struct {
	metricsRepo port.QueryMetricsRepository
	alertRepo   port.PerformanceAlertRepository
	rateLimitRepo port.RateLimitRepository
	monitor     port.PerformanceMonitor
	log         *slog.Logger
}

// NewPerformanceHandlers creates new performance handlers
func NewPerformanceHandlers(
	metricsRepo port.QueryMetricsRepository,
	alertRepo port.PerformanceAlertRepository,
	rateLimitRepo port.RateLimitRepository,
	monitor port.PerformanceMonitor,
	log *slog.Logger,
) *PerformanceHandlers {
	return &PerformanceHandlers{
		metricsRepo: metricsRepo,
		alertRepo:   alertRepo,
		rateLimitRepo: rateLimitRepo,
		monitor:     monitor,
		log:         log,
	}
}

// GetHealthStatus returns overall performance health
// GET /performance/health
func (h *PerformanceHandlers) GetHealthStatus(c echo.Context) error {
	health, err := h.monitor.GetHealthStatus(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get health status", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get health status"})
	}

	return c.JSON(http.StatusOK, health)
}

// GetSlowQueries returns slow queries exceeding threshold
// GET /performance/queries/slow?limit=50&offset=0
func (h *PerformanceHandlers) GetSlowQueries(c echo.Context) error {
	limit := 50
	offset := 0

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

	queries, err := h.metricsRepo.ListSlowQueries(c.Request().Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to get slow queries", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get queries"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":    len(queries),
		"limit":    limit,
		"offset":   offset,
		"queries":  queries,
	})
}

// GetQueryStats returns aggregated query statistics
// GET /performance/queries/stats
func (h *PerformanceHandlers) GetQueryStats(c echo.Context) error {
	stats, err := h.metricsRepo.GetQueryStats(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get query stats", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get stats"})
	}

	return c.JSON(http.StatusOK, stats)
}

// GetPerformanceAlerts returns active performance alerts
// GET /performance/alerts?resolved=false
func (h *PerformanceHandlers) GetPerformanceAlerts(c echo.Context) error {
	alerts, err := h.alertRepo.ListActiveAlerts(c.Request().Context())
	if err != nil {
		h.log.Error("failed to get performance alerts", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get alerts"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":  len(alerts),
		"alerts": alerts,
	})
}

// ResolveAlert marks alert as resolved
// PUT /performance/alerts/:id/resolve
func (h *PerformanceHandlers) ResolveAlert(c echo.Context) error {
	alertID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid alert ID"})
	}

	err = h.alertRepo.ResolveAlert(c.Request().Context(), alertID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "alert not found"})
		}
		h.log.Error("failed to resolve alert", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to resolve alert"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "resolved"})
}

// GetPerformanceReport generates performance analysis report
// GET /performance/report?hours_back=24
func (h *PerformanceHandlers) GetPerformanceReport(c echo.Context) error {
	hoursBack := 24
	if h := c.QueryParam("hours_back"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
			hoursBack = parsed
		}
	}

	report, err := h.monitor.GeneratePerformanceReport(c.Request().Context(), hoursBack)
	if err != nil {
		h.log.Error("failed to generate performance report", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate report"})
	}

	return c.JSON(http.StatusOK, report)
}

// ListRateLimitRules returns all active rate limit rules
// GET /performance/rate-limits/rules
func (h *PerformanceHandlers) ListRateLimitRules(c echo.Context) error {
	rules, err := h.rateLimitRepo.ListRules(c.Request().Context())
	if err != nil {
		h.log.Error("failed to list rate limit rules", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get rules"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count": len(rules),
		"rules": rules,
	})
}

// GetRateLimitViolations returns recent violations for IP
// GET /performance/rate-limits/violations?client_ip=1.2.3.4&minutes_back=60
func (h *PerformanceHandlers) GetRateLimitViolations(c echo.Context) error {
	clientIP := c.QueryParam("client_ip")
	if clientIP == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "client_ip required"})
	}

	minutesBack := 60
	if m := c.QueryParam("minutes_back"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 {
			minutesBack = parsed
		}
	}

	violations, err := h.rateLimitRepo.GetViolations(c.Request().Context(), clientIP, minutesBack)
	if err != nil {
		h.log.Error("failed to get violations", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get violations"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"count":        len(violations),
		"client_ip":    clientIP,
		"violations":   violations,
	})
}
