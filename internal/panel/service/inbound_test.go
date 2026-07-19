package service

import (
	"testing"

	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
)

func TestCoreSupportsUDPNativeProtocols(t *testing.T) {
	// hysteria2/tuic are QUIC/UDP-native: the frontend submits them with
	// network "udp", which is not a stream transport. The sing-box core must
	// accept them regardless of the (irrelevant) network value.
	cases := []struct {
		name     string
		core     domain.CoreType
		proto    domain.Protocol
		network  string
		security domain.Security
		wantErr  bool
	}{
		{"singbox hysteria2 udp", domain.CoreSingbox, domain.ProtoHysteria2, "udp", domain.SecurityNone, false},
		{"singbox tuic udp", domain.CoreSingbox, domain.ProtoTUIC, "udp", domain.SecurityNone, false},
		{"xray hysteria2 udp rejected", domain.CoreXray, domain.ProtoHysteria2, "udp", domain.SecurityNone, true},
		{"xray tuic udp rejected", domain.CoreXray, domain.ProtoTUIC, "udp", domain.SecurityNone, true},
		{"singbox vless udp rejected", domain.CoreSingbox, domain.ProtoVLESS, "udp", domain.SecurityNone, true},
		{"singbox vless ws ok", domain.CoreSingbox, domain.ProtoVLESS, "ws", domain.SecurityTLS, false},
		{"xray vless tcp ok", domain.CoreXray, domain.ProtoVLESS, "tcp", domain.SecurityReality, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := coreSupports(tc.core, tc.proto, tc.network, tc.security)
			if tc.wantErr && err == nil {
				t.Fatalf("coreSupports(%v, %v, %q, %q) = nil, want error", tc.core, tc.proto, tc.network, tc.security)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("coreSupports(%v, %v, %q, %q) = %v, want nil", tc.core, tc.proto, tc.network, tc.security, err)
			}
		})
	}
}

// TestCoreSupportsRejectsOffMatrixCombos asserts the guard rejects representative
// combinations that are NOT in either core's capability matrix, and accepts a few
// that are. This locks in the guard side of the save-then-reject contract: the
// matrix test proves accepted combos render; this proves rejected combos never
// reach the renderer in the first place.
func TestCoreSupportsRejectsOffMatrixCombos(t *testing.T) {
	cases := []struct {
		name     string
		core     domain.CoreType
		proto    domain.Protocol
		network  string
		security domain.Security
		wantErr  bool
	}{
		// --- Rejected: protocol not on the core ---
		{"xray tuic not a protocol", domain.CoreXray, domain.ProtoTUIC, "tcp", domain.SecurityTLS, true},
		{"xray wireguard not a protocol", domain.CoreXray, domain.ProtoWireGuard, "tcp", domain.SecurityNone, true},
		{"xray shadowtls not a protocol", domain.CoreXray, domain.ProtoShadowTLS, "tcp", domain.SecurityNone, true},
		{"xray anytls not a protocol", domain.CoreXray, domain.ProtoAnyTLS, "tcp", domain.SecurityTLS, true},
		{"xray hysteria not a protocol", domain.CoreXray, domain.ProtoHysteria, "tcp", domain.SecurityTLS, true},

		// --- Rejected: transport not on the core ---
		{"xray quic transport", domain.CoreXray, domain.ProtoVLESS, "quic", domain.SecurityReality, true},
		{"singbox xhttp transport", domain.CoreSingbox, domain.ProtoVLESS, "xhttp", domain.SecurityNone, true},
		{"xray unknown transport", domain.CoreXray, domain.ProtoVMess, "bogus-transport", domain.SecurityNone, true},

		// --- Valid: present in the matrix ---
		{"xray vless xhttp none ok", domain.CoreXray, domain.ProtoVLESS, "xhttp", domain.SecurityNone, false},
		{"xray vmess ws tls ok", domain.CoreXray, domain.ProtoVMess, "ws", domain.SecurityTLS, false},
		{"xray hysteria2 tls ok", domain.CoreXray, domain.ProtoHysteria2, "", domain.SecurityTLS, false},
		{"singbox vless quic reality ok", domain.CoreSingbox, domain.ProtoVLESS, "quic", domain.SecurityReality, false},
		{"singbox hysteria2 udp tls ok", domain.CoreSingbox, domain.ProtoHysteria2, "udp", domain.SecurityTLS, false},
		{"singbox anytls tcp tls ok", domain.CoreSingbox, domain.ProtoAnyTLS, "tcp", domain.SecurityTLS, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := coreSupports(tc.core, tc.proto, tc.network, tc.security)
			if tc.wantErr && err == nil {
				t.Fatalf("coreSupports(%v, %v, %q, %q) = nil, want error", tc.core, tc.proto, tc.network, tc.security)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("coreSupports(%v, %v, %q, %q) = %v, want nil", tc.core, tc.proto, tc.network, tc.security, err)
			}
		})
	}
}

func TestProvisionSecuritySyncsRealitySNI(t *testing.T) {
	in := &domain.Inbound{
		Security: domain.SecurityReality,
		SNI:      []string{"www.apple.com"},
		Raw: map[string]any{
			"reality": map[string]any{
				"private_key":  "testpriv",
				"server_names": []any{"www.microsoft.com"},
				"dest":         "www.microsoft.com:443",
			},
		},
	}
	provisionSecurity(in)
	p := reality.ParseParams(in.Raw["reality"])
	if len(p.ServerNames) != 1 || p.ServerNames[0] != "www.apple.com" {
		t.Fatalf("server_names = %v, want [www.apple.com]", p.ServerNames)
	}
	if p.Dest != "www.apple.com:443" {
		t.Fatalf("dest = %q, want www.apple.com:443", p.Dest)
	}
}
