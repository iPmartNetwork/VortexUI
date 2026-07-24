package domain

import (
	"time"

	"github.com/google/uuid"
)

// WireGuardMeshPeer represents a node in a WireGuard mesh network.
type WireGuardMeshPeer struct {
	ID         uuid.UUID `json:"id"`
	MeshID     uuid.UUID `json:"mesh_id"`
	NodeID     uuid.UUID `json:"node_id"`
	PublicKey  string    `json:"public_key"`
	PrivateKey string    `json:"private_key"`
	Endpoint   string    `json:"endpoint"`    // host:port
	Address    string    `json:"address"`     // allocated mesh IP
	KeepAlive  int       `json:"keepalive"`   // seconds, 0 = disabled
	CreatedAt  time.Time `json:"created_at"`
}

// WireGuardMesh defines a mesh network topology between nodes.
type WireGuardMesh struct {
	ID        uuid.UUID           `json:"id"`
	Name      string              `json:"name"`
	Peers     []WireGuardMeshPeer `json:"peers"`
	CIDR      string              `json:"cidr"` // e.g. "10.10.0.0/16"
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// WireGuardMeshLink represents a connection between two mesh peers (full mesh topology).
type WireGuardMeshLink struct {
	FromNodeID uuid.UUID `json:"from_node_id"`
	ToNodeID   uuid.UUID `json:"to_node_id"`
	FromPubKey string    `json:"from_public_key"`
	ToPubKey   string    `json:"to_public_key"`
	Endpoint   string    `json:"endpoint"`
}
