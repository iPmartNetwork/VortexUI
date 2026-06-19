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

// EnsurePeers makes sure each user has a peer on the WG inbound, allocating a
// keypair and the next free /32 in the inbound's subnet. Returns all peers.
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
	out := make([]domain.WireGuardPeer, 0, len(have))
	for _, p := range have {
		out = append(out, *p)
	}
	return out, nil
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
