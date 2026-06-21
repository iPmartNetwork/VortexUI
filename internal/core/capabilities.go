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
}

// capabilities is the per-core matrix. Keep these values in lockstep with the
// renderers — every entry here must be renderable by its core, and every branch
// the renderer supports must appear here (no dead branches, no save-then-reject).
var capabilities = map[domain.CoreType]Capability{
	domain.CoreXray: {
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "http", "h2", "xhttp"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
		UDPNative:  nil,
	},
	domain.CoreSingbox: {
		Protocols:  []domain.Protocol{domain.ProtoVMess, domain.ProtoVLESS, domain.ProtoTrojan, domain.ProtoShadowsocks, domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard},
		Transports: []string{"tcp", "ws", "grpc", "httpupgrade", "http", "h2", "quic"},
		Securities: []domain.Security{domain.SecurityNone, domain.SecurityTLS, domain.SecurityReality},
		UDPNative:  []domain.Protocol{domain.ProtoHysteria2, domain.ProtoTUIC, domain.ProtoWireGuard},
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
	// UDP-native protocols carry no stream transport, so skip the transport check.
	if !IsUDPNative(c, proto) && !containsString(caps.Transports, network) {
		return fmt.Errorf("transport %q is not supported on the %s core", network, c)
	}
	if !containsSecurity(caps.Securities, security) {
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
