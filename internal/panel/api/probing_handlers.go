package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ProbingHandlers serves active probing protection endpoints.
type ProbingHandlers struct {
	Probing *service.ProbingService
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
	return c.JSON(http.StatusOK, echo.Map{"status": "unblocked"})
}
