package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// memWGRepo is an in-memory WireGuardPeerRepository for exercising the service
// without a database.
type memWGRepo struct {
	peers []*domain.WireGuardPeer
}

func (r *memWGRepo) Get(_ context.Context, inboundID, userID uuid.UUID) (*domain.WireGuardPeer, error) {
	for _, p := range r.peers {
		if p.InboundID == inboundID && p.UserID == userID {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *memWGRepo) Create(_ context.Context, p *domain.WireGuardPeer) error {
	cp := *p
	r.peers = append(r.peers, &cp)
	return nil
}

func (r *memWGRepo) ListByInbound(_ context.Context, inboundID uuid.UUID) ([]*domain.WireGuardPeer, error) {
	var out []*domain.WireGuardPeer
	for _, p := range r.peers {
		if p.InboundID == inboundID {
			out = append(out, p)
		}
	}
	return out, nil
}

func wgInbound() domain.Inbound {
	return domain.Inbound{
		ID:       uuid.New(),
		NodeID:   uuid.New(),
		Tag:      "wg-in",
		Protocol: domain.ProtoWireGuard,
		Port:     51820,
		Raw: map[string]any{
			"wireguard": map[string]any{
				"subnet":     "10.7.0.0/24",
				"public_key": "SERVER_PUBLIC_KEY",
			},
		},
	}
}

// EnsurePeers must return exactly one peer per passed (bound) user and must not
// include DB peers for users who are no longer bound.
func TestEnsurePeersReturnsOnlyBoundUsers(t *testing.T) {
	repo := &memWGRepo{}
	svc := NewWireGuardService(repo)
	in := wgInbound()
	u1 := &domain.User{ID: uuid.New(), Username: "alice"}
	u2 := &domain.User{ID: uuid.New(), Username: "bob"}

	// Bind both users.
	peers, err := svc.EnsurePeers(context.Background(), in, []*domain.User{u1, u2})
	if err != nil {
		t.Fatalf("ensure peers: %v", err)
	}
	if len(peers) != 2 {
		t.Fatalf("want 2 peers, got %d", len(peers))
	}

	// Now only u1 is bound: u2 should drop out of the returned set even though a
	// peer row still exists for key stability.
	peers, err = svc.EnsurePeers(context.Background(), in, []*domain.User{u1})
	if err != nil {
		t.Fatalf("ensure peers (rebound): %v", err)
	}
	if len(peers) != 1 {
		t.Fatalf("want 1 peer for the single bound user, got %d", len(peers))
	}
	if peers[0].UserID != u1.ID {
		t.Fatalf("want peer for u1, got %s", peers[0].UserID)
	}
	// The peer row for u2 is retained in the DB.
	if len(repo.peers) != 2 {
		t.Fatalf("want 2 retained peer rows, got %d", len(repo.peers))
	}
}

func TestClientConfigContents(t *testing.T) {
	repo := &memWGRepo{}
	svc := NewWireGuardService(repo)
	in := wgInbound()
	u := &domain.User{ID: uuid.New(), Username: "carol"}

	conf, err := svc.ClientConfig(context.Background(), in, u, "vpn.example.com")
	if err != nil {
		t.Fatalf("client config: %v", err)
	}

	peer, err := repo.Get(context.Background(), in.ID, u.ID)
	if err != nil {
		t.Fatalf("get peer: %v", err)
	}

	for _, want := range []string{
		"Address = " + peer.Address + "/32",
		"PublicKey = SERVER_PUBLIC_KEY",
		"Endpoint = vpn.example.com:51820",
		"DNS = 1.1.1.1",
		"MTU = 1420",
	} {
		if !strings.Contains(conf, want) {
			t.Errorf("conf missing %q\n---\n%s", want, conf)
		}
	}
}
