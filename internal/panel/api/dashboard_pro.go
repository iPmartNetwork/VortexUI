package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// DashboardProHandler serves advanced dashboard analytics endpoints.
type DashboardProHandler struct {
	svc *service.DashboardProService
}

// NewDashboardProHandler creates a new DashboardProHandler with the given service.
func NewDashboardProHandler(svc *service.DashboardProService) *DashboardProHandler {
	return &DashboardProHandler{svc: svc}
}

// Register mounts all dashboard pro routes on the given Echo group.
func (h *DashboardProHandler) Register(g *echo.Group) {
	dash := g.Group("/dashboard")
	dash.GET("/daily-check", h.DailyCheck)
	dash.GET("/isp-heatmap", h.ISPHeatmap)
	dash.GET("/geo-map", h.GeoMap)
	dash.GET("/revenue", h.Revenue)
	dash.GET("/sub-analytics", h.SubAnalytics)
}

// DailyCheck handles GET /api/v2/dashboard/daily-check.
// Returns aggregated node health, traffic anomaly status, cert validity, and
// diagnostic cards for actionable issues.
func (h *DashboardProHandler) DailyCheck(c echo.Context) error {
	widget, err := h.svc.DailyCheck(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"daily_check": widget})
}

// ISPHeatmap handles GET /api/v2/dashboard/isp-heatmap?isp=MCI&days=7.
// Returns a 7-day x 24-hour quality grid for the specified ISP.
func (h *DashboardProHandler) ISPHeatmap(c echo.Context) error {
	isp := c.QueryParam("isp")
	if isp == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "isp parameter is required")
	}

	days := 7
	if d := c.QueryParam("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	heatmap, err := h.svc.ISPHeatmap(c.Request().Context(), isp, days)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"heatmap": heatmap})
}

// GeoMap handles GET /api/v2/dashboard/geo-map.
// Returns node locations with live status for geographic map visualization.
func (h *DashboardProHandler) GeoMap(c echo.Context) error {
	nodes, err := h.svc.GeoMap(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"nodes": nodes})
}

// Revenue handles GET /api/v2/dashboard/revenue?from=...&to=...&admin_id=...
// Returns aggregated revenue data with per-admin breakdowns and time series.
func (h *DashboardProHandler) Revenue(c echo.Context) error {
	from, to, err := parseTimeRange(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var adminID *uuid.UUID
	if aid := c.QueryParam("admin_id"); aid != "" {
		parsed, err := uuid.Parse(aid)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid admin_id")
		}
		adminID = &parsed
	}

	report, err := h.svc.Revenue(c.Request().Context(), adminID, from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"revenue": report})
}

// SubAnalytics handles GET /api/v2/dashboard/sub-analytics?from=...&to=...
// Returns subscription fetch analytics grouped by format, ISP, and time.
func (h *DashboardProHandler) SubAnalytics(c echo.Context) error {
	from, to, err := parseTimeRange(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	report, err := h.svc.SubAnalytics(c.Request().Context(), from, to)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"sub_analytics": report})
}

// parseTimeRange extracts from/to query parameters as time.Time values.
// Defaults to the last 30 days if not specified.
func parseTimeRange(c echo.Context) (time.Time, time.Time, error) {
	now := time.Now()
	from := now.AddDate(0, 0, -30)
	to := now

	if f := c.QueryParam("from"); f != "" {
		parsed, err := time.Parse(time.RFC3339, f)
		if err != nil {
			// Try date-only format
			parsed, err = time.Parse("2006-01-02", f)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
		}
		from = parsed
	}

	if t := c.QueryParam("to"); t != "" {
		parsed, err := time.Parse(time.RFC3339, t)
		if err != nil {
			// Try date-only format
			parsed, err = time.Parse("2006-01-02", t)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
			// End of day for date-only format
			parsed = parsed.Add(24*time.Hour - time.Second)
		}
		to = parsed
	}

	return from, to, nil
}
