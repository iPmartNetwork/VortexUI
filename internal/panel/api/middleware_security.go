package api

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// SecurityMiddleware integrates advanced security checks into the request pipeline.
type SecurityMiddleware struct {
	svc *service.SecurityAdvancedService
}

// NewSecurityMiddleware creates the middleware with the security service.
func NewSecurityMiddleware(svc *service.SecurityAdvancedService) *SecurityMiddleware {
	return &SecurityMiddleware{svc: svc}
}

// IPAccessCheck returns middleware that verifies the client IP is not banned
// and is whitelisted (if a whitelist is configured for the admin).
func (m *SecurityMiddleware) IPAccessCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			adminID := getAdminIDFromContext(c)

			allowed, reason, err := m.svc.CheckIPAccess(c.Request().Context(), adminID, ip)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "security check failed")
			}
			if !allowed {
				return echo.NewHTTPError(http.StatusForbidden, reason)
			}

			return next(c)
		}
	}
}

// AccountLockoutCheck returns middleware that blocks locked-out accounts.
func (m *SecurityMiddleware) AccountLockoutCheck() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			adminID := getAdminIDFromContext(c)
			if adminID == uuid.Nil {
				return next(c)
			}

			locked, err := m.svc.CheckAccountLocked(c.Request().Context(), adminID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "lockout check failed")
			}
			if locked {
				return echo.NewHTTPError(http.StatusTooManyRequests, "account temporarily locked due to failed login attempts")
			}

			return next(c)
		}
	}
}

// SessionTracker returns middleware that updates session last_active timestamp.
func (m *SecurityMiddleware) SessionTracker() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			sessionID := getSessionIDFromContext(c)
			if sessionID != uuid.Nil {
				_ = m.svc.TouchSession(c.Request().Context(), sessionID)
			}
			return next(c)
		}
	}
}

// TokenScopeCheck returns middleware that validates token scopes for the endpoint.
func (m *SecurityMiddleware) TokenScopeCheck(requiredScope string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			scopes := getTokenScopesFromContext(c)
			if scopes == nil {
				// No scoped token, use full access (session-based auth).
				return next(c)
			}
			if !m.svc.CheckTokenScope(scopes, requiredScope) {
				return echo.NewHTTPError(http.StatusForbidden, "insufficient token scope: "+requiredScope+" required")
			}
			return next(c)
		}
	}
}

// LoginAuditRecorder returns middleware that records login attempts on auth endpoints.
func (m *SecurityMiddleware) LoginAuditRecorder() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			// Only audit POST /api/*/login paths
			if c.Request().Method != http.MethodPost || !strings.Contains(c.Path(), "login") {
				return err
			}

			entry := &domain.LoginAuditEntry{
				IPAddress: c.RealIP(),
				UserAgent: c.Request().UserAgent(),
				Username:  c.Get("auth_username_attempt").(string),
			}

			if adminVal := c.Get("admin_id"); adminVal != nil {
				if aid, ok := adminVal.(uuid.UUID); ok {
					entry.AdminID = &aid
					entry.Success = true
				}
			}
			if err != nil {
				entry.Success = false
				entry.FailureReason = err.Error()
			}

			_ = m.svc.RecordLogin(c.Request().Context(), entry)
			return err
		}
	}
}

// --- context helpers ---

func getAdminIDFromContext(c echo.Context) uuid.UUID {
	if val := c.Get("admin_id"); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func getSessionIDFromContext(c echo.Context) uuid.UUID {
	if val := c.Get("session_id"); val != nil {
		switch v := val.(type) {
		case uuid.UUID:
			return v
		case string:
			if id, err := uuid.Parse(v); err == nil {
				return id
			}
		}
	}
	return uuid.Nil
}

func getTokenScopesFromContext(c echo.Context) []string {
	if val := c.Get("token_scopes"); val != nil {
		if scopes, ok := val.([]string); ok {
			return scopes
		}
	}
	return nil
}
