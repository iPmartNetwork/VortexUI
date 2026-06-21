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
			Action    string   `json:"action"`
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
	// sing-box >= 1.12 removed the legacy `block`/`dns` outbounds; only `direct`
	// is synthesized now. A blackhole/dns target becomes a rule action instead.
	if tags["direct"] != "direct" {
		t.Errorf("default direct outbound missing: %v", tags)
	}
	if _, ok := tags["block"]; ok {
		t.Errorf("legacy block outbound must not be emitted: %v", tags)
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

func TestBuilderRejectAndHijackDNSActions(t *testing.T) {
	// A blackhole outbound and a dns outbound must NOT be emitted as outbounds;
	// rules targeting them (or the well-known "block" tag) become modern rule
	// actions: reject / hijack-dns.
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "drop", Protocol: domain.OutBlackhole, Enabled: true},
			{Tag: "dns-out", Protocol: domain.OutDNS, Enabled: true},
		},
		Routing: []domain.RoutingRule{
			{Priority: 1, Domains: []string{"ads.example.com"}, OutboundTag: "drop", Enabled: true},
			{Priority: 2, Port: "53", OutboundTag: "dns-out", Enabled: true},
			{Priority: 3, IP: []string{"1.2.3.4/32"}, OutboundTag: "block", Enabled: true},
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
	for _, o := range c.Outbounds {
		if o.Tag == "drop" || o.Tag == "dns-out" {
			t.Errorf("blackhole/dns must not be emitted as an outbound, got %q (%s)", o.Tag, o.Type)
		}
		if o.Type == "block" || o.Type == "dns" {
			t.Errorf("legacy outbound type %q must not be emitted", o.Type)
		}
	}
	if len(c.Route.Rules) != 3 {
		t.Fatalf("want 3 route rules, got %d", len(c.Route.Rules))
	}
	if c.Route.Rules[0].Action != "reject" || c.Route.Rules[0].Outbound != "" {
		t.Errorf("rule[0] should be reject action with no outbound: %+v", c.Route.Rules[0])
	}
	if c.Route.Rules[1].Action != "hijack-dns" || c.Route.Rules[1].Outbound != "" {
		t.Errorf("rule[1] should be hijack-dns action with no outbound: %+v", c.Route.Rules[1])
	}
	if c.Route.Rules[2].Action != "reject" || c.Route.Rules[2].Outbound != "" {
		t.Errorf("rule[2] (block target) should be reject action with no outbound: %+v", c.Route.Rules[2])
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

func TestBuilderWireGuardEndpoint(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()
	inID := uuid.New()
	in := domain.Inbound{
		ID: inID, Tag: "wg0", Protocol: domain.ProtoWireGuard, Port: 51820,
		Raw: map[string]any{"wireguard": map[string]any{
			"private_key": "SRV_PRIV",
			"public_key":  "SRV_PUB",
			"listen_port": 51820,
			"subnet":      "10.7.0.0/24",
		}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		WireGuardPeers: map[string][]domain.WireGuardPeer{
			"wg0": {
				{InboundID: inID, UserID: u1, PublicKey: "PEER1_PUB", Address: "10.7.0.2"},
				{InboundID: inID, UserID: u2, PublicKey: "PEER2_PUB", Address: "10.7.0.3"},
			},
		},
	}
	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var p struct {
		Inbounds  []map[string]any `json:"inbounds"`
		Endpoints []struct {
			Type       string   `json:"type"`
			Tag        string   `json:"tag"`
			Address    []string `json:"address"`
			PrivateKey string   `json:"private_key"`
			ListenPort int      `json:"listen_port"`
			Peers      []struct {
				PublicKey  string   `json:"public_key"`
				AllowedIPs []string `json:"allowed_ips"`
			} `json:"peers"`
		} `json:"endpoints"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("generated config invalid JSON: %v\n%s", err, raw)
	}
	// WireGuard must NOT appear as a regular inbound.
	if len(p.Inbounds) != 0 {
		t.Errorf("wireguard must render as endpoint, not inbound: %+v", p.Inbounds)
	}
	if len(p.Endpoints) != 1 {
		t.Fatalf("want 1 endpoint, got %d", len(p.Endpoints))
	}
	ep := p.Endpoints[0]
	if ep.Type != "wireguard" || ep.Tag != "wg0" || ep.PrivateKey != "SRV_PRIV" || ep.ListenPort != 51820 {
		t.Errorf("endpoint header wrong: %+v", ep)
	}
	if len(ep.Address) != 1 || ep.Address[0] != "10.7.0.1/24" {
		t.Errorf("server address = %v, want [10.7.0.1/24]", ep.Address)
	}
	if len(ep.Peers) != 2 {
		t.Fatalf("want 2 peers, got %d", len(ep.Peers))
	}
	if ep.Peers[0].PublicKey != "PEER1_PUB" || len(ep.Peers[0].AllowedIPs) != 1 || ep.Peers[0].AllowedIPs[0] != "10.7.0.2/32" {
		t.Errorf("peer[0] wrong: %+v", ep.Peers[0])
	}
}

// TestBuilderGeoBlockAllowedCountries verifies GeoBlockRules for an
// AllowedCountries policy render into sing-box route rules such that (a) the
// allowed-country traffic is routed to a real egress (not a reject action) and
// (b) the rest of the inbound's traffic is rejected. sing-box maps the "blocked"
// target to a reject action; the allow rule must keep a resolvable outbound or
// buildRoute would drop it and reject everything.
func TestBuilderGeoBlockAllowedCountries(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-geo", Protocol: domain.ProtoVLESS, Port: 443,
		GeoPolicy: &domain.GeoPolicy{AllowedCountries: []string{"IR"}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		Routing:  core.GeoBlockRules(in),
	}
	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var c routedConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse: %v", err)
	}

	var allow, blockRest *struct {
		Inbound   []string `json:"inbound"`
		Domain    []string `json:"domain"`
		IPCIDR    []string `json:"ip_cidr"`
		Port      []int    `json:"port"`
		PortRange []string `json:"port_range"`
		Network   string   `json:"network"`
		Outbound  string   `json:"outbound"`
		Action    string   `json:"action"`
	}
	for i := range c.Route.Rules {
		rule := &c.Route.Rules[i]
		switch {
		case len(rule.IPCIDR) == 1 && rule.IPCIDR[0] == "geoip:IR":
			allow = rule
		case len(rule.IPCIDR) == 0 && rule.Action == "reject":
			blockRest = rule
		}
	}

	// (b) Non-IR traffic on the inbound must be rejected.
	if blockRest == nil {
		t.Fatal("missing catch-all reject rule for non-allowed traffic")
	}
	// (a) IR traffic must route to a real egress, not be rejected/dropped.
	if allow == nil {
		t.Fatal("allow rule for geoip:IR was dropped — allowed traffic would be rejected")
	}
	if allow.Action == "reject" {
		t.Error("allow rule must not be a reject action")
	}
	if allow.Outbound == "" {
		t.Error("allow rule has no outbound — sing-box would drop it and reject allowed traffic")
	}
}

// TestBuilderGeoBlockBlockedCountries verifies a BlockedCountries policy renders
// a reject rule matching the blocked country's source IPs.
func TestBuilderGeoBlockBlockedCountries(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-geo", Protocol: domain.ProtoVLESS, Port: 443,
		GeoPolicy: &domain.GeoPolicy{BlockedCountries: []string{"CN"}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		Routing:  core.GeoBlockRules(in),
	}
	raw, err := Builder{APIPort: 9090}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var c routedConfig
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var found bool
	for _, rule := range c.Route.Rules {
		if len(rule.IPCIDR) == 1 && rule.IPCIDR[0] == "geoip:CN" && rule.Action == "reject" {
			found = true
		}
	}
	if !found {
		t.Error("expected a reject rule matching geoip:CN")
	}
}

// TestBuilderVLESSFlowOnlyForValidCombos verifies xtls-rprx-vision is emitted
// only for VLESS over raw TCP with TLS or REALITY. On ws (or any non-TCP
// transport) or with security=none the flow key must be omitted entirely,
// otherwise sing-box rejects the config.
func TestBuilderVLESSFlowOnlyForValidCombos(t *testing.T) {
	flowOf := func(t *testing.T, in domain.Inbound) (string, bool) {
		t.Helper()
		u := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
		raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{
			Inbounds:       []domain.Inbound{in},
			UsersByInbound: map[string][]*domain.User{in.Tag: {u}},
		})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		var p parsedConfig
		if err := json.Unmarshal(raw, &p); err != nil {
			t.Fatalf("parse: %v", err)
		}
		for _, i := range p.Inbounds {
			if i.Tag != in.Tag {
				continue
			}
			if len(i.Users) == 0 {
				t.Fatalf("inbound %q has no users", in.Tag)
			}
			flow, present := i.Users[0]["flow"]
			if !present {
				return "", false
			}
			str, _ := flow.(string)
			return str, true
		}
		t.Fatalf("inbound %q missing from output", in.Tag)
		return "", false
	}

	// ws transport: flow must be omitted even though Flow is set.
	if flow, present := flowOf(t, domain.Inbound{
		Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws",
		Security: domain.SecurityTLS, SNI: []string{"ex.com"}, Flow: "xtls-rprx-vision",
	}); present {
		t.Errorf("ws inbound emitted flow %q, want it omitted", flow)
	}

	// security=none on tcp: flow must be omitted.
	if flow, present := flowOf(t, domain.Inbound{
		Tag: "vless-tcp-none", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityNone, Flow: "xtls-rprx-vision",
	}); present {
		t.Errorf("tcp+none inbound emitted flow %q, want it omitted", flow)
	}

	// vless + tcp + reality WITH flow: must emit xtls-rprx-vision.
	flow, present := flowOf(t, domain.Inbound{
		Tag: "vless-reality", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.apple.com"},
		Flow: "xtls-rprx-vision",
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK", "dest": "www.apple.com:443",
		}},
	})
	if !present || flow != "xtls-rprx-vision" {
		t.Errorf("vless+tcp+reality flow = %q (present=%v), want xtls-rprx-vision", flow, present)
	}
}

// TestBuilderSkipsRealityInboundWithoutPrivateKey verifies that a REALITY
// inbound with no private key is dropped from the output entirely (sing-box
// would reject reality.enabled:true with an empty private_key), while a sibling
// inbound that has a private key still renders.
func TestBuilderSkipsRealityInboundWithoutPrivateKey(t *testing.T) {
	u := func() *domain.User {
		return &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	}
	bad := domain.Inbound{
		Tag: "reality-nokey", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.apple.com"},
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "", "dest": "www.apple.com:443",
		}},
	}
	good := domain.Inbound{
		Tag: "reality-ok", Protocol: domain.ProtoVLESS, Port: 8443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.apple.com"},
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK", "dest": "www.apple.com:443",
		}},
	}
	raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{
		Inbounds: []domain.Inbound{bad, good},
		UsersByInbound: map[string][]*domain.User{
			"reality-nokey": {u()},
			"reality-ok":    {u()},
		},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var p struct {
		Inbounds []struct {
			Tag string `json:"tag"`
			TLS struct {
				Reality struct {
					Enabled bool   `json:"enabled"`
					PriKey  string `json:"private_key"`
				} `json:"reality"`
			} `json:"tls"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("parse: %v", err)
	}
	tags := map[string]bool{}
	for _, in := range p.Inbounds {
		tags[in.Tag] = true
		if in.TLS.Reality.Enabled && in.TLS.Reality.PriKey == "" {
			t.Errorf("inbound %q emitted reality.enabled with empty private_key", in.Tag)
		}
	}
	if tags["reality-nokey"] {
		t.Error("reality inbound with empty private key must be skipped")
	}
	if !tags["reality-ok"] {
		t.Error("reality inbound with a private key must render")
	}
}

// inboundRaw is a permissive view of a rendered sing-box inbound used by the
// added-protocol tests below, exposing the protocol-specific fields each one
// emits alongside the common header.
type inboundRaw struct {
	Type       string           `json:"type"`
	Tag        string           `json:"tag"`
	ListenPort int              `json:"listen_port"`
	Version    int              `json:"version"`
	UpMbps     int              `json:"up_mbps"`
	DownMbps   int              `json:"down_mbps"`
	Obfs       string           `json:"obfs"`
	StrictMode bool             `json:"strict_mode"`
	Padding    []string         `json:"padding_scheme"`
	Users      []map[string]any `json:"users"`
	TLS        map[string]any   `json:"tls"`
	Handshake  struct {
		Server     string `json:"server"`
		ServerPort int    `json:"server_port"`
	} `json:"handshake"`
}

func buildInbounds(t *testing.T, in domain.Inbound, users []*domain.User) []inboundRaw {
	t.Helper()
	raw, err := Builder{APIPort: 9090}.Build(&core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{in.Tag: users},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var p struct {
		Inbounds []inboundRaw `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("generated config invalid JSON: %v\n%s", err, raw)
	}
	return p.Inbounds
}

func ssUser() *domain.User {
	return &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{TrojanPass: "secretpw"}}
}

// TestBuilderHysteriaV1 verifies a minimal Hysteria v1 inbound renders the
// "hysteria" type with bandwidth, the per-user auth_str (from the trojan
// password), the obfs password from Raw, and a TLS block (Hysteria mandates TLS).
func TestBuilderHysteriaV1(t *testing.T) {
	u1, u2 := ssUser(), ssUser()
	in := domain.Inbound{
		Tag: "hy1", Protocol: domain.ProtoHysteria, Port: 8443,
		Security: domain.SecurityTLS, SNI: []string{"ex.com"},
		Raw: map[string]any{"hysteria": map[string]any{"up_mbps": 200, "down_mbps": 300, "obfs": "obfspw"}},
	}
	ins := buildInbounds(t, in, []*domain.User{u1, u2})
	if len(ins) != 1 {
		t.Fatalf("want 1 inbound, got %d", len(ins))
	}
	got := ins[0]
	if got.Type != "hysteria" || got.ListenPort != 8443 {
		t.Fatalf("hysteria header wrong: %+v", got)
	}
	if got.UpMbps != 200 || got.DownMbps != 300 {
		t.Errorf("bandwidth = up:%d down:%d, want 200/300", got.UpMbps, got.DownMbps)
	}
	if got.Obfs != "obfspw" {
		t.Errorf("obfs = %q, want obfspw", got.Obfs)
	}
	if got.TLS["enabled"] != true {
		t.Errorf("hysteria must carry a TLS block: %v", got.TLS)
	}
	if len(got.Users) != 2 {
		t.Fatalf("want 2 users, got %d", len(got.Users))
	}
	for i, want := range []string{u1.ID.String(), u2.ID.String()} {
		if got.Users[i]["name"] != want {
			t.Errorf("user[%d] name = %v, want %v", i, got.Users[i]["name"], want)
		}
		if got.Users[i]["auth_str"] != "secretpw" {
			t.Errorf("user[%d] auth_str = %v, want secretpw", i, got.Users[i]["auth_str"])
		}
	}
}

// TestBuilderHysteriaV1DefaultBandwidth verifies up/down default to non-zero
// when Raw omits them, so the inbound stays valid.
func TestBuilderHysteriaV1DefaultBandwidth(t *testing.T) {
	in := domain.Inbound{
		Tag: "hy1", Protocol: domain.ProtoHysteria, Port: 8443,
		Security: domain.SecurityTLS, SNI: []string{"ex.com"},
	}
	got := buildInbounds(t, in, []*domain.User{ssUser()})
	if len(got) != 1 || got[0].UpMbps <= 0 || got[0].DownMbps <= 0 {
		t.Fatalf("expected non-zero default bandwidth, got %+v", got)
	}
}

// TestBuilderHysteriaV1SkippedWithoutTLS verifies a Hysteria inbound with
// security=none is skipped (Hysteria mandates TLS) rather than emitting a
// broken block.
func TestBuilderHysteriaV1SkippedWithoutTLS(t *testing.T) {
	in := domain.Inbound{
		Tag: "hy1", Protocol: domain.ProtoHysteria, Port: 8443,
		Security: domain.SecurityNone,
	}
	if got := buildInbounds(t, in, []*domain.User{ssUser()}); len(got) != 0 {
		t.Errorf("hysteria without TLS must be skipped, got %+v", got)
	}
}

// TestBuilderAnyTLS verifies AnyTLS renders the "anytls" type with per-user
// passwords (from the trojan password), an optional padding_scheme from Raw,
// and a TLS block (AnyTLS mandates TLS).
func TestBuilderAnyTLS(t *testing.T) {
	u1, u2 := ssUser(), ssUser()
	in := domain.Inbound{
		Tag: "any1", Protocol: domain.ProtoAnyTLS, Port: 8443,
		Security: domain.SecurityTLS, SNI: []string{"ex.com"},
		Raw: map[string]any{"anytls": map[string]any{"padding_scheme": []any{"stop=8", "0=30-30"}}},
	}
	ins := buildInbounds(t, in, []*domain.User{u1, u2})
	if len(ins) != 1 {
		t.Fatalf("want 1 inbound, got %d", len(ins))
	}
	got := ins[0]
	if got.Type != "anytls" || got.ListenPort != 8443 {
		t.Fatalf("anytls header wrong: %+v", got)
	}
	if got.TLS["enabled"] != true {
		t.Errorf("anytls must carry a TLS block: %v", got.TLS)
	}
	if len(got.Padding) != 2 || got.Padding[0] != "stop=8" {
		t.Errorf("padding_scheme = %v, want [stop=8 0=30-30]", got.Padding)
	}
	if len(got.Users) != 2 {
		t.Fatalf("want 2 users, got %d", len(got.Users))
	}
	if got.Users[0]["name"] != u1.ID.String() || got.Users[0]["password"] != "secretpw" {
		t.Errorf("user[0] = %v", got.Users[0])
	}
}

// TestBuilderAnyTLSSkippedWithoutTLS verifies AnyTLS with security=none is
// skipped (it mandates TLS).
func TestBuilderAnyTLSSkippedWithoutTLS(t *testing.T) {
	in := domain.Inbound{
		Tag: "any1", Protocol: domain.ProtoAnyTLS, Port: 8443, Security: domain.SecurityNone,
	}
	if got := buildInbounds(t, in, []*domain.User{ssUser()}); len(got) != 0 {
		t.Errorf("anytls without TLS must be skipped, got %+v", got)
	}
}

// TestBuilderShadowTLS verifies ShadowTLS renders the "shadowtls" type at
// version 3 with a handshake target (from Raw), per-user passwords (from the
// trojan password), and NO separate tls block (it fronts its own handshake).
func TestBuilderShadowTLS(t *testing.T) {
	u1, u2 := ssUser(), ssUser()
	in := domain.Inbound{
		Tag: "stls", Protocol: domain.ProtoShadowTLS, Port: 443,
		Security: domain.SecurityNone,
		Raw: map[string]any{"shadowtls": map[string]any{
			"handshake_server": "www.apple.com", "handshake_port": 443, "version": 3,
		}},
	}
	ins := buildInbounds(t, in, []*domain.User{u1, u2})
	if len(ins) != 1 {
		t.Fatalf("want 1 inbound, got %d", len(ins))
	}
	got := ins[0]
	if got.Type != "shadowtls" || got.ListenPort != 443 {
		t.Fatalf("shadowtls header wrong: %+v", got)
	}
	if got.Version != 3 {
		t.Errorf("version = %d, want 3", got.Version)
	}
	if got.Handshake.Server != "www.apple.com" || got.Handshake.ServerPort != 443 {
		t.Errorf("handshake = %+v, want www.apple.com:443", got.Handshake)
	}
	if got.TLS != nil {
		t.Errorf("shadowtls must not carry a separate tls block, got %v", got.TLS)
	}
	if len(got.Users) != 2 {
		t.Fatalf("want 2 users, got %d", len(got.Users))
	}
	if got.Users[0]["name"] != u1.ID.String() || got.Users[0]["password"] != "secretpw" {
		t.Errorf("user[0] = %v", got.Users[0])
	}
}

// TestBuilderShadowTLSSkippedWithoutHandshake verifies a ShadowTLS inbound with
// no handshake target (no Raw handshake_server and no SNI) is skipped, since
// sing-box cannot front a handshake without one.
func TestBuilderShadowTLSSkippedWithoutHandshake(t *testing.T) {
	in := domain.Inbound{
		Tag: "stls", Protocol: domain.ProtoShadowTLS, Port: 443, Security: domain.SecurityNone,
	}
	if got := buildInbounds(t, in, []*domain.User{ssUser()}); len(got) != 0 {
		t.Errorf("shadowtls without handshake must be skipped, got %+v", got)
	}
}
