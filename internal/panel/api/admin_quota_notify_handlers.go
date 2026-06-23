package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// AdminQuotaNotifyHandlers serves reseller quota alert configuration.
type AdminQuotaNotifyHandlers struct {
	Svc *service.AdminQuotaNotifyService
}

func (h *AdminQuotaNotifyHandlers) GetConfig(c echo.Context) error {
	cfg, err := h.Svc.GetConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "config failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *AdminQuotaNotifyHandlers) UpdateConfig(c echo.Context) error {
	var cfg domain.AdminQuotaNotifyConfig
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Svc.UpdateConfig(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *AdminQuotaNotifyHandlers) ListEvents(c echo.Context) error {
	events, err := h.Svc.ListEvents(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}
