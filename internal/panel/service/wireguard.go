package service

import (
	"context"
	"fmt"
	"net"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core/wireguard"
	"github.com/vortexui/vortexui/internal/domain"
)

// WireGuardPeerRepository persists per-user WireGuard peers for an inbound.
type WireGuardPeerRepository interface {
	Get(ctx context.Context, inboundID, userID uuid.UUID) (*domain.WireGuardPeer, error)
	Create(ctx context.Context, p *domain.WireGuardPeer) error
	ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.WireGuardPeer, error)
}

// WireGuardService provisions and lists WireGuard peers.
type WireGuardService struct {
	repo WireGuardPeerRepository
}

// NewWireGuardService wires the service.
func NewWireGuardService(repo WireGuardPeerRepository) *WireGuardService {
	return &WireGuardService{repo: repo}
}

// EnsurePeers makes sure each passed user has a peer on the WG inbound,
// allocating a keypair and the next free /32 in the inbound's subnet for any
// user that lacks one. The returned slice contains exactly one peer per passed
// user (and nothing for users not in the slice), so the rendered server config
// includes only currently-bound users and drops anyone who has been unbound.
// Existing peer rows for unbound users are left in the DB for key stability.
func (s *WireGuardService) EnsurePeers(ctx context.Context, in domain.Inbound, users []*domain.User) ([]domain.WireGuardPeer, error) {
	subnet := wgSubnet(in)
	existing, err := s.repo.ListByInbound(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{}
	have := map[uuid.UUID]*domain.WireGuardPeer{}
	for _, p := range existing {
		used[p.Address] = true
		have[p.UserID] = p
	}
	for _, u := range users {
		if _, ok := have[u.ID]; ok {
			continue
		}
		ip, err := nextFreeIP(subnet, used)
		if err != nil {
			return nil, err
		}
		priv, pub, err := wireguard.GenerateKeypair()
		if err != nil {
			return nil, err
		}
		p := &domain.WireGuardPeer{InboundID: in.ID, UserID: u.ID, PrivateKey: priv, PublicKey: pub, Address: ip}
		if err := s.repo.Create(ctx, p); err != nil {
			return nil, err
		}
		used[ip] = true
		have[u.ID] = p
	}
	// Return exactly one peer per passed user, preserving the caller's order.
	out := make([]domain.WireGuardPeer, 0, len(users))
	for _, u := range users {
		if p, ok := have[u.ID]; ok {
			out = append(out, *p)
		}
	}
	return out, nil
}

// ClientConfig returns the WireGuard .conf text for one user on a WG inbound.
// endpointHost is the node's public host (caller resolves it). It ensures the
// user has a peer first.
func (s *WireGuardService) ClientConfig(ctx context.Context, in domain.Inbound, user *domain.User, endpointHost string) (string, error) {
	if _, err := s.EnsurePeers(ctx, in, []*domain.User{user}); err != nil {
		return "", err
	}
	peer, err := s.repo.Get(ctx, in.ID, user.ID)
	if err != nil {
		return "", err
	}
	wg, _ := in.Raw["wireguard"].(map[string]any)
	serverPub, _ := wg["public_key"].(string)
	dns, _ := wg["dns"].(string)
	if dns == "" {
		dns = "1.1.1.1"
	}
	mtu := 1420
	switch v := wg["mtu"].(type) {
	case int:
		if v != 0 {
			mtu = v
		}
	case float64:
		if v != 0 {
			mtu = int(v)
		}
	}
	listenPort := in.Port
	switch v := wg["listen_port"].(type) {
	case int:
		if v != 0 {
			listenPort = v
		}
	case float64:
		if v != 0 {
			listenPort = int(v)
		}
	}
	conf := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = %s
MTU = %d

[Peer]
PublicKey = %s
Endpoint = %s:%d
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`, peer.PrivateKey, peer.Address, dns, mtu, serverPub, endpointHost, listenPort)
	return conf, nil
}

func wgSubnet(in domain.Inbound) string {
	if m, ok := in.Raw["wireguard"].(map[string]any); ok {
		if s, _ := m["subnet"].(string); s != "" {
			return s
		}
	}
	return "10.7.0.0/24"
}

// nextFreeIP returns the next unused host IP in cidr, skipping the network
// address and the .1 (reserved for the server). IPv4 only.
func nextFreeIP(cidr string, used map[string]bool) (string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", err
	}
	base := ipnet.IP.Mask(ipnet.Mask).To4()
	if base == nil {
		return "", fmt.Errorf("only IPv4 subnets are supported")
	}
	// Reserve the network address (base) and the server address (base + 1).
	server := make(net.IP, len(base))
	copy(server, base)
	server[len(server)-1]++
	reserved := map[string]bool{base.String(): true, server.String(): true}

	ip := make(net.IP, len(base))
	copy(ip, base)
	for {
		// advance to the next address
		for i := len(ip) - 1; i >= 0; i-- {
			ip[i]++
			if ip[i] != 0 {
				break
			}
		}
		if !ipnet.Contains(ip) {
			break
		}
		cand := ip.String()
		if reserved[cand] || used[cand] {
			continue
		}
		return cand, nil
	}
	return "", fmt.Errorf("wireguard subnet %s is full", cidr)
}
