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

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/api"
	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/platform/postgres"
	"github.com/vortexui/vortexui/internal/platform/redis"
	"github.com/vortexui/vortexui/internal/stats"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

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

	if err := run(ctx, log, cfg); err != nil {
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
func run(ctx context.Context, log *slog.Logger, cfg *config.Panel) error {
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
	h := hub.New(hub.Options{
		Dialer: nodeDialer(tlsFiles),
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
	if rc, err := redis.Open(ctx, cfg.RedisURL); err != nil {
		log.Warn("redis unavailable; login rate limiting and device limits disabled", "err", err)
	} else {
		defer rc.Close()
		limiter = rc.RateLimiter()
		devices = rc.Devices()
	}

	// 4. Services + 5. HTTP API.
	issuer := auth.NewIssuer([]byte(cfg.JWTSecret), cfg.JWTTTL)
	// Enforcement loop: disables + de-provisions users who hit their cap/expiry.
	enforcer := service.NewEnforcer(users, h, time.Minute, log)
	go enforcer.Run(ctx)

	// Reset loop: zeroes used traffic on schedule and re-activates quota-limited
	// users — the complement to enforcement.
	resetter := service.NewResetter(users, h, time.Hour, log)
	go resetter.Run(ctx)

	authSvc := service.NewAuthService(admins, issuer)
	userSvc := service.NewUserService(users, h)
	subSvc := service.NewSubscriptionService(users, nodes)
	syncSvc := service.NewSyncService(store.Inbounds(), users, h)
	nodeSvc := service.NewNodeService(nodes, h)
	inboundSvc := service.NewInboundService(store.Inbounds(), syncSvc)
	adminSvc := service.NewAdminService(admins)

	// Wire failover migration into the hub now that its dependencies exist (the
	// migration service provisions onto the target via the hub itself).
	migration := service.NewMigrationService(store.Inbounds(), users, users, users, h, log)
	h.SetOnFailover(func(fctx context.Context, failed, target *domain.Node) {
		if err := migration.Migrate(fctx, failed, target); err != nil {
			log.Error("failover migration failed", "failed", failed.Name, "err", err)
		}
	})

	// On (re)connect: repopulate the node with its own config, then shed any
	// temporary copies that failover parked on other nodes while it was down.
	h.SetOnConnect(func(cctx context.Context, node *domain.Node) {
		if err := syncSvc.Resync(cctx, node.ID); err != nil {
			log.Error("node resync on connect failed", "node", node.Name, "err", err)
		}
		if err := migration.MigrateBack(cctx, node); err != nil {
			log.Error("migrate-back failed", "node", node.Name, "err", err)
		}
	})
	router := api.NewRouter(api.Deps{
		Handlers: &api.Handlers{
			Auth: authSvc, Users: userSvc, Sub: subSvc,
			Nodes: nodeSvc, Inbounds: inboundSvc, Admins: adminSvc, Devices: devices,
			Repo: users, Traffic: traffic,
		},
		Issuer:  issuer,
		Auth:    authSvc,
		Limiter: limiter,
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
