package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// PortalHandlers serves end-user self-service endpoints.
type PortalHandlers struct {
	Portal *service.PortalService
	Issuer *auth.Issuer
}

// --- Portal Auth ---

type portalLoginRequest struct {
	Token string `json:"token"` // subscription token
}

// PortalLogin authenticates an end-user by their subscription token.
func (h *PortalHandlers) PortalLogin(c echo.Context) error {
	var req portalLoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Token == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "token is required")
	}
	user, err := h.Portal.LoginByToken(c.Request().Context(), req.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	// Issue a portal-scoped JWT (shorter TTL, user-level claims).
	jwt, err := h.Issuer.IssuePortal(user.ID, user.Username, 24*time.Hour)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "token generation failed")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"token":    jwt,
		"user_id":  user.ID,
		"username": user.Username,
	})
}

// --- Portal Dashboard ---

// PortalDashboard returns the user's usage and status.
func (h *PortalHandlers) PortalDashboard(c echo.Context) error {
	userID := portalUserID(c)
	user, err := h.Portal.GetUsage(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	return c.JSON(http.StatusOK, echo.Map{
		"username":       user.Username,
		"status":         user.Status,
		"data_limit":     user.DataLimit,
		"used_traffic":   user.UsedTraffic,
		"expire_at":      user.ExpireAt,
		"device_limit":   user.DeviceLimit,
		"reset_strategy": user.ResetStrategy,
		"sub_token":      user.SubToken,
		"created_at":     user.CreatedAt,
	})
}

// --- Portal Plans ---

// PortalListPlans returns enabled plans for purchase.
func (h *PortalHandlers) PortalListPlans(c echo.Context) error {
	plans, err := h.Portal.ListPlans(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list plans")
	}
	return c.JSON(http.StatusOK, echo.Map{"plans": plans})
}

// --- Portal Tickets ---

type createTicketRequest struct {
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Priority string `json:"priority"`
}

// PortalCreateTicket opens a new support ticket.
func (h *PortalHandlers) PortalCreateTicket(c echo.Context) error {
	var req createTicketRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	ticket, err := h.Portal.CreateTicket(c.Request().Context(), service.CreateTicketInput{
		UserID:   portalUserID(c),
		Subject:  req.Subject,
		Body:     req.Body,
		Priority: domain.TicketPriority(req.Priority),
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusCreated, echo.Map{"ticket": ticket})
}

// PortalListTickets lists the user's tickets.
func (h *PortalHandlers) PortalListTickets(c echo.Context) error {
	tickets, err := h.Portal.ListUserTickets(c.Request().Context(), portalUserID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tickets")
	}
	return c.JSON(http.StatusOK, echo.Map{"tickets": tickets})
}

// PortalGetTicket returns a single ticket with messages.
func (h *PortalHandlers) PortalGetTicket(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ticket, err := h.Portal.GetTicket(c.Request().Context(), id, portalUserID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ticket": ticket})
}

type replyTicketRequest struct {
	Body string `json:"body"`
}

// PortalReplyTicket adds a reply to a ticket.
func (h *PortalHandlers) PortalReplyTicket(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req replyTicketRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if err := h.Portal.ReplyTicket(c.Request().Context(), id, portalUserID(c), req.Body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "replied"})
}

// --- Admin Ticket Management ---

// AdminListTickets lists all tickets (admin view).
func (h *PortalHandlers) AdminListTickets(c echo.Context) error {
	status := c.QueryParam("status")
	tickets, total, err := h.Portal.ListAllTickets(c.Request().Context(), status, 50, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list tickets")
	}
	return c.JSON(http.StatusOK, echo.Map{"tickets": tickets, "total": total})
}

type adminReplyRequest struct {
	Body string `json:"body"`
}

// AdminReplyTicket adds an admin reply to a ticket.
func (h *PortalHandlers) AdminReplyTicket(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req adminReplyRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if err := h.Portal.AdminReply(c.Request().Context(), id, claims.AdminID, req.Body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "replied"})
}

// AdminCloseTicket closes a ticket.
func (h *PortalHandlers) AdminCloseTicket(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Portal.CloseTicket(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"status": "closed"})
}

// portalUserID extracts the user UUID from portal JWT claims stored in context.
func portalUserID(c echo.Context) uuid.UUID {
	if claims, ok := c.Get("portal_claims").(*auth.PortalClaims); ok {
		return claims.UserID
	}
	return uuid.Nil
}
