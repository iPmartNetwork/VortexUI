package service

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"github.com/vortexui/vortexui/internal/core/wireguard"
	"github.com/vortexui/vortexui/internal/domain"
)

// WireGuardPeerRepository persists per-user WireGuard peers for an inbound.
type WireGuardPeerRepository interface {
	Get(ctx context.Context, inboundID, userID uuid.UUID) (*domain.WireGuardPeer, error)
	Create(ctx context.Context, p *domain.WireGuardPeer) error
	Update(ctx context.Context, p *domain.WireGuardPeer) error
	Delete(ctx context.Context, inboundID, userID uuid.UUID) error
	ListByInbound(ctx context.Context, inboundID uuid.UUID) ([]*domain.WireGuardPeer, error)
}

// WireGuardMeshRepository persists WireGuard mesh networks.
type WireGuardMeshRepository interface {
	CreateMesh(ctx context.Context, mesh *domain.WireGuardMesh) error
	GetMesh(ctx context.Context, id uuid.UUID) (*domain.WireGuardMesh, error)
	ListMeshes(ctx context.Context) ([]*domain.WireGuardMesh, error)
	AddMeshPeer(ctx context.Context, peer *domain.WireGuardMeshPeer) error
	ListMeshPeers(ctx context.Context, meshID uuid.UUID) ([]*domain.WireGuardMeshPeer, error)
}

// WireGuardService provisions and lists WireGuard peers.
type WireGuardService struct {
	repo     WireGuardPeerRepository
	meshRepo WireGuardMeshRepository
}

// NewWireGuardService wires the service.
func NewWireGuardService(repo WireGuardPeerRepository, meshRepo WireGuardMeshRepository) *WireGuardService {
	return &WireGuardService{repo: repo, meshRepo: meshRepo}
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

// AllocateIP allocates the next available IP from a CIDR pool.
// Supports IPv4 only currently. Skips network address and .1 (server).
func (s *WireGuardService) AllocateIP(ctx context.Context, in domain.Inbound) (string, error) {
	subnet := wgSubnet(in)
	existing, err := s.repo.ListByInbound(ctx, in.ID)
	if err != nil {
		return "", err
	}
	used := map[string]bool{}
	for _, p := range existing {
		used[p.Address] = true
	}
	return nextFreeIP(subnet, used)
}

// RepairPeers detects duplicate or out-of-range addresses and reassigns them.
func (s *WireGuardService) RepairPeers(ctx context.Context, in domain.Inbound) (*domain.WireGuardRepairReport, error) {
	subnet := wgSubnet(in)
	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet %s: %w", subnet, err)
	}

	peers, err := s.repo.ListByInbound(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	report := &domain.WireGuardRepairReport{
		InboundID: in.ID,
	}

	// Track addresses; detect duplicates and out-of-range
	seen := map[string]*domain.WireGuardPeer{}
	var needRepair []*domain.WireGuardPeer
	used := map[string]bool{}

	for _, p := range peers {
		ip := net.ParseIP(p.Address)
		outOfRange := ip == nil || !ipnet.Contains(ip)
		_, duplicate := seen[p.Address]

		if outOfRange {
			report.OutOfRange++
			needRepair = append(needRepair, p)
		} else if duplicate {
			report.Duplicates++
			needRepair = append(needRepair, p)
		} else {
			seen[p.Address] = p
			used[p.Address] = true
		}
	}

	// Reassign addresses for conflicting peers
	for _, p := range needRepair {
		newIP, err := nextFreeIP(subnet, used)
		if err != nil {
			return nil, fmt.Errorf("subnet full during repair: %w", err)
		}
		oldAddr := p.Address
		p.Address = newIP
		used[newIP] = true
		if err := s.repo.Update(ctx, p); err != nil {
			return nil, fmt.Errorf("update peer %s: %w", p.UserID, err)
		}
		report.Reassigned = append(report.Reassigned, domain.WireGuardReassign{
			UserID:     p.UserID,
			OldAddress: oldAddr,
			NewAddress: newIP,
		})
	}

	return report, nil
}

// GenerateQR produces a QR code PNG of the user's WireGuard client config.
func (s *WireGuardService) GenerateQR(ctx context.Context, in domain.Inbound, user *domain.User, endpointHost string) ([]byte, error) {
	conf, err := s.ClientConfig(ctx, in, user, endpointHost)
	if err != nil {
		return nil, err
	}

	qr, err := qrcode.New(conf, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("generate QR code: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, qr.Image(256)); err != nil {
		return nil, fmt.Errorf("encode QR PNG: %w", err)
	}
	return buf.Bytes(), nil
}

// UpdatePeerSettings updates the MTU and DNS for a specific peer.
func (s *WireGuardService) UpdatePeerSettings(ctx context.Context, inboundID, userID uuid.UUID, settings domain.WireGuardPeerSettings) (*domain.WireGuardPeer, error) {
	peer, err := s.repo.Get(ctx, inboundID, userID)
	if err != nil {
		return nil, fmt.Errorf("get peer: %w", err)
	}

	if settings.MTU > 0 {
		peer.MTU = settings.MTU
	}
	if settings.DNS != "" {
		peer.DNS = settings.DNS
	}

	if err := s.repo.Update(ctx, peer); err != nil {
		return nil, fmt.Errorf("update peer: %w", err)
	}
	return peer, nil
}

// ListPeers returns all peers for an inbound with their current stats.
func (s *WireGuardService) ListPeers(ctx context.Context, inboundID uuid.UUID) ([]*domain.WireGuardPeer, error) {
	return s.repo.ListByInbound(ctx, inboundID)
}

// CreateMesh creates a new WireGuard mesh network between the specified nodes.
// It generates keypairs for each node and assigns IPs from the mesh CIDR.
func (s *WireGuardService) CreateMesh(ctx context.Context, name string, cidr string, nodeIDs []uuid.UUID, endpoints map[uuid.UUID]string) (*domain.WireGuardMesh, error) {
	if cidr == "" {
		cidr = "10.10.0.0/16"
	}

	mesh := &domain.WireGuardMesh{
		ID:        uuid.New(),
		Name:      name,
		CIDR:      cidr,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.meshRepo.CreateMesh(ctx, mesh); err != nil {
		return nil, fmt.Errorf("create mesh: %w", err)
	}

	// Allocate IPs and generate keypairs for each node
	used := map[string]bool{}
	for _, nodeID := range nodeIDs {
		ip, err := nextFreeIP(cidr, used)
		if err != nil {
			return nil, fmt.Errorf("allocate mesh IP for node %s: %w", nodeID, err)
		}
		used[ip] = true

		priv, pub, err := wireguard.GenerateKeypair()
		if err != nil {
			return nil, fmt.Errorf("generate keypair for node %s: %w", nodeID, err)
		}

		endpoint := endpoints[nodeID]
		peer := &domain.WireGuardMeshPeer{
			ID:         uuid.New(),
			MeshID:     mesh.ID,
			NodeID:     nodeID,
			PublicKey:  pub,
			PrivateKey: priv,
			Endpoint:   endpoint,
			Address:    ip,
			KeepAlive:  25,
			CreatedAt:  time.Now(),
		}

		if err := s.meshRepo.AddMeshPeer(ctx, peer); err != nil {
			return nil, fmt.Errorf("add mesh peer: %w", err)
		}
		mesh.Peers = append(mesh.Peers, *peer)
	}

	return mesh, nil
}

// GetMesh retrieves a mesh network by ID.
func (s *WireGuardService) GetMesh(ctx context.Context, meshID uuid.UUID) (*domain.WireGuardMesh, error) {
	mesh, err := s.meshRepo.GetMesh(ctx, meshID)
	if err != nil {
		return nil, err
	}
	peers, err := s.meshRepo.ListMeshPeers(ctx, meshID)
	if err != nil {
		return nil, err
	}
	for _, p := range peers {
		mesh.Peers = append(mesh.Peers, *p)
	}
	return mesh, nil
}

// ListMeshes returns all mesh networks.
func (s *WireGuardService) ListMeshes(ctx context.Context) ([]*domain.WireGuardMesh, error) {
	return s.meshRepo.ListMeshes(ctx)
}
