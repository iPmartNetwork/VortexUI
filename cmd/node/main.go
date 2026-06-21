// Command node is the VortexUI agent that runs on every proxy server. It serves
// the NodeService gRPC API over mTLS; the panel dials in to push config/users
// and to stream traffic. The agent drives a local Xray-core via the xray driver.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/singbox"
	"github.com/vortexui/vortexui/internal/core/xray"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// agentVersion is reported to the panel in health responses.
const agentVersion = "0.1.0"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, log); err != nil {
		log.Error("node agent exited with error", "err", err)
		os.Exit(1)
	}
	log.Info("node agent shut down cleanly")
}

func run(ctx context.Context, log *slog.Logger) error {
	cfg, err := config.LoadNode()
	if err != nil {
		return err
	}

	// Select the core driver this node runs. Both satisfy core.CoreDriver, so the
	// rest of the agent is engine-agnostic. The panel drives it through the
	// NodeServer; nothing starts until the first Sync.
	var driver core.CoreDriver
	switch cfg.Core {
	case "singbox":
		driver = singbox.New(singbox.Options{BinPath: cfg.CoreBin, ConfigPath: cfg.CoreConfig, APIPort: cfg.APIPort, OmitV2RayAPI: !cfg.SingboxV2RayAPI, Logger: log})
	case "xray", "":
		driver = xray.New(xray.Options{BinPath: cfg.CoreBin, ConfigPath: cfg.CoreConfig, APIPort: cfg.APIPort, Logger: log})
	default:
		return fmt.Errorf("unknown core %q (want xray|singbox)", cfg.Core)
	}
	defer driver.Stop(context.Background())

	srv := vgrpc.NewNodeServer(driver, agentVersion)

	creds, err := vgrpc.ServerCreds(vgrpc.TLSFiles{Cert: cfg.TLSCert, Key: cfg.TLSKey, CA: cfg.TLSCA})
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", cfg.ListenAddr)
	if err != nil {
		return err
	}

	// Serve in the background; stop gracefully when the signal context cancels.
	errCh := make(chan error, 1)
	go func() {
		log.Info("node agent listening (mTLS)", "addr", cfg.ListenAddr, "core", cfg.Core)
		errCh <- srv.Serve(lis,
			grpc.Creds(creds),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             10 * time.Second, // allow client pings every 10s
				PermitWithoutStream: true,
			}),
			grpc.KeepaliveParams(keepalive.ServerParameters{
				Time:    30 * time.Second, // server pings client every 30s
				Timeout: 10 * time.Second,
			}),
		)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutting down node agent")
		srv.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}
