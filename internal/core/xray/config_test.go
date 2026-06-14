package xray

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

func TestBuilder_RendersValidXrayConfig(t *testing.T) {
	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}

	in := domain.Inbound{
		Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "ws", Security: domain.SecurityTLS, SNI: []string{"example.com"}, Path: "/v",
	}
	cfg := &core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{"vless-ws": {u1, u2}},
	}

	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var parsed struct {
		API      apiConf `json:"api"`
		Policy   struct {
			Levels map[string]policyLevel `json:"levels"`
		} `json:"policy"`
		Inbounds []struct {
			Tag            string          `json:"tag"`
			Protocol       string          `json:"protocol"`
			Port           int             `json:"port"`
			Settings       json.RawMessage `json:"settings"`
			StreamSettings json.RawMessage `json:"streamSettings"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("generated config is not valid JSON: %v\n%s", err, raw)
	}

	// API inbound must be present and on loopback for runtime control.
	if parsed.API.Tag != APIInboundTag {
		t.Errorf("api tag = %q, want %q", parsed.API.Tag, APIInboundTag)
	}
	if !parsed.Policy.Levels["0"].StatsUserUplink {
		t.Error("per-user uplink stats must be enabled or traffic accounting breaks")
	}

	// Expect the reserved api inbound + the user inbound.
	var user *struct {
		Tag            string          `json:"tag"`
		Protocol       string          `json:"protocol"`
		Port           int             `json:"port"`
		Settings       json.RawMessage `json:"settings"`
		StreamSettings json.RawMessage `json:"streamSettings"`
	}
	for i := range parsed.Inbounds {
		if parsed.Inbounds[i].Tag == "vless-ws" {
			user = &parsed.Inbounds[i]
		}
	}
	if user == nil {
		t.Fatal("vless-ws inbound missing from output")
	}

	var settings struct {
		Decryption string           `json:"decryption"`
		Clients    []map[string]any `json:"clients"`
	}
	if err := json.Unmarshal(user.Settings, &settings); err != nil {
		t.Fatalf("settings parse: %v", err)
	}
	if settings.Decryption != "none" {
		t.Errorf("vless decryption = %q, want none", settings.Decryption)
	}
	if len(settings.Clients) != 2 {
		t.Fatalf("want 2 clients, got %d", len(settings.Clients))
	}
	// Client email must equal the user UUID so stats counters map back to users.
	if settings.Clients[0]["email"] != u1.ID.String() {
		t.Errorf("client email = %v, want %s", settings.Clients[0]["email"], u1.ID)
	}

	var ss struct {
		Network    string         `json:"network"`
		Security   string         `json:"security"`
		WSSettings map[string]any `json:"wsSettings"`
	}
	if err := json.Unmarshal(user.StreamSettings, &ss); err != nil {
		t.Fatalf("streamSettings parse: %v", err)
	}
	if ss.Network != "ws" || ss.Security != "tls" {
		t.Errorf("stream = %s/%s, want ws/tls", ss.Network, ss.Security)
	}
	if ss.WSSettings["path"] != "/v" {
		t.Errorf("ws path = %v, want /v", ss.WSSettings["path"])
	}
}

func TestBuilder_UnsupportedProtocol(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{{Tag: "wg", Protocol: domain.ProtoWireGuard, Port: 51820}},
	}
	if _, err := (Builder{APIPort: 1}).Build(cfg); err == nil {
		t.Fatal("expected error for unsupported protocol, got nil")
	}
}
