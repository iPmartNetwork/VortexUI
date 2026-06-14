package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/logbuf"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/subscription"
)

// Handlers bundles the service dependencies the HTTP routes need.
type Handlers struct {
	Auth      *service.AuthService
	Users     *service.UserService
	Sub       *service.SubscriptionService
	Nodes     *service.NodeService
	Inbounds  *service.InboundService
	Outbounds *service.OutboundService
	Routing   *service.RoutingService
	Balancers *service.BalancerService
	Admins    *service.AdminService
	Overview  *service.OverviewService
	Backup    *service.BackupService
	Devices   DeviceLimiter // optional; nil disables device-count enforcement
	Online    DeviceCounter // optional; nil disables the online-devices endpoint
	Logs      LogSource     // optional; nil disables the logs endpoint
	Repo      port.UserRepository
	Traffic   port.TrafficRepository
}

// DeviceLimiter caps the number of distinct devices a user may use within a
// window. *redis.DeviceTracker satisfies it.
type DeviceLimiter interface {
	Allow(ctx context.Context, userID, deviceID string, limit int, window time.Duration) (bool, error)
}

// DeviceCounter reports how many devices a user has been active on within a
// window. *redis.DeviceTracker satisfies it.
type DeviceCounter interface {
	Online(ctx context.Context, userID string, window time.Duration) (int, error)
}

// LogSource returns recent panel log entries at or above minLevel. *logbuf.Handler
// satisfies it.
type LogSource interface {
	Entries(minLevel slog.Level, limit int) []logbuf.Entry
}

// --- auth ---

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login authenticates an admin and returns a JWT.
func (h *Handlers) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	token, err := h.Auth.Login(c.Request().Context(), service.LoginInput{
		Username: req.Username, Password: req.Password, TOTPCode: req.TOTPCode,
	})
	if errors.Is(err, service.ErrInvalidCredentials) {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
	}
	return c.JSON(http.StatusOK, loginResponse{Token: token})
}

// --- users ---

type createUserRequest struct {
	Username      string     `json:"username"`
	Note          string     `json:"note"`
	DataLimit     int64      `json:"data_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	DeviceLimit   int        `json:"device_limit"`
	ResetStrategy string     `json:"reset_strategy"`
	InboundIDs    []string   `json:"inbound_ids"`
	OnHold        bool       `json:"on_hold"`
}

// CreateUser provisions a new service user across its bound inbounds.
func (h *Handlers) CreateUser(c echo.Context) error {
	var req createUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	inboundIDs, err := parseUUIDs(req.InboundIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound id")
	}
	u, err := h.Users.Create(c.Request().Context(), service.CreateUserInput{
		Username:      req.Username,
		Note:          req.Note,
		DataLimit:     req.DataLimit,
		ExpireAt:      req.ExpireAt,
		DeviceLimit:   req.DeviceLimit,
		ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
		InboundIDs:    inboundIDs,
		OnHold:        req.OnHold,
	})
	if u == nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	// u may be non-nil with a provisioning warning; report it without failing
	// the creation, which is already durable.
	resp := echo.Map{"user": u}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusCreated, resp)
}

type updateUserRequest struct {
	Note          string     `json:"note"`
	Status        string     `json:"status"`
	DataLimit     int64      `json:"data_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	DeviceLimit   int        `json:"device_limit"`
	ResetStrategy string     `json:"reset_strategy"`
}

// UpdateUser replaces a user's mutable metadata.
func (h *Handlers) UpdateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	var req updateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	status := domain.UserStatus(req.Status)
	if status == "" {
		status = domain.UserStatusActive
	}
	u, err := h.Users.Update(c.Request().Context(), id, service.UpdateUserInput{
		Note: req.Note, Status: status, DataLimit: req.DataLimit, ExpireAt: req.ExpireAt,
		DeviceLimit: req.DeviceLimit, ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
	})
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if u == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "update failed")
	}
	resp := echo.Map{"user": u}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// ListUsers returns a paginated, filterable list.
func (h *Handlers) ListUsers(c echo.Context) error {
	f := port.UserFilter{
		Search: c.QueryParam("search"),
		Status: domain.UserStatus(c.QueryParam("status")),
		Limit:  atoiDefault(c.QueryParam("limit"), 50),
		Offset: atoiDefault(c.QueryParam("offset"), 0),
	}
	users, total, err := h.Repo.List(c.Request().Context(), f)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"users": users, "total": total})
}

// GetUser fetches one user by id.
func (h *Handlers) GetUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	u, err := h.Repo.GetByID(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	return c.JSON(http.StatusOK, u)
}

// GetUserUsage returns a user's bucketed traffic series. Defaults to the last 7
// days in daily buckets; override with ?from=&to= (unix seconds) and ?bucket=.
func (h *Handlers) GetUserUsage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	now := time.Now()
	q := port.SeriesQuery{
		FromUnix: int64(atoiDefault(c.QueryParam("from"), int(now.AddDate(0, 0, -7).Unix()))),
		ToUnix:   int64(atoiDefault(c.QueryParam("to"), int(now.Unix()))),
		Bucket:   c.QueryParam("bucket"),
	}
	if q.Bucket == "" {
		q.Bucket = "1d"
	}
	points, err := h.Traffic.UsageSeries(c.Request().Context(), id, q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "usage query failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"points": points})
}

// DeleteUser de-provisions and removes a user.
func (h *Handlers) DeleteUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if err := h.Users.Delete(c.Request().Context(), id); errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// GetUserSubscription returns a user's subscription URLs (overall and per
// format) plus the individual share links, so the UI can show copyable links
// and render QR codes at any time. Absolute URLs are derived from the request
// host; the relative path is also included for proxy-agnostic use.
func (h *Handlers) GetUserSubscription(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	res, err := h.Sub.BuildForUser(c.Request().Context(), id)
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

// ResetUserUsage zeroes a user's used traffic now and re-activates them if they
// were quota-limited.
func (h *Handlers) ResetUserUsage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	u, err := h.Users.ResetUsage(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if u == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "reset failed")
	}
	resp := echo.Map{"user": u}
	if err != nil {
		resp["warning"] = err.Error()
	}
	return c.JSON(http.StatusOK, resp)
}

// RevokeUserSub rotates a user's subscription token, invalidating the old link.
func (h *Handlers) RevokeUserSub(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	u, err := h.Users.RevokeSubToken(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "revoke failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"user": u})
}

// GetUserOnline reports a user's live connection count (real, from the cores the
// user is bound to) plus recently-active device count (from the Redis tracker).
func (h *Handlers) GetUserOnline(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	resp := echo.Map{"window_seconds": int(deviceWindow.Seconds())}

	// Recently-active devices (subscription-fetch activity).
	if h.Online != nil {
		n, err := h.Online.Online(c.Request().Context(), id.String(), deviceWindow)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "online lookup failed")
		}
		resp["active_devices"] = n
		resp["device_tracking"] = true
	} else {
		resp["active_devices"] = 0
		resp["device_tracking"] = false
	}

	// Live connections reported by the cores (sum across the user's nodes).
	live, tracked, err := h.Users.LiveConnections(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "live connection lookup failed")
	}
	resp["live_connections"] = live
	resp["live_tracking"] = tracked

	return c.JSON(http.StatusOK, resp)
}

// --- helpers ---

func parseUUIDs(ss []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(ss))
	for _, s := range ss {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, nil
}

func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n := 0
	for _, r := range s {
		if r < '0' || r > '9' {
			return def
		}
		n = n*10 + int(r-'0')
	}
	return n
}

func errString(err error) string {
	if err == nil {
		return "bad request"
	}
	return err.Error()
}
