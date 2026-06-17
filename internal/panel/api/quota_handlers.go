package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// QuotaHandlers serves fair-use quota policy endpoints.
type QuotaHandlers struct {
	Quota *service.QuotaService
}

type createQuotaPolicyRequest struct {
	Name  string             `json:"name"`
	Tiers []domain.QuotaTier `json:"tiers"`
}

// CreateQuotaPolicy creates a new fair-use policy.
func (h *QuotaHandlers) CreateQuotaPolicy(c echo.Context) error {
	var req createQuotaPolicyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.Quota.CreatePolicy(c.Request().Context(), service.CreatePolicyInput{
		Name:  req.Name,
		Tiers: req.Tiers,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"policy": p})
}

// ListQuotaPolicies lists all quota policies.
func (h *QuotaHandlers) ListQuotaPolicies(c echo.Context) error {
	policies, err := h.Quota.ListPolicies(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"policies": policies})
}

type updateQuotaPolicyRequest struct {
	Name    string             `json:"name"`
	Tiers   []domain.QuotaTier `json:"tiers"`
	Enabled bool               `json:"enabled"`
}

// UpdateQuotaPolicy updates an existing quota policy.
func (h *QuotaHandlers) UpdateQuotaPolicy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateQuotaPolicyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p, err := h.Quota.UpdatePolicy(c.Request().Context(), id, req.Name, req.Tiers, req.Enabled)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"policy": p})
}

// DeleteQuotaPolicy removes a quota policy.
func (h *QuotaHandlers) DeleteQuotaPolicy(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Quota.DeletePolicy(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
