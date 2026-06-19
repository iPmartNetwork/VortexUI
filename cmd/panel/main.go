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
	"syscall"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/events"
	"github.com/vortexui/vortexui/internal/geoip"
	"github.com/vortexui/vortexui/internal/logbuf"
	"github.com/vortexui/vortexui/internal/notify"
	"github.com/vortexui/vortexui/internal/panel/api"
	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
	"github.com/vortexui/vortexui/internal/platform/redis"
	"github.com/vortexui/vortexui/internal/stats"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// version is the panel build version. It defaults to the contents of the VERSION
// file and is overridden at build time via -ldflags "-X main.version=...".
var version = "1.2.0"

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
	agg := stats.New(users, traffic)
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

	h := hub.New(hub.Options{
		Dialer: localAwareDialer(tlsFiles, localID, localDriver),
		Nodes:  nodes,
		Ingest: agg.Ingest,
		Logger: log,
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
	enforcer := service.NewEnforcer(users, h, time.Minute, log)
	enforcer.SetPublisher(bus)
	go enforcer.Run(ctx)

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
	go shareGuard.Run(ctx)
	subSvc := service.NewSubscriptionService(users, nodes)
	syncSvc := service.NewSyncService(store.Inbounds(), users, h, store.Outbounds(), store.Routing(), store.Balancers())
	nodeSvc := service.NewNodeService(nodes, h)
	nodeSvc.SetLogQuerier(h)
	nodeSvc.SetCoreController(h)
	inboundSvc := service.NewInboundService(store.Inbounds(), nodes, syncSvc)
	outboundSvc := service.NewOutboundService(store.Outbounds(), syncSvc)
	routingSvc := service.NewRoutingService(store.Routing(), syncSvc)
	balancerSvc := service.NewBalancerService(store.Balancers(), syncSvc)
	adminSvc := service.NewAdminService(admins)
	overviewSvc := service.NewOverviewService(users, nodes)
	backupSvc := service.NewBackupService(nodes, store.Inbounds(), store.Outbounds(), store.Routing(), store.Balancers(), users, store.Backup())

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
	tokenSvc := service.NewAPITokenService(store.APITokens())

	// Construct all feature services.
	portalSvc := service.NewPortalService(users, store.Tickets(), store.Plans())
	planSvc := service.NewPlanService(store.Plans(), userSvc)
	realitySvc := service.NewRealityScannerService(store.RealityScans(), nodes)
	quotaSvc := service.NewQuotaService(store.QuotaPolicies())
	relaySvc := service.NewRelayService(store.RelayChains(), nodes)
	decoySvc := service.NewDecoyService(store.DecoySites())
	analyticsSvc := service.NewAnalyticsService(store.Analytics())
	probingSvc := service.NewProbingService(store.Probing(), log)
	familySvc := service.NewFamilyService(store.Families(), users)
	referralSvc := service.NewReferralService(store.Referrals(), users)
	dohSvc := service.NewDoHService(store.DoH())
	sniSvc := service.NewSNIService(store.SNIDomains())
	tlsTricksSvc := service.NewTLSTricksService(store.TLSTricks())
	fpSvc := service.NewFingerprintService(store.Fingerprints())
	fedSvc := service.NewFederationService(store.Federation())
	deepLinkSvc := service.NewDeepLinkService(store.DeepLinks())
	quotaNotifySvc := service.NewQuotaNotifyService(store.QuotaNotify())
	subSettingsSvc := service.NewSubSettingsService(store.SubSettings())

	// GeoIP resolver for the "Traffic by Country" analytics. Optional: an empty
	// or unopenable DB path degrades gracefully to a disabled resolver.
	geoResolver, gerr := geoip.Open(cfg.GeoIPDB)
	if gerr != nil {
		log.Warn("geoip database unavailable; Traffic by Country disabled", "path", cfg.GeoIPDB, "err", gerr)
		geoResolver = geoip.Disabled()
	}
	defer func() { _ = geoResolver.Close() }()
	geoSvc := service.NewGeoService(geoResolver, store.UserGeo())

	router := api.NewRouter(api.Deps{
		Handlers: &api.Handlers{
			Auth: authSvc, Users: userSvc, Sub: subSvc,
			Nodes: nodeSvc, Inbounds: inboundSvc, Admins: adminSvc, Devices: devices,
			Outbounds: outboundSvc, Routing: routingSvc, Balancers: balancerSvc,
			Overview: overviewSvc, Backup: backupSvc,
			Plans:    planSvc,
			Online: online, Logs: logBuf, Audit: store.Audit(),
			Repo: users, Traffic: traffic,
			Throttle: api.NewLoginThrottle(5, 15*time.Minute),
			Events:   bus,
			SubSettings: subSettingsSvc,
			Geo:         geoSvc,
		},
		APITokens:   &api.APITokenHandlers{Svc: tokenSvc},
		Portal:      &api.PortalHandlers{Portal: portalSvc, Issuer: issuer},
		Reality:     &api.RealityHandlers{Scanner: realitySvc},
		Quota:       &api.QuotaHandlers{Quota: quotaSvc},
		Relay:       &api.RelayHandlers{Relay: relaySvc},
		Decoy:       &api.DecoyHandlers{Decoy: decoySvc},
		Analytics:   &api.AnalyticsHandlers{Analytics: analyticsSvc},
		Migration:   &api.MigrationHandlers{Migration: migration},
		Probing:     &api.ProbingHandlers{Probing: probingSvc},
		Family:      &api.FamilyHandlers{Family: familySvc},
		Referral:    &api.ReferralHandlers{Referral: referralSvc},
		DoH:         &api.DoHHandlers{DoH: dohSvc},
		SNI:         &api.SNIHandlers{SNI: sniSvc},
		TLSTricks:   &api.TLSTricksHandlers{Tricks: tlsTricksSvc},
		Fingerprint: &api.FingerprintHandlers{FP: fpSvc},
		Federation:  &api.FederationHandlers{Fed: fedSvc},
		DeepLink:    &api.DeepLinkHandlers{DeepLink: deepLinkSvc},
		QuotaNotify: &api.QuotaNotifyHandlers{QN: quotaNotifySvc},
		SubSettings: &api.SubSettingsHandlers{Svc: subSettingsSvc},
		Monitor:     &api.MonitorHandlers{Hub: h, Nodes: nodes, Monitor: monitorAdapter{store.Monitor()}},
		Issuer:      issuer,
		Auth:        authSvc,
		Limiter:     limiter,
		Audit:       store.Audit(),
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
