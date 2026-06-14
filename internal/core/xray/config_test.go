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
	if _, err := (Builder{APIPort: 1}).Build(cfg); err == nil {
		t.Fatal("expected error for unsupported protocol, got nil")
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
