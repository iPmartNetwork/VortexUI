package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SubSettingsHandlers serves the panel-configurable subscription settings.
type SubSettingsHandlers struct {
	Svc *service.SubSettingsService
}

// GetSubSettings returns the current subscription settings.
func (h *SubSettingsHandlers) GetSubSettings(c echo.Context) error {
	cfg, err := h.Svc.GetConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

// UpdateSubSettings saves the subscription settings.
func (h *SubSettingsHandlers) UpdateSubSettings(c echo.Context) error {
	var cfg domain.SubSettings
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Svc.UpdateConfig(c.Request().Context(), &cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}
