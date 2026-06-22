package xray

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/warp"
)

func TestBuilder_RendersValidXrayConfig(t *testing.T) {
	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}

	in := domain.Inbound{
		Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "ws", Security: domain.SecurityTLS, SNI: []string{"example.com"}, Path: "/v",
		Raw: map[string]any{"tls": map[string]any{
			"certificate": "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
			"key":         "-----BEGIN PRIVATE KEY-----\nMIIB\n-----END PRIVATE KEY-----",
		}},
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
		API    apiConf `json:"api"`
		Policy struct {
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
	// Unsupported protocols are now skipped (not fatal) so one bad inbound
	// doesn't crash the core. The config is still valid — it just won't contain
	// the unsupported inbound.
	raw, err := (Builder{APIPort: 1}).Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(raw) == 0 {
		t.Fatal("expected non-empty config")
	}
}

// fullRouting is a parse target for the routing/outbound/observatory output.
type fullRouting struct {
	Outbounds []struct {
		Protocol string          `json:"protocol"`
		Tag      string          `json:"tag"`
		Settings json.RawMessage `json:"settings"`
	} `json:"outbounds"`
	Routing struct {
		Balancers []struct {
			Tag      string   `json:"tag"`
			Selector []string `json:"selector"`
			Strategy struct {
				Type string `json:"type"`
			} `json:"strategy"`
		} `json:"balancers"`
		Rules []struct {
			Type        string   `json:"type"`
			InboundTag  []string `json:"inboundTag"`
			Domain      []string `json:"domain"`
			Port        string   `json:"port"`
			OutboundTag string   `json:"outboundTag"`
			BalancerTag string   `json:"balancerTag"`
		} `json:"rules"`
	} `json:"routing"`
	Observatory *struct {
		SubjectSelector []string `json:"subjectSelector"`
		ProbeURL        string   `json:"probeUrl"`
		ProbeInterval   string   `json:"probeInterval"`
	} `json:"observatory"`
}

func TestBuilder_APIRuleRoutesToAPIOutbound(t *testing.T) {
	// Regression: the API inbound must route to the auto-created outbound whose
	// tag equals api.tag (APIInboundTag), not a non-existent "api" tag.
	raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r fullRouting
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(r.Routing.Rules) == 0 {
		t.Fatal("no routing rules emitted")
	}
	api := r.Routing.Rules[0]
	if len(api.InboundTag) != 1 || api.InboundTag[0] != APIInboundTag {
		t.Errorf("first rule inboundTag = %v, want [%s]", api.InboundTag, APIInboundTag)
	}
	if api.OutboundTag != APIInboundTag {
		t.Errorf("API rule outboundTag = %q, want %q", api.OutboundTag, APIInboundTag)
	}
	// Default direct/blocked outbounds must always be present.
	tags := map[string]bool{}
	for _, o := range r.Outbounds {
		tags[o.Tag] = true
	}
	if !tags["direct"] || !tags["blocked"] {
		t.Errorf("default outbounds missing: %v", tags)
	}
}

func TestBuilder_OutboundsAndRoutingPriority(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "proxy-de", Protocol: domain.OutVLESS, Address: "de.example.com", Port: 443, UUID: "11111111-1111-1111-1111-111111111111", Network: "tcp", Security: domain.SecurityTLS, SNI: "de.example.com", Enabled: true},
			{Tag: "disabled", Protocol: domain.OutFreedom, Enabled: false},
		},
		Routing: []domain.RoutingRule{
			{Priority: 20, InboundTags: []string{"vless-ws"}, OutboundTag: "proxy-de", Enabled: true},
			{Priority: 10, Domains: []string{"geosite:category-ads"}, OutboundTag: "blocked", Enabled: true},
			{Priority: 5, Domains: []string{"x"}, OutboundTag: "direct", Enabled: false}, // skipped
		},
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r fullRouting
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}

	// proxy-de must be rendered with vnext settings; disabled one skipped.
	var foundProxy bool
	for _, o := range r.Outbounds {
		if o.Tag == "disabled" {
			t.Error("disabled outbound should not be rendered")
		}
		if o.Tag == "proxy-de" {
			foundProxy = true
			if o.Protocol != "vless" {
				t.Errorf("proxy-de protocol = %q", o.Protocol)
			}
		}
	}
	if !foundProxy {
		t.Fatal("proxy-de outbound missing")
	}

	// Rules: API rule first, then enabled rules in ascending priority
	// (ads@10 before proxy-de@20); the disabled rule is absent.
	if len(r.Routing.Rules) != 3 {
		t.Fatalf("want 3 rules (api + 2 enabled), got %d", len(r.Routing.Rules))
	}
	if r.Routing.Rules[1].OutboundTag != "blocked" {
		t.Errorf("rule[1] (priority 10) outbound = %q, want blocked", r.Routing.Rules[1].OutboundTag)
	}
	if r.Routing.Rules[2].OutboundTag != "proxy-de" {
		t.Errorf("rule[2] (priority 20) outbound = %q, want proxy-de", r.Routing.Rules[2].OutboundTag)
	}
}

func TestBuilder_BalancerWithObservatory(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "proxy-a", Protocol: domain.OutVLESS, Address: "a", Port: 1, UUID: "u", Enabled: true},
			{Tag: "proxy-b", Protocol: domain.OutVLESS, Address: "b", Port: 1, UUID: "u", Enabled: true},
		},
		Balancers: []domain.Balancer{
			{Tag: "lb", Selectors: []string{"proxy-"}, Strategy: domain.BalancerLeastPing, Enabled: true},
		},
		Routing: []domain.RoutingRule{
			{Priority: 1, InboundTags: []string{"vless-ws"}, BalancerTag: "lb", Enabled: true},
		},
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r fullRouting
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(r.Routing.Balancers) != 1 || r.Routing.Balancers[0].Tag != "lb" {
		t.Fatalf("balancer not rendered: %+v", r.Routing.Balancers)
	}
	if r.Routing.Balancers[0].Strategy.Type != "leastPing" {
		t.Errorf("strategy = %q, want leastPing", r.Routing.Balancers[0].Strategy.Type)
	}
	// leastPing requires an observatory covering the balancer's selectors.
	if r.Observatory == nil {
		t.Fatal("leastPing balancer must produce an observatory")
	}
	if len(r.Observatory.SubjectSelector) != 1 || r.Observatory.SubjectSelector[0] != "proxy-" {
		t.Errorf("observatory subjectSelector = %v", r.Observatory.SubjectSelector)
	}
	if r.Observatory.ProbeInterval == "" || r.Observatory.ProbeURL == "" {
		t.Error("observatory must have default probe url/interval")
	}
	// The rule must target the balancer, not an outbound.
	last := r.Routing.Rules[len(r.Routing.Rules)-1]
	if last.BalancerTag != "lb" || last.OutboundTag != "" {
		t.Errorf("rule should target balancer: %+v", last)
	}
}

func TestBuilder_RealitySettings(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-reality", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "tcp", Security: domain.SecurityReality, SNI: []string{"www.microsoft.com"},
		Raw: map[string]any{
			"reality": map[string]any{
				"private_key": "QABC_privatekey",
				"short_ids":   []any{"", "0123abcd"},
				"dest":        "www.microsoft.com:443",
			},
		},
	}
	raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{"vless-reality": {{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}}},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var parsed struct {
		Inbounds []struct {
			Tag            string          `json:"tag"`
			StreamSettings json.RawMessage `json:"streamSettings"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var streamRaw json.RawMessage
	for _, i := range parsed.Inbounds {
		if i.Tag == "vless-reality" {
			streamRaw = i.StreamSettings
		}
	}
	if streamRaw == nil {
		t.Fatal("reality inbound missing")
	}
	var ss struct {
		Security        string `json:"security"`
		RealitySettings struct {
			PrivateKey  string   `json:"privateKey"`
			ShortIds    []string `json:"shortIds"`
			ServerNames []string `json:"serverNames"`
			Dest        string   `json:"dest"`
		} `json:"realitySettings"`
	}
	if err := json.Unmarshal(streamRaw, &ss); err != nil {
		t.Fatalf("stream parse: %v", err)
	}
	if ss.Security != "reality" {
		t.Errorf("security = %q, want reality", ss.Security)
	}
	if ss.RealitySettings.PrivateKey != "QABC_privatekey" {
		t.Errorf("privateKey = %q", ss.RealitySettings.PrivateKey)
	}
	if len(ss.RealitySettings.ShortIds) != 2 {
		t.Errorf("shortIds = %v, want 2", ss.RealitySettings.ShortIds)
	}
	if len(ss.RealitySettings.ServerNames) != 1 || ss.RealitySettings.ServerNames[0] != "www.microsoft.com" {
		t.Errorf("serverNames = %v", ss.RealitySettings.ServerNames)
	}
	if ss.RealitySettings.Dest != "www.microsoft.com:443" {
		t.Errorf("dest = %q", ss.RealitySettings.Dest)
	}
}

func TestBuilder_InvalidOutboundErrors(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "bad", Protocol: domain.OutVLESS, Enabled: true}, // missing address/port
		},
	}
	if _, err := (Builder{APIPort: 1}).Build(cfg); err == nil {
		t.Fatal("expected error for proxy outbound without endpoint")
	}
}

func TestBuilder_DirectLeadsSoDefaultEgressIsSafe(t *testing.T) {
	// "blocked" sorts before "direct" alphabetically (as the DB returns them);
	// the builder must still place "direct" first so Xray's default outbound is
	// direct egress, never the blackhole.
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "blocked", Protocol: domain.OutBlackhole, Enabled: true},
			{Tag: "aaa-proxy", Protocol: domain.OutVLESS, Address: "x", Port: 1, UUID: "u", Enabled: true},
		},
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r fullRouting
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(r.Outbounds) == 0 || r.Outbounds[0].Tag != "direct" {
		t.Fatalf("first (default) outbound = %q, want direct", firstTag(r.Outbounds))
	}
	if r.Outbounds[0].Protocol != "freedom" {
		t.Errorf("default outbound protocol = %q, want freedom", r.Outbounds[0].Protocol)
	}
}

func firstTag(outs []struct {
	Protocol string          `json:"protocol"`
	Tag      string          `json:"tag"`
	Settings json.RawMessage `json:"settings"`
}) string {
	if len(outs) == 0 {
		return ""
	}
	return outs[0].Tag
}

// TestBuilder_GeoBlockAllowedCountries verifies that GeoBlockRules for an
// AllowedCountries policy render into xray routing such that (a) allowed-country
// traffic is routed to a real egress (not dropped, not blocked) and (b) the rest
// of the inbound's traffic is sent to the "blocked" outbound. The risk being
// guarded against: if the "allow" rule lost its target it would be omitted and
// the catch-all would block everything, breaking allowed-country access.
func TestBuilder_GeoBlockAllowedCountries(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-geo", Protocol: domain.ProtoVLESS, Port: 443,
		GeoPolicy: &domain.GeoPolicy{AllowedCountries: []string{"IR"}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		Routing:  core.GeoBlockRules(in),
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r struct {
		Routing struct {
			Rules []struct {
				InboundTag  []string `json:"inboundTag"`
				IP          []string `json:"ip"`
				OutboundTag string   `json:"outboundTag"`
			} `json:"rules"`
		} `json:"routing"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}

	var allow, blockRest *struct {
		InboundTag  []string `json:"inboundTag"`
		IP          []string `json:"ip"`
		OutboundTag string   `json:"outboundTag"`
	}
	for i := range r.Routing.Rules {
		rule := &r.Routing.Rules[i]
		if len(rule.InboundTag) != 1 || rule.InboundTag[0] != "vless-geo" {
			continue
		}
		if len(rule.IP) == 1 && rule.IP[0] == "geoip:IR" {
			allow = rule
		} else if len(rule.IP) == 0 && rule.OutboundTag == "blocked" {
			blockRest = rule
		}
	}

	// (b) Non-IR traffic on the inbound must be blocked.
	if blockRest == nil {
		t.Fatal("missing catch-all rule sending non-allowed traffic to blocked")
	}
	// (a) IR traffic must be routed to a real egress, NOT dropped or blocked.
	if allow == nil {
		t.Fatal("allow rule for geoip:IR was dropped — allowed traffic would be blocked")
	}
	if allow.OutboundTag == "" {
		t.Error("allow rule has no outboundTag — xray would omit it and block allowed traffic")
	}
	if allow.OutboundTag == "blocked" {
		t.Errorf("allow rule routes allowed traffic to %q (must not be blocked)", allow.OutboundTag)
	}
}

// TestBuilder_GeoBlockBlockedCountries verifies a BlockedCountries policy renders
// a rule sending traffic from the blocked country's source IPs to "blocked".
func TestBuilder_GeoBlockBlockedCountries(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-geo", Protocol: domain.ProtoVLESS, Port: 443,
		GeoPolicy: &domain.GeoPolicy{BlockedCountries: []string{"CN"}},
	}
	cfg := &core.GeneratedConfig{
		Inbounds: []domain.Inbound{in},
		Routing:  core.GeoBlockRules(in),
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r struct {
		Routing struct {
			Rules []struct {
				InboundTag  []string `json:"inboundTag"`
				IP          []string `json:"ip"`
				OutboundTag string   `json:"outboundTag"`
			} `json:"rules"`
		} `json:"routing"`
	}
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var found bool
	for _, rule := range r.Routing.Rules {
		if len(rule.IP) == 1 && rule.IP[0] == "geoip:CN" && rule.OutboundTag == "blocked" {
			found = true
		}
	}
	if !found {
		t.Error("expected a rule sending geoip:CN traffic to blocked")
	}
}

// TestBuilder_VLESSFlowOnlyForValidCombos verifies xtls-rprx-vision is emitted
// only for VLESS over raw TCP with TLS or REALITY. On ws (or any non-TCP
// transport) or with security=none the flow must be omitted entirely (not even
// "flow":""), otherwise xray rejects the config.
func TestBuilder_VLESSFlowOnlyForValidCombos(t *testing.T) {
	flowOf := func(t *testing.T, in domain.Inbound) (string, bool) {
		t.Helper()
		u := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
		raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{
			Inbounds:       []domain.Inbound{in},
			UsersByInbound: map[string][]*domain.User{in.Tag: {u}},
		})
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		var parsed struct {
			Inbounds []struct {
				Tag      string          `json:"tag"`
				Settings json.RawMessage `json:"settings"`
			} `json:"inbounds"`
		}
		if err := json.Unmarshal(raw, &parsed); err != nil {
			t.Fatalf("parse: %v", err)
		}
		for _, i := range parsed.Inbounds {
			if i.Tag != in.Tag {
				continue
			}
			var s struct {
				Clients []map[string]any `json:"clients"`
			}
			if err := json.Unmarshal(i.Settings, &s); err != nil {
				t.Fatalf("settings parse: %v", err)
			}
			if len(s.Clients) == 0 {
				t.Fatalf("inbound %q has no clients", in.Tag)
			}
			flow, present := s.Clients[0]["flow"]
			if !present {
				return "", false
			}
			str, _ := flow.(string)
			return str, true
		}
		t.Fatalf("inbound %q missing from output", in.Tag)
		return "", false
	}

	tlsCert := map[string]any{"tls": map[string]any{
		"certificate": "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
		"key":         "-----BEGIN PRIVATE KEY-----\nMIIB\n-----END PRIVATE KEY-----",
	}}

	// ws transport: flow must be omitted even though Flow is set.
	if flow, present := flowOf(t, domain.Inbound{
		Tag: "vless-ws", Protocol: domain.ProtoVLESS, Port: 443, Network: "ws",
		Security: domain.SecurityTLS, SNI: []string{"ex.com"}, Flow: "xtls-rprx-vision",
		Raw: tlsCert,
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
		Security: domain.SecurityReality, SNI: []string{"www.microsoft.com"},
		Flow: "xtls-rprx-vision",
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK", "dest": "www.microsoft.com:443",
		}},
	})
	if !present || flow != "xtls-rprx-vision" {
		t.Errorf("vless+tcp+reality flow = %q (present=%v), want xtls-rprx-vision", flow, present)
	}

	// vless + tcp + reality with NO flow: must auto-default to xtls-rprx-vision
	// (REALITY best practice).
	autoFlow, autoPresent := flowOf(t, domain.Inbound{
		Tag: "vless-reality-auto", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.microsoft.com"},
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK", "dest": "www.microsoft.com:443",
		}},
	})
	if !autoPresent || autoFlow != "xtls-rprx-vision" {
		t.Errorf("vless+tcp+reality (blank flow) = %q (present=%v), want auto xtls-rprx-vision", autoFlow, autoPresent)
	}

	// An explicitly-set non-default flow on reality must be preserved as-is.
	customFlow, customPresent := flowOf(t, domain.Inbound{
		Tag: "vless-reality-custom", Protocol: domain.ProtoVLESS, Port: 443, Network: "tcp",
		Security: domain.SecurityReality, SNI: []string{"www.microsoft.com"},
		Flow: "xtls-rprx-direct",
		Raw: map[string]any{"reality": map[string]any{
			"private_key": "PK", "dest": "www.microsoft.com:443",
		}},
	})
	if !customPresent || customFlow != "xtls-rprx-direct" {
		t.Errorf("explicit flow = %q (present=%v), want xtls-rprx-direct preserved", customFlow, customPresent)
	}
}

// ssSettingsFor builds the config for a single SS inbound and returns its parsed
// settings block, so SS-shape assertions stay terse.
func ssSettingsFor(t *testing.T, in domain.Inbound, users []*domain.User) struct {
	Method   string           `json:"method"`
	Password string           `json:"password"`
	Network  string           `json:"network"`
	Clients  []map[string]any `json:"clients"`
} {
	t.Helper()
	raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{in.Tag: users},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var parsed struct {
		Inbounds []struct {
			Tag      string          `json:"tag"`
			Settings json.RawMessage `json:"settings"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	var settings struct {
		Method   string           `json:"method"`
		Password string           `json:"password"`
		Network  string           `json:"network"`
		Clients  []map[string]any `json:"clients"`
	}
	found := false
	for _, i := range parsed.Inbounds {
		if i.Tag == in.Tag {
			if err := json.Unmarshal(i.Settings, &settings); err != nil {
				t.Fatalf("settings parse: %v", err)
			}
			found = true
		}
	}
	if !found {
		t.Fatalf("inbound %q missing from output", in.Tag)
	}
	return settings
}

// TestBuilder_Shadowsocks2022MultiUser verifies that a Shadowsocks-2022 inbound
// with two bound users renders a server-level method+password plus a per-user
// clients array (length 2), each entry carrying that user's PSK and UUID email.
// Regression guard: the old renderer silently dropped every user past users[0].
func TestBuilder_Shadowsocks2022MultiUser(t *testing.T) {
	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{SSMethod: "2022-blake3-aes-256-gcm", ShadowsocksP: "psk-user-1"}}
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{SSMethod: "2022-blake3-aes-256-gcm", ShadowsocksP: "psk-user-2"}}

	in := domain.Inbound{
		Tag: "ss-2022", Protocol: domain.ProtoShadowsocks, Port: 8388,
		Network: "tcp", Security: domain.SecurityNone,
		Raw: map[string]any{"ss": map[string]any{"password": "server-psk"}},
	}
	s := ssSettingsFor(t, in, []*domain.User{u1, u2})

	if s.Method != "2022-blake3-aes-256-gcm" {
		t.Errorf("method = %q, want 2022-blake3-aes-256-gcm", s.Method)
	}
	if s.Password != "server-psk" {
		t.Errorf("server password = %q, want server-psk", s.Password)
	}
	if len(s.Clients) != 2 {
		t.Fatalf("want 2 clients, got %d", len(s.Clients))
	}
	// Each client must carry its own PSK and UUID email so multiple users work
	// and stats counters map back to users.
	byEmail := map[string]string{}
	for _, c := range s.Clients {
		email, _ := c["email"].(string)
		pw, _ := c["password"].(string)
		byEmail[email] = pw
	}
	if byEmail[u1.ID.String()] != "psk-user-1" {
		t.Errorf("client[%s] password = %q, want psk-user-1", u1.ID, byEmail[u1.ID.String()])
	}
	if byEmail[u2.ID.String()] != "psk-user-2" {
		t.Errorf("client[%s] password = %q, want psk-user-2", u2.ID, byEmail[u2.ID.String()])
	}
}

// TestBuilder_ShadowsocksLegacySingleCredential verifies that a legacy-cipher SS
// inbound renders a single method+password and NO clients array (which xray
// rejects on legacy ciphers), and does not crash when multiple users are bound.
func TestBuilder_ShadowsocksLegacySingleCredential(t *testing.T) {
	u1 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{SSMethod: "aes-256-gcm", ShadowsocksP: "shared-pass"}}
	u2 := &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{SSMethod: "aes-256-gcm", ShadowsocksP: "other-pass"}}

	in := domain.Inbound{
		Tag: "ss-legacy", Protocol: domain.ProtoShadowsocks, Port: 8389,
		Network: "tcp", Security: domain.SecurityNone,
	}
	s := ssSettingsFor(t, in, []*domain.User{u1, u2})

	if s.Method != "aes-256-gcm" {
		t.Errorf("method = %q, want aes-256-gcm", s.Method)
	}
	if s.Password != "shared-pass" {
		t.Errorf("password = %q, want shared-pass (first user's credential)", s.Password)
	}
	if len(s.Clients) != 0 {
		t.Errorf("legacy SS must not emit a clients array, got %d entries", len(s.Clients))
	}
	if s.Network != "tcp,udp" {
		t.Errorf("network = %q, want tcp,udp", s.Network)
	}
}

// streamSettingsFor builds the config for a single inbound and returns its
// parsed streamSettings as a generic map, so transport-shape assertions on the
// new TLS-ALPN / tcp-header / xhttp-mode rendering stay terse.
func streamSettingsFor(t *testing.T, in domain.Inbound, users []*domain.User) map[string]any {
	t.Helper()
	raw, err := Builder{APIPort: 10085}.Build(&core.GeneratedConfig{
		Inbounds:       []domain.Inbound{in},
		UsersByInbound: map[string][]*domain.User{in.Tag: users},
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var parsed struct {
		Inbounds []struct {
			Tag            string          `json:"tag"`
			StreamSettings json.RawMessage `json:"streamSettings"`
		} `json:"inbounds"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, i := range parsed.Inbounds {
		if i.Tag == in.Tag {
			var ss map[string]any
			if err := json.Unmarshal(i.StreamSettings, &ss); err != nil {
				t.Fatalf("streamSettings parse: %v", err)
			}
			return ss
		}
	}
	t.Fatalf("inbound %q missing from output", in.Tag)
	return nil
}

func vlessUser() *domain.User {
	return &domain.User{ID: uuid.New(), Proxies: domain.UserCredentials{VLESSUUID: uuid.New()}}
}

// TestBuilder_TLSALPN verifies a TLS inbound with Raw["tls"].alpn renders the
// list into tlsSettings.alpn.
func TestBuilder_TLSALPN(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-tls-alpn", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "ws", Security: domain.SecurityTLS, SNI: []string{"ex.com"}, Path: "/v",
		Raw: map[string]any{"tls": map[string]any{
			"certificate": "-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----",
			"key":         "-----BEGIN PRIVATE KEY-----\nMIIB\n-----END PRIVATE KEY-----",
			"alpn":        []any{"h2", "http/1.1"},
		}},
	}
	ss := streamSettingsFor(t, in, []*domain.User{vlessUser()})
	tls, ok := ss["tlsSettings"].(map[string]any)
	if !ok {
		t.Fatalf("tlsSettings missing or wrong type: %v", ss["tlsSettings"])
	}
	alpn, ok := tls["alpn"].([]any)
	if !ok || len(alpn) != 2 || alpn[0] != "h2" || alpn[1] != "http/1.1" {
		t.Errorf("tlsSettings.alpn = %v, want [h2 http/1.1]", tls["alpn"])
	}
}

// TestBuilder_TCPHeaderDefault verifies a plain tcp inbound renders
// tcpSettings.header.type == "none".
func TestBuilder_TCPHeaderDefault(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-tcp", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "tcp", Security: domain.SecurityNone,
	}
	ss := streamSettingsFor(t, in, []*domain.User{vlessUser()})
	tcp, ok := ss["tcpSettings"].(map[string]any)
	if !ok {
		t.Fatalf("tcpSettings missing or wrong type: %v", ss["tcpSettings"])
	}
	header, ok := tcp["header"].(map[string]any)
	if !ok || header["type"] != "none" {
		t.Errorf("tcpSettings.header = %v, want {type:none}", tcp["header"])
	}
}

// TestBuilder_TCPHeaderHTTPOverride verifies an operator-supplied tcp header
// (Raw["tcp"].header) is rendered verbatim into tcpSettings.
func TestBuilder_TCPHeaderHTTPOverride(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-tcp-http", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "tcp", Security: domain.SecurityNone,
		Raw: map[string]any{"tcp": map[string]any{
			"header": map[string]any{
				"type":    "http",
				"request": map[string]any{"path": []any{"/"}},
			},
		}},
	}
	ss := streamSettingsFor(t, in, []*domain.User{vlessUser()})
	tcp, ok := ss["tcpSettings"].(map[string]any)
	if !ok {
		t.Fatalf("tcpSettings missing or wrong type: %v", ss["tcpSettings"])
	}
	header, ok := tcp["header"].(map[string]any)
	if !ok || header["type"] != "http" {
		t.Fatalf("tcpSettings.header = %v, want {type:http,...}", tcp["header"])
	}
	if _, ok := header["request"].(map[string]any); !ok {
		t.Errorf("tcpSettings.header.request missing: %v", header)
	}
}

// TestBuilder_XHTTPModeOverride verifies an xhttp inbound with Raw["xhttp"].mode
// renders that mode (instead of the "auto" default).
func TestBuilder_XHTTPModeOverride(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-xhttp", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "xhttp", Security: domain.SecurityNone, Path: "/x",
		Raw: map[string]any{"xhttp": map[string]any{"mode": "packet-up"}},
	}
	ss := streamSettingsFor(t, in, []*domain.User{vlessUser()})
	x, ok := ss["xhttpSettings"].(map[string]any)
	if !ok {
		t.Fatalf("xhttpSettings missing or wrong type: %v", ss["xhttpSettings"])
	}
	if x["mode"] != "packet-up" {
		t.Errorf("xhttpSettings.mode = %v, want packet-up", x["mode"])
	}
	if x["path"] != "/x" {
		t.Errorf("xhttpSettings.path = %v, want /x (existing path logic must be kept)", x["path"])
	}
}

// TestBuilder_XHTTPModeDefault verifies an xhttp inbound with no Raw override
// keeps the "auto" default mode.
func TestBuilder_XHTTPModeDefault(t *testing.T) {
	in := domain.Inbound{
		Tag: "vless-xhttp-def", Protocol: domain.ProtoVLESS, Port: 443,
		Network: "xhttp", Security: domain.SecurityNone,
	}
	ss := streamSettingsFor(t, in, []*domain.User{vlessUser()})
	x, ok := ss["xhttpSettings"].(map[string]any)
	if !ok {
		t.Fatalf("xhttpSettings missing or wrong type: %v", ss["xhttpSettings"])
	}
	if x["mode"] != "auto" {
		t.Errorf("xhttpSettings.mode = %v, want auto (default)", x["mode"])
	}
}

// TestBuilder_WireguardOutbound verifies a wireguard/WARP outbound renders as an
// xray wireguard outbound whose settings carry the secretKey, address and a
// peers entry. With endpoint/publicKey left unset the peer must default to
// Cloudflare's WARP endpoint and public key, and the reserved bytes pass through.
func TestBuilder_WireguardOutbound(t *testing.T) {
	cfg := &core.GeneratedConfig{
		Outbounds: []domain.Outbound{
			{Tag: "warp", Protocol: domain.OutWireguard, Enabled: true, Raw: map[string]any{
				"wireguard": map[string]any{
					"private_key": "secret-key",
					"address":     []any{"172.16.0.2/32", "fd01::1/128"},
					"reserved":    []any{float64(171), float64(48), float64(225)},
				},
			}},
		},
	}
	raw, err := Builder{APIPort: 10085}.Build(cfg)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	var r fullRouting
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("parse: %v", err)
	}

	var wg *struct {
		Protocol string          `json:"protocol"`
		Tag      string          `json:"tag"`
		Settings json.RawMessage `json:"settings"`
	}
	for i := range r.Outbounds {
		if r.Outbounds[i].Tag == "warp" {
			wg = &r.Outbounds[i]
		}
	}
	if wg == nil {
		t.Fatal("warp outbound missing from output")
	}
	if wg.Protocol != "wireguard" {
		t.Errorf("protocol = %q, want wireguard", wg.Protocol)
	}

	var s struct {
		SecretKey string           `json:"secretKey"`
		Address   []string         `json:"address"`
		Reserved  []int            `json:"reserved"`
		Peers     []map[string]any `json:"peers"`
	}
	if err := json.Unmarshal(wg.Settings, &s); err != nil {
		t.Fatalf("settings parse: %v", err)
	}
	if s.SecretKey != "secret-key" {
		t.Errorf("secretKey = %q, want secret-key", s.SecretKey)
	}
	if len(s.Address) != 2 || s.Address[0] != "172.16.0.2/32" {
		t.Errorf("address = %v, want [172.16.0.2/32 fd01::1/128]", s.Address)
	}
	if len(s.Reserved) != 3 || s.Reserved[0] != 171 {
		t.Errorf("reserved = %v, want [171 48 225]", s.Reserved)
	}
	if len(s.Peers) != 1 {
		t.Fatalf("want 1 peer, got %d", len(s.Peers))
	}
	if s.Peers[0]["publicKey"] != warp.DefaultPublicKey {
		t.Errorf("peer publicKey = %v, want Cloudflare default", s.Peers[0]["publicKey"])
	}
	if s.Peers[0]["endpoint"] != warp.DefaultEndpoint {
		t.Errorf("peer endpoint = %v, want Cloudflare default", s.Peers[0]["endpoint"])
	}
}
