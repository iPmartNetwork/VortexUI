package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// --- nodes ---

type createNodeRequest struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Core         string   `json:"core"`
	EnabledCores []string `json:"enabled_cores"`
	UsageRatio   float64  `json:"usage_ratio"`
	Endpoint     string  `json:"endpoint"`
	Region       string  `json:"region"`
	CountryCode  string  `json:"country_code"`
	LocationAuto *bool   `json:"location_auto"`
}

// CreateNode registers a new node and brings it under live management.
func (h *Handlers) CreateNode(c echo.Context) error {
	var req createNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	n, err := h.Nodes.Create(c.Request().Context(), service.CreateNodeInput{
		Name: req.Name, Address: req.Address, Core: domain.CoreType(req.Core),
		EnabledCores: coreTypesFromStrings(req.EnabledCores),
		UsageRatio: req.UsageRatio, Endpoint: req.Endpoint,
		Region: req.Region, CountryCode: req.CountryCode, LocationAuto: req.LocationAuto,
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

// ListNodes returns all nodes. Resellers only see nodes on their allowlist.
func (h *Handlers) ListNodes(c echo.Context) error {
	nodes, err := h.Nodes.List(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo && h.Admins != nil {
		allowed, aerr := h.Admins.NodeIDsForAdmin(c.Request().Context(), claims.AdminID)
		if aerr != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
		}
		allow := make(map[uuid.UUID]struct{}, len(allowed))
		for _, id := range allowed {
			allow[id] = struct{}{}
		}
		filtered := nodes[:0]
		for _, n := range nodes {
			if _, ok := allow[n.ID]; ok {
				filtered = append(filtered, n)
			}
		}
		nodes = filtered
	}
	for _, n := range nodes {
		enrichNode(n, h)
	}
	var usersByNode map[uuid.UUID]int
	if h.Counters != nil {
		usersByNode, _ = h.Counters.UsersCountByNode(c.Request().Context())
	}
	items := make([]domain.NodeListItem, len(nodes))
	for i, n := range nodes {
		items[i] = domain.NodeListItem{
			Node:       *n,
			Location:   service.NodeDisplayLocation(n),
			UsersCount: usersByNode[n.ID],
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"nodes": items})
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
	enrichNode(n, h)
	return c.JSON(http.StatusOK, n)
}

type updateNodeRequest struct {
	Name         string   `json:"name"`
	Address      string   `json:"address"`
	Core         string   `json:"core"`
	EnabledCores []string `json:"enabled_cores"`
	UsageRatio   float64  `json:"usage_ratio"`
	Endpoint     string  `json:"endpoint"`
	Region       *string `json:"region"`
	CountryCode  *string `json:"country_code"`
	LocationAuto *bool   `json:"location_auto"`
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
	in := service.UpdateNodeInput{
		Name: req.Name, Address: req.Address, UsageRatio: req.UsageRatio, Endpoint: req.Endpoint,
		Core: domain.CoreType(req.Core),
		EnabledCores: coreTypesFromStrings(req.EnabledCores),
	}
	if req.Region != nil {
		in.RegionSet = true
		in.Region = *req.Region
	}
	if req.CountryCode != nil {
		in.CountrySet = true
		in.CountryCode = *req.CountryCode
	}
	if req.LocationAuto != nil {
		in.LocationSet = true
		in.LocationAuto = req.LocationAuto
	}
	n, err := h.Nodes.Update(c.Request().Context(), id, in)
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
	enrichNode(n, h)
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
	Core       string            `json:"core"`
	Protocol   string            `json:"protocol"`
	Listen     string            `json:"listen"`
	Port       int               `json:"port"`
	PortEnd    int               `json:"port_end"`
	Network    string            `json:"network"`
	Security   string            `json:"security"`
	SNI        []string          `json:"sni"`
	Path       string            `json:"path"`
	Host       []string          `json:"host"`
	Flow       string            `json:"flow"`
	Raw        map[string]any    `json:"raw"`
	Enabled    bool              `json:"enabled"`
	SpeedLimit int64             `json:"speed_limit"`
	Notes      string            `json:"notes"`
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
		NodeID: nodeID, Tag: req.Tag, Core: domain.CoreType(req.Core),
		Protocol: domain.Protocol(req.Protocol),
		Listen: req.Listen, Port: req.Port, PortEnd: req.PortEnd, Network: req.Network,
		Security: domain.Security(req.Security), SNI: req.SNI, Path: req.Path,
		Host: req.Host, Flow: req.Flow, Raw: req.Raw, Enabled: req.Enabled,
		SpeedLimit: req.SpeedLimit, Notes: req.Notes, GeoPolicy: req.GeoPolicy,
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

// ListInbounds returns inbounds for a node (?node_id=...) or the full fleet when
// node_id is omitted. Resellers only see inbounds on their admin allowlist.
func (h *Handlers) ListInbounds(c echo.Context) error {
	nodeIDParam := c.QueryParam("node_id")
	if nodeIDParam == "" {
		return h.listInboundsFleet(c)
	}
	nodeID, err := uuid.Parse(nodeIDParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
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

func (h *Handlers) listInboundsFleet(c echo.Context) error {
	items, err := h.Inbounds.ListFleet(c.Request().Context())
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
		filtered := items[:0]
		for _, item := range items {
			if _, ok := allow[item.ID]; ok {
				filtered = append(filtered, item)
			}
		}
		items = filtered
	}
	h.enrichFleetHealth(items)
	return c.JSON(http.StatusOK, echo.Map{"inbounds": items})
}

// enrichFleetHealth sets the Health field on each InboundListItem based on
// the hosting node's live status from the Hub.
//
// Health values:
//   - "healthy"  — node connected, core running, CPU < 90%, memory < 90%
//   - "degraded" — node connected but high CPU (≥90%) or high memory (≥90%)
//   - "down"     — node disconnected or core not running
func (h *Handlers) enrichFleetHealth(items []domain.InboundListItem) {
	if h.Hub == nil || len(items) == 0 {
		return
	}
	// Build a per-node health cache so we call Hub.Live at most once per node.
	type nodeHealth struct {
		status domain.NodeStatus
		health domain.NodeHealth
		ok     bool
	}
	cache := make(map[uuid.UUID]*nodeHealth)
	for i := range items {
		nid := items[i].NodeID
		if _, exists := cache[nid]; !exists {
			st, hl, _, ok := h.Hub.Live(nid)
			cache[nid] = &nodeHealth{status: st, health: hl, ok: ok}
		}
		nh := cache[nid]
		if !nh.ok || nh.status != domain.NodeConnected || !nh.health.CoreRunning {
			items[i].Health = "down"
		} else if nh.health.CPUPercent >= 90 || nh.health.MemPercent >= 90 {
			items[i].Health = "degraded"
		} else {
			items[i].Health = "healthy"
		}
	}
}

type updateInboundRequest struct {
	Listen     string            `json:"listen"`
	Port       int               `json:"port"`
	PortEnd    int               `json:"port_end"`
	Network    string            `json:"network"`
	Security   string            `json:"security"`
	Core       *string           `json:"core"`
	SNI        []string          `json:"sni"`
	Path       string            `json:"path"`
	Host       []string          `json:"host"`
	Flow       string            `json:"flow"`
	Raw        *map[string]any   `json:"raw"`
	Enabled    bool              `json:"enabled"`
	SpeedLimit int64             `json:"speed_limit"`
	Notes      string            `json:"notes"`
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
		Listen: req.Listen, Port: req.Port, PortEnd: req.PortEnd, Network: req.Network,
		Security: domain.Security(req.Security), Core: optionalCoreType(req.Core), SNI: req.SNI, Path: req.Path,
		Host: req.Host, Flow: req.Flow, Raw: req.Raw, Enabled: req.Enabled,
		SpeedLimit: req.SpeedLimit, Notes: req.Notes, GeoPolicy: req.GeoPolicy,
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

// CloneInbound duplicates an inbound with a new tag and port.
func (h *Handlers) CloneInbound(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req struct {
		Port int `json:"port"`
	}
	_ = c.Bind(&req) // port is optional
	in, err := h.Inbounds.Clone(c.Request().Context(), id, req.Port)
	if in == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	resp := echo.Map{"inbound": in}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

// BulkInboundAction applies enable/disable/delete to multiple inbounds.
func (h *Handlers) BulkInboundAction(c echo.Context) error {
	var req struct {
		IDs    []string `json:"ids"`
		Action string   `json:"action"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if len(req.IDs) == 0 || req.Action == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "ids and action are required")
	}
	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}
	affected, err := h.Inbounds.BulkAction(c.Request().Context(), ids, req.Action)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"affected": affected})
}

// CheckInboundPort reports whether a port is available on a node.
func (h *Handlers) CheckInboundPort(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	portStr := c.QueryParam("port")
	port := 0
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	}
	if port <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid port")
	}
	available, tag, err := h.Inbounds.CheckPort(c.Request().Context(), nodeID, port)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "check failed")
	}
	resp := echo.Map{"available": available}
	if tag != "" {
		resp["conflict_tag"] = tag
	}
	return c.JSON(http.StatusOK, resp)
}

func coreTypesFromStrings(ss []string) []domain.CoreType {
	if len(ss) == 0 {
		return nil
	}
	out := make([]domain.CoreType, 0, len(ss))
	for _, s := range ss {
		if s == "" {
			continue
		}
		out = append(out, domain.CoreType(s))
	}
	return out
}

func optionalCoreType(s *string) *domain.CoreType {
	if s == nil {
		return nil
	}
	ct := domain.CoreType(*s)
	return &ct
}

// GetInboundOnline returns the approximate online connection count for an inbound.
// It sums the per-user connection counts on the inbound's hosting node.
func (h *Handlers) GetInboundOnline(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ctx := c.Request().Context()
	nodeID, err := h.Inbounds.GetNodeIDForInbound(ctx, id)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"count": 0})
	}
	if h.Hub == nil {
		return c.JSON(http.StatusOK, echo.Map{"count": 0})
	}
	stats, err := h.Hub.OnlineStats(ctx, nodeID)
	if err != nil || stats == nil {
		return c.JSON(http.StatusOK, echo.Map{"count": 0})
	}
	total := 0
	for _, n := range stats {
		total += n
	}
	return c.JSON(http.StatusOK, echo.Map{"count": total})
}

// GetInboundStats returns per-inbound traffic statistics (last 30 days).
func (h *Handlers) GetInboundStats(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if h.InboundTraffic == nil {
		return c.JSON(http.StatusOK, &domain.InboundTrafficStats{InboundID: id})
	}
	stats, err := h.InboundTraffic.GetStats(c.Request().Context(), id, 30)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "stats failed")
	}
	return c.JSON(http.StatusOK, stats)
}

// GetCertStatus returns the TLS certificate status for an inbound.
func (h *Handlers) GetCertStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if h.CertStatus == nil {
		return c.JSON(http.StatusOK, &domain.InboundCertStatus{Status: "none"})
	}
	status, err := h.CertStatus.GetStatus(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "cert status failed")
	}
	return c.JSON(http.StatusOK, status)
}

// GetShareLink generates a protocol-specific share URI for an inbound.
func (h *Handlers) GetShareLink(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if h.ShareLinks == nil {
		return c.JSON(http.StatusOK, echo.Map{"error": "share links not configured"})
	}
	link, err := h.ShareLinks.GenerateLink(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusOK, echo.Map{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, echo.Map{"link": link})
}
