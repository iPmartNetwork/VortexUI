package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SNIHandlers serves Multi-Domain SNI Routing + SSL endpoints.
type SNIHandlers struct {
	SNI      *service.SNIService
	Inbounds port.InboundRepository
	Resync   *service.FleetResync
}

func (h *SNIHandlers) resyncInbound(ctx context.Context, inboundID uuid.UUID) {
	if h == nil || h.Resync == nil || h.Inbounds == nil {
		return
	}
	in, err := h.Inbounds.GetByID(ctx, inboundID)
	if err != nil || in == nil {
		_ = h.Resync.All(ctx)
		return
	}
	_ = h.Resync.Node(ctx, &in.NodeID)
}

// --- Domains ---

type addDomainRequest struct {
	InboundID string `json:"inbound_id"`
	Domain    string `json:"domain"`
	AutoCert  bool   `json:"auto_cert"`
}

func (h *SNIHandlers) AddDomain(c echo.Context) error {
	var req addDomainRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	inbID, err := uuid.Parse(req.InboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	d, err := h.SNI.AddDomain(c.Request().Context(), service.AddDomainInput{
		InboundID: inbID, Domain: req.Domain, AutoCert: req.AutoCert,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	h.resyncInbound(c.Request().Context(), inbID)
	return c.JSON(http.StatusCreated, echo.Map{"domain": d})
}

func (h *SNIHandlers) ListDomains(c echo.Context) error {
	var inbID *uuid.UUID
	if v := c.QueryParam("inbound_id"); v != "" {
		id, err := uuid.Parse(v)
		if err == nil {
			inbID = &id
		}
	}
	domains, err := h.SNI.ListDomains(c.Request().Context(), inbID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"domains": domains})
}

func (h *SNIHandlers) DeleteDomain(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	d, _ := h.SNI.GetDomain(c.Request().Context(), id)
	if err := h.SNI.DeleteDomain(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if d != nil {
		h.resyncInbound(c.Request().Context(), d.InboundID)
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Certificates ---

type issueCertRequest struct {
	Domain    string `json:"domain"`
	Wildcard  bool   `json:"wildcard"`
	Issuer    string `json:"issuer"`
	AutoRenew bool   `json:"auto_renew"`
}

func (h *SNIHandlers) IssueCert(c echo.Context) error {
	var req issueCertRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	cert, err := h.SNI.IssueCert(c.Request().Context(), service.IssueCertInput{
		Domain: req.Domain, Wildcard: req.Wildcard, Issuer: req.Issuer, AutoRenew: req.AutoRenew,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if h.Resync != nil {
		_ = h.Resync.All(c.Request().Context())
	}
	return c.JSON(http.StatusCreated, echo.Map{"certificate": cert})
}

func (h *SNIHandlers) ListCerts(c echo.Context) error {
	certs, err := h.SNI.ListCerts(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"certificates": certs})
}

func (h *SNIHandlers) DeleteCert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.SNI.DeleteCert(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if h.Resync != nil {
		_ = h.Resync.All(c.Request().Context())
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *SNIHandlers) RenewCert(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	cert, err := h.SNI.RenewCert(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if h.Resync != nil {
		_ = h.Resync.All(c.Request().Context())
	}
	return c.JSON(http.StatusOK, echo.Map{"certificate": cert})
}

// --- SNI Routes ---

type addRouteRequest struct {
	InboundID string `json:"inbound_id"`
	SNI       string `json:"sni"`
	Action    string `json:"action"`
	TargetTag string `json:"target_tag"`
	Priority  int    `json:"priority"`
}

func (h *SNIHandlers) AddRoute(c echo.Context) error {
	var req addRouteRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	inbID, err := uuid.Parse(req.InboundID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound_id")
	}
	r, err := h.SNI.AddRoute(c.Request().Context(), service.AddRouteInput{
		InboundID: inbID, SNI: req.SNI, Action: req.Action, TargetTag: req.TargetTag, Priority: req.Priority,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	h.resyncInbound(c.Request().Context(), inbID)
	return c.JSON(http.StatusCreated, echo.Map{"route": r})
}

func (h *SNIHandlers) ListRoutes(c echo.Context) error {
	inbID, err := uuid.Parse(c.QueryParam("inbound_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "inbound_id required")
	}
	routes, err := h.SNI.ListRoutes(c.Request().Context(), inbID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"routes": routes})
}

func (h *SNIHandlers) DeleteRoute(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	r, _ := h.SNI.GetRoute(c.Request().Context(), id)
	if err := h.SNI.DeleteRoute(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if r != nil {
		h.resyncInbound(c.Request().Context(), r.InboundID)
	}
	return c.NoContent(http.StatusNoContent)
}
