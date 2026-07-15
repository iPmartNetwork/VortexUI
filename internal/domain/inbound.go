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
	// ProtoHysteria is the original Hysteria v1 protocol (sing-box "hysteria"),
	// distinct from the newer ProtoHysteria2. Like hysteria2/tuic it is
	// UDP-native (carries no stream transport).
	ProtoHysteria Protocol = "hysteria"
	// ProtoShadowTLS fronts a real TLS handshake to disguise the inbound
	// (sing-box "shadowtls", v3). It is TCP-based and supplies its own TLS
	// layer via the handshake target, so it carries no separate tls block.
	ProtoShadowTLS Protocol = "shadowtls"
	// ProtoAnyTLS is the AnyTLS protocol (sing-box "anytls"); TCP-based and
	// requires a TLS layer.
	ProtoAnyTLS Protocol = "anytls"
	// ProtoSocks is a SOCKS5 utility proxy inbound. It is a plain TCP proxy
	// (security none) that carries NO stream transport — the network value is
	// irrelevant for it. Per-user auth reuses the existing credentials.
	ProtoSocks Protocol = "socks"
	// ProtoHTTP is an HTTP CONNECT utility proxy inbound. Like socks it is a
	// plain TCP proxy (security none) carrying NO stream transport.
	//
	// NOTE: the Protocol value "http" lives in a DIFFERENT namespace from the
	// "http"/"h2" stream-transport token (which is keyed on Inbound.Network in
	// streamSettings/transportBlock). Protocol is Inbound.Protocol; the transport
	// is Inbound.Network — separate fields, so there is no collision.
	ProtoHTTP Protocol = "http"
	// ProtoNaive is a NaiveProxy inbound (sing-box "naive"). It MANDATES TLS and
	// carries no stream transport of its own.
	ProtoNaive Protocol = "naive"
	// ProtoDokodemo is the xray dokodemo-door inbound: a transparent/redirect
	// listener with NO per-user auth. It carries no stream transport; its target
	// address/port and network come from Inbound.Raw["dokodemo"]. NOTE the xray
	// WIRE protocol name is "dokodemo-door" (mapped in the renderer), not the
	// "dokodemo" value used here.
	ProtoDokodemo Protocol = "dokodemo"
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
	// Core optionally overrides the parent node's default engine for this inbound.
	Core     CoreType  `json:"core,omitempty"`
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

	// GeoPolicy restricts which countries can connect to this inbound.
	// Empty = no restriction (all countries allowed).
	GeoPolicy *GeoPolicy `json:"geo_policy,omitempty"`

	// EvasionProfileID optionally links a reusable DPI-evasion policy (fragment,
	// reality keys, fingerprint) so operators apply hardening with one click.
	EvasionProfileID *uuid.UUID `json:"evasion_profile_id,omitempty"`

	// Raw lets advanced operators override generated settings with native
	// Xray/sing-box JSON for the parts the abstraction does not yet model.
	Raw map[string]any `json:"raw,omitempty"`

	Enabled bool `json:"enabled"`
}

// InboundListItem is an inbound row with its parent node name for fleet views.
type InboundListItem struct {
	Inbound
	NodeName string `json:"node_name"`
}

// UserProxy binds a User to an Inbound. This many-to-many join is what makes the
// model user-centric instead of inbound-centric.
type UserProxy struct {
	UserID    uuid.UUID `json:"user_id"`
	InboundID uuid.UUID `json:"inbound_id"`
}
