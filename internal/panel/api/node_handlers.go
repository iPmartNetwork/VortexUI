package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// --- nodes ---

type createNodeRequest struct {
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	Core       string  `json:"core"`
	UsageRatio float64 `json:"usage_ratio"`
	Endpoint   string  `json:"endpoint"`
}

// CreateNode registers a new node and brings it under live management.
func (h *Handlers) CreateNode(c echo.Context) error {
	var req createNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	n, err := h.Nodes.Create(c.Request().Context(), service.CreateNodeInput{
		Name: req.Name, Address: req.Address, Core: domain.CoreType(req.Core), UsageRatio: req.UsageRatio, Endpoint: req.Endpoint,
	})
	if n == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	resp := echo.Map{"node": n}
	if err != nil {
		resp["warning"] = err.Error() // saved but not yet managed
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListNodes returns all nodes.
func (h *Handlers) ListNodes(c echo.Context) error {
	nodes, err := h.Nodes.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"nodes": nodes})
}

// GetNode returns one node.
func (h *Handlers) GetNode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	n, err := h.Nodes.Get(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	return c.JSON(http.StatusOK, n)
}

type updateNodeRequest struct {
	Name       string  `json:"name"`
	Address    string  `json:"address"`
	UsageRatio float64 `json:"usage_ratio"`
	Endpoint   string  `json:"endpoint"`
}

// UpdateNode edits a node and re-establishes its hub connection.
func (h *Handlers) UpdateNode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	n, err := h.Nodes.Update(c.Request().Context(), id, service.UpdateNodeInput{
		Name: req.Name, Address: req.Address, UsageRatio: req.UsageRatio, Endpoint: req.Endpoint,
	})
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	if n == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update failed")
	}
	resp := echo.Map{"node": n}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteNode deregisters and removes a node.
func (h *Handlers) DeleteNode(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Nodes.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// GetNodeLogs returns recent core log lines from a node (?limit=). Fails with
// 502 when the node is unreachable or its agent cannot provide logs.
func (h *Handlers) GetNodeLogs(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	limit := atoiDefault(c.QueryParam("limit"), 200)
	lines, err := h.Nodes.Logs(c.Request().Context(), id, limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "node logs unavailable: "+err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"lines": lines})
}

// GetNodeStatus returns a node's real-time system snapshot: version, IP,
// uptime-like data, connections — everything the UI "Manage" panel needs.
func (h *Handlers) GetNodeStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	n, err := h.Nodes.Get(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "node not found")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"id":            n.ID,
		"name":          n.Name,
		"address":       n.Address,
		"core":          n.Core,
		"core_version":  n.CoreVer,
		"agent_version": n.AgentVer,
		"status":        n.Status,
		"last_seen":     n.LastSeen,
		"health":        n.Health,
	})
}

// RestartNodeCore restarts a node's proxy engine.
func (h *Handlers) RestartNodeCore(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Nodes.RestartCore(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "restart failed: "+err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "message": "core restarted"})
}

type updateGeoRequest struct {
	GeoipURL   string `json:"geoip_url"`   // empty = Iran-rules default
	GeositeURL string `json:"geosite_url"` // empty = Iran-rules default
}

// UpdateNodeGeo refreshes a node's geoip/geosite routing databases. With no body
// it uses the Iran-focused defaults (geoip:ir / geosite:ir and friends).
func (h *Handlers) UpdateNodeGeo(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateGeoRequest
	_ = c.Bind(&req) // body is optional; empty falls back to defaults
	geoip, geosite, err := h.Nodes.UpdateGeo(c.Request().Context(), id, req.GeoipURL, req.GeositeURL)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "geo update failed: "+err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "geoip_bytes": geoip, "geosite_bytes": geosite})
}

// StopNodeCore stops a node's proxy engine.
func (h *Handlers) StopNodeCore(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Nodes.StopCore(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "stop failed: "+err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ok": true, "message": "core stopped"})
}

// --- inbounds ---

type createInboundRequest struct {
	NodeID     string            `json:"node_id"`
	Tag        string            `json:"tag"`
	Protocol   string            `json:"protocol"`
	Listen     string            `json:"listen"`
	Port       int               `json:"port"`
	Network    string            `json:"network"`
	Security   string            `json:"security"`
	SNI        []string          `json:"sni"`
	Path       string            `json:"path"`
	Host       []string          `json:"host"`
	Flow       string            `json:"flow"`
	Raw        map[string]any    `json:"raw"`
	Enabled    bool              `json:"enabled"`
	SpeedLimit int64             `json:"speed_limit"`
	GeoPolicy  *domain.GeoPolicy `json:"geo_policy"`
}

// CreateInbound adds an inbound to a node and resyncs the node's core.
func (h *Handlers) CreateInbound(c echo.Context) error {
	var req createInboundRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	in, err := h.Inbounds.Create(c.Request().Context(), service.CreateInboundInput{
		NodeID: nodeID, Tag: req.Tag, Protocol: domain.Protocol(req.Protocol),
		Listen: req.Listen, Port: req.Port, Network: req.Network,
		Security: domain.Security(req.Security), SNI: req.SNI, Path: req.Path,
		Host: req.Host, Flow: req.Flow, Raw: req.Raw, Enabled: req.Enabled,
		SpeedLimit: req.SpeedLimit, GeoPolicy: req.GeoPolicy,
	})
	if in == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	resp := echo.Map{"inbound": in}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListInbounds returns the inbounds for a node (?node_id=...). Resellers only
// see inbounds on their admin allowlist.
func (h *Handlers) ListInbounds(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id query param required")
	}
	ins, err := h.Inbounds.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		allowed, aerr := h.Admins.InboundIDsForAdmin(c.Request().Context(), claims.AdminID)
		if aerr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
		}
		allow := make(map[uuid.UUID]struct{}, len(allowed))
		for _, id := range allowed {
			allow[id] = struct{}{}
		}
		filtered := ins[:0]
		for _, in := range ins {
			if _, ok := allow[in.ID]; ok {
				filtered = append(filtered, in)
			}
		}
		ins = filtered
	}
	return c.JSON(http.StatusOK, echo.Map{"inbounds": ins})
}

type updateInboundRequest struct {
	Listen     string            `json:"listen"`
	Port       int               `json:"port"`
	Network    string            `json:"network"`
	Security   string            `json:"security"`
	SNI        []string          `json:"sni"`
	Path       string            `json:"path"`
	Host       []string          `json:"host"`
	Flow       string            `json:"flow"`
	Raw        map[string]any    `json:"raw"`
	Enabled    bool              `json:"enabled"`
	SpeedLimit int64             `json:"speed_limit"`
	GeoPolicy  *domain.GeoPolicy `json:"geo_policy"`
}

// UpdateInbound edits an inbound and resyncs its node.
func (h *Handlers) UpdateInbound(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateInboundRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	in, err := h.Inbounds.Update(c.Request().Context(), id, service.UpdateInboundInput{
		Listen: req.Listen, Port: req.Port, Network: req.Network,
		Security: domain.Security(req.Security), SNI: req.SNI, Path: req.Path,
		Host: req.Host, Flow: req.Flow, Raw: req.Raw, Enabled: req.Enabled,
		SpeedLimit: req.SpeedLimit, GeoPolicy: req.GeoPolicy,
	})
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "inbound not found")
	}
	if in == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update failed")
	}
	resp := echo.Map{"inbound": in}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteInbound removes an inbound and resyncs its node.
func (h *Handlers) DeleteInbound(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Inbounds.Delete(c.Request().Context(), id); errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "inbound not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}
