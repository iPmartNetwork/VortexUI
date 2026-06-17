package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// RelayHandlers serves CDN/Relay chain management endpoints.
type RelayHandlers struct {
	Relay *service.RelayService
}

type createChainRequest struct {
	Name   string           `json:"name"`
	NodeID string           `json:"node_id"`
	Hops   []domain.RelayHop `json:"hops"`
}

// CreateChain creates a new relay chain.
func (h *RelayHandlers) CreateChain(c echo.Context) error {
	var req createChainRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	chain, err := h.Relay.CreateChain(c.Request().Context(), service.CreateChainInput{
		Name:   req.Name,
		NodeID: nodeID,
		Hops:   req.Hops,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"chain": chain})
}

// ListChains lists all relay chains.
func (h *RelayHandlers) ListChains(c echo.Context) error {
	chains, err := h.Relay.ListChains(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"chains": chains})
}

type updateChainRequest struct {
	Name    string            `json:"name"`
	Hops    []domain.RelayHop `json:"hops"`
	Enabled bool              `json:"enabled"`
}

// UpdateChain updates a relay chain.
func (h *RelayHandlers) UpdateChain(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateChainRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	chain, err := h.Relay.UpdateChain(c.Request().Context(), id, req.Name, req.Hops, req.Enabled)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"chain": chain})
}

// DeleteChain removes a relay chain.
func (h *RelayHandlers) DeleteChain(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Relay.DeleteChain(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
