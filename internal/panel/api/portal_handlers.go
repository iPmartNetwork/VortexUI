package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/subscription"
)

// PortalHandlers serves end-user self-service endpoints.
type PortalHandlers struct {
	Portal          *service.PortalService
	Admins          *service.AdminService
	Issuer          *auth.Issuer
	ResellerPayment *service.ResellerPaymentService
	WalletBilling   *service.WalletBillingService
	Sub             *service.SubscriptionService
	Traffic         port.TrafficRepository
	Users           *service.UserService
	Online          DeviceCounter // optional; HWID fallback when live IP stats unavailable
	DeepLink        *service.DeepLinkService
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

// PortalSubscription returns copyable subscription URLs and share links for the
// authenticated portal user.
func (h *PortalHandlers) PortalSubscription(c echo.Context) error {
	if h.Sub == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "subscription service unavailable")
	}
	userID := portalUserID(c)
	res, err := h.Sub.BuildForUser(c.Request().Context(), userID)
	if errors.Is(err, domain.ErrNotFound) || res == nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "subscription build failed")
	}

	links := make([]string, 0, len(res.Proxies))
	for _, p := range res.Proxies {
		if l := subscription.ShareLink(p); l != "" {
			links = append(links, l)
		}
	}

	base := c.Scheme() + "://" + c.Request().Host
	path := "/sub/" + res.User.SubToken
	return c.JSON(http.StatusOK, echo.Map{
		"token":             res.User.SubToken,
		"subscription_path": path,
		"subscription_url":  base + path,
		"formats": echo.Map{
			"auto":    base + path,
			"clash":   base + path + "?format=clash",
			"singbox": base + path + "?format=singbox",
			"base64":  base + path + "?format=base64",
		},
		"links": links,
	})
}

// PortalUsage returns the authenticated user's bucketed traffic series.
// Defaults to the last 7 days in daily buckets.
func (h *PortalHandlers) PortalUsage(c echo.Context) error {
	if h.Traffic == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "traffic service unavailable")
	}
	userID := portalUserID(c)
	now := time.Now()
	q := port.SeriesQuery{
		FromUnix: int64(atoiDefault(c.QueryParam("from"), int(now.AddDate(0, 0, -7).Unix()))),
		ToUnix:   int64(atoiDefault(c.QueryParam("to"), int(now.Unix()))),
		Bucket:   c.QueryParam("bucket"),
	}
	if q.Bucket == "" {
		q.Bucket = "1d"
	}
	points, err := h.Traffic.UsageSeries(c.Request().Context(), userID, q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "usage query failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"points": points})
}

// PortalOnline reports live connections and active devices for the portal user.
func (h *PortalHandlers) PortalOnline(c echo.Context) error {
	if h.Users == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "user service unavailable")
	}
	userID := portalUserID(c)
	resp := echo.Map{"window_seconds": int(deviceWindow.Seconds())}

	ips, ipTracked, err := h.Users.OnlineIPList(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "online ip lookup failed")
	}
	if ipTracked {
		resp["active_devices"] = len(ips)
		resp["device_tracking"] = true
	} else if h.Online != nil {
		n, err := h.Online.Online(c.Request().Context(), userID.String(), deviceWindow)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "online lookup failed")
		}
		resp["active_devices"] = n
		resp["device_tracking"] = true
	} else {
		resp["active_devices"] = 0
		resp["device_tracking"] = false
	}

	live, tracked, err := h.Users.LiveConnections(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "live connection lookup failed")
	}
	resp["live_connections"] = live
	resp["live_tracking"] = tracked

	return c.JSON(http.StatusOK, resp)
}

// PortalDeepLink returns a client deep-link URL when deep links are enabled.
func (h *PortalHandlers) PortalDeepLink(c echo.Context) error {
	if h.DeepLink == nil || h.Portal == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "deep link service unavailable")
	}
	user, err := h.Portal.GetUsage(c.Request().Context(), portalUserID(c))
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	link, err := h.DeepLink.GenerateDeepLink(c.Request().Context(), user.SubToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"deep_link": link})
}

// --- Portal Plans ---

// PortalListPlans returns enabled plans for purchase (filtered by reseller allowlist).
func (h *PortalHandlers) PortalListPlans(c echo.Context) error {
	userID := portalUserID(c)
	plans, err := h.Portal.ListPlans(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list plans")
	}
	if h.Admins != nil && userID != uuid.Nil {
		user, uerr := h.Portal.GetUsage(c.Request().Context(), userID)
		if uerr == nil && user.AdminID != nil {
			plans, err = h.Admins.FilterPlans(c.Request().Context(), *user.AdminID, plans)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "failed to filter plans")
			}
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"plans": plans})
}

// PortalPaymentInfo returns card/crypto details for manual payment methods.
func (h *PortalHandlers) PortalPaymentInfo(c echo.Context) error {
	userID := portalUserID(c)
	var adminID *uuid.UUID
	if userID != uuid.Nil {
		user, err := h.Portal.GetUsage(c.Request().Context(), userID)
		if err == nil && user.AdminID != nil {
			adminID = user.AdminID
		}
	}
	s, err := billingSettingsForReseller(c.Request().Context(), h.ResellerPayment, h.WalletBilling, adminID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "payment info failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"settings": s})
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

// AdminGetTicket returns a single ticket with messages (admin view).
func (h *PortalHandlers) AdminGetTicket(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	ticket, err := h.Portal.AdminGetTicket(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"ticket": ticket})
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
