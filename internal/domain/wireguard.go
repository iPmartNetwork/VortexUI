package domain

import (
	"time"

	"github.com/google/uuid"
)

// WireGuardPeer is one client's peer entry on a WireGuard server inbound: a
// keypair plus the tunnel IP assigned to that user within the inbound's subnet.
type WireGuardPeer struct {
	InboundID     uuid.UUID  `json:"inbound_id"`
	UserID        uuid.UUID  `json:"user_id"`
	PrivateKey    string     `json:"private_key"`
	PublicKey     string     `json:"public_key"`
	Address       string     `json:"address"` // e.g. "10.7.0.2"
	MTU           int        `json:"mtu"`
	DNS           string     `json:"dns"`
	LastHandshake *time.Time `json:"last_handshake,omitempty"`
	TxBytes       int64      `json:"tx_bytes"`
	RxBytes       int64      `json:"rx_bytes"`
}

// WireGuardPeerSettings holds per-peer configurable settings (MTU and DNS).
type WireGuardPeerSettings struct {
	MTU int    `json:"mtu"`
	DNS string `json:"dns"`
}

// WireGuardRepairReport describes the result of a peer repair operation.
type WireGuardRepairReport struct {
	InboundID    uuid.UUID            `json:"inbound_id"`
	Duplicates   int                  `json:"duplicates"`
	OutOfRange   int                  `json:"out_of_range"`
	Reassigned   []WireGuardReassign  `json:"reassigned"`
}

// WireGuardReassign records an IP reassignment during repair.
type WireGuardReassign struct {
	UserID     uuid.UUID `json:"user_id"`
	OldAddress string    `json:"old_address"`
	NewAddress string    `json:"new_address"`
}
