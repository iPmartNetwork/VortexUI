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
}

// CreateNode registers a new node and brings it under live management.
func (h *Handlers) CreateNode(c echo.Context) error {
	var req createNodeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	n, err := h.Nodes.Create(c.Request().Context(), service.CreateNodeInput{
		Name: req.Name, Address: req.Address, Core: domain.CoreType(req.Core), UsageRatio: req.UsageRatio,
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
		Name: req.Name, Address: req.Address, UsageRatio: req.UsageRatio,
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

// --- inbounds ---

type createInboundRequest struct {
	NodeID   string         `json:"node_id"`
	Tag      string         `json:"tag"`
	Protocol string         `json:"protocol"`
	Listen   string         `json:"listen"`
	Port     int            `json:"port"`
	Network  string         `json:"network"`
	Security string         `json:"security"`
	SNI      []string       `json:"sni"`
	Path     string         `json:"path"`
	Host     []string       `json:"host"`
	Flow     string         `json:"flow"`
	Raw      map[string]any `json:"raw"`
	Enabled  bool           `json:"enabled"`
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

// ListInbounds returns the inbounds for a node (?node_id=...).
func (h *Handlers) ListInbounds(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id query param required")
	}
	ins, err := h.Inbounds.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"inbounds": ins})
}

type updateInboundRequest struct {
	Listen   string         `json:"listen"`
	Port     int            `json:"port"`
	Network  string         `json:"network"`
	Security string         `json:"security"`
	SNI      []string       `json:"sni"`
	Path     string         `json:"path"`
	Host     []string       `json:"host"`
	Flow     string         `json:"flow"`
	Raw      map[string]any `json:"raw"`
	Enabled  bool           `json:"enabled"`
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
