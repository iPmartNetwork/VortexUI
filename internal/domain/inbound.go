package domain

import "github.com/google/uuid"

// Protocol enumerates the transport-layer proxy protocols an inbound can speak.
type Protocol string

const (
	ProtoVMess       Protocol = "vmess"
	ProtoVLESS       Protocol = "vless"
	ProtoTrojan      Protocol = "trojan"
	ProtoShadowsocks Protocol = "shadowsocks"
	ProtoHysteria2   Protocol = "hysteria2"
	ProtoTUIC        Protocol = "tuic"
	ProtoWireGuard   Protocol = "wireguard"
)

// Security is the TLS-layer obfuscation applied to an inbound.
type Security string

const (
	SecurityNone    Security = "none"
	SecurityTLS     Security = "tls"
	SecurityReality Security = "reality"
)

// Inbound is a listening endpoint on a node. Users are attached to inbounds via
// UserProxy bindings; the same user can be bound to many inbounds at once.
type Inbound struct {
	ID       uuid.UUID `json:"id"`
	NodeID   uuid.UUID `json:"node_id"`
	Tag      string    `json:"tag"` // unique per node; used as the core inbound tag
	Protocol Protocol  `json:"protocol"`
	Listen   string    `json:"listen"`
	Port     int       `json:"port"`

	// Transport / TLS layer.
	Network   string   `json:"network"` // tcp, ws, grpc, httpupgrade, xhttp
	Security  Security `json:"security"`
	SNI       []string `json:"sni,omitempty"`
	Path      string   `json:"path,omitempty"`
	Host      []string `json:"host,omitempty"`
	Flow      string   `json:"flow,omitempty"`

	// SpeedLimit sets per-user download speed in bytes/sec for this inbound.
	// 0 means unlimited. Applied via Xray's policy.levels or sing-box's
	// speed_limit field.
	SpeedLimit int64 `json:"speed_limit,omitempty"`

	// EvasionProfileID optionally links a reusable DPI-evasion policy (fragment,
	// reality keys, fingerprint) so operators apply hardening with one click.
	EvasionProfileID *uuid.UUID `json:"evasion_profile_id,omitempty"`

	// Raw lets advanced operators override generated settings with native
	// Xray/sing-box JSON for the parts the abstraction does not yet model.
	Raw map[string]any `json:"raw,omitempty"`

	Enabled bool `json:"enabled"`
}

// UserProxy binds a User to an Inbound. This many-to-many join is what makes the
// model user-centric instead of inbound-centric.
type UserProxy struct {
	UserID    uuid.UUID `json:"user_id"`
	InboundID uuid.UUID `json:"inbound_id"`
}
