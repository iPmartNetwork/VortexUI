package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimitConfig defines per-endpoint rate limit settings.
type RateLimitConfig struct {
	// Requests per window.
	Limit int
	// Window duration.
	Window time.Duration
}

// DefaultRateLimits defines sensible per-endpoint defaults.
var DefaultRateLimits = map[string]RateLimitConfig{
	"POST /api/login":         {Limit: 5, Window: time.Minute},  // 5 login attempts per minute
	"POST /api/users":         {Limit: 30, Window: time.Minute}, // 30 user creates per minute
	"POST /api/users/bulk":    {Limit: 5, Window: time.Minute},  // 5 bulk creates per minute
	"POST /api/inbounds":      {Limit: 20, Window: time.Minute}, // 20 inbound creates per minute
	"POST /api/inbounds/bulk": {Limit: 5, Window: time.Minute},  // 5 bulk actions per minute
	"GET /sub/":               {Limit: 60, Window: time.Minute}, // 60 subscription fetches per minute
}

// RateLimitMiddleware applies per-endpoint rate limiting based on client IP.
// It uses the existing RateLimiter interface declared in middleware.go. If
// limiter is nil, the middleware is a no-op (fail open for environments
// without Redis).
func RateLimitMiddleware(limiter RateLimiter) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if limiter == nil {
				return next(c)
			}
			method := c.Request().Method
			path := c.Path()
			lookupKey := method + " " + path

			cfg, ok := DefaultRateLimits[lookupKey]
			if !ok {
				// No specific limit for this endpoint — apply a generous default.
				cfg = RateLimitConfig{Limit: 120, Window: time.Minute}
			}

			ip := c.RealIP()
			key := fmt.Sprintf("rl:%s:%s", ip, lookupKey)

			allowed, _, err := limiter.Allow(c.Request().Context(), key, cfg.Limit, cfg.Window)
			if err != nil {
				// On Redis error, fail open (allow request).
				return next(c)
			}
			if !allowed {
				return echo.NewHTTPError(http.StatusTooManyRequests, "rate limit exceeded")
			}
			return next(c)
		}
	}
}
