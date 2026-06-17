package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// DoHHandlers serves DNS-over-HTTPS configuration endpoints.
type DoHHandlers struct {
	DoH *service.DoHService
}

// GetDoHConfig returns the current DoH config.
func (h *DoHHandlers) GetDoHConfig(c echo.Context) error {
	cfg, err := h.DoH.GetConfig(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

type updateDoHConfigRequest struct {
	Enabled        bool     `json:"enabled"`
	ListenAddr     string   `json:"listen_addr"`
	UpstreamDNS    []string `json:"upstream_dns"`
	BlockAds       bool     `json:"block_ads"`
	BlockMalware   bool     `json:"block_malware"`
	CustomBlocklist []string `json:"custom_blocklist"`
	LogQueries     bool     `json:"log_queries"`
	CacheTTL       int      `json:"cache_ttl"`
}

// UpdateDoHConfig saves the DoH configuration.
func (h *DoHHandlers) UpdateDoHConfig(c echo.Context) error {
	var req updateDoHConfigRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	cfg := &domain.DoHConfig{
		Enabled:         req.Enabled,
		ListenAddr:      req.ListenAddr,
		UpstreamDNS:     req.UpstreamDNS,
		BlockAds:        req.BlockAds,
		BlockMalware:    req.BlockMalware,
		CustomBlocklist: req.CustomBlocklist,
		LogQueries:      req.LogQueries,
		CacheTTL:        req.CacheTTL,
	}
	if err := h.DoH.UpdateConfig(c.Request().Context(), cfg); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"config": cfg})
}

// GetDoHStats returns DNS resolution statistics.
func (h *DoHHandlers) GetDoHStats(c echo.Context) error {
	stats, err := h.DoH.GetStats(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"stats": stats})
}

// GetDoHLogs returns recent DNS query logs.
func (h *DoHHandlers) GetDoHLogs(c echo.Context) error {
	logs, err := h.DoH.GetQueryLogs(c.Request().Context(), 100)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"logs": logs})
}
