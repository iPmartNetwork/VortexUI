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
	Version    string // panel build version, threaded to Handlers for GET /api/version
	Handlers   *Handlers
	APITokens  *APITokenHandlers
	Portal     *PortalHandlers
	Reality    *RealityHandlers
	CleanIP    *CleanIPHandlers
	SubHosts   *SubHostHandlers
	RoutingPacks *RoutingPackHandlers
	Quota      *QuotaHandlers
	Relay      *RelayHandlers
	Decoy      *DecoyHandlers
	Analytics  *AnalyticsHandlers
	Migration  *MigrationHandlers
	Probing    *ProbingHandlers
	Family     *FamilyHandlers
	Referral   *ReferralHandlers
	DoH        *DoHHandlers
	SNI        *SNIHandlers
	TLSTricks  *TLSTricksHandlers
	Fingerprint *FingerprintHandlers
	Federation *FederationHandlers
	DeepLink   *DeepLinkHandlers
	QuotaNotify *QuotaNotifyHandlers
	IPLimit    *IPLimitHandlers
	SubSettings *SubSettingsHandlers
	Monitor    *MonitorHandlers
	Issuer     *auth.Issuer
	Auth       *service.AuthService
	Limiter    RateLimiter   // nil disables login rate limiting
	Audit      AuditRecorder // nil disables audit logging
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

	// Thread the build version into the handlers so GET /api/version can report it.
	if d.Handlers != nil {
		d.Handlers.Version = d.Version
	}

	api := e.Group("/api")
	api.GET("/health", func(c echo.Context) error { return c.JSON(200, echo.Map{"status": "ok"}) }) // Throttle login by client IP to blunt credential brute-forcing.
	api.POST("/login", d.Handlers.Login, RateLimit(d.Limiter, 10, time.Minute, func(c echo.Context) string {
		return "login:" + c.RealIP()
	}))

	// Public subscription endpoint, authenticated only by the opaque token.
	e.GET("/sub/:token", d.Handlers.Subscribe)
	e.GET("/sub/:token/info", d.Handlers.SubscriptionInfoPage)
	e.GET("/sub/:token/usage", d.Handlers.SubscriptionUsage)
	e.GET("/sub/:token/wireguard", d.Handlers.SubscribeWireGuard)

	// Authenticated subtree. The audit middleware records every mutating request.
	authed := api.Group("", RequireAuth(d.Issuer), Audit(d.Audit))

	// Dashboard overview: user aggregates + node fleet health.
	authed.GET("/overview", d.Handlers.GetOverview, RequirePermission(d.Auth, domain.PermSystemRead))
	authed.GET("/traffic/series", d.Handlers.GetTrafficSeries, RequirePermission(d.Auth, domain.PermSystemRead))
	// Live event stream (SSE). Token may be passed as ?access_token= since the
	// browser EventSource API cannot set an Authorization header.
	authed.GET("/events/stream", d.Handlers.StreamEvents, RequirePermission(d.Auth, domain.PermSystemRead))
	// Recent panel logs (in-memory ring buffer).
	authed.GET("/logs", d.Handlers.GetLogs, RequirePermission(d.Auth, domain.PermSystemRead))
	// Live system info (process/memory).
	authed.GET("/system", d.Handlers.GetSystem, RequirePermission(d.Auth, domain.PermSystemRead))
	// Panel build version (for the UI footer).
	authed.GET("/version", d.Handlers.GetVersion, RequirePermission(d.Auth, domain.PermSystemRead))

	// Live connection monitor.
	authed.GET("/monitor/connections", d.Monitor.GetLiveConnections, RequirePermission(d.Auth, domain.PermSystemRead))

	users := authed.Group("/users")
	users.GET("", d.Handlers.ListUsers, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id", d.Handlers.GetUser, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/usage", d.Handlers.GetUserUsage, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/sub", d.Handlers.GetUserSubscription, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/online", d.Handlers.GetUserOnline, RequirePermission(d.Auth, domain.PermUserRead))
	users.GET("/:id/online-ips", d.Handlers.GetUserOnlineIPs, RequirePermission(d.Auth, domain.PermUserRead))
	users.POST("", d.Handlers.CreateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/bulk", d.Handlers.BulkCreateUsers, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/import", d.Handlers.ImportUsers, RequirePermission(d.Auth, domain.PermUserWrite))
	users.PUT("/:id", d.Handlers.UpdateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/reset", d.Handlers.ResetUserUsage, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/revoke-sub", d.Handlers.RevokeUserSub, RequirePermission(d.Auth, domain.PermUserWrite))
	users.DELETE("/:id", d.Handlers.DeleteUser, RequirePermission(d.Auth, domain.PermUserWrite))

	nodes := authed.Group("/nodes")
	nodes.GET("", d.Handlers.ListNodes, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id", d.Handlers.GetNode, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/logs", d.Handlers.GetNodeLogs, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/status", d.Handlers.GetNodeStatus, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.POST("/:id/restart", d.Handlers.RestartNodeCore, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("/:id/geo-update", d.Handlers.UpdateNodeGeo, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("/:id/stop", d.Handlers.StopNodeCore, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("", d.Handlers.CreateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.PUT("/:id", d.Handlers.UpdateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.DELETE("/:id", d.Handlers.DeleteNode, RequirePermission(d.Auth, domain.PermNodeWrite))

	// Per-core capability matrix (protocols/transports/securities/udp-native),
	// consumed by the UI to filter inbound-form options to the node's core.
	authed.GET("/capabilities", d.Handlers.GetCapabilities, RequirePermission(d.Auth, domain.PermInboundRead))

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

	// Audit log of mutating admin actions.
	authed.GET("/audit", d.Handlers.ListAudit, RequirePermission(d.Auth, domain.PermAdminManage))

	// Self-service account actions: any authenticated admin manages their own 2FA.
	account := authed.Group("/account")
	account.GET("", d.Handlers.GetAccount)
	account.POST("/password", d.Handlers.ChangePassword)
	account.POST("/2fa/setup", d.Handlers.SetupTOTP)
	account.POST("/2fa/confirm", d.Handlers.ConfirmTOTP)
	account.POST("/2fa/disable", d.Handlers.DisableTOTP)

	// API Tokens (Personal Access Tokens)
	tokens := authed.Group("/tokens")
	tokens.GET("", d.APITokens.ListAPITokens)
	tokens.POST("", d.APITokens.CreateAPIToken)
	tokens.DELETE("/:id", d.APITokens.DeleteAPIToken)

	// Plans + Orders (admin)
	plans := authed.Group("/plans")
	plans.GET("", d.Handlers.ListPlans, RequirePermission(d.Auth, domain.PermAdminManage))
	plans.POST("", d.Handlers.CreatePlan, RequirePermission(d.Auth, domain.PermAdminManage))
	plans.DELETE("/:id", d.Handlers.DeletePlan, RequirePermission(d.Auth, domain.PermAdminManage))

	orders := authed.Group("/orders")
	orders.GET("", d.Handlers.ListOrders, RequirePermission(d.Auth, domain.PermAdminManage))

	// Public payment endpoints (no auth — user self-purchase)
	e.GET("/api/shop/plans", d.Handlers.PublicPlans)
	e.POST("/api/shop/purchase", d.Handlers.InitPurchase)
	e.GET("/api/payment/callback", d.Handlers.PaymentCallback)
	e.POST("/api/payment/ipn/nowpayments", d.Handlers.NowPaymentsIPN)

	// Update checker (admin)
	authed.GET("/update/check", d.Handlers.CheckUpdate, RequirePermission(d.Auth, domain.PermAdminManage))

	// --- Portal: end-user self-service ---
	e.POST("/api/portal/login", d.Portal.PortalLogin)
	portal := e.Group("/api/portal", RequirePortalAuth(d.Issuer))
	portal.GET("/dashboard", d.Portal.PortalDashboard)
	portal.GET("/plans", d.Portal.PortalListPlans)
	portal.GET("/tickets", d.Portal.PortalListTickets)
	portal.POST("/tickets", d.Portal.PortalCreateTicket)
	portal.GET("/tickets/:id", d.Portal.PortalGetTicket)
	portal.POST("/tickets/:id/reply", d.Portal.PortalReplyTicket)

	// --- Admin ticket management ---
	tickets := authed.Group("/tickets", RequirePermission(d.Auth, domain.PermUserWrite))
	tickets.GET("", d.Portal.AdminListTickets)
	tickets.GET("/:id", d.Portal.AdminGetTicket)
	tickets.POST("/:id/reply", d.Portal.AdminReplyTicket)
	tickets.POST("/:id/close", d.Portal.AdminCloseTicket)

	// --- Reality Scanner ---
	reality := authed.Group("/reality")
	reality.POST("/scan", d.Reality.ScanReality, RequirePermission(d.Auth, domain.PermInboundWrite))
	reality.GET("/results", d.Reality.GetCachedScans, RequirePermission(d.Auth, domain.PermInboundRead))

	// --- Clean-IP Scanner ---
	cleanip := authed.Group("/clean-ip")
	cleanip.POST("/scan", d.CleanIP.Scan, RequirePermission(d.Auth, domain.PermInboundWrite))
	cleanip.GET("/results", d.CleanIP.GetCached, RequirePermission(d.Auth, domain.PermInboundRead))

	// --- Subscription Hosts (Marzban-style per-inbound overrides) ---
	hosts := authed.Group("/sub-hosts")
	hosts.GET("", d.SubHosts.ListSubHosts, RequirePermission(d.Auth, domain.PermInboundRead))
	hosts.POST("", d.SubHosts.CreateSubHost, RequirePermission(d.Auth, domain.PermInboundWrite))
	hosts.PUT("/:id", d.SubHosts.UpdateSubHost, RequirePermission(d.Auth, domain.PermInboundWrite))
	hosts.DELETE("/:id", d.SubHosts.DeleteSubHost, RequirePermission(d.Auth, domain.PermInboundWrite))
	hosts.POST("/reorder", d.SubHosts.ReorderSubHosts, RequirePermission(d.Auth, domain.PermInboundWrite))

	// --- Smart-routing rule packs (reusable routing rule sets) ---
	packs := authed.Group("/routing-packs")
	packs.GET("", d.RoutingPacks.ListPacks, RequirePermission(d.Auth, domain.PermInboundRead))
	packs.POST("", d.RoutingPacks.CreatePack, RequirePermission(d.Auth, domain.PermInboundWrite))
	packs.PUT("/:id", d.RoutingPacks.UpdatePack, RequirePermission(d.Auth, domain.PermInboundWrite))
	packs.DELETE("/:id", d.RoutingPacks.DeletePack, RequirePermission(d.Auth, domain.PermInboundWrite))
	packs.POST("/apply", d.RoutingPacks.ApplyToNode, RequirePermission(d.Auth, domain.PermInboundWrite))
	packs.PUT("/default", d.RoutingPacks.SetDefault, RequirePermission(d.Auth, domain.PermInboundWrite))
	packs.GET("/user/:user_id", d.RoutingPacks.GetUserPack, RequirePermission(d.Auth, domain.PermUserRead))
	packs.PUT("/user/:user_id", d.RoutingPacks.SetUserPack, RequirePermission(d.Auth, domain.PermUserWrite))

	// --- Smart Quota (Fair Use) ---
	quota := authed.Group("/quota", RequirePermission(d.Auth, domain.PermAdminManage))
	quota.GET("", d.Quota.ListQuotaPolicies)
	quota.POST("", d.Quota.CreateQuotaPolicy)
	quota.PUT("/:id", d.Quota.UpdateQuotaPolicy)
	quota.DELETE("/:id", d.Quota.DeleteQuotaPolicy)

	// --- CDN/Relay Chain Builder ---
	relays := authed.Group("/relays", RequirePermission(d.Auth, domain.PermNodeWrite))
	relays.GET("", d.Relay.ListChains)
	relays.POST("", d.Relay.CreateChain)
	relays.PUT("/:id", d.Relay.UpdateChain)
	relays.DELETE("/:id", d.Relay.DeleteChain)

	// --- Decoy Website ---
	decoys := authed.Group("/decoys", RequirePermission(d.Auth, domain.PermNodeWrite))
	decoys.GET("", d.Decoy.ListDecoys)
	decoys.POST("", d.Decoy.CreateDecoy)
	decoys.PUT("/:id", d.Decoy.UpdateDecoy)
	decoys.DELETE("/:id", d.Decoy.DeleteDecoy)

	// --- Advanced Analytics ---
	authed.GET("/analytics", d.Analytics.GetAnalytics, RequirePermission(d.Auth, domain.PermSystemRead))
	authed.GET("/analytics/export", d.Analytics.ExportAnalyticsCSV, RequirePermission(d.Auth, domain.PermSystemRead))

	// --- Node Auto-Migration ---
	migration := authed.Group("/migration", RequirePermission(d.Auth, domain.PermAdminManage))
	migration.GET("/policy", d.Migration.GetMigrationPolicy)
	migration.PUT("/policy", d.Migration.UpdateMigrationPolicy)
	migration.GET("/events", d.Migration.ListMigrationEvents)

	// --- Active Probing Protection ---
	probing := authed.Group("/probing", RequirePermission(d.Auth, domain.PermAdminManage))
	probing.GET("/policy", d.Probing.GetProbingPolicy)
	probing.PUT("/policy", d.Probing.UpdateProbingPolicy)
	probing.GET("/events", d.Probing.ListProbeEvents)
	probing.GET("/blocked", d.Probing.ListBlockedIPs)
	probing.POST("/unblock", d.Probing.UnblockIP)

	// --- Family/Group Subscriptions ---
	families := authed.Group("/families", RequirePermission(d.Auth, domain.PermUserWrite))
	families.GET("", d.Family.ListGroups)
	families.POST("", d.Family.CreateGroup)
	families.GET("/:id", d.Family.GetGroup)
	families.DELETE("/:id", d.Family.DeleteGroup)
	families.POST("/:id/members", d.Family.AddMember)
	families.DELETE("/:id/members/:uid", d.Family.RemoveMember)

	// --- Invite/Referral System ---
	referrals := authed.Group("/referrals", RequirePermission(d.Auth, domain.PermAdminManage))
	referrals.GET("/config", d.Referral.GetReferralConfig)
	referrals.PUT("/config", d.Referral.UpdateReferralConfig)
	referrals.GET("/codes", d.Referral.ListReferralCodes)
	referrals.GET("/events", d.Referral.ListReferralEvents)

	// Portal referral endpoints (end-user)
	portal.GET("/referral/code", d.Referral.GetMyCode)
	portal.POST("/referral/apply", d.Referral.ApplyReferral)

	// --- DNS-over-HTTPS ---
	doh := authed.Group("/doh", RequirePermission(d.Auth, domain.PermAdminManage))
	doh.GET("/config", d.DoH.GetDoHConfig)
	doh.PUT("/config", d.DoH.UpdateDoHConfig)
	doh.GET("/stats", d.DoH.GetDoHStats)
	doh.GET("/logs", d.DoH.GetDoHLogs)

	// --- Multi-Domain SNI Routing + SSL ---
	sni := authed.Group("/sni", RequirePermission(d.Auth, domain.PermInboundWrite))
	sni.GET("/domains", d.SNI.ListDomains)
	sni.POST("/domains", d.SNI.AddDomain)
	sni.DELETE("/domains/:id", d.SNI.DeleteDomain)
	sni.GET("/certs", d.SNI.ListCerts)
	sni.POST("/certs", d.SNI.IssueCert)
	sni.POST("/certs/:id/renew", d.SNI.RenewCert)
	sni.DELETE("/certs/:id", d.SNI.DeleteCert)
	sni.GET("/routes", d.SNI.ListRoutes)
	sni.POST("/routes", d.SNI.AddRoute)
	sni.DELETE("/routes/:id", d.SNI.DeleteRoute)

	// --- Fragment/TLS Tricks Manager ---
	tricks := authed.Group("/tls-tricks", RequirePermission(d.Auth, domain.PermInboundWrite))
	tricks.GET("", d.TLSTricks.ListProfiles)
	tricks.GET("/presets", d.TLSTricks.GetPresets)
	tricks.POST("", d.TLSTricks.CreateProfile)
	tricks.POST("/preset", d.TLSTricks.CreateFromPreset)
	tricks.PUT("/:id", d.TLSTricks.UpdateProfile)
	tricks.DELETE("/:id", d.TLSTricks.DeleteProfile)

	// --- Client Fingerprint Validator ---
	fp := authed.Group("/fingerprint", RequirePermission(d.Auth, domain.PermAdminManage))
	fp.GET("/policy", d.Fingerprint.GetPolicy)
	fp.PUT("/policy", d.Fingerprint.UpdatePolicy)
	fp.GET("/rules", d.Fingerprint.ListRules)
	fp.POST("/rules", d.Fingerprint.CreateRule)
	fp.DELETE("/rules/:id", d.Fingerprint.DeleteRule)
	fp.GET("/events", d.Fingerprint.ListEvents)

	// --- Multi-Panel Federation ---
	fed := authed.Group("/federation", RequirePermission(d.Auth, domain.PermAdminManage))
	fed.GET("/config", d.Federation.GetConfig)
	fed.PUT("/config", d.Federation.UpdateConfig)
	fed.GET("/peers", d.Federation.ListPeers)
	fed.POST("/peers", d.Federation.AddPeer)
	fed.DELETE("/peers/:id", d.Federation.DeletePeer)
	fed.GET("/events", d.Federation.ListSyncEvents)

	// --- Deep Link + QR ---
	dl := authed.Group("/deeplink", RequirePermission(d.Auth, domain.PermAdminManage))
	dl.GET("/config", d.DeepLink.GetConfig)
	dl.PUT("/config", d.DeepLink.UpdateConfig)
	dl.GET("/generate", d.DeepLink.GenerateLink)

	// --- Subscription auto-update settings ---
	subset := authed.Group("/sub-settings", RequirePermission(d.Auth, domain.PermAdminManage))
	subset.GET("", d.SubSettings.GetSubSettings)
	subset.PUT("", d.SubSettings.UpdateSubSettings)

	// --- Smart Quota Notifications ---
	qn := authed.Group("/quota-notify", RequirePermission(d.Auth, domain.PermAdminManage))
	qn.GET("/config", d.QuotaNotify.GetConfig)
	qn.PUT("/config", d.QuotaNotify.UpdateConfig)
	qn.GET("/events", d.QuotaNotify.ListEvents)

	// --- IP-limit enforcement ---
	iplimit := authed.Group("/ip-limit", RequirePermission(d.Auth, domain.PermAdminManage))
	iplimit.GET("/policy", d.IPLimit.GetPolicy)
	iplimit.PUT("/policy", d.IPLimit.UpdatePolicy)
	iplimit.GET("/events", d.IPLimit.ListEvents)

	return e
}
