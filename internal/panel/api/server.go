package api

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
)

// Deps are the dependencies needed to build the HTTP API.
type Deps struct {
	Version            string // panel build version, threaded to Handlers for GET /api/version
	Handlers           *Handlers
	APITokens          *APITokenHandlers
	Portal             *PortalHandlers
	Reality            *RealityHandlers
	CleanIP            *CleanIPHandlers
	SubHosts           *SubHostHandlers
	RoutingPacks       *RoutingPackHandlers
	Quota              *QuotaHandlers
	Relay              *RelayHandlers
	Decoy              *DecoyHandlers
	Analytics          *AnalyticsHandlers
	Migration          *MigrationHandlers
	Probing            *ProbingHandlers
	Family             *FamilyHandlers
	Referral           *ReferralHandlers
	DoH                *DoHHandlers
	SNI                *SNIHandlers
	TLSTricks          *TLSTricksHandlers
	Fingerprint        *FingerprintHandlers
	Federation         *FederationHandlers
	DeepLink           *DeepLinkHandlers
	QuotaNotify        *QuotaNotifyHandlers
	AdminQuotaNotify   *AdminQuotaNotifyHandlers
	IPLimit            *IPLimitHandlers
	SubSettings        *SubSettingsHandlers
	Monitor            *MonitorHandlers
	WalletBilling      *WalletBillingHandlers
	PaymentConfig      *PaymentConfigHandlers
	ProtocolGroups     *ProtocolGroupHandlers
	Issuer             *auth.Issuer
	PanelAuth          Authenticator
	Auth               *service.AuthService
	Limiter            RateLimiter   // nil disables login rate limiting
	Audit              AuditRecorder // nil disables audit logging
	IPGuard            *IPGuard      // nil disables IP access control
	PanelSettings      *PanelSettingsHandlers
	AuditService       *service.AuditService            // for advanced audit operations
	SessionService     *service.SessionService          // for session management
	MetricsService     *service.MetricsCollectorService // for metrics collection
	HealthService      *service.HealthCheckService      // for health checks
	LoggerService      *service.StructuredLoggerService // for structured logging
	TraceService       *service.TraceManagerService     // for distributed tracing
	PrometheusExporter *service.PrometheusExporter      // for Prometheus export

	// PHASE 3A - Authentication & Hardening (Security)
	TOTPRepository           *postgres.TOTPRepository
	MFARepository            *postgres.MFARepository
	PasswordPolicyRepository *postgres.PasswordPolicyRepository
	IPAccessRuleRepository   *postgres.IPAccessRuleRepository
	LoginAttemptRepository   *postgres.LoginAttemptRepository

	TOTPService           *service.TOTPService
	PasswordPolicyService *service.PasswordPolicyService
	IPValidatorService    *service.IPValidatorService

	MFAHandlers       *MFAHandlers
	PasswordHandlers  *PasswordHandlers
	IPControlHandlers *IPControlHandlers

	MFAMiddleware          *MFAMiddleware
	IPValidationMiddleware *IPValidationMiddleware

	// PHASE 3C - Audit & Compliance
	AuditEventRepository      *postgres.AuditEventRepository
	ComplianceEventRepository *postgres.ComplianceEventRepository
	AuditReportRepository     *postgres.AuditReportRepository
	AuditPolicyRepository     *postgres.AuditPolicyRepository
	AuditArchiveRepository    *postgres.AuditArchiveRepository

	ReportGeneratorService   port.ReportGenerator
	ComplianceCheckerService port.ComplianceChecker

	AuditEventHandlers *AuditEventHandlers
	ComplianceHandlers *ComplianceHandlers
	ReportHandlers     *ReportHandlers

	// PHASE 3B - Performance Optimization
	QueryMetricsRepository     port.QueryMetricsRepository
	RateLimitRepository        port.RateLimitRepository
	PerformanceAlertRepository port.PerformanceAlertRepository

	CacheService       port.CacheService
	PerformanceMonitor port.PerformanceMonitor
	RateLimiterService port.RateLimiter

	PerformanceHandlers *PerformanceHandlers

	// PHASE 3D - Security Hardening & Defense
	SecurityThreatRepository port.SecurityThreatRepository
	SecurityPolicyRepository port.SecurityPolicyRepository
	ThreatDetector           port.ThreatDetector
	AnomalyDetector          port.AnomalyDetector

	SecurityHardeningHandlers *SecurityHandlers

	// PHASE 8-14: Enterprise Feature Handlers
	ConfigManagement *ConfigManagementHandler  // Phase 8: config validation/versioning
	WireGuardMgmt    *WireGuardHandler         // Phase 9: WireGuard peer management
	SecurityAdvanced *SecurityAdvancedHandler   // Phase 10: advanced security
	Jobs             *JobsHandler              // Phase 11: background job status
	Docs             *DocsHandler              // Phase 12: API docs / Swagger UI
	RateLimits       *RateLimitHandler         // Phase 12: rate limit dashboard
	WebhookTest      *WebhookTestHandler       // Phase 12: webhook test endpoint
	PortalPro        *PortalProHandler         // Phase 14: portal pro features
}

// NewRouter builds the Echo instance with all routes and middleware mounted.
// Public routes: POST /api/login. Everything else requires a valid token and,
// where it mutates, the matching RBAC permission.
func NewRouter(d Deps) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level:     5,
		MinLength: 1024, // only compress responses larger than 1 KB
	}))
	e.Use(middleware.Logger())
	if d.IPGuard != nil {
		e.Use(d.IPGuard.Middleware())
	}

	// Thread the build version into the handlers so GET /api/version can report it.
	if d.Handlers != nil {
		d.Handlers.Version = d.Version
	}

	api := e.Group("/api")
	api.GET("/health", func(c echo.Context) error { return c.JSON(200, echo.Map{"status": "ok"}) })
	api.GET("/version", d.Handlers.GetVersion) // Panel build version (public, no auth needed)

	// Prometheus metrics endpoint (unauthenticated for scraper access)
	if d.PrometheusExporter != nil {
		prometheusHandlers := NewPrometheusHandlers(d.PrometheusExporter)
		api.GET("/metrics", prometheusHandlers.GetMetrics)
		e.GET("/metrics", prometheusHandlers.GetMetrics) // Also available at root for compatibility
	}
	api.POST("/login", d.Handlers.Login, RateLimit(d.Limiter, 10, time.Minute, func(c echo.Context) string {
		return "login:" + c.RealIP()
	}))

	// Public subscription endpoint, authenticated only by the opaque token.
	e.GET("/sub/:token", d.Handlers.Subscribe)
	e.GET("/sub/:token/info", d.Handlers.SubscriptionInfoPage)
	e.GET("/sub/:token/usage", d.Handlers.SubscriptionUsage)
	e.GET("/sub/:token/wireguard", d.Handlers.SubscribeWireGuard)
	e.GET("/sub/:token/shop", d.Handlers.SubscriptionShop)
	e.POST("/sub/:token/switch", d.Handlers.ReportSwitch, RateLimit(d.Limiter, 30, time.Minute, func(c echo.Context) string {
		return "switch:" + c.Param("token")
	}))

	// Authenticated subtree. The audit middleware records every mutating request.
	authed := api.Group("", RequireAuth(d.PanelAuth), RequireActiveAdmin(d.Handlers.Admins), Audit(d.Audit))

	// Wire advanced audit middleware if available
	if d.AuditService != nil {
		authed.Use(AuditMiddleware(d.AuditService, slog.Default()))
	}

	// Wire session management middleware if available
	if d.SessionService != nil {
		authed.Use(SessionMiddleware(d.SessionService))
	}

	// Wire metrics collection middleware if available
	if d.MetricsService != nil {
		authed.Use(MetricsMiddleware(d.MetricsService))
	}

	// Wire trace middleware if available
	if d.TraceService != nil {
		authed.Use(TraceMiddleware(d.TraceService, slog.Default()))
	}

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

	// Panel-wide settings (persisted in DB).
	if d.PanelSettings != nil {
		authed.GET("/settings", d.PanelSettings.GetSettings, RequirePermission(d.Auth, domain.PermSystemRead))
		authed.PUT("/settings", d.PanelSettings.UpdateSettings, RequirePermission(d.Auth, domain.PermSystemRead))
	}

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
	users.POST("/bulk-delete", d.Handlers.BulkDeleteUsers, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/import", d.Handlers.ImportUsers, RequirePermission(d.Auth, domain.PermUserWrite))
	users.PUT("/:id", d.Handlers.UpdateUser, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/reset", d.Handlers.ResetUserUsage, RequirePermission(d.Auth, domain.PermUserWrite))
	users.POST("/:id/revoke-sub", d.Handlers.RevokeUserSub, RequirePermission(d.Auth, domain.PermUserWrite))
	users.DELETE("/:id", d.Handlers.DeleteUser, RequirePermission(d.Auth, domain.PermUserWrite))

	nodes := authed.Group("/nodes")
	nodes.GET("/enrollment", d.Handlers.GetNodeEnrollment, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.GET("", d.Handlers.ListNodes, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id", d.Handlers.GetNode, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/logs", d.Handlers.GetNodeLogs, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/status", d.Handlers.GetNodeStatus, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.GET("/:id/debug", d.Handlers.GetNodeDebugBundle, RequirePermission(d.Auth, domain.PermNodeRead))
	nodes.POST("/:id/test", d.Handlers.TestNodeConnection, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("/:id/restart", d.Handlers.RestartNodeCore, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("/:id/geo-update", d.Handlers.UpdateNodeGeo, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("/:id/stop", d.Handlers.StopNodeCore, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.POST("", d.Handlers.CreateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.PUT("/:id", d.Handlers.UpdateNode, RequirePermission(d.Auth, domain.PermNodeWrite))
	nodes.DELETE("/:id", d.Handlers.DeleteNode, RequirePermission(d.Auth, domain.PermNodeWrite))

	// Per-core capability matrix (protocols/transports/securities/udp-native),
	// consumed by the UI to filter inbound-form options to the node's core.
	authed.GET("/capabilities", d.Handlers.GetCapabilities, RequirePermission(d.Auth, domain.PermInboundRead))

	// AI auto-config: rule-based recommendation engine for optimal proxy settings.
	authed.GET("/auto-config", d.Handlers.GetAutoConfig, RequirePermission(d.Auth, domain.PermInboundRead))

	inbounds := authed.Group("/inbounds")
	inbounds.GET("", d.Handlers.ListInbounds, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.GET("/check-port", d.Handlers.CheckInboundPort, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.POST("", d.Handlers.CreateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.POST("/bulk", d.Handlers.BulkInboundAction, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.PUT("/:id", d.Handlers.UpdateInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.DELETE("/:id", d.Handlers.DeleteInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.POST("/:id/clone", d.Handlers.CloneInbound, RequirePermission(d.Auth, domain.PermInboundWrite))
	inbounds.GET("/:id/stats", d.Handlers.GetInboundStats, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.GET("/:id/online", d.Handlers.GetInboundOnline, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.GET("/:id/share-link", d.Handlers.GetShareLink, RequirePermission(d.Auth, domain.PermInboundRead))
	inbounds.GET("/:id/cert-status", d.Handlers.GetCertStatus, RequirePermission(d.Auth, domain.PermInboundRead))

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

	// Auto-Protocol Switching: protocol groups, ISP profiles, and switch events.
	if d.ProtocolGroups != nil {
		pg := authed.Group("/protocol-groups")
		pg.GET("", d.ProtocolGroups.ListProtocolGroups, RequirePermission(d.Auth, domain.PermInboundRead))
		pg.GET("/:id", d.ProtocolGroups.GetProtocolGroup, RequirePermission(d.Auth, domain.PermInboundRead))
		pg.POST("", d.ProtocolGroups.CreateProtocolGroup, RequirePermission(d.Auth, domain.PermInboundWrite))
		pg.PUT("/:id", d.ProtocolGroups.UpdateProtocolGroup, RequirePermission(d.Auth, domain.PermInboundWrite))
		pg.DELETE("/:id", d.ProtocolGroups.DeleteProtocolGroup, RequirePermission(d.Auth, domain.PermInboundWrite))
		pg.POST("/:id/reorder", d.ProtocolGroups.ReorderInbounds, RequirePermission(d.Auth, domain.PermInboundWrite))

		isp := authed.Group("/isp-profiles")
		isp.GET("", d.ProtocolGroups.ListISPProfiles, RequirePermission(d.Auth, domain.PermInboundRead))
		isp.POST("", d.ProtocolGroups.CreateISPProfile, RequirePermission(d.Auth, domain.PermInboundWrite))
		isp.PUT("/:id", d.ProtocolGroups.UpdateISPProfile, RequirePermission(d.Auth, domain.PermInboundWrite))
		isp.DELETE("/:id", d.ProtocolGroups.DeleteISPProfile, RequirePermission(d.Auth, domain.PermInboundWrite))

		sw := authed.Group("/switch-events")
		sw.POST("", d.ProtocolGroups.RecordSwitchEvent, RequirePermission(d.Auth, domain.PermInboundWrite))
		sw.GET("/summary", d.ProtocolGroups.GetSwitchSummary, RequirePermission(d.Auth, domain.PermInboundRead))
	}

	// REALITY key generation helper for building reality inbounds.
	authed.GET("/reality/keypair", d.Handlers.GenerateReality, RequirePermission(d.Auth, domain.PermInboundWrite))

	// Configuration backup/restore — sudo-level (admin:manage); restore is destructive.
	backup := authed.Group("/backup", RequirePermission(d.Auth, domain.PermAdminManage))
	backup.GET("/manifest", d.Handlers.GetBackupManifest)
	backup.GET("", d.Handlers.GetBackup)
	backup.POST("/restore", d.Handlers.RestoreBackup)

	// Admin + role management, gated behind the admin:manage permission (sudo bypasses).
	admins := authed.Group("/admins", RequirePermission(d.Auth, domain.PermAdminManage))
	admins.GET("", d.Handlers.ListAdmins)
	admins.GET("/usage", d.Handlers.ListResellerQuotaUsage)
	admins.POST("", d.Handlers.CreateAdmin)
	admins.PUT("/:id", d.Handlers.UpdateAdmin)
	admins.GET("/:id/inbounds", d.Handlers.GetAdminInbounds)
	admins.GET("/:id/nodes", d.Handlers.GetAdminNodes)
	admins.GET("/:id/plans", d.Handlers.GetAdminPlans)
	admins.GET("/:id/quota", d.Handlers.GetAdminQuotaUsage)
	admins.GET("/:id/wallet", d.Handlers.GetAdminWallet)
	admins.DELETE("/:id", d.Handlers.DeleteAdmin)
	admins.POST("/:id/wallet", d.Handlers.TopUpAdminWallet)
	admins.POST("/:id/impersonate", d.Handlers.ImpersonateAdmin)
	admins.POST("/:id/unsuspend", d.Handlers.UnsuspendAdmin)
	admins.POST("/:id/quota-adjust", d.Handlers.AdjustAdminQuota)

	roles := authed.Group("/roles", RequirePermission(d.Auth, domain.PermAdminManage))
	roles.GET("", d.Handlers.ListRoles)
	roles.POST("", d.Handlers.CreateRole)
	roles.PUT("/:id", d.Handlers.UpdateRole)
	roles.DELETE("/:id", d.Handlers.DeleteRole)

	// Audit log of mutating admin actions (sudo: all; reseller: own scope).
	authed.GET("/audit", d.Handlers.ListAudit, RequirePermission(d.Auth, domain.PermSystemRead))

	// Self-service account actions: any authenticated admin manages their own 2FA.
	account := authed.Group("/account")
	account.GET("", d.Handlers.GetAccount)
	account.GET("/quota", d.Handlers.GetAccountQuota)
	account.GET("/dashboard", d.Handlers.GetResellerDashboard)
	account.GET("/export/users", d.Handlers.ExportAccountUsers)
	account.GET("/backup/users", d.Handlers.ExportAccountUsersBackup, RequirePermission(d.Auth, domain.PermUserRead))
	account.POST("/backup/users/restore", d.Handlers.RestoreAccountUsersBackup, RequirePermission(d.Auth, domain.PermUserWrite))
	account.GET("/wallet", d.Handlers.GetAccountWallet)
	account.GET("/wallet/export", d.Handlers.ExportAccountWallet)
	account.GET("/wallet/packages", d.Handlers.ListAccountWalletPackages)
	account.GET("/wallet/payment-info", d.Handlers.GetAccountPaymentInfo)
	account.GET("/wallet/deposits", d.Handlers.ListAccountWalletDeposits)
	account.POST("/wallet/deposits", d.Handlers.InitAccountWalletDeposit)
	account.GET("/sub-admins", d.Handlers.ListSubAdmins)
	account.POST("/sub-admins", d.Handlers.CreateSubAdmin)
	account.GET("/branding", d.Handlers.GetAccountBranding)
	account.PUT("/branding", d.Handlers.UpdateAccountBranding)
	account.GET("/webhook", d.Handlers.GetAccountWebhook)
	account.PUT("/webhook", d.Handlers.UpdateAccountWebhook)
	account.POST("/stop-impersonate", d.Handlers.StopImpersonation)
	account.POST("/password", d.Handlers.ChangePassword)
	account.POST("/2fa/setup", d.Handlers.SetupTOTP)
	account.POST("/2fa/confirm", d.Handlers.ConfirmTOTP)
	account.POST("/2fa/disable", d.Handlers.DisableTOTP)

	// Per-reseller payment configuration (gated by "billing" reseller setting).
	if d.PaymentConfig != nil {
		account.GET("/payment-config", d.PaymentConfig.GetPaymentConfig)
		account.PUT("/payment-config", d.PaymentConfig.SavePaymentConfig)
	}

	// API Tokens (Personal Access Tokens)
	tokens := authed.Group("/tokens")
	tokens.GET("", d.APITokens.ListAPITokens)
	tokens.POST("", d.APITokens.CreateAPIToken)
	tokens.DELETE("/:id", d.APITokens.DeleteAPIToken)

	// PHASE 3A - Security & Authentication Routes (MFA, Password, IP Access Control)
	if d.MFAHandlers != nil && d.MFAMiddleware != nil {
		mfa := authed.Group("/mfa", d.MFAMiddleware.ValidateMFAIfEnabled())
		mfa.POST("/setup", d.MFAHandlers.SetupTOTP)
		mfa.POST("/verify", d.MFAHandlers.VerifyTOTP)
		mfa.POST("/backup-codes", d.MFAHandlers.GenerateBackupCodes)
		mfa.POST("/disable", d.MFAHandlers.DisableMFA)
		mfa.GET("/status", d.MFAHandlers.GetMFAStatus)
	}

	if d.PasswordHandlers != nil {
		pwd := authed.Group("/account/password")
		pwd.POST("/change", d.PasswordHandlers.ChangePassword)
		pwd.GET("/policy", d.PasswordHandlers.GetPasswordPolicy)
		pwd.PUT("/policy", d.PasswordHandlers.UpdatePasswordPolicy, RequirePermission(d.Auth, domain.PermAdminManage))
		pwd.GET("/status", d.PasswordHandlers.GetPasswordStatus)
	}

	if d.IPControlHandlers != nil {
		ipRules := authed.Group("/ip-rules", d.IPValidationMiddleware.ValidateClientIP())
		ipRules.POST("", d.IPControlHandlers.CreateRule)
		ipRules.GET("", d.IPControlHandlers.ListRules)
		ipRules.PUT("/:id", d.IPControlHandlers.UpdateRule)
		ipRules.DELETE("/:id", d.IPControlHandlers.DeleteRule)
		ipRules.GET("/global", d.IPControlHandlers.GetGlobalRules, RequirePermission(d.Auth, domain.PermAdminManage))
	}

	// Plans + Orders (admin)
	plans := authed.Group("/plans")
	plans.GET("", d.Handlers.ListPlans, RequirePermission(d.Auth, domain.PermUserRead))
	plans.POST("", d.Handlers.CreatePlan, RequirePermission(d.Auth, domain.PermUserWrite))
	plans.DELETE("/:id", d.Handlers.DeletePlan, RequirePermission(d.Auth, domain.PermUserWrite))

	orders := authed.Group("/orders")
	orders.GET("", d.Handlers.ListOrders, RequirePermission(d.Auth, domain.PermUserRead))
	orders.GET("/pending", d.Handlers.ListPendingOrders, RequirePermission(d.Auth, domain.PermUserRead))
	orders.POST("/:id/review", d.Handlers.ReviewOrder, RequirePermission(d.Auth, domain.PermUserWrite))

	// Wallet billing (sudo)
	if d.WalletBilling != nil {
		billing := authed.Group("/billing", RequirePermission(d.Auth, domain.PermAdminManage))
		billing.GET("/settings", d.WalletBilling.GetBillingSettings)
		billing.PUT("/settings", d.WalletBilling.UpdateBillingSettings)
		billing.GET("/deposits", d.WalletBilling.ListWalletDeposits)
		billing.POST("/deposits/:id/review", d.WalletBilling.ReviewWalletDeposit)
		wp := billing.Group("/wallet-packages")
		wp.GET("", d.WalletBilling.ListWalletPackages)
		wp.POST("", d.WalletBilling.CreateWalletPackage)
		wp.PUT("/:id", d.WalletBilling.UpdateWalletPackage)
		wp.DELETE("/:id", d.WalletBilling.DeleteWalletPackage)
	}

	// Public payment endpoints (no auth — user self-purchase)
	e.GET("/api/shop/plans", d.Handlers.PublicPlans)
	e.POST("/api/shop/purchase", d.Handlers.InitPurchase)
	e.GET("/api/payment/callback", d.Handlers.PaymentCallback)
	e.GET("/api/payment/wallet/callback", d.Handlers.WalletDepositCallback)
	e.POST("/api/payment/ipn/nowpayments", d.Handlers.NowPaymentsIPN)
	e.GET("/api/portal/branding", d.Handlers.PublicPortalBranding)
	e.GET("/payment/result", d.Handlers.PaymentResult)

	// Update checker (admin)
	authed.GET("/update/check", d.Handlers.CheckUpdate, RequirePermission(d.Auth, domain.PermAdminManage))

	// --- Portal: end-user self-service ---
	e.POST("/api/portal/login", d.Portal.PortalLogin)
	portal := e.Group("/api/portal", RequirePortalAuth(d.Issuer))
	portal.GET("/dashboard", d.Portal.PortalDashboard)
	portal.GET("/subscription", d.Portal.PortalSubscription)
	portal.GET("/usage", d.Portal.PortalUsage)
	portal.GET("/online", d.Portal.PortalOnline)
	portal.GET("/deeplink", d.Portal.PortalDeepLink)
	portal.GET("/plans", d.Portal.PortalListPlans)
	portal.GET("/payment-info", d.Portal.PortalPaymentInfo)
	portal.GET("/switch-history", d.Portal.PortalSwitchHistory)
	portal.GET("/connection-stats", d.Portal.PortalConnectionStats)
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
	cleanip.POST("/throughput", d.CleanIP.Throughput, RequirePermission(d.Auth, domain.PermInboundWrite))
	cleanip.POST("/throughput/all", d.CleanIP.ThroughputAll, RequirePermission(d.Auth, domain.PermInboundWrite))
	cleanip.GET("/schedule", d.CleanIP.GetSchedule, RequirePermission(d.Auth, domain.PermInboundRead))
	cleanip.PUT("/schedule", d.CleanIP.UpdateSchedule, RequirePermission(d.Auth, domain.PermInboundWrite))

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
	packs.GET("/default", d.RoutingPacks.GetDefault, RequirePermission(d.Auth, domain.PermInboundRead))
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
	probing.POST("/report", d.Probing.ReportProbe, RequirePermission(d.Auth, domain.PermNodeWrite))

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
	tricks.POST("/:id/apply", d.TLSTricks.ApplyToInbounds)
	tricks.DELETE("/:id", d.TLSTricks.DeleteProfile)

	// --- Client Fingerprint Validator ---
	fp := authed.Group("/fingerprint", RequirePermission(d.Auth, domain.PermAdminManage))
	fp.GET("/policy", d.Fingerprint.GetPolicy)
	fp.PUT("/policy", d.Fingerprint.UpdatePolicy)
	fp.GET("/rules", d.Fingerprint.ListRules)
	fp.POST("/rules", d.Fingerprint.CreateRule)
	fp.DELETE("/rules/:id", d.Fingerprint.DeleteRule)
	fp.GET("/events", d.Fingerprint.ListEvents)
	fp.POST("/report", d.Fingerprint.Report, RequirePermission(d.Auth, domain.PermNodeWrite))

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

	// --- Reseller quota alerts ---
	if d.AdminQuotaNotify != nil {
		aqn := authed.Group("/admin-quota-notify", RequirePermission(d.Auth, domain.PermAdminManage))
		aqn.GET("/config", d.AdminQuotaNotify.GetConfig)
		aqn.PUT("/config", d.AdminQuotaNotify.UpdateConfig)
		aqn.GET("/events", d.AdminQuotaNotify.ListEvents)
	}

	// --- IP-limit enforcement ---
	iplimit := authed.Group("/ip-limit", RequirePermission(d.Auth, domain.PermAdminManage))
	iplimit.GET("/policy", d.IPLimit.GetPolicy)
	iplimit.PUT("/policy", d.IPLimit.UpdatePolicy)
	iplimit.GET("/events", d.IPLimit.ListEvents)

	// --- Observability & Monitoring (PHASE 2) ---
	if d.MetricsService != nil {
		metricsHandlers := NewMetricsHandlers(d.MetricsService)
		authed.GET("/observability/metrics", metricsHandlers.GetMetrics, RequirePermission(d.Auth, domain.PermSystemRead))
	}

	if d.HealthService != nil {
		healthHandlers := NewHealthHandlers(d.HealthService)
		authed.GET("/health/check", healthHandlers.GetHealth, RequirePermission(d.Auth, domain.PermSystemRead))
		authed.GET("/health/component/:component", healthHandlers.GetHealthComponent, RequirePermission(d.Auth, domain.PermSystemRead))
		// Allow unauthenticated liveness check for Kubernetes probes
		api.GET("/health/live", healthHandlers.GetHealth)
		api.GET("/health/ready", healthHandlers.GetHealth)
	}

	if d.LoggerService != nil {
		logHandlers := NewLogHandlers(d.LoggerService)
		authed.GET("/observability/logs", logHandlers.GetLogs, RequirePermission(d.Auth, domain.PermSystemRead))
	}

	if d.TraceService != nil {
		traceHandlers := NewTraceHandlers(d.TraceService)
		authed.GET("/observability/trace", traceHandlers.GetTraceContext, RequirePermission(d.Auth, domain.PermSystemRead))
	}

	// --- PHASE 3C: Audit & Compliance ---
	if d.AuditEventHandlers != nil {
		auditEvents := authed.Group("/audit/events", RequirePermission(d.Auth, domain.PermSystemRead))
		auditEvents.GET("", d.AuditEventHandlers.ListEvents)
		auditEvents.GET("/:id", d.AuditEventHandlers.GetEvent)
		auditEvents.GET("/admin/:admin_id", d.AuditEventHandlers.GetEventsByAdmin)
		auditEvents.GET("/search", d.AuditEventHandlers.SearchEvents)
		auditEvents.GET("/count", d.AuditEventHandlers.GetEventCount)
		auditEvents.DELETE("/old", d.AuditEventHandlers.DeleteOldEvents, RequirePermission(d.Auth, domain.PermAdminManage))
	}

	if d.ComplianceHandlers != nil {
		compliance := authed.Group("/compliance", RequirePermission(d.Auth, domain.PermSystemRead))
		compliance.GET("/status", d.ComplianceHandlers.GetComplianceStatus)
		compliance.GET("/framework/:framework", d.ComplianceHandlers.CheckFramework)
		compliance.GET("/events", d.ComplianceHandlers.ListComplianceEvents)
		compliance.GET("/events/:id", d.ComplianceHandlers.GetComplianceEvent)
		compliance.GET("/non-compliant", d.ComplianceHandlers.GetNonCompliantItems)
		compliance.PUT("/events/:id/verify", d.ComplianceHandlers.VerifyComplianceEvent, RequirePermission(d.Auth, domain.PermAdminManage))
		compliance.GET("/certificate/:framework", d.ComplianceHandlers.GenerateComplianceCertificate)
	}

	if d.ReportHandlers != nil {
		reports := authed.Group("/reports", RequirePermission(d.Auth, domain.PermSystemRead))
		reports.POST("/daily", d.ReportHandlers.GenerateDailyReport)
		reports.POST("/custom", d.ReportHandlers.GenerateCustomReport)
		reports.POST("/compliance", d.ReportHandlers.GenerateComplianceReport)
		reports.GET("", d.ReportHandlers.ListReports)
		reports.GET("/:id", d.ReportHandlers.GetReport)
		reports.PUT("/:id/approve", d.ReportHandlers.ApproveReport, RequirePermission(d.Auth, domain.PermAdminManage))
		reports.GET("/:id/export", d.ReportHandlers.ExportReport)
		reports.DELETE("/:id", d.ReportHandlers.DeleteReport, RequirePermission(d.Auth, domain.PermAdminManage))
	}

	// Routes for PHASE 3B: Performance Optimization
	if d.PerformanceHandlers != nil {
		perf := authed.Group("/performance", RequirePermission(d.Auth, domain.PermSystemRead))
		perf.GET("/health", d.PerformanceHandlers.GetHealthStatus)
		perf.GET("/queries/slow", d.PerformanceHandlers.GetSlowQueries)
		perf.GET("/queries/stats", d.PerformanceHandlers.GetQueryStats)
		perf.GET("/alerts", d.PerformanceHandlers.GetPerformanceAlerts)
		perf.PUT("/alerts/:id/resolve", d.PerformanceHandlers.ResolveAlert, RequirePermission(d.Auth, domain.PermAdminManage))
		perf.GET("/report", d.PerformanceHandlers.GetPerformanceReport)
		perf.GET("/rate-limits/rules", d.PerformanceHandlers.ListRateLimitRules)
		perf.GET("/rate-limits/violations", d.PerformanceHandlers.GetRateLimitViolations)
	}

	// Routes for PHASE 3D: Security Hardening & Defense
	if d.SecurityHardeningHandlers != nil {
		sec := authed.Group("/security", RequirePermission(d.Auth, domain.PermSystemRead))
		sec.GET("/threats", d.SecurityHardeningHandlers.GetSecurityThreats)
		sec.GET("/threats/blocked", d.SecurityHardeningHandlers.GetBlockedThreats)
		sec.GET("/threats/count", d.SecurityHardeningHandlers.GetThreatCount)
		sec.GET("/policy", d.SecurityHardeningHandlers.GetSecurityPolicy)
		sec.PUT("/policy", d.SecurityHardeningHandlers.UpdateSecurityPolicy, RequirePermission(d.Auth, domain.PermAdminManage))
		sec.GET("/score", d.SecurityHardeningHandlers.GetSecurityScore)
		sec.GET("/compliance/validate", d.SecurityHardeningHandlers.ValidateCompliance)
		sec.GET("/reputation/:ip", d.SecurityHardeningHandlers.GetIPReputation)
	}

	// --- API v2 routes ---
	v2 := e.Group("/api/v2", RequireAuth(d.PanelAuth), RequireActiveAdmin(d.Handlers.Admins), Audit(d.Audit))
	// Subscription format template settings
	v2sub := v2.Group("/sub-settings", RequirePermission(d.Auth, domain.PermAdminManage))
	v2sub.GET("", d.SubSettings.GetSubSettings)
	v2sub.PUT("", d.SubSettings.UpdateSubSettings)

	// --- Phase 8-14: Enterprise Feature Handlers ---

	// Phase 8: Config Management (validation, versioning, diff, rollback, export/import, auto-update)
	if d.ConfigManagement != nil {
		d.ConfigManagement.Register(v2)
	}

	// Phase 9: WireGuard Management (peers, repair, QR, mesh)
	if d.WireGuardMgmt != nil {
		d.WireGuardMgmt.Register(v2)
	}

	// Phase 10: Advanced Security (sessions, login audit, security audit, IP whitelist, bans, scoped tokens)
	if d.SecurityAdvanced != nil {
		d.SecurityAdvanced.Register(v2)
	}

	// Phase 11: Background Jobs (status polling, list)
	if d.Jobs != nil {
		d.Jobs.Register(v2)
	}

	// Phase 12: API Docs (Swagger UI, rate limits, webhook test)
	if d.Docs != nil {
		d.Docs.Register(v2)
	}
	if d.RateLimits != nil {
		d.RateLimits.Register(v2)
	}
	if d.WebhookTest != nil {
		d.WebhookTest.Register(v2)
	}

	// Phase 14: Portal Pro (push subscribe, speed test, guides, setup wizard)
	if d.PortalPro != nil {
		d.PortalPro.Register(v2)
	}

	return e
}
