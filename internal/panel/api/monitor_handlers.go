package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// MonitorLiveUser is a traffic-derived active user.
type MonitorLiveUser struct {
	UserID   string
	Username string
	NodeID   string
	LastSeen time.Time
}

// MonitorSource yields recently-active users from traffic samples.
type MonitorSource interface {
	RecentActive(ctx context.Context, window time.Duration) ([]MonitorLiveUser, error)
}

// MonitorHandlers serves the live connection monitor.
type MonitorHandlers struct {
	Hub     *hub.Hub
	Nodes   port.NodeRepository
	Users   port.UserRepository
	Monitor MonitorSource // optional traffic-derived fallback
}

type liveConnection struct {
	UserID         string `json:"user_id"`
	Username       string `json:"username"`
	NodeName       string `json:"node_name"`
	IP             string `json:"ip"`
	Protocol       string `json:"protocol"`
	ConnectedSince string `json:"connected_since"`
}

func (h *MonitorHandlers) ownedUserSet(ctx context.Context, adminID uuid.UUID) (map[string]struct{}, error) {
	if h.Users == nil {
		return nil, nil
	}
	users, _, err := h.Users.List(ctx, port.UserFilter{AdminID: &adminID, Limit: 100_000})
	if err != nil {
		return nil, err
	}
	set := make(map[string]struct{}, len(users))
	for _, u := range users {
		set[u.ID.String()] = struct{}{}
		set[u.Username] = struct{}{}
	}
	return set, nil
}

func connectionOwned(set map[string]struct{}, userID, username string) bool {
	if set == nil {
		return true
	}
	if userID != "" {
		if _, ok := set[userID]; ok {
			return true
		}
	}
	if username != "" {
		if _, ok := set[username]; ok {
			return true
		}
	}
	return false
}

// GetLiveConnections aggregates online users. It tries the core stats API first
// (precise, per-connection) and falls back to recent traffic samples (works on
// any core version). Non-sudo admins only see their own users.
func (h *MonitorHandlers) GetLiveConnections(c echo.Context) error {
	ctx := c.Request().Context()

	var owned map[string]struct{}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		var err error
		owned, err = h.ownedUserSet(ctx, claims.AdminID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
		}
	}

	nodes, _ := h.Nodes.List(ctx)
	nodeNames := map[string]string{}
	for _, n := range nodes {
		nodeNames[n.ID.String()] = n.Name
	}

	var connections []liveConnection

	for _, n := range nodes {
		stats, err := h.Hub.OnlineStats(ctx, n.ID)
		if err == nil && len(stats) > 0 {
			for userID, count := range stats {
				if !connectionOwned(owned, userID, userID) {
					continue
				}
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

	if len(connections) == 0 && h.Monitor != nil {
		active, err := h.Monitor.RecentActive(ctx, 3*time.Minute)
		if err == nil {
			for _, a := range active {
				name := a.Username
				if name == "" {
					name = a.UserID
				}
				if !connectionOwned(owned, a.UserID, name) {
					continue
				}
				connections = append(connections, liveConnection{
					UserID:         a.UserID,
					Username:       name,
					NodeName:       nodeNames[a.NodeID],
					ConnectedSince: a.LastSeen.Format(time.RFC3339),
				})
			}
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"connections": connections,
		"total":       len(connections),
	})
}
