package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// policyError maps a service error to an HTTP error: validation -> 400,
// missing entity -> 404, anything else -> 500.
func policyError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, "not found")
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, errString(err))
	}
}

// --- outbounds ---

type outboundRequest struct {
	NodeID   string         `json:"node_id"`
	Tag      string         `json:"tag"`
	Protocol string         `json:"protocol"`
	Address  string         `json:"address"`
	Port     int            `json:"port"`
	UUID     string         `json:"uuid"`
	Password string         `json:"password"`
	Username string         `json:"username"`
	Method   string         `json:"method"`
	Flow     string         `json:"flow"`
	Network  string         `json:"network"`
	Security string         `json:"security"`
	SNI      string         `json:"sni"`
	Path     string         `json:"path"`
	Host     string         `json:"host"`
	Raw      map[string]any `json:"raw"`
	Enabled  *bool          `json:"enabled"`
}

// CreateOutbound adds an egress handler to a node and resyncs its core.
func (h *Handlers) CreateOutbound(c echo.Context) error {
	var req outboundRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	o, err := h.Outbounds.Create(c.Request().Context(), service.CreateOutboundInput{
		NodeID: nodeID, Tag: req.Tag, Protocol: domain.OutboundProtocol(req.Protocol),
		Address: req.Address, Port: req.Port, UUID: req.UUID, Password: req.Password,
		Username: req.Username, Method: req.Method, Flow: req.Flow, Network: req.Network,
		Security: domain.Security(req.Security), SNI: req.SNI, Path: req.Path,
		Host: req.Host, Raw: req.Raw, Enabled: req.Enabled,
	})
	if o == nil {
		return policyError(err)
	}
	resp := echo.Map{"outbound": o}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListOutbounds returns the outbounds for a node (?node_id=...).
func (h *Handlers) ListOutbounds(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id query param required")
	}
	outs, err := h.Outbounds.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"outbounds": outs})
}

// UpdateOutbound edits an outbound and resyncs its node.
func (h *Handlers) UpdateOutbound(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req outboundRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	o, err := h.Outbounds.Update(c.Request().Context(), id, service.UpdateOutboundInput{
		Protocol: domain.OutboundProtocol(req.Protocol), Address: req.Address, Port: req.Port,
		UUID: req.UUID, Password: req.Password, Username: req.Username, Method: req.Method,
		Flow: req.Flow, Network: req.Network, Security: domain.Security(req.Security),
		SNI: req.SNI, Path: req.Path, Host: req.Host, Raw: req.Raw, Enabled: req.Enabled != nil && *req.Enabled,
	})
	if o == nil {
		return policyError(err)
	}
	resp := echo.Map{"outbound": o}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteOutbound removes an outbound and resyncs its node.
func (h *Handlers) DeleteOutbound(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Outbounds.Delete(c.Request().Context(), id); err != nil {
		return policyError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- routing rules ---

type routingRequest struct {
	NodeID      string   `json:"node_id"`
	Priority    int      `json:"priority"`
	Name        string   `json:"name"`
	InboundTags []string `json:"inbound_tags"`
	Domains     []string `json:"domains"`
	IP          []string `json:"ip"`
	Port        string   `json:"port"`
	Protocols   []string `json:"protocols"`
	Network     string   `json:"network"`
	OutboundTag string   `json:"outbound_tag"`
	BalancerTag string   `json:"balancer_tag"`
	Enabled     *bool    `json:"enabled"`
}

func (req routingRequest) toInput() service.RoutingRuleInput {
	return service.RoutingRuleInput{
		Priority: req.Priority, Name: req.Name, InboundTags: req.InboundTags,
		Domains: req.Domains, IP: req.IP, Port: req.Port, Protocols: req.Protocols,
		Network: req.Network, OutboundTag: req.OutboundTag, BalancerTag: req.BalancerTag,
		Enabled: req.Enabled,
	}
}

// CreateRoutingRule adds a routing rule to a node and resyncs its core.
func (h *Handlers) CreateRoutingRule(c echo.Context) error {
	var req routingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	rule, err := h.Routing.Create(c.Request().Context(), nodeID, req.toInput())
	if rule == nil {
		return policyError(err)
	}
	resp := echo.Map{"rule": rule}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListRoutingRules returns a node's routing rules (?node_id=...).
func (h *Handlers) ListRoutingRules(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id query param required")
	}
	rules, err := h.Routing.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"rules": rules})
}

// UpdateRoutingRule edits a routing rule and resyncs its node.
func (h *Handlers) UpdateRoutingRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req routingRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	rule, err := h.Routing.Update(c.Request().Context(), id, req.toInput())
	if rule == nil {
		return policyError(err)
	}
	resp := echo.Map{"rule": rule}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteRoutingRule removes a routing rule and resyncs its node.
func (h *Handlers) DeleteRoutingRule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Routing.Delete(c.Request().Context(), id); err != nil {
		return policyError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- balancers ---

type balancerRequest struct {
	NodeID        string   `json:"node_id"`
	Tag           string   `json:"tag"`
	Selectors     []string `json:"selectors"`
	Strategy      string   `json:"strategy"`
	Observe       bool     `json:"observe"`
	ProbeURL      string   `json:"probe_url"`
	ProbeInterval string   `json:"probe_interval"`
	Enabled       *bool    `json:"enabled"`
}

func (req balancerRequest) toInput() service.BalancerInput {
	return service.BalancerInput{
		Tag: req.Tag, Selectors: req.Selectors, Strategy: domain.BalancerStrategy(req.Strategy),
		Observe: req.Observe, ProbeURL: req.ProbeURL, ProbeInterval: req.ProbeInterval,
		Enabled: req.Enabled,
	}
}

// CreateBalancer adds a balancer to a node and resyncs its core.
func (h *Handlers) CreateBalancer(c echo.Context) error {
	var req balancerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	nodeID, err := uuid.Parse(req.NodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid node_id")
	}
	b, err := h.Balancers.Create(c.Request().Context(), nodeID, req.toInput())
	if b == nil {
		return policyError(err)
	}
	resp := echo.Map{"balancer": b}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

// ListBalancers returns a node's balancers (?node_id=...).
func (h *Handlers) ListBalancers(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id query param required")
	}
	bals, err := h.Balancers.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"balancers": bals})
}

// UpdateBalancer edits a balancer and resyncs its node.
func (h *Handlers) UpdateBalancer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req balancerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	b, err := h.Balancers.Update(c.Request().Context(), id, req.toInput())
	if b == nil {
		return policyError(err)
	}
	resp := echo.Map{"balancer": b}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// DeleteBalancer removes a balancer and resyncs its node.
func (h *Handlers) DeleteBalancer(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Balancers.Delete(c.Request().Context(), id); err != nil {
		return policyError(err)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- overview ---
// GetOverview returns the dashboard summary: user aggregates and node fleet
// connectivity/health.
func (h *Handlers) GetOverview(c echo.Context) error {
	ov, err := h.Overview.Build(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "overview failed")
	}
	return c.JSON(http.StatusOK, ov)
}

// --- reality key generation ---
// GenerateReality returns a fresh REALITY X25519 keypair and a short ID so the
// operator can populate a reality inbound. Nothing is persisted; the caller
// stores the chosen material on the inbound.
func (h *Handlers) GenerateReality(c echo.Context) error {
	kp, err := reality.GenerateKeypair()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "keygen failed")
	}
	shortID, err := reality.ShortID(8)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "keygen failed")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"private_key": kp.PrivateKey,
		"public_key":  kp.PublicKey,
		"short_id":    shortID,
	})
}

// --- backup / restore ---

// GetBackup exports the full proxy configuration as a downloadable JSON document.
func (h *Handlers) GetBackup(c echo.Context) error {
	b, err := h.Backup.Export(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "export failed")
	}
	c.Response().Header().Set("Content-Disposition", `attachment; filename="vortexui-backup.json"`)
	return c.JSON(http.StatusOK, b)
}

// RestoreBackup replaces the entire proxy configuration with the posted snapshot.
// Destructive: it wipes the current config and re-applies the document in one
// transaction. Live cores reconcile on the next node (re)connect.
func (h *Handlers) RestoreBackup(c echo.Context) error {
	var b domain.Backup
	if err := c.Bind(&b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid backup document")
	}
	if err := h.Backup.Restore(c.Request().Context(), &b); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusOK, echo.Map{
		"restored": echo.Map{
			"nodes": len(b.Nodes), "inbounds": len(b.Inbounds), "outbounds": len(b.Outbounds),
			"routing": len(b.Routing), "balancers": len(b.Balancers),
			"users": len(b.Users), "bindings": len(b.Bindings),
		},
	})
}

// --- logs ---

// GetLogs returns recent panel log entries (in-memory ring buffer). Filter with
// ?level=debug|info|warn|error and cap with ?limit=. Node/core logs are not
// included here — those require the node agent to stream them.
func (h *Handlers) GetLogs(c echo.Context) error {
	if h.Logs == nil {
		return c.JSON(http.StatusOK, echo.Map{"entries": []any{}})
	}
	limit := atoiDefault(c.QueryParam("limit"), 200)
	entries := h.Logs.Entries(parseLogLevel(c.QueryParam("level")), limit)
	return c.JSON(http.StatusOK, echo.Map{"entries": entries})
}

// parseLogLevel maps a query value to an slog level, defaulting to Info.
func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
