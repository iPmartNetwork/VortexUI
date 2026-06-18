package api

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
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
// It tries hub.OnlineStats first (detailed per-user data from gRPC agents).
// If that returns nothing, falls back to node health.connections count.
func (h *MonitorHandlers) GetLiveConnections(c echo.Context) error {
	ctx := c.Request().Context()

	nodes, err := h.Nodes.List(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"connections": []any{}, "total": 0})
	}

	var connections []liveConnection
	var totalFromHealth int

	for _, n := range nodes {
		// Try detailed stats from the agent.
		stats, err := h.Hub.OnlineStats(context.Background(), n.ID)
		if err == nil && len(stats) > 0 {
			for userID, count := range stats {
				for i := 0; i < count; i++ {
					connections = append(connections, liveConnection{
						UserID:   userID,
						Username: userID,
						NodeName: n.Name,
					})
				}
			}
			continue
		}

		// Fallback: use the node's health connections count from the hub.
		status, health, err := h.Hub.Status(n.ID)
		if err == nil && status == domain.NodeConnected && health.Connections > 0 {
			totalFromHealth += health.Connections
			// We don't have per-user detail, just show aggregate.
			connections = append(connections, liveConnection{
				UserID:   "",
				Username: "(aggregated)",
				NodeName: n.Name,
				Protocol: string(n.Core),
			})
		}
	}

	total := len(connections)
	if totalFromHealth > 0 && total == 0 {
		total = totalFromHealth
	}
	// If we only have aggregated data, set total to health sum.
	if totalFromHealth > 0 {
		total = totalFromHealth
	}

	return c.JSON(http.StatusOK, echo.Map{
		"connections": connections,
		"total":       total,
	})
}
