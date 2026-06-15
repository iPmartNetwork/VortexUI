package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/singbox"
	"github.com/vortexui/vortexui/internal/core/xray"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/panel/service"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// Compile-time proof that the concrete gRPC client satisfies the hub's narrow
// NodeConn port, and that the hub in turn satisfies the user service's NodeOps
// port. If a signature ever drifts, the build breaks here rather than at runtime.
var (
	_ hub.NodeConn          = (*vgrpc.NodeClient)(nil)
	_ hub.NodeConn          = (*hub.LocalConn)(nil)
	_ service.NodeOps       = (*hub.Hub)(nil)
	_ service.NodeRegistrar = (*hub.Hub)(nil)
	_ service.Syncer        = (*hub.Hub)(nil)
)

// nodeDialer builds the hub.Dialer used in production: every node connection is
// an mTLS gRPC link. The node's address host is used as the expected TLS server
// name so the node certificate's SAN must include that IP or hostname.
func nodeDialer(tls vgrpc.TLSFiles) hub.Dialer {
	return func(n *domain.Node) (hub.NodeConn, error) {
		// Extract host from "host:port" address for TLS ServerName verification.
		host := n.Address
		if idx := len(host) - 1; idx > 0 {
			for i := idx; i >= 0; i-- {
				if host[i] == ':' {
					host = host[:i]
					break
				}
			}
		}
		creds, err := vgrpc.ClientCreds(tls, host)
		if err != nil {
			return nil, err
		}
		return vgrpc.Dial(n.ID, n.Address, creds)
	}
}

// localAwareDialer wraps the gRPC dialer so the one local node (if configured)
// is served by an in-process LocalConn instead of dialing a remote agent.
func localAwareDialer(tls vgrpc.TLSFiles, localID uuid.UUID, localDriver core.CoreDriver) hub.Dialer {
	grpcDial := nodeDialer(tls)
	return func(n *domain.Node) (hub.NodeConn, error) {
		if localDriver != nil && n.ID == localID {
			return hub.NewLocalConn(n.ID, localDriver), nil
		}
		return grpcDial(n)
	}
}

// buildLocalDriver constructs the in-process core driver for the local node,
// mirroring the node agent's selection logic.
func buildLocalDriver(cfg *config.Panel, log *slog.Logger) (core.CoreDriver, error) {
	switch cfg.Core {
	case "singbox":
		return singbox.New(singbox.Options{BinPath: cfg.CoreBin, ConfigPath: cfg.CoreConfig, APIPort: cfg.CoreAPIPort, Logger: log}), nil
	case "xray", "":
		return xray.New(xray.Options{BinPath: cfg.CoreBin, ConfigPath: cfg.CoreConfig, APIPort: cfg.CoreAPIPort, Logger: log}), nil
	default:
		return nil, fmt.Errorf("unknown core %q (want xray|singbox)", cfg.Core)
	}
}

// ensureLocalNode returns the local node's DB record, creating it on first run
// and keeping its address/core in sync with config on subsequent starts. The
// record lets inbounds/users attach to the local node exactly like a remote one.
func ensureLocalNode(ctx context.Context, nodes port.NodeRepository, cfg *config.Panel) (*domain.Node, error) {
	list, err := nodes.List(ctx)
	if err != nil {
		return nil, err
	}
	core := domain.CoreType(cfg.Core)
	if core == "" {
		core = domain.CoreXray
	}
	for _, n := range list {
		if n.Name == cfg.LocalNodeName {
			if n.Address != cfg.LocalNodeHost || n.Core != core {
				n.Address = cfg.LocalNodeHost
				n.Core = core
				if err := nodes.Update(ctx, n); err != nil {
					return nil, err
				}
			}
			return n, nil
		}
	}
	n := &domain.Node{
		ID:         uuid.New(),
		Name:       cfg.LocalNodeName,
		Address:    cfg.LocalNodeHost,
		Core:       core,
		Status:     domain.NodeDisconnected,
		UsageRatio: 1,
		CreatedAt:  time.Now(),
	}
	if err := nodes.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}
