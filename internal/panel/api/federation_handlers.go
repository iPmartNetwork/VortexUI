package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

type FederationHandlers struct {
	Fed *service.FederationService
}

func (h *FederationHandlers) GetConfig(c echo.Context) error {
	cfg, _ := h.Fed.GetConfig(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

func (h *FederationHandlers) UpdateConfig(c echo.Context) error {
	var cfg domain.FederationConfig
	if err := c.Bind(&cfg); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Fed.UpdateConfig(c.Request().Context(), &cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

type addPeerRequest struct {
	Name      string `json:"name"`
	Endpoint  string `json:"endpoint"`
	APIKey    string `json:"api_key"`
	SyncUsers bool   `json:"sync_users"`
	SyncNodes bool   `json:"sync_nodes"`
}

func (h *FederationHandlers) AddPeer(c echo.Context) error {
	var req addPeerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.Fed.AddPeer(c.Request().Context(), req.Name, req.Endpoint, req.APIKey, req.SyncUsers, req.SyncNodes)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"peer": p})
}

func (h *FederationHandlers) ListPeers(c echo.Context) error {
	peers, _ := h.Fed.ListPeers(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"peers": peers})
}

func (h *FederationHandlers) DeletePeer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	_ = h.Fed.DeletePeer(c.Request().Context(), id)
	return c.NoContent(http.StatusNoContent)
}

func (h *FederationHandlers) ListSyncEvents(c echo.Context) error {
	events, _ := h.Fed.ListSyncEvents(c.Request().Context(), 50)
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}
