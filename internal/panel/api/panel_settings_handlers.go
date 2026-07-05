package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// PanelSettingsHandlers serves GET/PUT /api/settings.
type PanelSettingsHandlers struct {
	Svc *service.PanelSettingsService
}

// GetSettings returns persisted panel settings.
func (h *PanelSettingsHandlers) GetSettings(c echo.Context) error {
	cfg, err := h.Svc.Get(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"settings": cfg})
}

// UpdateSettings saves panel settings (sudo or permitted admins).
func (h *PanelSettingsHandlers) UpdateSettings(c echo.Context) error {
	var cfg domain.PanelSettings
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	out, err := h.Svc.Update(c.Request().Context(), cfg)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"settings": out})
}
