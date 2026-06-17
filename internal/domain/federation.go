package domain

import (
	"time"

	"github.com/google/uuid"
)

// FederationPeerStatus tracks the connection state to a peer panel.
type FederationPeerStatus string

const (
	PeerConnected    FederationPeerStatus = "connected"
	PeerDisconnected FederationPeerStatus = "disconnected"
	PeerSyncing      FederationPeerStatus = "syncing"
)

// FederationPeer represents a remote VortexUI panel in the federation.
type FederationPeer struct {
	ID        uuid.UUID            `json:"id"`
	Name      string               `json:"name"`
	Endpoint  string               `json:"endpoint"` // https://panel2.example.com
	APIKey    string               `json:"api_key"`
	Status    FederationPeerStatus `json:"status"`
	SyncUsers bool                 `json:"sync_users"`
	SyncNodes bool                 `json:"sync_nodes"`
	LastSync  *time.Time           `json:"last_sync,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
}

// FederationConfig stores global federation settings.
type FederationConfig struct {
	Enabled     bool   `json:"enabled"`
	ClusterName string `json:"cluster_name"`
	SSOEnabled  bool   `json:"sso_enabled"`
	SyncInterval int   `json:"sync_interval"` // seconds
	SharedSecret string `json:"shared_secret,omitempty"`
}

// DefaultFederationConfig returns sensible defaults.
func DefaultFederationConfig() FederationConfig {
	return FederationConfig{
		Enabled:      false,
		ClusterName:  "vortex-cluster",
		SyncInterval: 60,
	}
}

// FederationSyncEvent records a sync operation between peers.
type FederationSyncEvent struct {
	ID         uuid.UUID `json:"id"`
	PeerID     uuid.UUID `json:"peer_id"`
	PeerName   string    `json:"peer_name,omitempty"`
	Direction  string    `json:"direction"` // "push" | "pull"
	EntityType string    `json:"entity_type"` // "users" | "nodes"
	Count      int       `json:"count"`
	Status     string    `json:"status"` // "success" | "failed"
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
