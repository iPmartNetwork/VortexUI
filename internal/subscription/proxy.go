// Package subscription turns a user's bound inbounds into the various client
// configuration formats (share links, Clash, sing-box). It is pure: it takes
// normalized Proxy values and renders bytes, with no knowledge of storage or
// HTTP, so every renderer is unit-tested in isolation.
package subscription

import "github.com/vortexui/vortexui/internal/domain"

// Proxy is a single, fully-resolved outbound endpoint a client can connect to.
// It flattens a user's credentials together with one inbound's transport and the
// node's public host, which is everything each renderer needs.
type Proxy struct {
	Name     string          // display label, e.g. "alice @ de-1"
	Protocol domain.Protocol
	Host     string          // node public host/IP
	Port     int

	// Transport / TLS.
	Network    string // tcp, ws, grpc
	Security   string // none, tls, reality
	SNI        string
	Path       string
	HostHeader string
	Flow       string
	// AllowInsecure tells clients to skip TLS certificate verification — set for
	// inbounds using an auto-generated self-signed certificate so the handshake
	// succeeds instead of timing out on an untrusted cert.
	AllowInsecure bool

	// Credentials (only the field relevant to Protocol is populated).
	UUID     string // vmess / vless
	Password string // trojan / shadowsocks
	SSMethod string // shadowsocks cipher

	// REALITY parameters, populated only when Security == "reality". The client
	// needs the public key (paired with the inbound's private key) and a short
	// ID to complete the REALITY handshake.
	PublicKey   string
	ShortID     string
	Fingerprint string // uTLS fingerprint, e.g. "chrome"

	// ECH enables Encrypted Client Hello when the host supports it.
	// When true, sing-box clients will attempt ECH negotiation.
	ECH bool

	// Hysteria2-specific options sourced from the inbound's Raw["hysteria2"] block.
	Hy2Obfs string // Salamander obfuscation password
	Hy2Up   int    // upstream bandwidth hint (Mbps)
	Hy2Down int    // downstream bandwidth hint (Mbps)

	// Optional Marzban-style host overrides projected from a SubHost. They are
	// additive: a zero-value Proxy (empty ALPN, Mux=false, empty Fragment)
	// renders byte-identically to before these fields existed, so inbounds with
	// no enabled hosts see no change.
	ALPN     []string // negotiated ALPN protocols, e.g. ["h2","http/1.1"]
	Mux      bool     // enable client-side stream multiplexing (smux/multiplex)
	Fragment string   // TLS-hello fragment setting "length,interval,packet"

	// Padding carries the random TLS padding size range (e.g. "100-200"). When
	// non-empty, renderers emit client-side padding to defeat length-based DPI.
	Padding string

	// Port hopping: when PortEnd > 0 the inbound listens on [Port, PortEnd] and
	// clients rotate ports at HopInterval seconds.
	PortEnd     int // 0 = single-port
	HopInterval int // seconds; 0 = disabled

	// GroupName is the ProtocolGroup this proxy belongs to (empty = ungrouped).
	// Used by renderers to emit per-group urltest/fallback outbounds for
	// auto-protocol switching.
	GroupName string
}
