package core

import (
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

// TestSupports exercises the single per-core capability matrix: representative
// allowed combos must pass and representative rejected combos must fail with a
// clear error. These mirror exactly what the renderers emit today and lock in
// the quic/xhttp/udp reconciliation.
func TestSupports(t *testing.T) {
	cases := []struct {
		name     string
		core     domain.CoreType
		proto    domain.Protocol
		network  string
		security domain.Security
		wantErr  bool
	}{
		// Allowed combos.
		{"xray vless ws tls", domain.CoreXray, domain.ProtoVLESS, "ws", domain.SecurityTLS, false},
		{"xray vless tcp reality", domain.CoreXray, domain.ProtoVLESS, "tcp", domain.SecurityReality, false},
		{"xray vless xhttp none", domain.CoreXray, domain.ProtoVLESS, "xhttp", domain.SecurityNone, false},
		{"singbox hysteria2 udp-native any network", domain.CoreSingbox, domain.ProtoHysteria2, "udp", domain.SecurityTLS, false},
		{"singbox tuic udp-native empty network", domain.CoreSingbox, domain.ProtoTUIC, "", domain.SecurityTLS, false},
		{"singbox wireguard udp-native", domain.CoreSingbox, domain.ProtoWireGuard, "udp", domain.SecurityNone, false},
		{"singbox vless quic reality", domain.CoreSingbox, domain.ProtoVLESS, "quic", domain.SecurityReality, false},
		{"empty network defaults to tcp", domain.CoreXray, domain.ProtoTrojan, "", domain.SecurityTLS, false},
		{"unknown core defaults to xray", domain.CoreType("bogus"), domain.ProtoVLESS, "tcp", domain.SecurityTLS, false},

		// Rejected combos.
		{"xray hysteria2 unsupported protocol", domain.CoreXray, domain.ProtoHysteria2, "udp", domain.SecurityNone, true},
		{"xray vless quic unsupported transport", domain.CoreXray, domain.ProtoVLESS, "quic", domain.SecurityTLS, true},
		{"singbox vless xhttp unsupported transport", domain.CoreSingbox, domain.ProtoVLESS, "xhttp", domain.SecurityTLS, true},
		{"xray unknown security", domain.CoreXray, domain.ProtoVLESS, "tcp", domain.Security("bogus"), true},

		// Per-protocol constraints: socks/http carry no transport and allow only
		// security none; naive carries no transport and mandates tls.
		{"xray socks no-transport none", domain.CoreXray, domain.ProtoSocks, "", domain.SecurityNone, false},
		{"xray http no-transport none", domain.CoreXray, domain.ProtoHTTP, "", domain.SecurityNone, false},
		{"xray socks ignores network", domain.CoreXray, domain.ProtoSocks, "quic", domain.SecurityNone, false},
		{"xray socks rejects tls override", domain.CoreXray, domain.ProtoSocks, "", domain.SecurityTLS, true},
		{"singbox socks no-transport none", domain.CoreSingbox, domain.ProtoSocks, "", domain.SecurityNone, false},
		{"singbox http no-transport none", domain.CoreSingbox, domain.ProtoHTTP, "", domain.SecurityNone, false},
		{"singbox naive mandates tls", domain.CoreSingbox, domain.ProtoNaive, "", domain.SecurityTLS, false},
		{"singbox naive rejects none", domain.CoreSingbox, domain.ProtoNaive, "", domain.SecurityNone, true},
		{"xray naive unsupported protocol", domain.CoreXray, domain.ProtoNaive, "", domain.SecurityTLS, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Supports(tc.core, tc.proto, tc.network, tc.security)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Supports(%v, %v, %q, %q) = nil, want error", tc.core, tc.proto, tc.network, tc.security)
				}
				return
			}
			if err != nil {
				t.Fatalf("Supports(%v, %v, %q, %q) = %v, want nil", tc.core, tc.proto, tc.network, tc.security, err)
			}
		})
	}
}

// TestCapabilitiesDefaultsToXray confirms unknown/empty cores fall back to the
// xray matrix, matching the guard's historical default.
func TestCapabilitiesDefaultsToXray(t *testing.T) {
	got := Capabilities(domain.CoreType("nope"))
	want := Capabilities(domain.CoreXray)
	if len(got.Protocols) != len(want.Protocols) || len(got.Transports) != len(want.Transports) {
		t.Fatalf("unknown core did not default to xray: got %+v", got)
	}
}

// TestIsUDPNative verifies the UDP-native classification per core.
func TestIsUDPNative(t *testing.T) {
	if !IsUDPNative(domain.CoreSingbox, domain.ProtoHysteria2) {
		t.Fatal("singbox hysteria2 should be UDP-native")
	}
	if IsUDPNative(domain.CoreSingbox, domain.ProtoVLESS) {
		t.Fatal("singbox vless should not be UDP-native")
	}
	if IsUDPNative(domain.CoreXray, domain.ProtoHysteria2) {
		t.Fatal("xray has no UDP-native protocols")
	}
}

// TestSkipsTransport verifies that both UDP-native protocols and NoTransport
// utility proxies (socks/http/naive) skip the stream-transport check, while
// regular protocols do not.
func TestSkipsTransport(t *testing.T) {
	cases := []struct {
		core  domain.CoreType
		proto domain.Protocol
		want  bool
	}{
		{domain.CoreSingbox, domain.ProtoHysteria2, true}, // UDP-native
		{domain.CoreSingbox, domain.ProtoSocks, true},     // NoTransport
		{domain.CoreSingbox, domain.ProtoHTTP, true},      // NoTransport
		{domain.CoreSingbox, domain.ProtoNaive, true},     // NoTransport
		{domain.CoreSingbox, domain.ProtoVLESS, false},
		{domain.CoreXray, domain.ProtoSocks, true},
		{domain.CoreXray, domain.ProtoHTTP, true},
		{domain.CoreXray, domain.ProtoVLESS, false},
	}
	for _, tc := range cases {
		if got := SkipsTransport(tc.core, tc.proto); got != tc.want {
			t.Errorf("SkipsTransport(%v, %v) = %v, want %v", tc.core, tc.proto, got, tc.want)
		}
	}
}

// TestAllowedSecurities verifies the per-protocol security override falls back
// to the core-wide list when no constraint applies, and returns the narrowed set
// when one does.
func TestAllowedSecurities(t *testing.T) {
	// Constrained protocols return their override only.
	if got := AllowedSecurities(domain.CoreSingbox, domain.ProtoSocks); len(got) != 1 || got[0] != domain.SecurityNone {
		t.Errorf("socks AllowedSecurities = %v, want [none]", got)
	}
	if got := AllowedSecurities(domain.CoreSingbox, domain.ProtoNaive); len(got) != 1 || got[0] != domain.SecurityTLS {
		t.Errorf("naive AllowedSecurities = %v, want [tls]", got)
	}
	// Unconstrained protocols fall back to the core-wide list.
	if got := AllowedSecurities(domain.CoreSingbox, domain.ProtoVLESS); len(got) != len(Capabilities(domain.CoreSingbox).Securities) {
		t.Errorf("vless AllowedSecurities = %v, want core-wide list", got)
	}
}
