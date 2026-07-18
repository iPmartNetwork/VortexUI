package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// GetAutoConfig returns AI-recommended proxy settings for the requesting client.
// Query params:
//   - isp: ISP preset name (e.g. "hamrah_aval", "irancell", "mokhaberat")
//   - cdn: "true" if the node is behind a CDN
func (h *Handlers) GetAutoConfig(c echo.Context) error {
	isp := c.QueryParam("isp")
	hasCDN := c.QueryParam("cdn") == "true"
	engine := service.NewAutoConfigEngine()
	rec := engine.Recommend(domain.ISPPreset(isp), hasCDN)
	return c.JSON(http.StatusOK, rec)
}
