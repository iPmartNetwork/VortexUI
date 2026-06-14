package grpc

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	genv1 "github.com/vortexui/vortexui/internal/transport/genv1"
)

// NodeClient is the panel's handle to one remote node. It hides the generated
// stub behind domain-typed methods and owns the dialed connection.
type NodeClient struct {
	nodeID uuid.UUID
	conn   *grpc.ClientConn
	rpc    genv1.NodeServiceClient
}

// Dial opens an mTLS connection to a node agent.
func Dial(nodeID uuid.UUID, address string, creds credentials.TransportCredentials) (*NodeClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("dial node %s: %w", address, err)
	}
	return &NodeClient{nodeID: nodeID, conn: conn, rpc: genv1.NewNodeServiceClient(conn)}, nil
}

// Close releases the underlying connection.
func (c *NodeClient) Close() error { return c.conn.Close() }

// Sync pushes the full desired state to the node.
func (c *NodeClient) Sync(ctx context.Context, cfg *core.GeneratedConfig, coreType domain.CoreType) error {
	req := &genv1.SyncRequest{
		Core:           coreTypeToProto(coreType),
		LogLevel:       cfg.LogLevel,
		UsersByInbound: make(map[string]*genv1.UserList, len(cfg.UsersByInbound)),
	}
	for _, in := range cfg.Inbounds {
		req.Inbounds = append(req.Inbounds, inboundToSpec(in))
	}
	for tag, users := range cfg.UsersByInbound {
		list := &genv1.UserList{Users: make([]*genv1.UserSpec, 0, len(users))}
		for _, u := range users {
			list.Users = append(list.Users, userToSpec(u))
		}
		req.UsersByInbound[tag] = list
	}
	return ackErr(c.rpc.Sync(ctx, req))
}

// AddUser provisions a user on an inbound at runtime.
func (c *NodeClient) AddUser(ctx context.Context, inboundTag string, u *domain.User) error {
	return ackErr(c.rpc.AddUser(ctx, &genv1.AddUserRequest{InboundTag: inboundTag, User: userToSpec(u)}))
}

// RemoveUser deprovisions a user from an inbound at runtime.
func (c *NodeClient) RemoveUser(ctx context.Context, inboundTag string, userID uuid.UUID) error {
	return ackErr(c.rpc.RemoveUser(ctx, &genv1.RemoveUserRequest{InboundTag: inboundTag, UserId: userID.String()}))
}

// Health fetches a live snapshot for failover decisions.
func (c *NodeClient) Health(ctx context.Context) (domain.NodeHealth, error) {
	r, err := c.rpc.Health(ctx, &genv1.HealthRequest{})
	if err != nil {
		return domain.NodeHealth{}, err
	}
	return domain.NodeHealth{
		CPUPercent:  r.GetCpuPercent(),
		MemPercent:  r.GetMemPercent(),
		DiskPercent: r.GetDiskPercent(),
		CoreRunning: r.GetCoreRunning(),
		Connections: int(r.GetConnections()),
	}, nil
}

// ConsumeTraffic opens the long-lived traffic stream and hands every delta to
// ingest until ctx is cancelled or the stream ends. The caller (hub) is expected
// to call this in its own goroutine and reconnect on error with backoff.
func (c *NodeClient) ConsumeTraffic(ctx context.Context, ingest func(domain.TrafficDelta)) error {
	stream, err := c.rpc.StreamTraffic(ctx, &genv1.StreamTrafficRequest{})
	if err != nil {
		return fmt.Errorf("open traffic stream: %w", err)
	}
	for {
		d, err := stream.Recv()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return fmt.Errorf("recv traffic: %w", err)
		}
		ingest(trafficFromProto(d, c.nodeID))
	}
}

// ackErr collapses an (Ack, error) reply into a single error: transport errors
// pass through, and a non-ok Ack becomes an error carrying the node's message.
func ackErr(a *genv1.Ack, err error) error {
	if err != nil {
		return err
	}
	if a != nil && !a.GetOk() {
		return fmt.Errorf("node rejected request: %s", a.GetMessage())
	}
	return nil
}
