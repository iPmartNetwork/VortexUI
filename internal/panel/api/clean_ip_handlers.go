package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// CleanIPHandlers serves Clean-IP scanner endpoints.
type CleanIPHandlers struct {
	Scanner *service.CleanIPScannerService
}

type cleanIPScanRequest struct {
	IPs  []string `json:"ips"`
	Port int      `json:"port"`
}

// maxCleanIPScan caps the candidate list at the API boundary, mirroring the
// service-side cap so oversize requests are rejected early.
const maxCleanIPScan = 256

// Scan triggers a clean-IP scan over the supplied candidate IPs.
func (h *CleanIPHandlers) Scan(c echo.Context) error {
	var req cleanIPScanRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if len(req.IPs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "ips list is required")
	}
	if len(req.IPs) > maxCleanIPScan {
		return echo.NewHTTPError(http.StatusBadRequest, "max 256 IPs per scan")
	}
	results, err := h.Scanner.Scan(c.Request().Context(), req.IPs, req.Port)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}

// GetCached returns the last scan's scored results.
func (h *CleanIPHandlers) GetCached(c echo.Context) error {
	results, err := h.Scanner.GetCachedResults(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}
