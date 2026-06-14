package service

import (
	"context"
	"net"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/subscription"
)

// defaultStaleAfter is how long without a heartbeat before a node is pruned from
// subscriptions. Generous relative to the health interval so brief gaps don't
// flap a node out of clients' configs.
const defaultStaleAfter = 90 * time.Second

// SubscriptionService resolves a subscription token into the set of proxies a
// client should receive. It joins the user's credentials with each bound
// inbound's transport and the hosting node's public address, and prunes inbounds
// on nodes known to be unhealthy so clients aren't handed dead endpoints.
type SubscriptionService struct {
	users      port.UserRepository
	nodes      port.NodeRepository
	staleAfter time.Duration
	now        func() time.Time
}

// NewSubscriptionService wires the service.
func NewSubscriptionService(users port.UserRepository, nodes port.NodeRepository) *SubscriptionService {
	return &SubscriptionService{users: users, nodes: nodes, staleAfter: defaultStaleAfter, now: time.Now}
}

// SubResult bundles the resolved proxies with the owning user so the handler can
// render the body and emit usage headers.
type SubResult struct {
	User    *domain.User
	Proxies []subscription.Proxy
}

// Build looks up the user by token and assembles a Proxy per enabled inbound.
func (s *SubscriptionService) Build(ctx context.Context, token string) (*SubResult, error) {
	user, err := s.users.GetBySubToken(ctx, token)
	if err != nil {
		return nil, err
	}
	return s.buildFor(ctx, user)
}

// BuildForUser is the admin-facing counterpart of Build: it resolves a user's
// proxies by user id (not the opaque token) so the panel can show an operator a
// user's subscription links and QR codes on demand.
func (s *SubscriptionService) BuildForUser(ctx context.Context, userID uuid.UUID) (*SubResult, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildFor(ctx, user)
}

// buildFor assembles a Proxy per enabled inbound for a resolved user. Disabled
// inbounds and unreachable node lookups are skipped so one bad inbound never
// breaks the whole subscription.
func (s *SubscriptionService) buildFor(ctx context.Context, user *domain.User) (*SubResult, error) {
	inbounds, err := s.users.InboundsFor(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	now := s.now()
	cache := map[uuid.UUID]*nodeInfo{} // resolve each node once
	var proxies []subscription.Proxy
	for _, in := range inbounds {
		if !in.Enabled {
			continue
		}
		info, ok := cache[in.NodeID]
		if !ok {
			info = s.resolveNode(ctx, in.NodeID, user.Username, now)
			cache[in.NodeID] = info
		}
		if info == nil || !info.live {
			continue // node missing or known-unhealthy: prune from the subscription
		}
		proxies = append(proxies, buildProxy(user, in, info.host, info.name))
	}
	return &SubResult{User: user, Proxies: proxies}, nil
}

type nodeInfo struct {
	host string
	name string
	live bool
}

func (s *SubscriptionService) resolveNode(ctx context.Context, nodeID uuid.UUID, username string, now time.Time) *nodeInfo {
	node, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil
	}
	return &nodeInfo{
		host: hostOf(node.Address),
		name: username + " @ " + node.Name,
		live: node.Live(now, s.staleAfter),
	}
}

func buildProxy(u *domain.User, in domain.Inbound, host, name string) subscription.Proxy {
	p := subscription.Proxy{
		Name:       name,
		Protocol:   in.Protocol,
		Host:       host,
		Port:       in.Port,
		Network:    in.Network,
		Security:   string(in.Security),
		Path:       in.Path,
		Flow:       in.Flow,
		SNI:        first(in.SNI),
		HostHeader: first(in.Host),
	}
	switch in.Protocol {
	case domain.ProtoVLESS:
		p.UUID = u.Proxies.VLESSUUID.String()
	case domain.ProtoVMess:
		p.UUID = u.Proxies.VMessUUID.String()
	case domain.ProtoTrojan:
		p.Password = u.Proxies.TrojanPass
	case domain.ProtoShadowsocks:
		p.Password = u.Proxies.ShadowsocksP
		p.SSMethod = u.Proxies.SSMethod
	}
	// REALITY clients need the public key + short id; pull them from the
	// inbound's neutral reality params and prefer its server name as the SNI.
	if in.Security == domain.SecurityReality {
		rp := reality.ParseParams(in.Raw["reality"])
		p.PublicKey = rp.PublicKey
		p.ShortID = first(rp.ShortIDs)
		if len(rp.ServerNames) > 0 {
			p.SNI = rp.ServerNames[0]
		}
		p.Fingerprint = "chrome"
	}
	return p
}

// hostOf extracts the host from a "host:port" node address, tolerating a bare host.
func hostOf(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

func first(ss []string) string {
	if len(ss) > 0 {
		return ss[0]
	}
	return ""
}
