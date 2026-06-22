package domain

import "testing"

// TestOutbound_ValidateWireguard verifies the wireguard validation branch: the
// config lives in Raw["wireguard"] and must carry a non-empty private_key plus
// at least one address. Missing/empty material is rejected; a complete block
// passes.
func TestOutbound_ValidateWireguard(t *testing.T) {
	// No Raw at all -> error.
	o := Outbound{Tag: "warp", Protocol: OutWireguard}
	if err := o.Validate(); err == nil {
		t.Error("wireguard outbound without raw.wireguard must fail validation")
	}

	// Raw present but missing private_key -> error.
	o = Outbound{Tag: "warp", Protocol: OutWireguard, Raw: map[string]any{
		"wireguard": map[string]any{"address": []any{"172.16.0.2/32"}},
	}}
	if err := o.Validate(); err == nil {
		t.Error("wireguard outbound without private_key must fail validation")
	}

	// Raw present but no address -> error.
	o = Outbound{Tag: "warp", Protocol: OutWireguard, Raw: map[string]any{
		"wireguard": map[string]any{"private_key": "x"},
	}}
	if err := o.Validate(); err == nil {
		t.Error("wireguard outbound without address must fail validation")
	}

	// Complete config -> ok.
	o = Outbound{Tag: "warp", Protocol: OutWireguard, Raw: map[string]any{
		"wireguard": map[string]any{
			"private_key": "x",
			"address":     []any{"172.16.0.2/32"},
		},
	}}
	if err := o.Validate(); err != nil {
		t.Errorf("valid wireguard outbound must pass validation, got %v", err)
	}
}

// TestOutboundProtocol_WireguardMetadata verifies the protocol is recognised as
// valid and that it does NOT require the address/port endpoint (its endpoint
// lives in the wireguard config block).
func TestOutboundProtocol_WireguardMetadata(t *testing.T) {
	if !OutWireguard.Valid() {
		t.Error("OutWireguard must be a valid outbound protocol")
	}
	if OutWireguard.NeedsEndpoint() {
		t.Error("OutWireguard must not require an address/port endpoint")
	}
}
