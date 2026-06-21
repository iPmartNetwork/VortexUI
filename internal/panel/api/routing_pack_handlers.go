package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// RoutingPackHandlers serves smart-routing rule pack endpoints: listing built-in
// + custom packs, CRUD on custom packs, applying a pack to a node, and setting
// the global default selection.
type RoutingPackHandlers struct {
	Packs *service.RoutingPackService
}

// routingPackRequest is the JSON body for creating/updating a custom pack.
type routingPackRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Category    string               `json:"category"`
	Rules       []domain.RoutingRule `json:"rules"`
	Outbounds   []domain.Outbound    `json:"outbounds"`
}

func (r routingPackRequest) toInput() service.RoutingPackInput {
	return service.RoutingPackInput{
		Name:        r.Name,
		Description: r.Description,
		Category:    r.Category,
		Rules:       r.Rules,
		Outbounds:   r.Outbounds,
	}
}

// ListPacks returns the built-in packs merged with persisted custom packs.
func (h *RoutingPackHandlers) ListPacks(c echo.Context) error {
	packs, err := h.Packs.ListPacks(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"packs": packs})
}

// CreatePack creates a new custom routing pack.
func (h *RoutingPackHandlers) CreatePack(c echo.Context) error {
	var req routingPackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pack, err := h.Packs.Create(c.Request().Context(), req.toInput())
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"pack": pack})
}

// UpdatePack updates an existing custom routing pack.
func (h *RoutingPackHandlers) UpdatePack(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req routingPackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	pack, err := h.Packs.Update(c.Request().Context(), id, req.toInput())
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "pack not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"pack": pack})
}

// DeletePack removes a custom routing pack by ID.
func (h *RoutingPackHandlers) DeletePack(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Packs.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// applyPackRequest is the body for applying a pack to a node.
type applyPackRequest struct {
	NodeID string `json:"node_id"`
	PackID string `json:"pack_id"`
}

// ApplyToNode applies a pack's rules (and any outbounds) to a node. When the
// rules persist but the node resync fails, it returns 200 with a warning so the
// admin knows the change is saved but not yet live (Requirement 3.2.2).
func (h *RoutingPackHandlers) ApplyToNode(c echo.Context) error {
	var req applyPackRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	if req.PackID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "pack_id is required")
	}
	err = h.Packs.ApplyToNode(c.Request().Context(), nodeID, req.PackID)
	switch {
	case errors.Is(err, service.ErrPackResyncFailed):
		return c.JSON(http.StatusOK, echo.Map{
			"status":  "saved",
			"warning": "rules saved but node resync failed",
		})
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "pack not found")
	case err != nil:
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "applied"})
}

// setDefaultRequest is the body for setting the global default pack.
type setDefaultRequest struct {
	PackID string `json:"pack_id"`
}

// SetDefault persists the global default pack selection.
func (h *RoutingPackHandlers) SetDefault(c echo.Context) error {
	var req setDefaultRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Packs.SetGlobalDefault(c.Request().Context(), req.PackID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"pack_id": req.PackID})
}
