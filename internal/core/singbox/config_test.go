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
	// Unsupported protocols are now skipped (not fatal) so one bad inbound
	// doesn't crash the core.
	raw, err := (Builder{APIPort: 1}).Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected non-empty config")
	}
}

type routedConfig struct {
	Outbounds []struct {
		Type      string   `json:"type"`
		Tag       string   `json:"tag"`
		Outbounds []string `json:"outbounds"`
		URL       string   `json:"url"`
		Interval  string   `json:"interval"`
		Default   string   `json:"default"`
	} `json:"outbounds"`
	Route struct {
		Final string `json:"final"`
		Rules []struct {
			Inbound   []string `json:"inbound"`
			Domain    []string `json:"domain"`
			IPCIDR    []string `json:"ip_cidr"`
			Port      []int    `json:"port"`
			PortRange []string `json:"port_range"`
			Network   string   `json:"network"`
			Outbound  string   `json:"outbound"`
		} `json:"rules"`
	} `json:"route"`
}

func TestBuilderDefaultOutboundsWhenNoneConfigured(t *testing.T) {
	raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var c routedConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse: %v", err)
	}
	tags := map[string]string{}
	for _, o := range c.Outbounds {
		tags[o.Tag] = o.Type
	}
	if tags["direct"] != "direct" || tags["block"] != "block" {
		t.Errorf("default outbounds missing: %v", tags)
	}
	// No routing rules -> no rules, but final must still be a safe default.
	if len(c.Route.Rules) != 0 {
		t.Errorf("unexpected route rules: %+v", c.Route.Rules)
	}
	if c.Route.Final != "direct" {
		t.Errorf("route final = %q, want direct (deterministic default egress)", c.Route.Final)
	}
}

func TestBuilderOutboundsRoutingAndBalancerGroup(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "proxy-a", Protocol: domain.OutTrojan, Address: "a.example.com", Port: 443, Password: "pa", Security: domain.SecurityTLS, SNI: "a.example.com", Enabled: true},
			{Tag: "proxy-b", Protocol: domain.OutShadowsocks, Address: "b.example.com", Port: 8388, Password: "pb", Method: "2022-blake3-aes-128-gcm", Enabled: true},
		},
		Balancers: []domain.Balancer{
			{Tag: "auto", Selectors: []string{"proxy-"}, Strategy: domain.BalancerLeastPing, Enabled: true},
			{Tag: "pick", Selectors: []string{"proxy-a"}, Strategy: domain.BalancerRandom, Enabled: true},
		},
		Routing: []domain.RoutingRule{
			{Priority: 2, InboundTags: []string{"in"}, Port: "443,8388", BalancerTag: "auto", Enabled: true},
			{Priority: 1, IP: []string{"10.0.0.0/8"}, Network: "udp", OutboundTag: "direct", Enabled: true},
		},
	}
	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var c routedConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse: %v", err)
	}

	groups := map[string]struct {
		typ      string
		members  []string
		def      string
		url      string
		interval string
	}{}
	var sawTrojan, sawSS bool
	for _, o := range c.Outbounds {
		switch o.Tag {
		case "proxy-a":
			sawTrojan = o.Type == "trojan"
		case "proxy-b":
			sawSS = o.Type == "shadowsocks"
		case "auto", "pick":
			groups[o.Tag] = struct {
				typ      string
				members  []string
				def      string
				url      string
				interval string
			}{o.Type, o.Outbounds, o.Default, o.URL, o.Interval}
		}
	}
	if !sawTrojan || !sawSS {
		t.Errorf("proxy outbounds not rendered correctly: trojan=%v ss=%v", sawTrojan, sawSS)
	}
	// leastPing -> urltest group with both members and probe url/interval.
	auto := groups["auto"]
	if auto.typ != "urltest" {
		t.Errorf("auto group type = %q, want urltest", auto.typ)
	}
	if len(auto.members) != 2 {
		t.Errorf("auto members = %v, want [proxy-a proxy-b]", auto.members)
	}
	if auto.url == "" || auto.interval == "" {
		t.Error("urltest group needs url+interval")
	}
	// random -> selector group, single member, default set.
	pick := groups["pick"]
	if pick.typ != "selector" || len(pick.members) != 1 || pick.def != "proxy-a" {
		t.Errorf("pick group wrong: %+v", pick)
	}

	// Route: rules ordered by priority (udp/ip first), port spec parsed.
	if len(c.Route.Rules) != 2 {
		t.Fatalf("want 2 route rules, got %d", len(c.Route.Rules))
	}
	if c.Route.Rules[0].Network != "udp" || c.Route.Rules[0].Outbound != "direct" {
		t.Errorf("rule[0] wrong: %+v", c.Route.Rules[0])
	}
	if c.Route.Rules[1].Outbound != "auto" {
		t.Errorf("rule[1] should target balancer group 'auto', got %q", c.Route.Rules[1].Outbound)
	}
	if len(c.Route.Rules[1].Port) != 2 {
		t.Errorf("rule[1] ports = %v, want [443 8388]", c.Route.Rules[1].Port)
	}
	if c.Route.Final != "direct" {
		t.Errorf("route final = %q, want direct", c.Route.Final)
	}
}

func TestBuilderPortRangeParsing(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Routing: []domain.RoutingRule{
			{Priority: 1, Port: "1000-2000", InboundTags: []string{"in"}, OutboundTag: "direct", Enabled: true},
		},
	}
	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var c routedConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(c.Route.Rules) != 1 || len(c.Route.Rules[0].PortRange) != 1 || c.Route.Rules[0].PortRange[0] != "1000:2000" {
		t.Errorf("port_range not parsed: %+v", c.Route.Rules)
	}
}

func TestBuilderBalancerWithNoMembersErrors(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Balancers: []domain.Balancer{
			{Tag: "lb", Selectors: []string{"nonexistent-"}, Strategy: domain.BalancerRandom, Enabled: true},
		},
	}
	if _, err := (Builder{APIPort: 9090}).Build(cfg); err == nil {
		t.Fatal("expected error for balancer whose selectors match nothing")
	}
}

func TestBuilderRealityBlock(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-reality", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.apple.com"},
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK",
			"short_ids":   []any{"abcd"},
			"dest":        "www.apple.com:443",
		}},
	}
	raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{"vless-reality": {{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}}},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var p struct {
		Inbounds []struct {
			Tag string `json:"tag"`
			TLS struct {
				Reality struct {
					Enabled   bool     `json:"enabled"`
					PriKey    string   `json:"private_key"`
					ShortID   []string `json:"short_id"`
					Handshake struct {
						Server     string `json:"server"`
						ServerPort int    `json:"server_port"`
					} `json:"handshake"`
				} `json:"reality"`
			} `json:"tls"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("parse: %v", err)
	}
	r := p.Inbounds[0].TLS.Reality
	if !r.Enabled || r.PriKey != "PK" {
		t.Errorf("reality block wrong: %+v", r)
	}
	if len(r.ShortID) != 1 || r.ShortID[0] != "abcd" {
		t.Errorf("short_id = %v", r.ShortID)
	}
	if r.Handshake.Server != "www.apple.com" || r.Handshake.ServerPort != 443 {
		t.Errorf("handshake = %+v", r.Handshake)
	}
}
