package service

import (
	"testing"

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
