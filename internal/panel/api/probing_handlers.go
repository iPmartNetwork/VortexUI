package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ProbingHandlers serves active probing protection endpoints.
type ProbingHandlers struct {
	Probing *service.ProbingService
	Resync  *service.FleetResync
}

// GetProbingPolicy returns the current policy.
func (h *ProbingHandlers) GetProbingPolicy(c echo.Context) error {
	p, err := h.Probing.GetPolicy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

type updateProbingPolicyRequest struct {
	Enabled        bool     `json:"enabled"`
	Action         string   `json:"action"`
	BlockDuration  int      `json:"block_duration"`
	MaxProbePerMin int      `json:"max_probe_per_min"`
	WhitelistedIPs []string `json:"whitelisted_ips"`
	HoneypotHTML   string   `json:"honeypot_html"`
	NotifyTelegram bool     `json:"notify_telegram"`
}

// UpdateProbingPolicy saves the probing policy.
func (h *ProbingHandlers) UpdateProbingPolicy(c echo.Context) error {
	var req updateProbingPolicyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p := &domain.ProbingPolicy{
		Enabled:        req.Enabled,
		Action:         domain.ProbingAction(req.Action),
		BlockDuration:  req.BlockDuration,
		MaxProbePerMin: req.MaxProbePerMin,
		WhitelistedIPs: req.WhitelistedIPs,
		HoneypotHTML:   req.HoneypotHTML,
		NotifyTelegram: req.NotifyTelegram,
	}
	if err := h.Probing.UpdatePolicy(c.Request().Context(), p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if h.Resync != nil {
		_ = h.Resync.All(c.Request().Context())
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

// ListProbeEvents returns recent probing events.
func (h *ProbingHandlers) ListProbeEvents(c echo.Context) error {
	events, total, err := h.Probing.ListEvents(c.Request().Context(), 50, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"events": events, "total": total})
}

// ListBlockedIPs returns currently blocked IPs.
func (h *ProbingHandlers) ListBlockedIPs(c echo.Context) error {
	ips, err := h.Probing.ListBlockedIPs(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"blocked_ips": ips})
}

type unblockRequest struct {
	IP string `json:"ip"`
}

// UnblockIP removes an IP from the blocklist.
func (h *ProbingHandlers) UnblockIP(c echo.Context) error {
	var req unblockRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Probing.UnblockIP(c.Request().Context(), req.IP); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if h.Resync != nil {
		_ = h.Resync.All(c.Request().Context())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "unblocked"})
}

type probeReportRequest struct {
	SourceIP    string `json:"source_ip"`
	Port        int    `json:"port"`
	Method      string `json:"method"`
	Fingerprint string `json:"fingerprint"`
	NodeID      string `json:"node_id"`
}

// ReportProbe records a probing attempt from a node or edge reporter.
func (h *ProbingHandlers) ReportProbe(c echo.Context) error {
	var req probeReportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.SourceIP == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "source_ip is required")
	}
	var nodeID *uuid.UUID
	if req.NodeID != "" {
		id, err := uuid.Parse(req.NodeID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
		}
		nodeID = &id
	}
	if err := h.Probing.DetectProbe(c.Request().Context(), req.SourceIP, req.Port, req.Method, req.Fingerprint, nodeID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "recorded"})
}
