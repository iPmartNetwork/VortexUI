package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// DeviceHandler serves user device (HWID) management endpoints.
type DeviceHandler struct {
	svc *service.DeviceService
}

// NewDeviceHandler creates a new DeviceHandler with the given service dependency.
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// Register mounts all device routes on the given Echo group.
func (h *DeviceHandler) Register(g *echo.Group) {
	g.GET("/users/:id/devices", h.ListDevices)
	g.DELETE("/users/:id/devices/:hwid", h.RevokeDevice)
	g.POST("/users/bulk-hwid-reset", h.BulkResetHWIDs)
}

// ListDevices handles GET /api/v2/users/:id/devices.
// Returns all registered devices for the specified user.
func (h *DeviceHandler) ListDevices(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	devices, err := h.svc.ListDevices(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list devices")
	}
	if devices == nil {
		devices = []*domain.Device{}
	}
	return c.JSON(http.StatusOK, echo.Map{"devices": devices})
}

// RevokeDevice handles DELETE /api/v2/users/:id/devices/:hwid.
// Removes a specific HWID registration for the user.
func (h *DeviceHandler) RevokeDevice(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id")
	}

	hwid := c.Param("hwid")
	if hwid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "hwid is required")
	}

	if err := h.svc.RevokeDevice(c.Request().Context(), userID, hwid); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to revoke device")
	}
	return c.NoContent(http.StatusNoContent)
}

type bulkHWIDResetRequest struct {
	UserIDs []string `json:"user_ids"`
}

// BulkResetHWIDs handles POST /api/v2/users/bulk-hwid-reset.
// Removes all registered HWIDs for the specified users.
func (h *DeviceHandler) BulkResetHWIDs(c echo.Context) error {
	var req bulkHWIDResetRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if len(req.UserIDs) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "user_ids is required")
	}

	userIDs := make([]uuid.UUID, 0, len(req.UserIDs))
	for _, idStr := range req.UserIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid user id: "+idStr)
		}
		userIDs = append(userIDs, id)
	}

	if err := h.svc.BulkResetHWIDs(c.Request().Context(), userIDs); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to reset devices")
	}
	return c.JSON(http.StatusOK, echo.Map{"reset_count": len(userIDs)})
}
