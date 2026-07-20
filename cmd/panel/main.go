// Command panel is the VortexUI control plane: REST API, web dashboard, and the
// gRPC hub that node agents connect back to.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/acme"
	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/doh"
	"github.com/vortexui/vortexui/internal/decoy"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/geoip"
	"github.com/vortexui/vortexui/internal/logbuf"
	"github.com/vortexui/vortexui/internal/notify"
	"github.com/vortexui/vortexui/internal/panel/api"
	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/payment"
	"github.com/vortexui/vortexui/internal/platform/postgres"
	"github.com/vortexui/vortexui/internal/platform/redis"
	"github.com/vortexui/vortexui/internal/stats"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// version is the panel build version. It defaults to the contents of the VERSION
// file and is overridden at build time via -ldflags "-X main.version=...".
var version = "1.4.0"

func main() {
	logBuf := logbuf.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}), 2000)
	log := slog.New(logBuf)
	log.Info("starting VortexUI panel", "version", version)

	// Root context cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Subcommands run a one-shot task and exit instead of starting the server.
	if len(os.Args) > 1 && os.Args[1] == "admin" {
		if err := runAdmin(ctx, os.Args[2:]); err != nil {
			log.Error("admin command failed", "err", err)
			os.Exit(1)
		}
		return
	}

	cfg, err := config.LoadPanel()
	if err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(1)
	}
	log.Info("starting VortexUI panel", "config", cfg.String())

	if err := run(ctx, log, logBuf, cfg); err != nil {
		log.Error("panel exited with error", "err", err)
		os.Exit(1)
	}
	log.Info("panel shut down cleanly")
}

// run wires the whole control plane and blocks until ctx is cancelled. Kept
// separate from main so it is testable and returns errors instead of os.Exit.
//
// Assembly order mirrors the dependency graph: storage → stats aggregator →
// node hub → services → HTTP API.
func run(ctx context.Context, log *slog.Logger, logBuf *logbuf.Handler, cfg *config.Panel) error {
	// 1. Storage. Apply pending migrations first (embedded, idempotent) so a
	// fresh database is ready without a separate migration step.
	if err := postgres.Migrate(cfg.DatabaseURL); err != nil {
		return err
	}
	log.Info("database migrations applied")
	store, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer store.Close()
	users, nodes, traffic, admins := store.Users(), store.Nodes(), store.Traffic(), store.Admins()

	// 2. Stats aggregator: folds node traffic deltas into the users table and
	// the time series. Runs as a single background consumer.
	agg := stats.New(users, traffic, store.InboundTraffic())
	go func() {
		if err := agg.Run(ctx); err != nil && ctx.Err() == nil {
			log.Error("aggregator stopped", "err", err)
		}
	}()

	// 3. Node hub: dials every node over mTLS, drains traffic into the
	// aggregator, and supervises health/failover.
	tlsFiles := vgrpc.TLSFiles{Cert: cfg.TLSCert, Key: cfg.TLSKey, CA: cfg.TLSCA}

	// Optional in-process local node: build its driver and ensure its DB record
	// exists before the hub loads nodes, so the dialer can route it in-process.
	var localID uuid.UUID
	var localDriver core.CoreDriver
	if cfg.LocalNode {
		drv, err := buildLocalDriver(cfg, log)
		if err != nil {
			return err
		}
		localDriver = drv
		defer localDriver.Stop(context.Background())
		localNode, err := ensureLocalNode(ctx, nodes, cfg)
		if err != nil {
			return err
		}
		localID = localNode.ID
		log.Info("local node enabled", "name", localNode.Name, "core", cfg.Core, "id", localID)
	}

	decoySrv := decoy.NewServer()
	defer func() { _ = decoySrv.Stop(context.Background()) }()

	h := hub.New(hub.Options{
		Dialer:              localAwareDialer(tlsFiles, localID, localDriver, decoySrv),
		Nodes:               nodes,
		Ingest:              agg.Ingest,
		Logger:              log,
		RecentTraffic:       store.Monitor(),
		RecentTrafficWindow: 2 * time.Minute,
		AutoRecoverCore:         cfg.AutoRecoverCore,
		AutoRecoverCoreAfter:    cfg.AutoRecoverCoreAfter,
		AutoRecoverCoreCooldown: cfg.AutoRecoverCoreCooldown,
		AutoRecoverHub:          cfg.AutoRecoverHub,
		AutoRecoverHubAfter:     cfg.AutoRecoverHubAfter,
		AutoRecoverHubCooldown:  cfg.AutoRecoverHubCooldown,
	})
	defer h.Close()
	if list, err := nodes.List(ctx); err != nil {
		log.Error("load nodes failed", "err", err)
	} else {
		for _, n := range list {
			if err := h.Register(ctx, n); err != nil {
				log.Error("register node failed", "node", n.Name, "err", err)
			}
		}
	}
	h.StartWatchdog(ctx)
	if cfg.AutoRecoverCore || cfg.AutoRecoverHub {
		log.Info("node auto-recovery enabled",
			"core", cfg.AutoRecoverCore,
			"core_after", cfg.AutoRecoverCoreAfter,
			"hub", cfg.AutoRecoverHub,
			"hub_after", cfg.AutoRecoverHubAfter,
		)
	}

	// Redis powers login rate limiting. Optional: if unreachable the panel still
	// runs, just without brute-force throttling.
	var limiter api.RateLimiter
	var devices api.DeviceLimiter
	var online api.DeviceCounter
	if rc, err := redis.Open(ctx, cfg.RedisURL); err != nil {
		log.Warn("redis unavailable; login rate limiting and device limits disabled", "err", err)
	} else {
		defer func() { _ = rc.Close() }()
		limiter = rc.RateLimiter()
		devices = rc.Devices()
		online = rc.Devices()
	}

	// Event bus + optional outbound notifiers (webhook, Telegram). Subscribers
	// start first so no event is missed once the loops below begin publishing.
	bus := events.New(log)
	h.SetOnDisconnectAlert(func(_ context.Context, node *domain.Node, diag domain.NodeDiagnostics, since time.Duration) {
		bus.Publish(events.Event{
			Type:     events.NodeDisconnectAlert,
			NodeID:   node.ID.String(),
			NodeName: node.Name,
			Message:  diag.Message,
			Data: map[string]any{
				"diagnostics":   string(diag.Code),
				"since_minutes": since.Minutes(),
				"network_ok":    diag.NetworkReachable,
				"ca_match":      diag.CAMatch,
			},
		})
	})
	h.SetOnAutoRecover(func(_ context.Context, node *domain.Node, action string) {
		bus.Publish(events.Event{
			Type:     events.NodeAutoRecover,
			NodeID:   node.ID.String(),
			NodeName: node.Name,
			Message:  action,
			Data:     map[string]any{"action": action},
		})
	})
	if cfg.WebhookURL != "" {
		wh := notify.NewWebhook(cfg.WebhookURL, cfg.WebhookSecret, log)
		go wh.Run(ctx, bus.Subscribe(256))
		log.Info("webhook notifier enabled", "url", cfg.WebhookURL)
	}
	if cfg.TelegramToken != "" && cfg.TelegramChatID != "" {
		tg := notify.NewTelegram(cfg.TelegramToken, cfg.TelegramChatID, log)
		go tg.Run(ctx, bus.Subscribe(256))
		log.Info("telegram notifier enabled")

		// Interactive bot (long-polling) for admin commands.
		botAdapter := service.NewBotAdapter(users, nodes)
		bot := notify.NewTelegramBot(cfg.TelegramToken, notify.ParseChatID(cfg.TelegramChatID), botAdapter, log)
		go bot.Run(ctx)
		log.Info("telegram bot enabled (interactive commands)")
	}

	// 4. Services + 5. HTTP API.
	issuer := auth.NewIssuer([]byte(cfg.JWTSecret), cfg.JWTTTL)
	// Enforcement loop: disables + de-provisions users who hit their cap/expiry.
	// 15s keeps overshoot small; traffic flush also triggers an immediate tick.
	enforcer := service.NewEnforcer(users, h, 15*time.Second, log)
	enforcer.SetPublisher(bus)
	go enforcer.Run(ctx)
	agg.AfterFlush = func(fctx context.Context) {
		// Do not block traffic aggregation on RemoveUser RPCs.
		go func() {
			tctx, cancel := context.WithTimeout(context.WithoutCancel(fctx), 30*time.Second)
			defer cancel()
			if err := enforcer.Tick(tctx); err != nil {
				log.Warn("post-flush enforcement failed", "err", err)
			}
		}()
	}

	// Reset loop: zeroes used traffic on schedule and re-activates quota-limited
	// users — the complement to enforcement.
	resetter := service.NewResetter(users, h, time.Hour, log)
	resetter.SetPublisher(bus)

	// Expiry warning loop: alerts admins 3 days before user subscriptions expire.
	expiryWarner := service.NewExpiryWarner(store.Users(), log)
	expiryWarner.SetPublisher(bus)
	go expiryWarner.Run(ctx)
	go resetter.Run(ctx)

	authSvc := service.NewAuthService(admins, issuer)
	userSvc := service.NewUserService(users, h)
	userSvc.SetPublisher(bus)
	userSvc.SetOnlineQuerier(h)

	// Account-sharing guard: alerts (and optionally limits) users online from
	// more distinct IPs than their device limit allows.
	shareGuard := service.NewShareGuard(userSvc, users, h, 2*time.Minute, log)
	shareGuard.SetPublisher(bus)
	shareGuard.SetAutoLimit(cfg.ShareAutoLimit)
	// Wire IP-limit enforcement: ShareGuard reads the policy each pass and, when
	// enabled, branches on the configured action (warn/disable/kill), recording
	// each action. With no enabled policy its legacy detection-only behavior is
	// unchanged. The node repo supplies the per-node core kind (xray vs sing-box).
	ipLimitSvc := service.NewIPLimitService(store.IPLimits())
	shareGuard.SetIPLimit(store.IPLimits(), nodes)
	go shareGuard.Run(ctx)

	adminQuotaWarner := service.NewAdminQuotaWarner(admins, users, store.AdminQuotaNotify(), log)
	adminQuotaWarner.SetPublisher(bus)
	go adminQuotaWarner.Run(ctx)

	subSvc := service.NewSubscriptionService(users, nodes, store.SubHosts())
	subSvc.SetTLSTricks(store.TLSTricks())
	subSvc.SetProtocolGroups(store.ProtocolGroups(), store.ISPProfiles())
	switchEventSvc := service.NewSwitchEventService(store.SwitchEvents())
	switchEventSvc.SetPublisher(bus)
	subSvc.SetSwitchEvents(store.SwitchEvents())
	wgSvc := service.NewWireGuardService(store.WireGuardPeers())
	syncSvc := service.NewSyncService(store.Inbounds(), users, h, store.Outbounds(), store.Routing(), store.Balancers())
	syncSvc.SetWireGuard(wgSvc)
	protocolGroupSvc := service.NewProtocolGroupService(store.ProtocolGroups(), store.ISPProfiles(), store.Inbounds(), nodes, syncSvc)
	protocolGroupSvc.SetPublisher(bus)
	// Provisioning a user onto a node that hosts a WireGuard inbound requires a
	// full resync (the only path that computes WG peers); wire it in.
	userSvc.SetResyncer(syncSvc)
	nodeSvc := service.NewNodeService(nodes, h)
	nodeSvc.SetLogQuerier(h)
	nodeSvc.SetCoreController(h)
	inboundSvc := service.NewInboundService(store.Inbounds(), nodes, syncSvc)
	outboundSvc := service.NewOutboundService(store.Outbounds(), syncSvc)
	routingSvc := service.NewRoutingService(store.Routing(), syncSvc)
	balancerSvc := service.NewBalancerService(store.Balancers(), syncSvc)
	adminSvc := service.NewAdminService(admins, users)
	resellerWH := service.NewResellerWebhookDispatcher(adminSvc, log)
	go resellerWH.Run(ctx, bus.Subscribe(256))
	resellerSuspender := service.NewResellerAutoSuspender(adminSvc, userSvc, users, log)
	resellerSuspender.SetPublisher(bus)
	go resellerSuspender.Run(ctx)
	overviewSvc := service.NewOverviewService(users, nodes)
	overviewSvc.SetWidgetDeps(service.WidgetDeps{
		Counters:     store.Dashboard(),
		Traffic:      traffic,
		Inbounds:     store.Inbounds(),
		Routing:      store.Routing(),
		Balancers:    store.Balancers(),
		RoutingPacks: store.RoutingPacks(),
		Probing:      store.Probing(),
	})
	backupSvc := service.NewBackupService(nodes, store.Inbounds(), store.Outbounds(), store.Routing(), store.Balancers(), users, store.Backup())
	backupSvc.SetAdminSource(store.Admins())
	backupSvc.SetPlanSource(store.Plans())
	backupSvc.SetDatabaseURL(cfg.DatabaseURL)
	backupSvc.SetResellerSources(store.Admins(), store.ResellerPayment())

	// Wire failover migration into the hub now that its dependencies exist (the
	// migration service provisions onto the target via the hub itself).
	migration := service.NewMigrationService(store.Inbounds(), nodes, users, h, log)
	migration.SetRepo(store.Migration())
	h.SetOnFailover(func(fctx context.Context, failed, target *domain.Node) {
		bus.Publish(events.Event{Type: events.NodeDown, NodeID: failed.ID.String(), NodeName: failed.Name})
		if err := migration.Migrate(fctx, failed, target); err != nil {
			log.Error("failover migration failed", "failed", failed.Name, "err", err)
		}
	})

	// On (re)connect: repopulate the node with its own config, then shed any
	// temporary copies that failover parked on other nodes while it was down.
	h.SetOnConnect(func(cctx context.Context, node *domain.Node) {
		bus.Publish(events.Event{Type: events.NodeUp, NodeID: node.ID.String(), NodeName: node.Name})
		if err := syncSvc.Resync(cctx, node.ID); err != nil {
			log.Error("node resync on connect failed", "node", node.Name, "err", err)
		}
		if err := migration.MigrateBack(cctx, node); err != nil {
			log.Error("migrate-back failed", "node", node.Name, "err", err)
		}
	})
	// Push the latest xray policy (incl. statsUserOnline) to every node once at
	// startup so live connection stats work without waiting for a disconnect edge.
	go func() {
		sctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		list, err := nodes.List(sctx)
		if err != nil {
			log.Warn("startup resync: list nodes failed", "err", err)
			return
		}
		for _, n := range list {
			if err := syncSvc.Resync(sctx, n.ID); err != nil {
				log.Warn("startup resync failed", "node", n.Name, "err", err)
			} else {
				log.Info("startup resync ok", "node", n.Name)
			}
		}
	}()
	tokenSvc := service.NewAPITokenService(store.APITokens())

	// Construct all feature services.
	portalSvc := service.NewPortalService(users, store.Tickets(), store.Plans())
	planSvc := service.NewPlanService(store.Plans(), userSvc)
	realitySvc := service.NewRealityScannerService(store.RealityScans(), nodes)
	cleanIPSvc := service.NewCleanIPScannerService(store.CleanIPScans())
	cleanIPSvc.SetScheduleRepo(store.CleanIPSchedule())
	go cleanIPSvc.RunScheduler(ctx)
	subHostSvc := service.NewSubHostService(store.SubHosts())
	routingPackSvc := service.NewRoutingPackService(store.RoutingPacks(), routingSvc, outboundSvc)
	// Embed a selected routing pack's rules into Clash/sing-box subscriptions
	// (per-user selection else global default). Nil-safe and fail-open: with no
	// pack selected, subscription output is unchanged.
	subSvc.SetRoutingPacks(routingPackSvc)
	quotaSvc := service.NewQuotaService(store.QuotaPolicies())
	relaySvc := service.NewRelayService(store.RelayChains(), nodes)
	decoySvc := service.NewDecoyService(store.DecoySites())
	analyticsSvc := service.NewAnalyticsService(store.Analytics())
	probingSvc := service.NewProbingService(store.Probing(), log)
	familySvc := service.NewFamilyService(store.Families(), users)
	referralSvc := service.NewReferralService(store.Referrals(), users)
	dohSvc := service.NewDoHService(store.DoH())

	ipGuard := api.NewIPGuard(os.Getenv("VORTEX_IP_WHITELIST"), os.Getenv("VORTEX_IP_BLACKLIST"))
	panelSettingsSvc := service.NewPanelSettingsService(store.PanelSettings(), service.PanelSettingsHooks{
		OnIPGuard: func(wl, bl string) {
			if strings.TrimSpace(wl) != "" {
				ipGuard.SetWhitelist(wl)
			}
			ipGuard.SetBlacklist(bl)
		},
	})
	if ps, err := panelSettingsSvc.Get(ctx); err == nil && ps != nil {
		if ps.IPWhitelist != "" {
			ipGuard.SetWhitelist(ps.IPWhitelist)
		}
		if ps.IPBlacklist != "" {
			ipGuard.SetBlacklist(ps.IPBlacklist)
		}
	}

	acmeMgr := acme.NewManager(acme.NewMemoryStore(), os.Getenv("VORTEX_ACME_EMAIL"), log)
	acmeMgr.SetCloudflare(cfg.CloudflareToken, cfg.CloudflareZoneID)

	sniSvc := service.NewSNIService(store.SNIDomains())
	sniSvc.SetCertIssuer(acmeMgr)
	tlsTricksSvc := service.NewTLSTricksService(store.TLSTricks())
	tlsTricksSvc.SetInboundLinker(store.TLSTricks())
	fpSvc := service.NewFingerprintService(store.Fingerprints())
	fedSvc := service.NewFederationService(store.Federation())

	securitySync := service.NewSecuritySyncHelper(store.Probing(), store.SNIDomains(), store.DecoySites())
	syncSvc.SetSecurity(securitySync)
	fleetResync := service.NewFleetResync(nodes, syncSvc)
	probingSvc.SetFleetResync(fleetResync)
	fpSvc.SetProbing(probingSvc)
	fpSvc.SetFleetResync(fleetResync)
	fpSvc.SetLogger(log)

	securityIngest := &service.SecurityIngestService{Probing: probingSvc, Fingerprint: fpSvc, Log: log}
	h.SetSecurityIngest(securityIngest.Ingest)
	probingSvc.SetPublisher(bus)

	dohServer := doh.NewServer(store.DoH(), log)
	dohSvc.SetRuntime(dohServer)
	if err := dohServer.Reload(ctx); err != nil {
		log.Warn("doh server reload failed", "err", err)
	}
	defer func() { _ = dohServer.Stop(context.Background()) }()

	go fedSvc.RunSyncWorker(ctx, log)
	deepLinkSvc := service.NewDeepLinkService(store.DeepLinks())
	quotaNotifySvc := service.NewQuotaNotifyService(store.QuotaNotify())
	adminQuotaNotifySvc := service.NewAdminQuotaNotifyService(store.AdminQuotaNotify())
	subSettingsSvc := service.NewSubSettingsService(store.SubSettings())

	// GeoIP resolver for the "Traffic by Country" analytics. Optional: an empty
	// or unopenable DB path degrades gracefully to a disabled resolver.
	geoResolver, gerr := geoip.Open(cfg.GeoIPDB)
	if gerr != nil {
		log.Warn("geoip database unavailable; Traffic by Country disabled", "path", cfg.GeoIPDB, "err", gerr)
		geoResolver = geoip.Disabled()
	}
	nodeSvc.SetGeoResolver(geoResolver)
	defer func() { _ = geoResolver.Close() }()
	geoSvc := service.NewGeoService(geoResolver, store.UserGeo())

	panelAuth := &auth.PanelAuth{JWT: issuer, Tokens: store.APITokens()}
	enrollSvc := &service.EnrollmentService{CAPath: cfg.TLSCA}

	var zarinPal *payment.ZarinPal
	if cfg.ZarinPalMerchantID != "" {
		zarinPal = payment.NewZarinPal(cfg.ZarinPalMerchantID)
		log.Info("ZarinPal gateway enabled")
	}
	var nowPayments *payment.NowPayments
	if cfg.NowPaymentsAPIKey != "" {
		nowPayments = payment.NewNowPayments(cfg.NowPaymentsAPIKey)
		log.Info("NowPayments gateway enabled")
	}
	walletBillingSvc := service.NewWalletBillingService(store.WalletBilling(), adminSvc, zarinPal, nowPayments)
	resellerPaymentSvc := service.NewResellerPaymentService(store.ResellerPayment())

	autoBackup := service.NewAutoBackup(backupSvc, 24*time.Hour, log)
	autoBackup.Settings = panelSettingsSvc
	autoBackup.DefaultTelegramToken = cfg.TelegramToken
	go autoBackup.Run(ctx)

	// Initialize observability services (metrics, health checks, structured logging, tracing)
	obsCfg := config.DefaultObservabilityConfig()
	metricsService, healthService, loggerService, traceService, prometheusExporter := config.InitializeObservability(obsCfg, log)

	// Register default health checks
	if healthService != nil {
		defaultChecks := service.NewDefaultHealthChecks(healthService, log)
		if err := defaultChecks.RegisterAll(); err != nil {
			log.Error("failed to register default health checks", "err", err)
		}
		
		// Start background health check poller (30 second interval)
		poller := service.NewHealthCheckPoller(healthService, 30*time.Second, log)
		poller.Start(ctx)
	}

	// Initialize PHASE 3A - Authentication & Hardening services
	totpRepo := postgres.NewTOTPRepository(store.Pool(), log)
	mfaRepo := postgres.NewMFARepository(store.Pool(), log)
	passwordPolicyRepo := postgres.NewPasswordPolicyRepository(store.Pool(), log)
	ipAccessRuleRepo := postgres.NewIPAccessRuleRepository(store.Pool(), log)
	loginAttemptRepo := postgres.NewLoginAttemptRepository(store.Pool(), log)

	// Start background cleanup worker for login attempts (deletes attempts > 24h old)
	go loginAttemptRepo.StartCleanupWorker(ctx, 1*time.Hour)

	// Initialize security services
	totpService := service.NewTOTPService(log)
	passwordPolicyService := service.NewPasswordPolicyService(log)
	ipValidatorService := service.NewIPValidatorService(log)

	// Initialize security handlers
	mfaHandlers := api.NewMFAHandlers(totpService, mfaRepo, log)
	passwordHandlers := api.NewPasswordHandlers(passwordPolicyService, passwordPolicyRepo, log)
	ipControlHandlers := api.NewIPControlHandlers(ipAccessRuleRepo, ipValidatorService, log)

	// Initialize security middleware
	mfaMiddleware := api.NewMFAMiddleware(mfaRepo, log)
	ipValidationMiddleware := api.NewIPValidationMiddleware(ipAccessRuleRepo, ipValidatorService, log)

	// Initialize PHASE 3C - Audit & Compliance services
	auditEventRepo := postgres.NewAuditEventRepository(store.Pool(), log)
	complianceEventRepo := postgres.NewComplianceEventRepository(store.Pool(), log)
	auditReportRepo := postgres.NewAuditReportRepository(store.Pool(), log)
	auditPolicyRepo := postgres.NewAuditPolicyRepository(store.Pool(), log)
	auditArchiveRepo := postgres.NewAuditArchiveRepository(store.Pool(), log)

	// Initialize audit and compliance services
	reportGeneratorService := service.NewReportGeneratorService(auditEventRepo, auditReportRepo, log)
	complianceCheckerService := service.NewComplianceCheckerService(complianceEventRepo, auditPolicyRepo, log)

	// Initialize audit and compliance handlers
	auditEventHandlers := api.NewAuditEventHandlers(auditEventRepo, log)
	complianceHandlers := api.NewComplianceHandlers(complianceEventRepo, complianceCheckerService, log)
	reportHandlers := api.NewReportHandlers(auditReportRepo, reportGeneratorService, log)

	// Initialize PHASE 3B - Performance Optimization services
	queryMetricsRepo := postgres.NewQueryMetricsRepository(store.Pool(), log)
	rateLimitRepo := postgres.NewRateLimitRepository(store.Pool(), log)
	performanceAlertRepo := postgres.NewPerformanceAlertRepository(store.Pool(), log)

	// Initialize cache service
	cacheService := service.NewInMemoryCacheService(104857600) // 100MB

	// Initialize performance monitor
	performanceMonitor := service.NewPerformanceMonitorService(queryMetricsRepo, performanceAlertRepo, cacheService, log)

	// Initialize rate limiter
	rateLimiterService := service.NewInMemoryRateLimiter(rateLimitRepo, log)

	// Initialize performance handlers
	performanceHandlers := api.NewPerformanceHandlers(queryMetricsRepo, performanceAlertRepo, rateLimitRepo, performanceMonitor, log)

	// Initialize PHASE 3D - Security Hardening & Defense services
	threatRepo := postgres.NewSecurityThreatRepository(store.Pool(), log)
	ipRepRepo := postgres.NewIPReputationRepository(store.Pool(), log)
	policyRepo := postgres.NewSecurityPolicyRepository(store.Pool(), log)

	threatDetector := service.NewThreatDetectionService(threatRepo, ipRepRepo, log)
	anomalyDetector := service.NewAnomalyDetectionService(log)

	// Initialize security handlers
	securityHandlers := api.NewSecurityHandlers(threatRepo, policyRepo, threatDetector, anomalyDetector, log)

	router := api.NewRouter(api.Deps{
		Version: version,
		Handlers: &api.Handlers{
			Auth: authSvc, Users: userSvc, Sub: subSvc,
			Nodes: nodeSvc, Hub: h, Enrollment: enrollSvc,
			Inbounds: inboundSvc, Admins: adminSvc, Devices: devices,
			Outbounds: outboundSvc, Routing: routingSvc, Balancers: balancerSvc,
			Overview: overviewSvc, Backup: backupSvc,
			Counters: store.Dashboard(),
			Plans:    planSvc,
			ZarinPal: zarinPal,
			NowPayments: nowPayments,
			WalletBilling: walletBillingSvc,
			ResellerPayment: resellerPaymentSvc,
			Online: online, Logs: logBuf, Audit: store.Audit(),
			Repo: users, Traffic: traffic,
			NodeRepo: nodes, WireGuard: wgSvc,
			Throttle: api.NewLoginThrottle(5, 15*time.Minute),
			Events:   bus,
			SubSettings: subSettingsSvc,
			Geo:         geoSvc,
			Issuer:      issuer,
			SwitchEvents: switchEventSvc,
		},
		APITokens:   &api.APITokenHandlers{Svc: tokenSvc},
		Portal: &api.PortalHandlers{
			Portal: portalSvc, Issuer: issuer, Admins: adminSvc,
			ResellerPayment: resellerPaymentSvc, WalletBilling: walletBillingSvc,
			Sub: subSvc, Traffic: traffic, Users: userSvc, Online: online, DeepLink: deepLinkSvc,
			SwitchEvents: switchEventSvc,
		},
		Reality:     &api.RealityHandlers{Scanner: realitySvc},
		CleanIP:     &api.CleanIPHandlers{Scanner: cleanIPSvc},
		SubHosts:    &api.SubHostHandlers{SubHosts: subHostSvc},
		RoutingPacks: &api.RoutingPackHandlers{Packs: routingPackSvc},
		Quota:       &api.QuotaHandlers{Quota: quotaSvc},
		Relay:       &api.RelayHandlers{Relay: relaySvc},
		Decoy:       &api.DecoyHandlers{Decoy: decoySvc, Resync: fleetResync},
		Analytics:   &api.AnalyticsHandlers{Analytics: analyticsSvc},
		Migration:   &api.MigrationHandlers{Migration: migration},
		Probing:     &api.ProbingHandlers{Probing: probingSvc, Resync: fleetResync},
		Family:      &api.FamilyHandlers{Family: familySvc},
		Referral:    &api.ReferralHandlers{Referral: referralSvc},
		DoH:         &api.DoHHandlers{DoH: dohSvc},
		SNI:         &api.SNIHandlers{SNI: sniSvc, Inbounds: store.Inbounds(), Resync: fleetResync},
		TLSTricks:   &api.TLSTricksHandlers{Tricks: tlsTricksSvc},
		Fingerprint: &api.FingerprintHandlers{FP: fpSvc},
		Federation:  &api.FederationHandlers{Fed: fedSvc},
		DeepLink:    &api.DeepLinkHandlers{DeepLink: deepLinkSvc},
		QuotaNotify: &api.QuotaNotifyHandlers{QN: quotaNotifySvc},
		AdminQuotaNotify: &api.AdminQuotaNotifyHandlers{Svc: adminQuotaNotifySvc},
		IPLimit:     &api.IPLimitHandlers{IPLimit: ipLimitSvc},
		SubSettings: &api.SubSettingsHandlers{Svc: subSettingsSvc},
		Monitor:     &api.MonitorHandlers{Hub: h, Nodes: nodes, Users: users, Monitor: monitorAdapter{store.Monitor()}},
		WalletBilling: &api.WalletBillingHandlers{Svc: walletBillingSvc},
		PaymentConfig: &api.PaymentConfigHandlers{Svc: resellerPaymentSvc, Admins: adminSvc},
		ProtocolGroups: &api.ProtocolGroupHandlers{Groups: protocolGroupSvc, Events: switchEventSvc},
		Issuer:      issuer,
		PanelAuth:   panelAuth,
		Auth:        authSvc,
		Limiter:     limiter,
		Audit:       store.Audit(),
		IPGuard:     ipGuard,
		PanelSettings: &api.PanelSettingsHandlers{Svc: panelSettingsSvc},
		// AuditService and SessionService are disabled: the underlying
		// AuditRepo/SessionRepo stubs need full sqlc query generation before
		// these services can be safely wired into the middleware chain.
		MetricsService: metricsService,
		HealthService: healthService,
		LoggerService: loggerService,
		TraceService: traceService,
		PrometheusExporter: prometheusExporter,
		// PHASE 3A - Security components
		TOTPRepository: totpRepo,
		MFARepository: mfaRepo,
		PasswordPolicyRepository: passwordPolicyRepo,
		IPAccessRuleRepository: ipAccessRuleRepo,
		LoginAttemptRepository: loginAttemptRepo,
		TOTPService: totpService,
		PasswordPolicyService: passwordPolicyService,
		IPValidatorService: ipValidatorService,
		MFAHandlers: mfaHandlers,
		PasswordHandlers: passwordHandlers,
		IPControlHandlers: ipControlHandlers,
		MFAMiddleware: mfaMiddleware,
		IPValidationMiddleware: ipValidationMiddleware,
		// PHASE 3C - Audit & Compliance components
		AuditEventRepository: auditEventRepo,
		ComplianceEventRepository: complianceEventRepo,
		AuditReportRepository: auditReportRepo,
		AuditPolicyRepository: auditPolicyRepo,
		AuditArchiveRepository: auditArchiveRepo,
		ReportGeneratorService: reportGeneratorService,
		ComplianceCheckerService: complianceCheckerService,
		AuditEventHandlers: auditEventHandlers,
		ComplianceHandlers: complianceHandlers,
		ReportHandlers: reportHandlers,
		// PHASE 3B - Performance Optimization components
		QueryMetricsRepository: queryMetricsRepo,
		RateLimitRepository: rateLimitRepo,
		PerformanceAlertRepository: performanceAlertRepo,
		CacheService: cacheService,
		PerformanceMonitor: performanceMonitor,
		RateLimiterService: rateLimiterService,
		PerformanceHandlers: performanceHandlers,
		// PHASE 3D - Security Hardening & Defense components
		ThreatDetector: threatDetector,
		AnomalyDetector: anomalyDetector,
		SecurityHardeningHandlers: securityHandlers,
	})

	srv := &http.Server{Addr: cfg.HTTPAddr, Handler: router, ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Info("HTTP API listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutCtx)
}

// monitorAdapter adapts postgres.MonitorRepo to api.MonitorSource.
type monitorAdapter struct{ repo *postgres.MonitorRepo }

func (m monitorAdapter) RecentActive(ctx context.Context, window time.Duration) ([]api.MonitorLiveUser, error) {
	users, err := m.repo.RecentActive(ctx, window)
	if err != nil {
		return nil, err
	}
	out := make([]api.MonitorLiveUser, len(users))
	for i, u := range users {
		out[i] = api.MonitorLiveUser{UserID: u.UserID, Username: u.Username, NodeID: u.NodeID, LastSeen: u.LastSeen}
	}
	return out, nil
}
