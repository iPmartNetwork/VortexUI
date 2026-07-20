package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// ProtocolGroupHandlers exposes REST endpoints for auto-protocol switching:
// ProtocolGroup CRUD, ISP Profiles, and SwitchEvent recording/summary.
type ProtocolGroupHandlers struct {
	Groups *service.ProtocolGroupService
	Events *service.SwitchEventService
}

// --- Protocol Group CRUD ---

type createProtocolGroupRequest struct {
	NodeID        uuid.UUID   `json:"node_id"`
	Name          string      `json:"name"`
	InboundIDs    []uuid.UUID `json:"inbound_ids"`
	ProbeURL      string      `json:"probe_url"`
	ProbeInterval int         `json:"probe_interval"`
	ProbeTimeout  int         `json:"probe_timeout"`
	MaxRetries    int         `json:"max_retries"`
}

// CreateProtocolGroup creates a new auto-protocol switching group.
func (h *ProtocolGroupHandlers) CreateProtocolGroup(c echo.Context) error {
	var req createProtocolGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	g := &domain.ProtocolGroup{
		NodeID:        req.NodeID,
		Name:          req.Name,
		InboundIDs:    req.InboundIDs,
		ProbeURL:      req.ProbeURL,
		ProbeInterval: req.ProbeInterval,
		ProbeTimeout:  req.ProbeTimeout,
		MaxRetries:    req.MaxRetries,
	}
	if err := h.Groups.Create(c.Request().Context(), g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"group": g})
}

// ListProtocolGroups returns all protocol groups for a node (?node_id=).
func (h *ProtocolGroupHandlers) ListProtocolGroups(c echo.Context) error {
	nodeID, err := uuid.Parse(c.QueryParam("node_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "node_id is required")
	}
	groups, err := h.Groups.ListByNode(c.Request().Context(), nodeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"groups": groups})
}

// GetProtocolGroup returns a single protocol group by ID.
func (h *ProtocolGroupHandlers) GetProtocolGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	g, err := h.Groups.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	return c.JSON(http.StatusOK, g)
}

type updateProtocolGroupRequest struct {
	Name          string      `json:"name"`
	InboundIDs    []uuid.UUID `json:"inbound_ids"`
	ProbeURL      string      `json:"probe_url"`
	ProbeInterval int         `json:"probe_interval"`
	ProbeTimeout  int         `json:"probe_timeout"`
	MaxRetries    int         `json:"max_retries"`
}

// UpdateProtocolGroup updates an existing protocol group.
func (h *ProtocolGroupHandlers) UpdateProtocolGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateProtocolGroupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	g, err := h.Groups.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "group not found")
	}
	g.Name = req.Name
	g.InboundIDs = req.InboundIDs
	g.ProbeURL = req.ProbeURL
	g.ProbeInterval = req.ProbeInterval
	g.ProbeTimeout = req.ProbeTimeout
	g.MaxRetries = req.MaxRetries
	if err := h.Groups.Update(c.Request().Context(), g); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"group": g})
}

// DeleteProtocolGroup removes a protocol group.
func (h *ProtocolGroupHandlers) DeleteProtocolGroup(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Groups.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

type reorderInboundsRequest struct {
	InboundIDs []uuid.UUID `json:"inbound_ids"`
}

// ReorderInbounds sets a new priority order for inbounds within a group.
func (h *ProtocolGroupHandlers) ReorderInbounds(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req reorderInboundsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Groups.ReorderInbounds(c.Request().Context(), id, req.InboundIDs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// --- ISP Profiles ---

type createISPProfileRequest struct {
	GroupID            uuid.UUID `json:"group_id"`
	ISPIdentifier      string    `json:"isp_identifier"`
	CountryCode        string    `json:"country_code"`
	PreferredProtocols []string  `json:"preferred_protocols"`
}

// CreateISPProfile creates a new ISP-specific protocol preference profile.
func (h *ProtocolGroupHandlers) CreateISPProfile(c echo.Context) error {
	var req createISPProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p := &domain.ISPProfile{
		GroupID:            req.GroupID,
		ISPIdentifier:      req.ISPIdentifier,
		CountryCode:        req.CountryCode,
		PreferredProtocols: req.PreferredProtocols,
	}
	if err := h.Groups.CreateISPProfile(c.Request().Context(), p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"profile": p})
}

// ListISPProfiles returns all ISP profiles for a group (?group_id=).
func (h *ProtocolGroupHandlers) ListISPProfiles(c echo.Context) error {
	groupID, err := uuid.Parse(c.QueryParam("group_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "group_id is required")
	}
	profiles, err := h.Groups.ListISPProfiles(c.Request().Context(), groupID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"profiles": profiles})
}

type updateISPProfileRequest struct {
	ISPIdentifier      string   `json:"isp_identifier"`
	CountryCode        string   `json:"country_code"`
	PreferredProtocols []string `json:"preferred_protocols"`
}

// UpdateISPProfile updates an existing ISP profile.
func (h *ProtocolGroupHandlers) UpdateISPProfile(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateISPProfileRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	p := &domain.ISPProfile{
		ID:                 id,
		ISPIdentifier:      req.ISPIdentifier,
		CountryCode:        req.CountryCode,
		PreferredProtocols: req.PreferredProtocols,
	}
	if err := h.Groups.UpdateISPProfile(c.Request().Context(), p); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"profile": p})
}

// DeleteISPProfile removes an ISP profile.
func (h *ProtocolGroupHandlers) DeleteISPProfile(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Groups.DeleteISPProfile(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Switch Events ---

type recordSwitchEventRequest struct {
	UserID         uuid.UUID `json:"user_id"`
	NodeID         uuid.UUID `json:"node_id"`
	GroupID        uuid.UUID `json:"group_id"`
	SourceProtocol string    `json:"source_protocol"`
	TargetProtocol string    `json:"target_protocol"`
	ISP            string    `json:"isp"`
}

// RecordSwitchEvent records a client-reported protocol switch.
func (h *ProtocolGroupHandlers) RecordSwitchEvent(c echo.Context) error {
	if h.Events == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "switch events not configured")
	}
	var req recordSwitchEventRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	e := &domain.SwitchEvent{
		UserID:         req.UserID,
		NodeID:         req.NodeID,
		GroupID:        req.GroupID,
		SourceProtocol: req.SourceProtocol,
		TargetProtocol: req.TargetProtocol,
		ISP:            req.ISP,
	}
	if err := h.Events.Record(c.Request().Context(), e); err != nil {
		if errors.Is(err, errors.New("")) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"event": e})
}

// GetSwitchSummary returns aggregated switch event data.
// Query params: node_id, user_id, isp, from, to (RFC3339).
func (h *ProtocolGroupHandlers) GetSwitchSummary(c echo.Context) error {
	if h.Events == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "switch events not configured")
	}
	filter := domain.SwitchEventFilter{}
	if v := c.QueryParam("node_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.NodeID = &id
		}
	}
	if v := c.QueryParam("user_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filter.UserID = &id
		}
	}
	filter.ISP = c.QueryParam("isp")
	if v := c.QueryParam("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.FromTime = t
		}
	}
	if v := c.QueryParam("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			filter.ToTime = t
		}
	}
	summary, err := h.Events.Summary(c.Request().Context(), filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "summary failed")
	}
	return c.JSON(http.StatusOK, summary)
}
