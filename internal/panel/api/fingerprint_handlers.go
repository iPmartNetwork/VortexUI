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
