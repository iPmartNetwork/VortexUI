package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// NotificationHandler serves notification channel management endpoints.
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler with the given service dependency.
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// Register mounts all notification routes on the given Echo group.
func (h *NotificationHandler) Register(g *echo.Group) {
	notif := g.Group("/notifications")
	notif.POST("/channels", h.CreateChannel)
	notif.GET("/channels", h.ListChannels)
	notif.PUT("/channels/:id", h.UpdateChannel)
	notif.DELETE("/channels/:id", h.DeleteChannel)
	notif.POST("/test", h.TestNotification)
}

// createChannelRequest is the request body for channel creation.
type createChannelRequest struct {
	Name      string         `json:"name"`
	Type      string         `json:"type"`
	Config    map[string]any `json:"config"`
	ScopeType string         `json:"scope_type"`
	ScopeID   string         `json:"scope_id,omitempty"`
	Events    []string       `json:"events"`
	Template  string         `json:"template,omitempty"`
	Enabled   bool           `json:"enabled"`
}

// CreateChannel handles POST /api/v2/notifications/channels.
func (h *NotificationHandler) CreateChannel(c echo.Context) error {
	var req createChannelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "name is required")
	}
	if req.Type != "telegram" && req.Type != "webhook" {
		return echo.NewHTTPError(http.StatusBadRequest, "type must be 'telegram' or 'webhook'")
	}

	ch := &domain.NotificationChannel{
		ID:        uuid.New(),
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		ScopeType: req.ScopeType,
		ScopeID:   req.ScopeID,
		Events:    req.Events,
		Template:  req.Template,
		Enabled:   req.Enabled,
	}

	if err := h.svc.CreateChannel(c.Request().Context(), ch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"channel": ch})
}

// ListChannels handles GET /api/v2/notifications/channels.
func (h *NotificationHandler) ListChannels(c echo.Context) error {
	channels, err := h.svc.ListChannels(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if channels == nil {
		channels = []*domain.NotificationChannel{}
	}
	return c.JSON(http.StatusOK, echo.Map{"channels": channels})
}

// UpdateChannel handles PUT /api/v2/notifications/channels/:id.
func (h *NotificationHandler) UpdateChannel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}

	var req createChannelRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}

	ch := &domain.NotificationChannel{
		ID:        id,
		Name:      req.Name,
		Type:      req.Type,
		Config:    req.Config,
		ScopeType: req.ScopeType,
		ScopeID:   req.ScopeID,
		Events:    req.Events,
		Template:  req.Template,
		Enabled:   req.Enabled,
	}

	if err := h.svc.UpdateChannel(c.Request().Context(), ch); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"channel": ch})
}

// DeleteChannel handles DELETE /api/v2/notifications/channels/:id.
func (h *NotificationHandler) DeleteChannel(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.svc.DeleteChannel(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// testNotificationRequest is the request body for sending a test notification.
type testNotificationRequest struct {
	ChannelID string `json:"channel_id"`
	Message   string `json:"message"`
}

// TestNotification handles POST /api/v2/notifications/test.
func (h *NotificationHandler) TestNotification(c echo.Context) error {
	var req testNotificationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.ChannelID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "channel_id is required")
	}

	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid channel_id")
	}

	if err := h.svc.SendTestNotification(c.Request().Context(), channelID, req.Message); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"success": true, "message": "Test notification sent"})
}
