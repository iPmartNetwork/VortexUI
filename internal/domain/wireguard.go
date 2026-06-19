package domain

import "github.com/google/uuid"

// WireGuardPeer is one client's peer entry on a WireGuard server inbound: a
// keypair plus the tunnel IP assigned to that user within the inbound's subnet.
type WireGuardPeer struct {
	InboundID  uuid.UUID `json:"inbound_id"`
	UserID     uuid.UUID `json:"user_id"`
	PrivateKey string    `json:"private_key"`
	PublicKey  string    `json:"public_key"`
	Address    string    `json:"address"` // e.g. "10.7.0.2"
}
