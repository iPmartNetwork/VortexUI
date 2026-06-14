package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// Handlers bundles the service dependencies the HTTP routes need.
type Handlers struct {
	Auth     *service.AuthService
	Users    *service.UserService
	Sub      *service.SubscriptionService
	Nodes    *service.NodeService
	Inbounds *service.InboundService
	Admins   *service.AdminService
	Devices  DeviceLimiter // optional; nil disables device-count enforcement
	Repo     port.UserRepository
	Traffic  port.TrafficRepository
}

// DeviceLimiter caps the number of distinct devices a user may use within a
// window. *redis.DeviceTracker satisfies it.
type DeviceLimiter interface {
	Allow(ctx context.Context, userID, deviceID string, limit int, window time.Duration) (bool, error)
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
