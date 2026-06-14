package main

import (
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/hub"
	"github.com/vortexui/vortexui/internal/panel/service"
	vgrpc "github.com/vortexui/vortexui/internal/transport/grpc"
)

// Compile-time proof that the concrete gRPC client satisfies the hub's narrow
// NodeConn port, and that the hub in turn satisfies the user service's NodeOps
// port. If a signature ever drifts, the build breaks here rather than at runtime.
var (
	_ hub.NodeConn        = (*vgrpc.NodeClient)(nil)
	_ service.NodeOps     = (*hub.Hub)(nil)
	_ service.NodeRegistrar = (*hub.Hub)(nil)
	_ service.Syncer      = (*hub.Hub)(nil)
)

// nodeDialer builds the hub.Dialer used in production: every node connection is
// an mTLS gRPC link. The node's Name is used as the expected TLS server name, so
// each node's certificate SAN must match its registered name.
func nodeDialer(tls vgrpc.TLSFiles) hub.Dialer {
	return func(n *domain.Node) (hub.NodeConn, error) {
		creds, err := vgrpc.ClientCreds(tls, n.Name)
		if err != nil {
			return nil, err
		}
		return vgrpc.Dial(n.ID, n.Address, creds)
	}
}
