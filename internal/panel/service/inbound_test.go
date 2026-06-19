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
		name    string
		core    domain.CoreType
		proto   domain.Protocol
		network string
		wantErr bool
	}{
		{"singbox hysteria2 udp", domain.CoreSingbox, domain.ProtoHysteria2, "udp", false},
		{"singbox tuic udp", domain.CoreSingbox, domain.ProtoTUIC, "udp", false},
		{"xray hysteria2 udp rejected", domain.CoreXray, domain.ProtoHysteria2, "udp", true},
		{"xray tuic udp rejected", domain.CoreXray, domain.ProtoTUIC, "udp", true},
		{"singbox vless udp rejected", domain.CoreSingbox, domain.ProtoVLESS, "udp", true},
		{"singbox vless ws ok", domain.CoreSingbox, domain.ProtoVLESS, "ws", false},
		{"xray vless tcp ok", domain.CoreXray, domain.ProtoVLESS, "tcp", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := coreSupports(tc.core, tc.proto, tc.network)
			if tc.wantErr && err == nil {
				t.Fatalf("coreSupports(%v, %v, %q) = nil, want error", tc.core, tc.proto, tc.network)
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("coreSupports(%v, %v, %q) = %v, want nil", tc.core, tc.proto, tc.network, err)
			}
		})
	}
}
