package domain

import "github.com/google/uuid"

// WireGuardMeshPeer represents a node in a WireGuard mesh network.
type WireGuardMeshPeer struct {
	NodeID     uuid.UUID `json:"node_id"`
	PublicKey  string    `json:"public_key"`
	Endpoint   string    `json:"endpoint"`    // host:port
	AllowedIPs string    `json:"allowed_ips"` // CIDR
	KeepAlive  int       `json:"keepalive"`   // seconds, 0 = disabled
}

// WireGuardMesh defines a mesh network topology between nodes.
type WireGuardMesh struct {
	ID    uuid.UUID           `json:"id"`
	Name  string              `json:"name"`
	Peers []WireGuardMeshPeer `json:"peers"`
	CIDR  string              `json:"cidr"` // e.g. "10.10.0.0/16"
}
