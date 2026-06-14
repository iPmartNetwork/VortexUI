package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	genv1 "github.com/vortexui/vortexui/internal/transport/genv1"
)

// NodeServer adapts a core.CoreDriver to the NodeService gRPC contract. It is
// the process boundary on each proxy server: the panel's calls land here and are
// translated into driver operations.
type NodeServer struct {
	genv1.UnimplementedNodeServiceServer
	driver   core.CoreDriver
	agentVer string

	mu  sync.Mutex
	srv *grpc.Server
}

// NewNodeServer wraps a driver. The driver owns the actual proxy engine.
func NewNodeServer(driver core.CoreDriver, agentVer string) *NodeServer {
	return &NodeServer{driver: driver, agentVer: agentVer}
}

// Serve registers the service and blocks serving on lis until GracefulStop is
// called or the listener fails. Pass server options such as mTLS credentials.
func (s *NodeServer) Serve(lis net.Listener, opts ...grpc.ServerOption) error {
	srv := grpc.NewServer(opts...)
	genv1.RegisterNodeServiceServer(srv, s)
	s.mu.Lock()
	s.srv = srv
	s.mu.Unlock()
	return srv.Serve(lis)
}

// GracefulStop stops the server, letting in-flight RPCs finish. Safe to call
// before Serve (no-op) and from another goroutine.
func (s *NodeServer) GracefulStop() {
	s.mu.Lock()
	srv := s.srv
	s.mu.Unlock()
	if srv != nil {
		srv.GracefulStop()
	}
}

func ack(ok bool, msg string) *genv1.Ack { return &genv1.Ack{Ok: ok, Message: msg} }

// Sync rebuilds the engine-neutral config from the request and (re)starts the
// core. Start is idempotent, so this doubles as a hot reload.
func (s *NodeServer) Sync(ctx context.Context, req *genv1.SyncRequest) (*genv1.Ack, error) {
	cfg := &core.GeneratedConfig{
		LogLevel:       req.GetLogLevel(),
		UsersByInbound: make(map[string][]*domain.User, len(req.GetUsersByInbound())),
	}
	for _, in := range req.GetInbounds() {
		cfg.Inbounds = append(cfg.Inbounds, inboundFromSpec(in))
	}
	for tag, list := range req.GetUsersByInbound() {
		users := make([]*domain.User, 0, len(list.GetUsers()))
		for _, u := range list.GetUsers() {
			users = append(users, userFromSpec(u))
		}
		cfg.UsersByInbound[tag] = users
	}
	if err := s.driver.Start(ctx, cfg); err != nil {
		return ack(false, fmt.Sprintf("sync: %v", err)), nil
	}
	return ack(true, ""), nil
}

func (s *NodeServer) AddUser(ctx context.Context, req *genv1.AddUserRequest) (*genv1.Ack, error) {
	if err := s.driver.AddUser(ctx, req.GetInboundTag(), userFromSpec(req.GetUser())); err != nil {
		return ack(false, err.Error()), nil
	}
	return ack(true, ""), nil
}

func (s *NodeServer) RemoveUser(ctx context.Context, req *genv1.RemoveUserRequest) (*genv1.Ack, error) {
	if err := s.driver.RemoveUser(ctx, req.GetInboundTag(), req.GetUserId()); err != nil {
		return ack(false, err.Error()), nil
	}
	return ack(true, ""), nil
}

// StreamTraffic bridges the driver's delta channel onto the gRPC stream. It ends
// when the driver channel closes or the panel disconnects (stream ctx done).
func (s *NodeServer) StreamTraffic(req *genv1.StreamTrafficRequest, stream grpc.ServerStreamingServer[genv1.TrafficDelta]) error {
	ch, err := s.driver.StreamTraffic(stream.Context())
	if err != nil {
		return err
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case d, ok := <-ch:
			if !ok {
				return nil
			}
			if err := stream.Send(trafficToProto(d)); err != nil {
				return err
			}
		}
	}
}

func (s *NodeServer) Health(ctx context.Context, _ *genv1.HealthRequest) (*genv1.HealthResponse, error) {
	h, err := s.driver.Health(ctx)
	if err != nil {
		return nil, err
	}
	ver, _ := s.driver.Version(ctx)
	return &genv1.HealthResponse{
		CpuPercent:  h.CPUPercent,
		MemPercent:  h.MemPercent,
		DiskPercent: h.DiskPercent,
		CoreRunning: h.CoreRunning,
		Connections: uint32(h.Connections),
		CoreVersion: ver,
		AgentVersion: s.agentVer,
	}, nil
}
