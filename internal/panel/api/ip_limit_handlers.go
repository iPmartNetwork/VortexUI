package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// IPLimitHandlers serves the IP-limit enforcement policy and event endpoints.
type IPLimitHandlers struct {
	IPLimit *service.IPLimitService
}

// GetPolicy returns the current enforcement policy.
func (h *IPLimitHandlers) GetPolicy(c echo.Context) error {
	pol, _ := h.IPLimit.GetPolicy(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"policy": pol})
}

// UpdatePolicy persists the enforcement policy.
func (h *IPLimitHandlers) UpdatePolicy(c echo.Context) error {
	var pol domain.IPLimitPolicy
	if err := c.Bind(&pol); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.IPLimit.UpdatePolicy(c.Request().Context(), &pol); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": pol})
}

// ListEvents returns the most recent enforcement events.
func (h *IPLimitHandlers) ListEvents(c echo.Context) error {
	events, _ := h.IPLimit.ListEvents(c.Request().Context(), 50)
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}
