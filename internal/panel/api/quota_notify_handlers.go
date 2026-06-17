package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

type QuotaNotifyHandlers struct {
	QN *service.QuotaNotifyService
}

func (h *QuotaNotifyHandlers) GetConfig(c echo.Context) error {
	cfg, _ := h.QN.GetConfig(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *QuotaNotifyHandlers) UpdateConfig(c echo.Context) error {
	var cfg domain.QuotaNotificationConfig
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.QN.UpdateConfig(c.Request().Context(), &cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *QuotaNotifyHandlers) ListEvents(c echo.Context) error {
	events, _ := h.QN.ListEvents(c.Request().Context(), 50)
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}
