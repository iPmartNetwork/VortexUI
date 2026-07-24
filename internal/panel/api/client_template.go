package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ClientTemplateHandler serves client template CRUD and approval queue endpoints.
type ClientTemplateHandler struct {
	svc *service.ClientTemplateService
}

// NewClientTemplateHandler creates a new ClientTemplateHandler with the given service dependency.
func NewClientTemplateHandler(svc *service.ClientTemplateService) *ClientTemplateHandler {
	return &ClientTemplateHandler{svc: svc}
}

// Register mounts all client template and approval routes on the given Echo group.
func (h *ClientTemplateHandler) Register(g *echo.Group) {
	templates := g.Group("/client-templates")
	templates.POST("", h.CreateTemplate)
	templates.GET("", h.ListTemplates)
	templates.GET("/:id", h.GetTemplate)
	templates.PUT("/:id", h.UpdateTemplate)
	templates.DELETE("/:id", h.DeleteTemplate)

	approvals := g.Group("/approvals")
	approvals.GET("", h.ListPendingApprovals)
	approvals.POST("/:id/approve", h.ApproveRequest)
	approvals.POST("/:id/reject", h.RejectRequest)

	g.POST("/templates/preview", h.PreviewTemplate)
}

// --- request/response types ---

type createTemplateRequest struct {
	Name           string         `json:"name"`
	ClientPattern  string         `json:"client_pattern"`
	RoutingRules   []any          `json:"routing_rules"`
	DNSSettings    map[string]any `json:"dns_settings"`
	CustomOutbounds []any         `json:"custom_outbounds"`
	Priority       int            `json:"priority"`
	Enabled        bool           `json:"enabled"`
}

type previewRequest struct {
	Template     string `json:"template"`
	SampleUserID string `json:"sample_user_id"`
}

// --- handlers ---

// CreateTemplate handles POST /api/v2/client-templates.
func (h *ClientTemplateHandler) CreateTemplate(c echo.Context) error {
	var req createTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.ClientPattern == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "client_pattern is required")
	}

	t := &domain.ClientTemplate{
		ID:              uuid.New(),
		Name:            req.Name,
		ClientPattern:   req.ClientPattern,
		RoutingRules:    req.RoutingRules,
		DNSSettings:     req.DNSSettings,
		CustomOutbounds: req.CustomOutbounds,
		Priority:        req.Priority,
		Enabled:         req.Enabled,
	}

	if t.RoutingRules == nil {
		t.RoutingRules = []any{}
	}
	if t.DNSSettings == nil {
		t.DNSSettings = map[string]any{}
	}
	if t.CustomOutbounds == nil {
		t.CustomOutbounds = []any{}
	}

	if err := h.svc.CreateTemplate(c.Request().Context(), t); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"template": t})
}

// ListTemplates handles GET /api/v2/client-templates.
func (h *ClientTemplateHandler) ListTemplates(c echo.Context) error {
	templates, err := h.svc.ListTemplates(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if templates == nil {
		templates = []*domain.ClientTemplate{}
	}
	return c.JSON(http.StatusOK, echo.Map{"templates": templates})
}

// GetTemplate handles GET /api/v2/client-templates/:id.
func (h *ClientTemplateHandler) GetTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	t, err := h.svc.GetTemplate(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	}
	return c.JSON(http.StatusOK, echo.Map{"template": t})
}

// UpdateTemplate handles PUT /api/v2/client-templates/:id.
func (h *ClientTemplateHandler) UpdateTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req createTemplateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	t := &domain.ClientTemplate{
		ID:              id,
		Name:            req.Name,
		ClientPattern:   req.ClientPattern,
		RoutingRules:    req.RoutingRules,
		DNSSettings:     req.DNSSettings,
		CustomOutbounds: req.CustomOutbounds,
		Priority:        req.Priority,
		Enabled:         req.Enabled,
	}

	if t.RoutingRules == nil {
		t.RoutingRules = []any{}
	}
	if t.DNSSettings == nil {
		t.DNSSettings = map[string]any{}
	}
	if t.CustomOutbounds == nil {
		t.CustomOutbounds = []any{}
	}

	if err := h.svc.UpdateTemplate(c.Request().Context(), t); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"template": t})
}

// DeleteTemplate handles DELETE /api/v2/client-templates/:id.
func (h *ClientTemplateHandler) DeleteTemplate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteTemplate(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// ListPendingApprovals handles GET /api/v2/approvals.
func (h *ClientTemplateHandler) ListPendingApprovals(c echo.Context) error {
	approvals, err := h.svc.ListPendingApprovals(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if approvals == nil {
		approvals = []*domain.SubscriptionApproval{}
	}
	return c.JSON(http.StatusOK, echo.Map{"approvals": approvals})
}

// ApproveRequest handles POST /api/v2/approvals/:id/approve.
func (h *ClientTemplateHandler) ApproveRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	// Extract admin ID from the authenticated context.
	adminID, err := extractAdminID(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "admin context required")
	}

	if err := h.svc.Approve(c.Request().Context(), id, adminID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"success": true, "message": "approved"})
}

// RejectRequest handles POST /api/v2/approvals/:id/reject.
func (h *ClientTemplateHandler) RejectRequest(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	if err := h.svc.Reject(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"success": true, "message": "rejected"})
}

// PreviewTemplate handles POST /api/v2/templates/preview.
func (h *ClientTemplateHandler) PreviewTemplate(c echo.Context) error {
	var req previewRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Template == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "template is required")
	}

	sampleID := uuid.New()
	if req.SampleUserID != "" {
		if parsed, err := uuid.Parse(req.SampleUserID); err == nil {
			sampleID = parsed
		}
	}

	result, err := h.svc.PreviewTemplate(req.Template, sampleID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"preview": result})
}
