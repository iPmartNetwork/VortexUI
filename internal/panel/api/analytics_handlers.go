package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// AnalyticsHandlers serves advanced analytics endpoints.
type AnalyticsHandlers struct {
	Analytics *service.AnalyticsService
}

// GetAnalytics returns full analytics overview for a time range.
func (h *AnalyticsHandlers) GetAnalytics(c echo.Context) error {
	// Default: last 7 days.
	to := time.Now()
	from := to.Add(-7 * 24 * time.Hour)

	if v := c.QueryParam("from"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			from = time.Unix(ts, 0)
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			to = time.Unix(ts, 0)
		}
	}

	var adminID *uuid.UUID
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		adminID = &claims.AdminID
	}

	overview, err := h.Analytics.GetOverview(c.Request().Context(), from, to, adminID)
	if err != nil || overview == nil {
		return c.JSON(http.StatusOK, echo.Map{
			"geo_breakdown": []any{},
			"top_users":     []any{},
			"peak_hours":    []any{},
			"total_up":      0,
			"total_down":    0,
		})
	}
	return c.JSON(http.StatusOK, overview)
}

// ExportAnalyticsCSV exports analytics data as CSV.
func (h *AnalyticsHandlers) ExportAnalyticsCSV(c echo.Context) error {
	to := time.Now()
	from := to.Add(-7 * 24 * time.Hour)

	if v := c.QueryParam("from"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			from = time.Unix(ts, 0)
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			to = time.Unix(ts, 0)
		}
	}

	var adminID *uuid.UUID
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		adminID = &claims.AdminID
	}

	overview, err := h.Analytics.GetOverview(c.Request().Context(), from, to, adminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	c.Response().Header().Set("Content-Type", "text/csv")
	c.Response().Header().Set("Content-Disposition", "attachment; filename=analytics.csv")
	c.Response().WriteHeader(http.StatusOK)

	// Write CSV header.
	w := c.Response().Writer
	_, _ = w.Write([]byte("type,country,connections,bytes_up,bytes_down\n"))
	for _, g := range overview.GeoBreakdown {
		line := "geo," + g.Country + "," +
			strconv.Itoa(g.Connections) + "," +
			strconv.FormatInt(g.BytesUp, 10) + "," +
			strconv.FormatInt(g.BytesDown, 10) + "\n"
		_, _ = w.Write([]byte(line))
	}
	_, _ = w.Write([]byte("\nuser_id,username,used_traffic\n"))
	for _, u := range overview.TopUsers {
		line := u.UserID + "," + u.Username + "," +
			strconv.FormatInt(u.UsedTraffic, 10) + "\n"
		_, _ = w.Write([]byte(line))
	}
	return nil
}
