package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/service"
)

// PortalProHandler serves portal user-facing endpoints.
type PortalProHandler struct {
	svc *service.PortalProService
}

// NewPortalProHandler creates the handler.
func NewPortalProHandler(svc *service.PortalProService) *PortalProHandler {
	return &PortalProHandler{svc: svc}
}

// Register mounts all portal routes on the given Echo group.
func (h *PortalProHandler) Register(g *echo.Group) {
	portal := g.Group("/portal")
	portal.POST("/push/subscribe", h.SubscribePush)
	portal.POST("/speed-test", h.SpeedTest)
	portal.GET("/guides", h.ListGuides)
	portal.GET("/setup-wizard", h.GetSetupWizard)
}

// --- request types ---

type pushSubscribeRequest struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

type speedTestRequest struct {
	NodeID       string `json:"node_id"`
	NodeEndpoint string `json:"node_endpoint"`
}

// --- handlers ---

// SubscribePush handles POST /api/v2/portal/push/subscribe.
func (h *PortalProHandler) SubscribePush(c echo.Context) error {
	var req pushSubscribeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Endpoint == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "endpoint is required")
	}

	userID := getUserIDFromContext(c)
	if userID == uuid.Nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
	}

	sub, err := h.svc.SubscribePush(c.Request().Context(), userID, req.Endpoint, req.P256dh, req.Auth)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, sub)
}

// SpeedTest handles POST /api/v2/portal/speed-test.
func (h *PortalProHandler) SpeedTest(c echo.Context) error {
	var req speedTestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	if req.NodeEndpoint == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "node_endpoint is required")
	}

	result, err := h.svc.SpeedTest(c.Request().Context(), nodeID, req.NodeEndpoint)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

// ListGuides handles GET /api/v2/portal/guides.
func (h *PortalProHandler) ListGuides(c echo.Context) error {
	guides, err := h.svc.ListGuides(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, guides)
}

// GetSetupWizard handles GET /api/v2/portal/setup-wizard.
func (h *PortalProHandler) GetSetupWizard(c echo.Context) error {
	// In production these flags would come from the user's actual state.
	hasSub := c.QueryParam("has_subscription") == "true"
	hasDevice := c.QueryParam("has_device") == "true"

	steps := h.svc.GetSetupWizard(c.Request().Context(), hasSub, hasDevice)
	return c.JSON(http.StatusOK, steps)
}

// --- helpers ---

func getUserIDFromContext(c echo.Context) uuid.UUID {
	if val := c.Get("user_id"); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}
