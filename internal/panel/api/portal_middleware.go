package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/auth"
)

// RequirePortalAuth is middleware that validates a portal JWT and stores portal
// claims in context. This is separate from the admin auth middleware.
func RequirePortalAuth(issuer *auth.Issuer) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			if header == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing token")
			}
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header")
			}
			claims, err := issuer.VerifyPortal(parts[1])
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
			}
			c.Set("portal_claims", claims)
			return next(c)
		}
	}
}
