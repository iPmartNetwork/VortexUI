package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// RateLimitHandler serves rate limit dashboard stats.
type RateLimitHandler struct {
	dashboard *service.RateLimitDashboard
}

// NewRateLimitHandler creates the handler.
func NewRateLimitHandler(dashboard *service.RateLimitDashboard) *RateLimitHandler {
	return &RateLimitHandler{dashboard: dashboard}
}

// Register mounts rate limit routes.
func (h *RateLimitHandler) Register(g *echo.Group) {
	g.GET("/rate-limits", h.GetRateLimits)
}

// GetRateLimits handles GET /api/v2/rate-limits.
func (h *RateLimitHandler) GetRateLimits(c echo.Context) error {
	stats := h.dashboard.GetStats(c.Request().Context())
	return c.JSON(http.StatusOK, stats)
}
