package domain

import "github.com/google/uuid"

// TorBridgeMode configures a node as a Tor pluggable transport bridge.
type TorBridgeMode string

const (
	TorBridgeDisabled  TorBridgeMode = ""
	TorBridgeObfs4     TorBridgeMode = "obfs4"
	TorBridgeSnowflake TorBridgeMode = "snowflake"
	TorBridgeMeek      TorBridgeMode = "meek"
)

// TorBridgeConfig holds configuration for running a node as a Tor bridge.
type TorBridgeConfig struct {
	NodeID      uuid.UUID     `json:"node_id"`
	Mode        TorBridgeMode `json:"mode"`
	BridgeLine  string        `json:"bridge_line,omitempty"`  // generated bridge line
	ORPort      int           `json:"or_port,omitempty"`      // onion router port
	PTPort      int           `json:"pt_port,omitempty"`      // pluggable transport port
	Fingerprint string        `json:"fingerprint,omitempty"`
	Enabled     bool          `json:"enabled"`
}
