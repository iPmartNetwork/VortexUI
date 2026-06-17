package domain

import (
	"time"

	"github.com/google/uuid"
)

// RelayHop represents a single hop in a relay chain.
type RelayHop struct {
	Type    string `json:"type"`    // "cdn" | "relay" | "worker"
	Address string `json:"address"` // IP/domain of the hop
	Port    int    `json:"port"`
	Protocol string `json:"protocol"` // "ws" | "grpc" | "tcp"
	SNI     string `json:"sni,omitempty"`
	Path    string `json:"path,omitempty"`
	Host    string `json:"host,omitempty"`
	Note    string `json:"note,omitempty"`
}

// RelayChain defines a multi-hop relay/CDN path from client to final node.
type RelayChain struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	NodeID    uuid.UUID  `json:"node_id"` // final destination node
	Hops      []RelayHop `json:"hops"`
	Enabled   bool       `json:"enabled"`
	CreatedAt time.Time  `json:"created_at"`
}
