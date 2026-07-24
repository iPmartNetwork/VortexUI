package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// TemplateHandler serves user template management endpoints.
type TemplateHandler struct {
	svc *service.TemplateService
}

// NewTemplateHandler creates a new TemplateHandler with the given service dependency.
func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

// Register mounts all template routes on the given Echo group.
func (h *TemplateHandler) Register(g *echo.Group) {
	templates := g.Group("/templates")
	templates.POST("", h.CreateTemplate)
	templates.GET("", h.ListTemplates)
	templates.GET("/:id", h.GetTemplate)
	templates.PUT("/:id", h.UpdateTemplate)
	templates.DELETE("/:id", h.DeleteTemplate)
	templates.POST("/:id/bulk-create", h.BulkCreate)

	// Clone lives under /users/:id/clone
	g.POST("/users/:id/clone", h.CloneUser)
}

// CreateTemplate handles POST /api/v2/templates.
func (h *TemplateHandler) CreateTemplate(c echo.Context) error {
	var t domain.UserTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.svc.Create(c.Request().Context(), &t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"template": t})
}

// ListTemplates handles GET /api/v2/templates.
// Filters by admin RBAC: non-sudo admins only see templates they are allowed.
func (h *TemplateHandler) ListTemplates(c echo.Context) error {
	claims := claimsFrom(c)

	var (
		templates []*domain.UserTemplate
		err       error
	)
	if claims != nil && !claims.Sudo {
		templates, err = h.svc.ListForAdmin(c.Request().Context(), claims.AdminID)
	} else {
		templates, err = h.svc.List(c.Request().Context())
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if templates == nil {
		templates = []*domain.UserTemplate{}
	}
	return c.JSON(http.StatusOK, echo.Map{"templates": templates})
}

// GetTemplate handles GET /api/v2/templates/:id.
func (h *TemplateHandler) GetTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	t, err := h.svc.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"template": t})
}

// UpdateTemplate handles PUT /api/v2/templates/:id.
func (h *TemplateHandler) UpdateTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var t domain.UserTemplate
	if err := c.Bind(&t); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	t.ID = id
	if err := h.svc.Update(c.Request().Context(), &t); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"template": t})
}

// DeleteTemplate handles DELETE /api/v2/templates/:id.
func (h *TemplateHandler) DeleteTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

type bulkCreateFromTemplateRequest struct {
	Count int `json:"count"`
}

// BulkCreate handles POST /api/v2/templates/:id/bulk-create.
func (h *TemplateHandler) BulkCreate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req bulkCreateFromTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Count < 1 || req.Count > 1000 {
		return echo.NewHTTPError(http.StatusBadRequest, "count must be between 1 and 1000")
	}

	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	users, err := h.svc.BulkCreate(c.Request().Context(), id, req.Count, claims.AdminID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "template not found")
		}
		// "admin is not allowed to use this template" maps to 403.
		if err.Error() == "admin is not allowed to use this template" {
			return echo.NewHTTPError(http.StatusForbidden, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"users": users})
}

// CloneUser handles POST /api/v2/users/:id/clone.
func (h *TemplateHandler) CloneUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	user, err := h.svc.CloneUser(c.Request().Context(), id, claims.AdminID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"user": user})
}
