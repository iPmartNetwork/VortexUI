package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SessionHandlers handles session-related API endpoints
type SessionHandlers struct {
	sessionService *service.SessionService
}

// NewSessionHandlers creates new session handlers
func NewSessionHandlers(sessionService *service.SessionService) *SessionHandlers {
	return &SessionHandlers{
		sessionService: sessionService,
	}
}

// ListSessions lists all active sessions for the current admin
// @Route GET /api/sessions
func (h *SessionHandlers) ListSessions(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, &domain.ErrorResponse{
			Code:    domain.ErrCodeUnauthorized,
			Message: "Unauthorized",
		})
	}

	sessions, err := h.sessionService.ListAdminSessions(c.Request().Context(), adminID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to list sessions",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"count":    len(sessions),
	})
}

// RevokeSession revokes a specific session
// @Route POST /api/sessions/:id/revoke
func (h *SessionHandlers) RevokeSession(c echo.Context) error {
	sessionID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, &domain.ErrorResponse{
			Code:    domain.ErrCodeInvalidInput,
			Message: "Invalid session ID",
		})
	}

	if err := h.sessionService.RevokeSession(c.Request().Context(), sessionID); err != nil {
		if err == domain.ErrNotFound {
			return c.JSON(http.StatusNotFound, &domain.ErrorResponse{
				Code:    domain.ErrCodeNotFound,
				Message: "Session not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to revoke session",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Session revoked successfully",
	})
}

// RevokeAllSessions revokes all sessions for the current admin
// @Route POST /api/sessions/revoke-all
func (h *SessionHandlers) RevokeAllSessions(c echo.Context) error {
	adminID, ok := c.Get("admin_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, &domain.ErrorResponse{
			Code:    domain.ErrCodeUnauthorized,
			Message: "Unauthorized",
		})
	}

	if err := h.sessionService.RevokeAdminSessions(c.Request().Context(), adminID); err != nil {
		return c.JSON(http.StatusInternalServerError, &domain.ErrorResponse{
			Code:    domain.ErrCodeDatabaseError,
			Message: "Failed to revoke sessions",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "All sessions revoked successfully",
	})
}

// GetCurrentSession gets the current session info
// @Route GET /api/sessions/current
func (h *SessionHandlers) GetCurrentSession(c echo.Context) error {
	sessionID, ok := c.Get("session_id").(uuid.UUID)
	if !ok {
		return c.JSON(http.StatusUnauthorized, &domain.ErrorResponse{
			Code:    domain.ErrCodeUnauthorized,
			Message: "Unauthorized",
		})
	}

	adminID, _ := c.Get("admin_id").(uuid.UUID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"session_id": sessionID,
		"admin_id":   adminID,
	})
}

// RegisterSessionHandlers registers session routes
func RegisterSessionHandlers(e *echo.Echo, handler *SessionHandlers) {
	sessionGroup := e.Group("/api/sessions")
	sessionGroup.GET("", handler.ListSessions)
	sessionGroup.GET("/current", handler.GetCurrentSession)
	sessionGroup.POST("/:id/revoke", handler.RevokeSession)
	sessionGroup.POST("/revoke-all", handler.RevokeAllSessions)
}
