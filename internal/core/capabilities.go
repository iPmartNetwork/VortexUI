// Package core — single per-core capability matrix.
//
// This is the ONE authoritative definition of what each proxy core supports:
// which protocols, transports, and securities are valid, and which protocols
// are UDP-native (carry no stream transport, so the network value is ignored).
//
// The guard in internal/panel/service derives its validation from this matrix
// (rather than a separate hand-maintained map), and the API/UI consume it too,
// so the three layers can never disagree. The values below mirror EXACTLY what
// the renderers emit today (internal/core/xray/config.go streamSettings +
// protocolSettings and internal/core/singbox/config.go transportBlock +
// buildInbound), and reconcile the prior quic/xhttp/udp inconsistencies.
package core

import (
	"fmt"

	"github.com/vortexui/vortexui/internal/domain"
)

// ProtocolConstraint narrows the core-wide capabilities for a single protocol.
// It captures the cases where a protocol is more restricted than its core: a
// utility proxy that carries no stream transport, or a protocol that only
// accepts a subset of the core's securities.
type ProtocolConstraint struct {
	// NoTransport marks a protocol that carries no stream transport (a TCP
	// utility proxy like socks/http, or a TLS-fronted proxy like naive); the
	// network value is irrelevant for it, so the transport check is skipped.
	NoTransport bool
	// Securities, when non-empty, is the ONLY set of securities allowed for this
	// protocol and overrides the core-wide Securities list.
	Securities []domain.Security
}

// Capability describes everything a single core supports.
type Capability struct {
	// Protocols are the inbound proxy protocols the core can run.
	Protocols []domain.Protocol
	// Transports are the stream transports (network values) the core can render.
	Transports []string
	// Securities are the TLS-layer modes the core supports.
	Securities []domain.Security
	// UDPNative lists protocols that have no stream transport (the network value
	// is irrelevant for them, so the transport check is skipped).
	UDPNative []domain.Protocol
	// Constraints narrows capabilities per protocol (no-transport utility proxies
	// and per-protocol security overrides). Protocols absent from this map use
	// the core-wide transports/securities unchanged.
	Constraints map[domain.Protocol]ProtocolConstraint
}

// capabilities is the per-core matrix. Keep these values in lockstep with the
// renderers — every entry here must be renderable by its core, and every branch
// the renderer supports must appear here (no dead branches, no save-then-reject).
var capabilities = map[domain.CoreType]Capability{
	domain.CoreXray: {
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks, domain.ProtoSocks, domain.ProtoHTTP, domain.ProtoDokodemo},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "http", "h2", "xhttp", "kcp"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
		UDPNative:  nil,
		Constraints: map[domain.Protocol]ProtocolConstraint{
			domain.ProtoSocks:    {NoTransport: true, Securities: []domain.Security{domain.SecurityNone}},
			domain.ProtoHTTP:     {NoTransport: true, Securities: []domain.Security{domain.SecurityNone}},
			domain.ProtoDokodemo: {NoTransport: true, Securities: []domain.Security{domain.SecurityNone}},
		},
	},
	domain.CoreSingbox: {
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks, domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard, domain.ProtoHysteria, domain.ProtoShadowTLS, domain.ProtoAnyTLS, domain.ProtoSocks, domain.ProtoHTTP, domain.ProtoNaive},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "http", "h2", "quic"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
		UDPNative:  []domain.Protocol{domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard, domain.ProtoHysteria},
		Constraints: map[domain.Protocol]ProtocolConstraint{
			domain.ProtoSocks: {NoTransport: true, Securities: []domain.Security{domain.SecurityNone}},
			domain.ProtoHTTP:  {NoTransport: true, Securities: []domain.Security{domain.SecurityNone}},
			domain.ProtoNaive: {NoTransport: true, Securities: []domain.Security{domain.SecurityTLS}},
		},
	},
}

// Capabilities returns the capability matrix for a core. Unknown or empty core
// types default to xray, matching the guard's historical default behavior.
func Capabilities(c domain.CoreType) Capability {
	if caps, ok := capabilities[c]; ok {
		return caps
	}
	return capabilities[domain.CoreXray]
}

// IsUDPNative reports whether the given protocol is UDP-native on the core (no
// stream transport, so the network value is meaningless for it).
func IsUDPNative(c domain.CoreType, proto domain.Protocol) bool {
	for _, p := range Capabilities(c).UDPNative {
		if p == proto {
			return true
		}
	}
	return false
}

// SkipsTransport reports whether the given protocol carries no stream transport
// on the core, so the network value is irrelevant and the transport check is
// skipped. This is true for UDP-native protocols (hysteria/tuic/...) and for any
// protocol whose constraint sets NoTransport (socks/http/naive utility proxies).
func SkipsTransport(c domain.CoreType, proto domain.Protocol) bool {
	return IsUDPNative(c, proto) || Capabilities(c).Constraints[proto].NoTransport
}

// AllowedSecurities returns the securities permitted for the given protocol on
// the core. When a per-protocol constraint defines a non-empty Securities
// override, that narrower set is returned; otherwise the core-wide list applies.
func AllowedSecurities(c domain.CoreType, proto domain.Protocol) []domain.Security {
	if con, ok := Capabilities(c).Constraints[proto]; ok && len(con.Securities) > 0 {
		return con.Securities
	}
	return Capabilities(c).Securities
}

// Supports reports whether a core can run the given protocol+transport+security
// combination, returning a clear error describing the first incompatibility (or
// nil if every part is allowed).
//
// An empty network is treated as "tcp". For UDP-native protocols the transport
// check is skipped because the network value is irrelevant. The security value
// is always validated against the core's supported list.
func Supports(c domain.CoreType, proto domain.Protocol, network string, security domain.Security) error {
	caps := Capabilities(c)
	if network == "" {
		network = "tcp"
	}
	if security == "" {
		security = domain.SecurityNone
	}

	if !containsProtocol(caps.Protocols, proto) {
		return fmt.Errorf("protocol %q is not supported on the %s core", proto, c)
	}
	// Protocols that carry no stream transport (UDP-native or NoTransport utility
	// proxies) ignore the network value, so skip the transport check for them.
	if !SkipsTransport(c, proto) && !containsString(caps.Transports, network) {
		return fmt.Errorf("transport %q is not supported on the %s core", network, c)
	}
	// Validate against the securities allowed for THIS protocol, which may be a
	// narrower per-protocol override (e.g. socks/http -> none, naive -> tls).
	if !containsSecurity(AllowedSecurities(c, proto), security) {
		return fmt.Errorf("security %q is not supported on the %s core", security, c)
	}
	return nil
}

func containsProtocol(list []domain.Protocol, v domain.Protocol) bool {
	for _, p := range list {
		if p == v {
			return true
		}
	}
	return false
}

func containsSecurity(list []domain.Security, v domain.Security) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}
