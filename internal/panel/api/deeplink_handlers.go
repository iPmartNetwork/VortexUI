package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

type DeepLinkHandlers struct {
	DeepLink *service.DeepLinkService
}

func (h *DeepLinkHandlers) GetConfig(c echo.Context) error {
	cfg, _ := h.DeepLink.GetConfig(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *DeepLinkHandlers) UpdateConfig(c echo.Context) error {
	var cfg domain.DeepLinkConfig
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.DeepLink.UpdateConfig(c.Request().Context(), &cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *DeepLinkHandlers) GenerateLink(c echo.Context) error {
	token := c.QueryParam("token")
	if token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token required")
	}
	link, err := h.DeepLink.GenerateDeepLink(c.Request().Context(), token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"deep_link": link})
}
