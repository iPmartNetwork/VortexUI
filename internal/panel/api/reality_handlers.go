package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// RealityHandlers serves Reality Scanner endpoints.
type RealityHandlers struct {
	Scanner *service.RealityScannerService
}

type scanRequest struct {
	NodeID string   `json:"node_id"`
	SNIs   []string `json:"snis"`
	Port   int      `json:"port"`
}

// ScanReality triggers a Reality SNI scan.
func (h *RealityHandlers) ScanReality(c echo.Context) error {
	var req scanRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	if len(req.SNIs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "snis list is required")
	}
	if len(req.SNIs) > 100 {
		return echo.NewHTTPError(http.StatusBadRequest, "max 100 SNIs per scan")
	}
	results, err := h.Scanner.Scan(c.Request().Context(), nodeID, req.SNIs, req.Port)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}

// GetCachedScans returns previously cached scan results.
func (h *RealityHandlers) GetCachedScans(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	results, err := h.Scanner.GetCachedResults(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}
