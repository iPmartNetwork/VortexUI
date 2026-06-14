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
	api.GET("/health", func(c echo.Context) error { return c.JSON(200, echo.Map{"status": "ok"}) }) // Throttle login by client IP to blunt credential brute-forcing.
	api.POST("/login", d.Handlers.Login, RateLimit(d.Limiter, 10, time.Minute, func(c echo.Context) string {
		return "login:" + c.RealIP()
	}))

	// Public subscription endpoint, authenticated only by the opaque token.
	e.GET("/sub/:token", d.Handlers.Subscribe)

	// Authenticated subtree.
	authed := api.Group("", RequireAuth(d.Issuer))

	// Dashboard overview: user aggregates + node fleet health.
	authed.GET("/overview", d.Handlers.GetOverview, RequirePermission(d.Auth, domain.PermSystemRead))
	// Recent panel logs (in-memory ring buffer).
	authed.GET("/logs", d.Handlers.GetLogs, RequirePermission(d.Auth, domain.PermSystemRead))

	users := authed.Group("/users")
	users.GET("", d.Handlers.ListUsers, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id", d.Handlers.GetUser, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/usage", d.Handlers.GetUserUsage, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/sub", d.Handlers.GetUserSubscription, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/online", d.Handlers.GetUserOnline, RequirePermission(d.Auth, domain.PermUserRead))
	users.POST("", d.Handlers.CreateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.PUT("/:id", d.Handlers.UpdateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/reset", d.Handlers.ResetUserUsage, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/revoke-sub", d.Handlers.RevokeUserSub, RequirePermission(d.Auth, domain.PermUserWrite))
	users.DELETE("/:id", d.Handlers.DeleteUser, RequirePermission(d.Auth, domain.PermUserWrite))

	nodes := authed.Group("/nodes")
	nodes.GET("", d.Handlers.ListNodes, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id", d.Handlers.GetNode, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/logs", d.Handlers.GetNodeLogs, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.POST("", d.Handlers.CreateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.PUT("/:id", d.Handlers.UpdateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.DELETE("/:id", d.Handlers.DeleteNode, RequirePermission(d.Auth, domain.PermNodeWrite))

	inbounds := authed.Group("/inbounds")
	inbounds.GET("", d.Handlers.ListInbounds, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.POST("", d.Handlers.CreateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.PUT("/:id", d.Handlers.UpdateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.DELETE("/:id", d.Handlers.DeleteInbound, RequirePermission(d.Auth, domain.PermInboundWrite))

	// Egress + steering policy share the inbound (core-config) permission.
	outbounds := authed.Group("/outbounds")
	outbounds.GET("", d.Handlers.ListOutbounds, RequirePermission(d.Auth, domain.PermInboundRead))
	outbounds.POST("", d.Handlers.CreateOutbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	outbounds.PUT("/:id", d.Handlers.UpdateOutbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	outbounds.DELETE("/:id", d.Handlers.DeleteOutbound, RequirePermission(d.Auth, domain.PermInboundWrite))

	routing := authed.Group("/routing")
	routing.GET("", d.Handlers.ListRoutingRules, RequirePermission(d.Auth, domain.PermInboundRead))
	routing.POST("", d.Handlers.CreateRoutingRule, RequirePermission(d.Auth, domain.PermInboundWrite))
	routing.PUT("/:id", d.Handlers.UpdateRoutingRule, RequirePermission(d.Auth, domain.PermInboundWrite))
	routing.DELETE("/:id", d.Handlers.DeleteRoutingRule, RequirePermission(d.Auth, domain.PermInboundWrite))

	balancers := authed.Group("/balancers")
	balancers.GET("", d.Handlers.ListBalancers, RequirePermission(d.Auth, domain.PermInboundRead))
	balancers.POST("", d.Handlers.CreateBalancer, RequirePermission(d.Auth, domain.PermInboundWrite))
	balancers.PUT("/:id", d.Handlers.UpdateBalancer, RequirePermission(d.Auth, domain.PermInboundWrite))
	balancers.DELETE("/:id", d.Handlers.DeleteBalancer, RequirePermission(d.Auth, domain.PermInboundWrite))

	// REALITY key generation helper for building reality inbounds.
	authed.GET("/reality/keypair", d.Handlers.GenerateReality, RequirePermission(d.Auth, domain.PermInboundWrite))

	// Configuration backup/restore — sudo-level (admin:manage); restore is destructive.
	backup := authed.Group("/backup", RequirePermission(d.Auth, domain.PermAdminManage))
	backup.GET("", d.Handlers.GetBackup)
	backup.POST("/restore", d.Handlers.RestoreBackup)

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
