package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// DecoyHandlers serves decoy website management endpoints.
type DecoyHandlers struct {
	Decoy *service.DecoyService
}

type createDecoyRequest struct {
	NodeID     string `json:"node_id"`
	Mode       string `json:"mode"`
	TargetURL  string `json:"target_url"`
	StaticHTML string `json:"static_html"`
}

// CreateDecoy creates a new decoy site config.
func (h *DecoyHandlers) CreateDecoy(c echo.Context) error {
	var req createDecoyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	var nodeID *uuid.UUID
	if req.NodeID != "" {
		id, err := uuid.Parse(req.NodeID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
		}
		nodeID = &id
	}
	d, err := h.Decoy.CreateDecoy(c.Request().Context(), service.CreateDecoyInput{
		NodeID:     nodeID,
		Mode:       domain.DecoyMode(req.Mode),
		TargetURL:  req.TargetURL,
		StaticHTML: req.StaticHTML,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"decoy": d})
}

// ListDecoys lists all decoy configurations.
func (h *DecoyHandlers) ListDecoys(c echo.Context) error {
	decoys, err := h.Decoy.ListDecoys(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"decoys": decoys})
}

type updateDecoyRequest struct {
	Mode       string `json:"mode"`
	TargetURL  string `json:"target_url"`
	StaticHTML string `json:"static_html"`
	Enabled    bool   `json:"enabled"`
}

// UpdateDecoy updates a decoy configuration.
func (h *DecoyHandlers) UpdateDecoy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateDecoyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	d, err := h.Decoy.UpdateDecoy(c.Request().Context(), id, domain.DecoyMode(req.Mode), req.TargetURL, req.StaticHTML, req.Enabled)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"decoy": d})
}

// DeleteDecoy removes a decoy configuration.
func (h *DecoyHandlers) DeleteDecoy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Decoy.DeleteDecoy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
