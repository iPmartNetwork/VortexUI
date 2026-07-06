// Package api is the panel's HTTP surface: a thin Echo layer that authenticates
// requests, enforces RBAC, and delegates all logic to the service layer.
package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// claimsKey is the echo context key under which verified claims are stored.
const claimsKey = "vortex.claims"

const auditInsertTimeout = 5 * time.Second

// Authenticator verifies bearer tokens. *auth.PanelAuth satisfies it.
type Authenticator interface {
	Verify(ctx context.Context, token string) (*auth.Claims, error)
}

// RequireAuth verifies the Authorization: Bearer token and stashes the claims
// for downstream handlers. Missing/invalid tokens get 401.
func RequireAuth(a Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			raw := bearerToken(c.Request().Header.Get("Authorization"))
			if raw == "" {
				// EventSource (SSE) cannot set headers, so it passes the token as
				// a query parameter instead.
				raw = c.QueryParam("access_token")
			}
			if raw == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing bearer token")
			}
			claims, err := a.Verify(c.Request().Context(), raw)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}
			c.Set(claimsKey, claims)
			return next(c)
		}
	}
}

// RequireActiveAdmin blocks suspended reseller accounts on authenticated routes.
func RequireActiveAdmin(admins *service.AdminService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := claimsFrom(c)
			if claims == nil || claims.Sudo {
				return next(c)
			}
			if err := admins.EnsureActiveAdmin(c.Request().Context(), claims.AdminID); errors.Is(err, service.ErrAdminSuspended) {
				return echo.NewHTTPError(http.StatusForbidden, "account suspended")
			} else if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "authorization failed")
			}
			return next(c)
		}
	}
}

// RequirePermission gates a route behind an RBAC permission, evaluated against
// the authenticated admin's role (sudo bypasses). Must run after RequireAuth.
func RequirePermission(authz *service.AuthService, p domain.Permission) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims := claimsFrom(c)
			if claims == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "not authenticated")
			}
			ok, err := authz.Authorize(c.Request().Context(), claims, p)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "authorization failed")
			}
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient permission")
			}
			return next(c)
		}
	}
}

// RateLimiter throttles requests by key. *redis.RateLimiter satisfies it; nil
// disables limiting (e.g. when Redis is not configured).
type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, time.Duration, error)
}

// RateLimit caps requests sharing a key (e.g. client IP) to limit per window. It
// fails open: if the limiter errors (Redis blip), the request proceeds rather
// than locking everyone out. A nil limiter is a no-op.
func RateLimit(rl RateLimiter, limit int, window time.Duration, keyFn func(echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl == nil {
				return next(c)
			}
			ok, retry, err := rl.Allow(c.Request().Context(), keyFn(c), limit, window)
			if err != nil {
				return next(c) // fail open
			}
			if !ok {
				c.Response().Header().Set("Retry-After", strconv.Itoa(int(retry.Seconds())))
				return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
			}
			return next(c)
		}
	}
}

// AuditRecorder persists and reads admin action audit entries.
// *postgres.AuditRepo satisfies it.
type AuditRecorder interface {
	Insert(ctx context.Context, e domain.AuditEntry) error
	List(ctx context.Context, limit, offset int) ([]domain.AuditEntry, error)
	ListForAdmin(ctx context.Context, adminID uuid.UUID, limit, offset int) ([]domain.AuditEntry, error)
}

// Audit records every authenticated mutating request (POST/PUT/DELETE) for
// accountability. It runs after the handler so it captures the final status
// (including permission denials), and writes asynchronously so logging never
// adds latency to the response.
func Audit(rec AuditRecorder) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if rec == nil {
				return err
			}
			switch c.Request().Method {
			case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
			default:
				return err
			}
			entry := domain.AuditEntry{
				ID:     uuid.New(),
				Method: c.Request().Method,
				Path:   c.Request().URL.Path,
				Status: c.Response().Status,
				IP:     c.RealIP(),
			}
			if claims := claimsFrom(c); claims != nil {
				id := claims.AdminID
				entry.AdminID = &id
				entry.ImpersonatorID = claims.ImpersonatorID
			}
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), auditInsertTimeout)
				defer cancel()
				_ = rec.Insert(ctx, entry)
			}()
			return err
		}
	}
}

func bearerToken(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}

func claimsFrom(c echo.Context) *auth.Claims {
	v, _ := c.Get(claimsKey).(*auth.Claims)
	return v
}
