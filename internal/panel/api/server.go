package api

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
)

// Deps are the dependencies needed to build the HTTP API.
type Deps struct {
	Handlers *Handlers
	Issuer   *auth.Issuer
	Auth     *service.AuthService
	Limiter  RateLimiter // nil disables login rate limiting
}

// NewRouter builds the Echo instance with all routes and middleware mounted.
// Public routes: POST /api/login. Everything else requires a valid token and,
// where it mutates, the matching RBAC permission.
func NewRouter(d Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	api := e.Group("/api")
	api.GET("/health", func(c echo.Context) error { return c.JSON(200, echo.Map{"status": "ok"}) })
	// Throttle login by client IP to blunt credential brute-forcing.
	api.POST("/login", d.Handlers.Login, RateLimit(d.Limiter, 10, time.Minute, func(c echo.Context) string {
		return "login:" + c.RealIP()
	}))

	// Public subscription endpoint, authenticated only by the opaque token.
	e.GET("/sub/:token", d.Handlers.Subscribe)

	// Authenticated subtree.
	authed := api.Group("", RequireAuth(d.Issuer))

	users := authed.Group("/users")
	users.GET("", d.Handlers.ListUsers, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id", d.Handlers.GetUser, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/usage", d.Handlers.GetUserUsage, RequirePermission(d.Auth, domain.PermUserRead))
	users.POST("", d.Handlers.CreateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.PUT("/:id", d.Handlers.UpdateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.DELETE("/:id", d.Handlers.DeleteUser, RequirePermission(d.Auth, domain.PermUserWrite))

	nodes := authed.Group("/nodes")
	nodes.GET("", d.Handlers.ListNodes, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id", d.Handlers.GetNode, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.POST("", d.Handlers.CreateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.PUT("/:id", d.Handlers.UpdateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.DELETE("/:id", d.Handlers.DeleteNode, RequirePermission(d.Auth, domain.PermNodeWrite))

	inbounds := authed.Group("/inbounds")
	inbounds.GET("", d.Handlers.ListInbounds, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.POST("", d.Handlers.CreateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.PUT("/:id", d.Handlers.UpdateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.DELETE("/:id", d.Handlers.DeleteInbound, RequirePermission(d.Auth, domain.PermInboundWrite))

	// Admin + role management, gated behind the admin:manage permission (sudo bypasses).
	admins := authed.Group("/admins", RequirePermission(d.Auth, domain.PermAdminManage))
	admins.GET("", d.Handlers.ListAdmins)
	admins.POST("", d.Handlers.CreateAdmin)
	admins.PUT("/:id", d.Handlers.UpdateAdmin)
	admins.DELETE("/:id", d.Handlers.DeleteAdmin)

	roles := authed.Group("/roles", RequirePermission(d.Auth, domain.PermAdminManage))
	roles.GET("", d.Handlers.ListRoles)
	roles.POST("", d.Handlers.CreateRole)

	// Self-service account actions: any authenticated admin manages their own 2FA.
	account := authed.Group("/account")
	account.POST("/2fa/setup", d.Handlers.SetupTOTP)
	account.POST("/2fa/confirm", d.Handlers.ConfirmTOTP)
	account.POST("/2fa/disable", d.Handlers.DisableTOTP)

	return e
}
