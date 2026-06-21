package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SubHostHandlers serves subscription host (per-inbound override) endpoints.
type SubHostHandlers struct {
	SubHosts *service.SubHostService
}

// subHostRequest is the JSON body for creating/updating a subscription host.
type subHostRequest struct {
	InboundID     string `json:"inbound_id"`
	Remark        string `json:"remark"`
	Address       string `json:"address"`
	Port          *int   `json:"port"`
	SNI           string `json:"sni"`
	HostHeader    string `json:"host"`
	Path          string `json:"path"`
	ALPN          string `json:"alpn"`
	Fingerprint   string `json:"fingerprint"`
	Security      string `json:"security"`
	AllowInsecure bool   `json:"allow_insecure"`
	MuxEnable     bool   `json:"mux_enable"`
	Fragment      string `json:"fragment"`
	Priority      int    `json:"priority"`
	Enabled       bool   `json:"enabled"`
}

func (r subHostRequest) toInput() service.SubHostInput {
	return service.SubHostInput{
		Remark:        r.Remark,
		Address:       r.Address,
		Port:          r.Port,
		SNI:           r.SNI,
		HostHeader:    r.HostHeader,
		Path:          r.Path,
		ALPN:          r.ALPN,
		Fingerprint:   r.Fingerprint,
		Security:      domain.HostSecurity(r.Security),
		AllowInsecure: r.AllowInsecure,
		MuxEnable:     r.MuxEnable,
		Fragment:      r.Fragment,
		Priority:      r.Priority,
		Enabled:       r.Enabled,
	}
}

// ListSubHosts returns the hosts bound to the inbound named by ?inbound_id=,
// in priority order.
func (h *SubHostHandlers) ListSubHosts(c echo.Context) error {
	inboundID, err := uuid.Parse(c.QueryParam("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	hosts, err := h.SubHosts.ListByInbound(c.Request().Context(), inboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"hosts": hosts})
}

// CreateSubHost creates a new subscription host bound to an inbound.
func (h *SubHostHandlers) CreateSubHost(c echo.Context) error {
	var req subHostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	inboundID, err := uuid.Parse(req.InboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	in := req.toInput()
	in.InboundID = inboundID
	host, err := h.SubHosts.Create(c.Request().Context(), in)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"host": host})
}

// UpdateSubHost updates an existing subscription host.
func (h *SubHostHandlers) UpdateSubHost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req subHostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	host, err := h.SubHosts.Update(c.Request().Context(), id, req.toInput())
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "host not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"host": host})
}

// DeleteSubHost removes a subscription host by ID.
func (h *SubHostHandlers) DeleteSubHost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.SubHosts.Delete(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// reorderSubHostsRequest carries the new ordering of host IDs.
type reorderSubHostsRequest struct {
	IDs []string `json:"ids"`
}

// ReorderSubHosts assigns ascending priorities to the supplied host IDs.
func (h *SubHostHandlers) ReorderSubHosts(c echo.Context) error {
	var req reorderSubHostsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	ordered := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid id in ids")
		}
		ordered = append(ordered, id)
	}
	if err := h.SubHosts.Reorder(c.Request().Context(), ordered); errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "host not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
