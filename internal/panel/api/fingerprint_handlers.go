package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

type FingerprintHandlers struct {
	FP *service.FingerprintService
}

type fingerprintReportRequest struct {
	ClientIP    string `json:"client_ip"`
	Fingerprint string `json:"fingerprint"`
	JA3Hash     string `json:"ja3_hash"`
	UserAgent   string `json:"user_agent"`
	NodeID      string `json:"node_id"`
}

func (h *FingerprintHandlers) Report(c echo.Context) error {
	var req fingerprintReportRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.ClientIP == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "client_ip is required")
	}
	var nodeID *uuid.UUID
	if req.NodeID != "" {
		id, err := uuid.Parse(req.NodeID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
		}
		nodeID = &id
	}
	action, err := h.FP.Report(c.Request().Context(), service.ReportInput{
		ClientIP:    req.ClientIP,
		Fingerprint: req.Fingerprint,
		JA3Hash:     req.JA3Hash,
		UserAgent:   req.UserAgent,
		NodeID:      nodeID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"action": action})
}

func (h *FingerprintHandlers) GetPolicy(c echo.Context) error {
	p, _ := h.FP.GetPolicy(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

func (h *FingerprintHandlers) UpdatePolicy(c echo.Context) error {
	var p domain.FingerprintPolicy
	if err := c.Bind(&p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.FP.UpdatePolicy(c.Request().Context(), &p); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

type createFPRuleRequest struct {
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	JA3Hash     string `json:"ja3_hash"`
	Action      string `json:"action"`
	Priority    int    `json:"priority"`
}

func (h *FingerprintHandlers) CreateRule(c echo.Context) error {
	var req createFPRuleRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	r, err := h.FP.CreateRule(c.Request().Context(), req.Name, req.Fingerprint, req.JA3Hash, req.Action, req.Priority)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"rule": r})
}

func (h *FingerprintHandlers) ListRules(c echo.Context) error {
	rules, _ := h.FP.ListRules(c.Request().Context())
	return c.JSON(http.StatusOK, echo.Map{"rules": rules})
}

func (h *FingerprintHandlers) DeleteRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	_ = h.FP.DeleteRule(c.Request().Context(), id)
	return c.NoContent(http.StatusNoContent)
}

func (h *FingerprintHandlers) ListEvents(c echo.Context) error {
	events, _ := h.FP.ListEvents(c.Request().Context(), 50)
	return c.JSON(http.StatusOK, echo.Map{"events": events})
}
