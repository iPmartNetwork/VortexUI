package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/logbuf"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/payment"
	"github.com/vortexui/vortexui/internal/subscription"
	"github.com/vortexui/vortexui/internal/updater"
)

// Handlers bundles the service dependencies the HTTP routes need.
type Handlers struct {
	Version   string // panel build version, surfaced by GET /api/version
	Auth      *service.AuthService
	Users     *service.UserService
	Sub       *service.SubscriptionService
	SubSettings SubUpdateIntervalSource // optional; nil -> hardcoded 12h interval
	Geo       *service.GeoService // optional; nil disables per-user geo recording
	Nodes     *service.NodeService
	Inbounds  *service.InboundService
	Outbounds *service.OutboundService
	Routing   *service.RoutingService
	Balancers *service.BalancerService
	Admins    *service.AdminService
	Overview  *service.OverviewService
	Backup    *service.BackupService
	Plans     PlanServiceInterface          // optional; nil disables plan/order endpoints
	ZarinPal  *payment.ZarinPal            // optional; nil disables ZarinPal payments
	NowPayments *payment.NowPayments       // optional; nil disables crypto payments
	Updater   *updater.Updater             // optional; nil disables update check
	Devices   DeviceLimiter // optional; nil disables device-count enforcement
	Online    DeviceCounter // optional; nil disables the online-devices endpoint
	Logs      LogSource     // optional; nil disables the logs endpoint
	Audit     AuditRecorder // optional; nil disables the audit log
	Repo      port.UserRepository
	Traffic   port.TrafficRepository
	NodeRepo  port.NodeRepository           // optional; resolves node host for WireGuard .conf
	WireGuard *service.WireGuardService     // optional; nil disables the WireGuard .conf endpoint
	Throttle  *LoginThrottle // optional; nil disables login brute-force protection
	Events    EventStream    // optional; nil disables the SSE live-events endpoint
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
	key := c.RealIP() + "|" + strings.ToLower(req.Username)
	if d := h.Throttle.RetryAfter(key); d > 0 {
		c.Response().Header().Set("Retry-After", strconv.Itoa(int(d.Seconds())+1))
		return echo.NewHTTPError(http.StatusTooManyRequests, "too many attempts, try again later")
	}
	token, err := h.Auth.Login(c.Request().Context(), service.LoginInput{
		Username: req.Username, Password: req.Password, TOTPCode: req.TOTPCode,
	})
	if errors.Is(err, service.ErrInvalidCredentials) {
		h.Throttle.Fail(key)
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "login failed")
	}
	h.Throttle.Reset(key)
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
	if err := h.validateInboundAccess(c, inboundIDs); err != nil {
		return err
	}
	// Tag the user with the creating admin (reseller ownership).
	adminID := creatorAdminID(c)
	u, err := h.Users.Create(c.Request().Context(), service.CreateUserInput{
		Username:      req.Username,
		Note:          req.Note,
		DataLimit:     req.DataLimit,
		ExpireAt:      req.ExpireAt,
		DeviceLimit:   req.DeviceLimit,
		ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
		InboundIDs:    inboundIDs,
		OnHold:        req.OnHold,
		AdminID:       adminID,
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

type bulkCreateUserRequest struct {
	Prefix        string     `json:"prefix"`     // username prefix, e.g. "vpn-"
	Count         int        `json:"count"`      // how many users to create
	Start         int        `json:"start"`      // starting sequence number (default 1)
	Pad           int        `json:"pad"`        // zero-pad width for the number, 0 = none
	Note          string     `json:"note"`       // shared note (the "plan/template")
	DataLimit     int64      `json:"data_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	DeviceLimit   int        `json:"device_limit"`
	ResetStrategy string     `json:"reset_strategy"`
	InboundIDs    []string   `json:"inbound_ids"`
	OnHold        bool       `json:"on_hold"`
}

// BulkCreateUsers provisions many users at once from a shared plan/template,
// 3x-ui style: a username prefix plus a sequential, optionally zero-padded
// counter. Each user is created independently; per-user failures are collected
// and reported without aborting the batch.
func (h *Handlers) BulkCreateUsers(c echo.Context) error {
	var req bulkCreateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Count < 1 || req.Count > 500 {
		return echo.NewHTTPError(http.StatusBadRequest, "count must be between 1 and 500")
	}
	if req.Prefix == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "prefix is required")
	}
	inboundIDs, err := parseUUIDs(req.InboundIDs)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound id")
	}
	if err := h.validateInboundAccess(c, inboundIDs); err != nil {
		return err
	}
	start := req.Start
	if start == 0 {
		start = 1
	}
	adminID := creatorAdminID(c)
	created := make([]*domain.User, 0, req.Count)
	failures := make([]echo.Map, 0)
	for i := 0; i < req.Count; i++ {
		num := start + i
		var username string
		if req.Pad > 0 {
			username = fmt.Sprintf("%s%0*d", req.Prefix, req.Pad, num)
		} else {
			username = fmt.Sprintf("%s%d", req.Prefix, num)
		}
		u, cerr := h.Users.Create(c.Request().Context(), service.CreateUserInput{
			Username:      username,
			Note:          req.Note,
			DataLimit:     req.DataLimit,
			ExpireAt:      req.ExpireAt,
			DeviceLimit:   req.DeviceLimit,
			ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
			InboundIDs:    inboundIDs,
			OnHold:        req.OnHold,
			AdminID:       adminID,
		})
		if u == nil {
			failures = append(failures, echo.Map{"username": username, "error": errString(cerr)})
			continue
		}
		created = append(created, u)
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"created":       created,
		"created_count": len(created),
		"failures":      failures,
	})
}

type importUsersRequest struct {
	Source     string          `json:"source"`      // "3xui" | "marzban"
	Data       json.RawMessage `json:"data"`        // raw export payload
	InboundIDs []string        `json:"inbound_ids"` // optional inbounds to bind imported users to
}

// importedUser is the normalized identity extracted from a foreign panel export.
type importedUser struct {
	Username      string
	DataLimit     int64
	ExpireAt      *time.Time
	DeviceLimit   int
	ResetStrategy string
}

// ImportUsers migrates users from another panel (3x-ui or Marzban). The foreign
// export is parsed into a normalized form and each user is created locally,
// optionally bound to the given inbounds. Imported credentials are regenerated
// (we cannot reuse foreign secrets safely), so clients receive fresh links.
// Per-user failures (e.g. duplicate usernames) are collected, not fatal.
func (h *Handlers) ImportUsers(c echo.Context) error {
	var req importUsersRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	var (
		parsed []importedUser
		err    error
	)
	switch req.Source {
	case "3xui", "3x-ui", "xui":
		parsed, err = parse3xui(req.Data)
	case "marzban":
		parsed, err = parseMarzban(req.Data)
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unknown source (expected 3xui or marzban)")
	}
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "could not parse export: "+err.Error())
	}
	inboundIDs, perr := parseUUIDs(req.InboundIDs)
	if perr != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound id")
	}
	if err := h.validateInboundAccess(c, inboundIDs); err != nil {
		return err
	}

	created := make([]*domain.User, 0, len(parsed))
	failures := make([]echo.Map, 0)
	adminID := creatorAdminID(c)
	for _, p := range parsed {
		if p.Username == "" {
			continue
		}
		u, cerr := h.Users.Create(c.Request().Context(), service.CreateUserInput{
			Username:      p.Username,
			Note:          "imported from " + req.Source,
			DataLimit:     p.DataLimit,
			ExpireAt:      p.ExpireAt,
			DeviceLimit:   p.DeviceLimit,
			ResetStrategy: domain.ResetStrategy(p.ResetStrategy),
			InboundIDs:    inboundIDs,
			AdminID:       adminID,
		})
		if u == nil {
			failures = append(failures, echo.Map{"username": p.Username, "error": errString(cerr)})
			continue
		}
		created = append(created, u)
	}
	return c.JSON(http.StatusCreated, echo.Map{
		"parsed":        len(parsed),
		"created":       created,
		"created_count": len(created),
		"failures":      failures,
	})
}

// parse3xui reads a 3x-ui inbound export. 3x-ui stores clients inside each
// inbound's settings JSON; we accept either the full inbounds list or a bare
// clients array. Fields: email, totalGB (bytes), expiryTime (ms epoch, 0 =
// never), limitIp (device cap), enable.
func parse3xui(data json.RawMessage) ([]importedUser, error) {
	type client struct {
		Email    string `json:"email"`
		TotalGB  int64  `json:"totalGB"`
		ExpiryMs int64  `json:"expiryTime"`
		LimitIP  int    `json:"limitIp"`
		Enable   *bool  `json:"enable"`
	}
	type inbound struct {
		Settings json.RawMessage `json:"settings"`
	}
	collect := func(cs []client) []importedUser {
		out := make([]importedUser, 0, len(cs))
		for _, cl := range cs {
			iu := importedUser{Username: cl.Email, DataLimit: cl.TotalGB, DeviceLimit: cl.LimitIP}
			if cl.ExpiryMs > 0 {
				t := time.UnixMilli(cl.ExpiryMs)
				iu.ExpireAt = &t
			}
			out = append(out, iu)
		}
		return out
	}
	// 3x-ui settings may be a JSON-encoded string or an object.
	parseSettings := func(raw json.RawMessage) []client {
		var asObj struct {
			Clients []client `json:"clients"`
		}
		if json.Unmarshal(raw, &asObj) == nil && len(asObj.Clients) > 0 {
			return asObj.Clients
		}
		var asStr string
		if json.Unmarshal(raw, &asStr) == nil {
			_ = json.Unmarshal([]byte(asStr), &asObj)
		}
		return asObj.Clients
	}

	// Try: bare clients array.
	var bare []client
	if json.Unmarshal(data, &bare) == nil && len(bare) > 0 {
		return collect(bare), nil
	}
	// Try: single settings object with clients.
	if cs := parseSettings(data); len(cs) > 0 {
		return collect(cs), nil
	}
	// Try: inbounds list, each with settings.
	var inbounds []inbound
	if err := json.Unmarshal(data, &inbounds); err != nil {
		// Try wrapper { "obj": [ ... ] } (3x-ui /panel/api export shape).
		var wrap struct {
			Obj []inbound `json:"obj"`
		}
		if json.Unmarshal(data, &wrap) == nil {
			inbounds = wrap.Obj
		} else {
			return nil, err
		}
	}
	var all []importedUser
	for _, ib := range inbounds {
		all = append(all, collect(parseSettings(ib.Settings))...)
	}
	return all, nil
}

// parseMarzban reads a Marzban users export. Accepts either the /api/users
// response ({ "users": [...] }) or a bare users array. Fields: username,
// data_limit (bytes, 0 = unlimited), expire (unix seconds, null/0 = never),
// data_limit_reset_strategy.
func parseMarzban(data json.RawMessage) ([]importedUser, error) {
	type muser struct {
		Username      string `json:"username"`
		DataLimit     int64  `json:"data_limit"`
		Expire        *int64 `json:"expire"`
		ResetStrategy string `json:"data_limit_reset_strategy"`
	}
	var users []muser
	if err := json.Unmarshal(data, &users); err != nil || len(users) == 0 {
		var wrap struct {
			Users []muser `json:"users"`
		}
		if werr := json.Unmarshal(data, &wrap); werr != nil {
			if err != nil {
				return nil, err
			}
			return nil, werr
		}
		users = wrap.Users
	}
	out := make([]importedUser, 0, len(users))
	for _, m := range users {
		iu := importedUser{Username: m.Username, DataLimit: m.DataLimit, ResetStrategy: m.ResetStrategy}
		if m.Expire != nil && *m.Expire > 0 {
			t := time.Unix(*m.Expire, 0)
			iu.ExpireAt = &t
		}
		out = append(out, iu)
	}
	return out, nil
}

type updateUserRequest struct {
	Note          string     `json:"note"`
	Status        string     `json:"status"`
	DataLimit     int64      `json:"data_limit"`
	ExpireAt      *time.Time `json:"expire_at"`
	DeviceLimit   int        `json:"device_limit"`
	ResetStrategy string     `json:"reset_strategy"`
	InboundIDs    *[]string  `json:"inbound_ids"` // nil/omitted = leave bindings unchanged
}

// UpdateUser replaces a user's mutable metadata.
func (h *Handlers) UpdateUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
	}
	var req updateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	status := domain.UserStatus(req.Status)
	if status == "" {
		status = domain.UserStatusActive
	}
	var inboundIDs []uuid.UUID
	if req.InboundIDs != nil {
		parsed, perr := parseUUIDs(*req.InboundIDs)
		if perr != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid inbound id")
		}
		inboundIDs = parsed
		if inboundIDs == nil {
			inboundIDs = []uuid.UUID{} // non-nil empty = clear all bindings
		}
		if err := h.validateInboundAccess(c, inboundIDs); err != nil {
			return err
		}
	}
	u, err := h.Users.Update(c.Request().Context(), id, service.UpdateUserInput{
		Note: req.Note, Status: status, DataLimit: req.DataLimit, ExpireAt: req.ExpireAt,
		DeviceLimit: req.DeviceLimit, ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
		InboundIDs: inboundIDs,
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

// ListUsers returns a paginated, filterable list. Non-sudo admins only see
// users they created (reseller scoping).
func (h *Handlers) ListUsers(c echo.Context) error {
	f := port.UserFilter{
		Search: c.QueryParam("search"),
		Status: domain.UserStatus(c.QueryParam("status")),
		Limit:  atoiDefault(c.QueryParam("limit"), 50),
		Offset: atoiDefault(c.QueryParam("offset"), 0),
	}
	// Reseller scoping: non-sudo admins only see their own users.
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		id := claims.AdminID
		f.AdminID = &id
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
	u, err := h.assertUserOwned(c, id)
	if err != nil {
		return err
	}
	// Include the user's current inbound bindings so the UI can pre-select them.
	inboundIDs := []string{}
	if ins, err := h.Repo.InboundsFor(c.Request().Context(), id); err == nil {
		for _, in := range ins {
			inboundIDs = append(inboundIDs, in.ID.String())
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"user": u, "inbound_ids": inboundIDs})
}

// GetUserUsage returns a user's bucketed traffic series. Defaults to the last 7
// days in daily buckets; override with ?from=&to= (unix seconds) and ?bucket=.
func (h *Handlers) GetUserUsage(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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

// GetTrafficSeries returns fleet-wide bucketed traffic for the dashboard chart.
// Defaults to the last hour in 1-minute buckets; override with ?from=&to= (unix
// seconds) and ?bucket=.
func (h *Handlers) GetTrafficSeries(c echo.Context) error {
	now := time.Now()
	q := port.SeriesQuery{
		FromUnix: int64(atoiDefault(c.QueryParam("from"), int(now.Add(-time.Hour).Unix()))),
		ToUnix:   int64(atoiDefault(c.QueryParam("to"), int(now.Unix()))),
		Bucket:   c.QueryParam("bucket"),
	}
	if q.Bucket == "" {
		q.Bucket = "1m"
	}
	points, err := h.Traffic.TotalSeries(c.Request().Context(), q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "traffic query failed")
	}
	return c.JSON(http.StatusOK, echo.Map{"points": points})
}

// DeleteUser de-provisions and removes a user.
func (h *Handlers) DeleteUser(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
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

// GetUserOnlineIPs lists the distinct source IPs a user is currently connected
// from (across all its nodes), with each IP's last-seen time. A high count is a
// strong signal of account sharing.
func (h *Handlers) GetUserOnlineIPs(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	if _, err := h.assertUserOwned(c, id); err != nil {
		return err
	}
	ips, tracked, err := h.Users.OnlineIPList(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "online ip lookup failed")
	}
	if ips == nil {
		ips = []service.OnlineIP{}
	}
	return c.JSON(http.StatusOK, echo.Map{"ips": ips, "count": len(ips), "tracking": tracked})
}

// --- helpers ---

// creatorAdminID tags new users with the authenticated admin (reseller ownership).
func creatorAdminID(c echo.Context) *uuid.UUID {
	if claims := claimsFrom(c); claims != nil {
		id := claims.AdminID
		return &id
	}
	return nil
}

// assertUserOwned loads a user and enforces reseller scoping: non-sudo admins
// may only access users they created.
func (h *Handlers) assertUserOwned(c echo.Context, id uuid.UUID) (*domain.User, error) {
	u, err := h.Repo.GetByID(c.Request().Context(), id)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
	}
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "fetch failed")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		if u.AdminID == nil || *u.AdminID != claims.AdminID {
			return nil, echo.NewHTTPError(http.StatusNotFound, "user not found")
		}
	}
	return u, nil
}

// validateInboundAccess ensures resellers only bind users to their allowlist.
func (h *Handlers) validateInboundAccess(c echo.Context, inboundIDs []uuid.UUID) error {
	claims := claimsFrom(c)
	if claims == nil || claims.Sudo {
		return nil
	}
	if err := h.Admins.ValidateInboundAccess(c.Request().Context(), claims.AdminID, inboundIDs); err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "inbound not allowed for your account")
	}
	return nil
}

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
