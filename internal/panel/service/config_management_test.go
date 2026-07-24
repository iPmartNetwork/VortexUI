package service

import (
	"encoding/json"
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

// Property 24: Config validation rejects invalid JSON — invalid protocol/transport
// combinations produce at least one validation error.
func TestProperty_ConfigValidationRejectsInvalid(t *testing.T) {
	v := NewConfigValidator()

	cases := []struct {
		name     string
		protocol domain.Protocol
		network  string
		security domain.Security
		config   map[string]any
	}{
		{"negative alterId", domain.ProtoVMess, "tcp", domain.SecurityNone, map[string]any{"alterId": -1}},
		{"invalid SS method", domain.ProtoShadowsocks, "tcp", domain.SecurityNone, map[string]any{"method": "invalid-cipher"}},
		{"hysteria2 zero up", domain.ProtoHysteria2, "", domain.SecurityTLS, map[string]any{"up_mbps": 0}},
		{"invalid grpc service", domain.ProtoVLESS, "grpc", domain.SecurityTLS, map[string]any{"serviceName": "bad name!"}},
		{"naive without TLS", domain.ProtoNaive, "", domain.SecurityNone, map[string]any{}},
		{"WG MTU too low", domain.ProtoWireGuard, "", domain.SecurityNone, map[string]any{"mtu": 1000}},
	}

	for _, c := range cases {
		errs := v.Validate(c.protocol, c.network, c.security, c.config)
		if len(errs) == 0 {
			t.Fatalf("case %q: expected validation errors, got none", c.name)
		}
	}
}

// Property 25: Config defaults produce valid configs — defaults for any combo
// pass validation with zero errors.
func TestProperty_ConfigDefaultsProduceValidConfigs(t *testing.T) {
	v := NewConfigValidator()

	combos := []struct {
		protocol domain.Protocol
		network  string
		security domain.Security
	}{
		{domain.ProtoVLESS, "tcp", domain.SecurityReality},
		{domain.ProtoVMess, "ws", domain.SecurityTLS},
		{domain.ProtoTrojan, "grpc", domain.SecurityTLS},
		{domain.ProtoShadowsocks, "tcp", domain.SecurityNone},
		{domain.ProtoHysteria2, "", domain.SecurityTLS},
		{domain.ProtoWireGuard, "", domain.SecurityNone},
	}

	for _, c := range combos {
		defaults := v.DefaultsFor(c.protocol, c.network, c.security)
		errs := v.Validate(c.protocol, c.network, c.security, defaults.Config)
		if len(errs) > 0 {
			t.Fatalf("defaults for %s/%s/%s failed validation: %v", c.protocol, c.network, c.security, errs)
		}
	}
}

// Property 26: Config version diff correctness — diff between two versions
// captures all added, removed, and modified fields.
func TestProperty_ConfigVersionDiffCorrectness(t *testing.T) {
	oldCfg := map[string]any{
		"path":      "/old",
		"host":      "example.com",
		"removedKey": "value",
	}
	newCfg := map[string]any{
		"path":     "/new",
		"host":     "example.com",
		"addedKey": "newvalue",
	}

	changes := computeChanges("", oldCfg, newCfg)

	hasModified := false
	hasAdded := false
	hasRemoved := false
	for _, c := range changes {
		switch c.Type {
		case "modified":
			hasModified = true
			if c.Path != "path" {
				continue
			}
			if c.OldValue != "/old" || c.NewValue != "/new" {
				t.Fatal("modified path values incorrect")
			}
		case "added":
			hasAdded = true
		case "removed":
			hasRemoved = true
		}
	}

	if !hasModified {
		t.Fatal("diff should detect modified field")
	}
	if !hasAdded {
		t.Fatal("diff should detect added field")
	}
	if !hasRemoved {
		t.Fatal("diff should detect removed field")
	}
}

// Property 27: Config export/import round-trip — exported JSON can be parsed
// back into the same config structure.
func TestProperty_ConfigExportImportRoundTrip(t *testing.T) {
	original := map[string]any{
		"protocol": "vless",
		"network":  "ws",
		"path":     "/secret",
		"security": "tls",
		"sni":      "example.com",
		"nested": map[string]any{
			"key1": "value1",
			"key2": float64(42),
		},
	}

	// Export (marshal)
	data, err := json.Marshal(map[string]any{"config": original})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Import (unmarshal)
	var imported struct {
		Config map[string]any `json:"config"`
	}
	if err := json.Unmarshal(data, &imported); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify key fields survived.
	if imported.Config["protocol"] != "vless" {
		t.Fatal("protocol mismatch after round-trip")
	}
	if imported.Config["path"] != "/secret" {
		t.Fatal("path mismatch after round-trip")
	}
	nested, ok := imported.Config["nested"].(map[string]any)
	if !ok {
		t.Fatal("nested map lost during round-trip")
	}
	if nested["key1"] != "value1" {
		t.Fatal("nested key1 mismatch")
	}
}
