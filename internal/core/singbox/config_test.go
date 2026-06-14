package singbox

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

type parsedConfig struct {
	Inbounds []struct {
		Type       string           `json:"type"`
		Tag        string           `json:"tag"`
		ListenPort int              `json:"listen_port"`
		Users      []map[string]any `json:"users"`
		TLS        map[string]any   `json:"tls"`
		Transport  map[string]any   `json:"transport"`
	} `json:"inbounds"`
	Experimental struct {
		V2RayAPI struct {
			Listen string `json:"listen"`
			Stats  struct {
				Enabled bool     `json:"enabled"`
				Users   []string `json:"users"`
			} `json:"stats"`
		} `json:"v2ray_api"`
	} `json:"experimental"`
}

func TestBuilderRendersValidSingboxConfig(t *testing.T) {
	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	in := domain.Inbound{Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws", Security: domain.SecurityTLS, SNI: []string{"ex.com"}, Path: "/v", Flow: "xtls-rprx-vision"}
	cfg := &core.GeneratedConfig{Inbounds: []domain.Inbound{in}, UsersByInbound: map[string][]*domain.User{"vless-ws": {u1, u2}}}

	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var p parsedConfig
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("generated config invalid JSON: %v\n%s", err, raw)
	}

	if len(p.Inbounds) != 1 || p.Inbounds[0].Type != "vless" || p.Inbounds[0].ListenPort != 443 {
		t.Fatalf("inbound wrong: %+v", p.Inbounds)
	}
	if len(p.Inbounds[0].Users) != 2 {
		t.Errorf("want 2 users, got %d", len(p.Inbounds[0].Users))
	}
	if p.Inbounds[0].Users[0]["name"] != u1.ID.String() {
		t.Errorf("user name (stats key) should be the UUID, got %v", p.Inbounds[0].Users[0]["name"])
	}
	if p.Inbounds[0].TLS["enabled"] != true || p.Inbounds[0].Transport["type"] != "ws" {
		t.Errorf("tls/transport wrong: tls=%v tr=%v", p.Inbounds[0].TLS, p.Inbounds[0].Transport)
	}
	// V2Ray API must enable stats and list both users, or per-user accounting fails.
	if !p.Experimental.V2RayAPI.Stats.Enabled || len(p.Experimental.V2RayAPI.Stats.Users) != 2 {
		t.Errorf("v2ray_api stats wrong: %+v", p.Experimental.V2RayAPI.Stats)
	}
}

func TestBuilderUnsupportedProtocol(t *testing.T) {
	cfg := &core.GeneratedConfig{Inbounds: []domain.Inbound{{Tag: "wg", Protocol: domain.ProtoWireGuard, Port: 1}}}
	if _, err := (Builder{APIPort: 1}).Build(cfg); err == nil {
		t.Fatal("expected error for unsupported protocol")
	}
}
