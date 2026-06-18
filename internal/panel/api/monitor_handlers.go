package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// MonitorHandlers serves the live connection monitor.
type MonitorHandlers struct {
	Hub   *hub.Hub
	Nodes port.NodeRepository
}

type liveConnection struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	NodeName       string `json:"node_name"`
	IP             string `json:"ip"`
	Protocol       string `json:"protocol"`
	ConnectedSince string `json:"connected_since"`
}

// GetLiveConnections aggregates online stats from all managed nodes.
// It tries hub.OnlineStats first (detailed per-user data from core stats API).
// If that returns nothing, queries active users from the database.
func (h *MonitorHandlers) GetLiveConnections(c echo.Context) error {
	ctx := c.Request().Context()

	nodes, err := h.Nodes.List(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"connections": []any{}, "total": 0})
	}

	var connections []liveConnection

	for _, n := range nodes {
		// Try detailed stats from the core stats API.
		stats, err := h.Hub.OnlineStats(ctx, n.ID)
		if err == nil && len(stats) > 0 {
			for userID, count := range stats {
				for i := 0; i < count; i++ {
					connections = append(connections, liveConnection{
						UserID:   userID,
						Username: userID,
						NodeName: n.Name,
						Protocol: string(n.Core),
					})
				}
			}
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"connections": connections,
		"total":       len(connections),
	})
}
