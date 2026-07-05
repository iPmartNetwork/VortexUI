package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
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

type cleanIPThroughputRequest struct {
	ID   string `json:"id"`
	Port int    `json:"port"`
}

// isValidPort reports whether p is a usable TCP port. 0 is accepted as
// "unset" (the service defaults it to 443).
func isValidPort(p int) bool {
	return p == 0 || (p > 0 && p <= 65535)
}

// Throughput runs a real download-speed test against one previously scanned
// candidate and persists the measured Mbps. It is intentionally on-demand
// (per-IP) rather than part of the bulk Scan, since a real transfer takes far
// longer than a latency probe.
//
// The target IP is resolved server-side from the cached scan row (by ID), not
// taken from the request body, so a caller can't point the probe at an IP
// other than the one that ID's cached result actually refers to.
func (h *CleanIPHandlers) Throughput(c echo.Context) error {
	var req cleanIPThroughputRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	id, err := uuid.Parse(req.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if !isValidPort(req.Port) {
		return echo.NewHTTPError(http.StatusBadRequest, "port must be between 1 and 65535")
	}
	mbps, err := h.Scanner.MeasureThroughput(c.Request().Context(), id, req.Port)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"throughput_mbps": mbps})
}

type cleanIPThroughputAllRequest struct {
	Port int `json:"port"`
}

// ThroughputAll runs the download-speed test against every reachable cached
// candidate and returns the refreshed, best-first result set.
func (h *CleanIPHandlers) ThroughputAll(c echo.Context) error {
	var req cleanIPThroughputAllRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if !isValidPort(req.Port) {
		return echo.NewHTTPError(http.StatusBadRequest, "port must be between 1 and 65535")
	}
	results, err := h.Scanner.MeasureAllThroughput(c.Request().Context(), req.Port)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"results": results})
}

// GetSchedule returns the current recurring-scan configuration.
func (h *CleanIPHandlers) GetSchedule(c echo.Context) error {
	sched, err := h.Scanner.GetSchedule(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, sched)
}

type cleanIPScheduleRequest struct {
	Enabled         bool     `json:"enabled"`
	IntervalMinutes int      `json:"interval_minutes"`
	Port            int      `json:"port"`
	IPs             []string `json:"ips"`
}

// UpdateSchedule validates and persists the recurring-scan configuration
// (enabled flag, cadence, port, and candidate list). The service applies
// the same SSRF guard and candidate cap used by Scan.
func (h *CleanIPHandlers) UpdateSchedule(c echo.Context) error {
	var req cleanIPScheduleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Port != 0 && !isValidPort(req.Port) {
		return echo.NewHTTPError(http.StatusBadRequest, "port must be between 1 and 65535")
	}
	if len(req.IPs) > maxCleanIPScan {
		return echo.NewHTTPError(http.StatusBadRequest, "max 256 IPs per schedule")
	}
	sched := &domain.CleanIPSchedule{
		Enabled:         req.Enabled,
		IntervalMinutes: req.IntervalMinutes,
		Port:            req.Port,
		IPs:             req.IPs,
	}
	if err := h.Scanner.UpdateSchedule(c.Request().Context(), sched); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true})
}
