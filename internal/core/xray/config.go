// Package xray implements core.CoreDriver for Xray-core. This file renders the
// engine-neutral core.GeneratedConfig into Xray's native JSON using typed
// structs (no map[string]any soup) so the output is predictable and unit-tested.
package xray

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
)

// APIInboundTag is the reserved inbound Xray exposes its gRPC API on. The driver
// dials this for runtime user mutations and stats queries.
const APIInboundTag = "vortex-api"

// Builder renders core.GeneratedConfig to Xray JSON. apiPort is where the local
// gRPC API inbound listens (loopback only).
type Builder struct {
	APIPort int
}

// Build implements core.Builder.
func (b Builder) Build(cfg *core.GeneratedConfig) ([]byte, error) {
	x := xrayConfig{
		Log:   logConf{Loglevel: orDefault(cfg.LogLevel, "warning")},
		API:   apiConf{Tag: APIInboundTag, Services: []string{"HandlerService", "StatsService", "LoggerService"}},
		Stats: struct{}{},
		Policy: policyConf{
			Levels: map[string]policyLevel{
				"0": {StatsUserUplink: true, StatsUserDownlink: true},
			},
			System: systemPolicy{StatsInboundUplink: true, StatsInboundDownlink: true},
		},
	}

	// Reserved loopback inbound carrying the Xray API.
	x.Inbounds = append(x.Inbounds, inbound{
		Tag: APIInboundTag, Listen: "127.0.0.1", Port: b.APIPort,
		Protocol: "dokodemo-door",
		Settings: mustRaw(map[string]any{"address": "127.0.0.1"}),
	})

	for _, in := range cfg.Inbounds {
		// Skip inbounds whose security material is missing rather than letting one
		// misconfigured inbound crash the entire engine.
		if !inboundUsable(in) {
			continue
		}
		users := cfg.UsersByInbound[in.Tag]
		built, err := b.buildInbound(in, users)
		if err != nil {
			return nil, fmt.Errorf("inbound %q: %w", in.Tag, err)
		}
		x.Inbounds = append(x.Inbounds, built)
	}

	outbounds, err := buildOutbounds(cfg.Outbounds)
	if err != nil {
		return nil, err
	}
	x.Outbounds = outbounds

	x.Routing = buildRouting(cfg.Routing, cfg.Balancers)
	x.Observatory = buildObservatory(cfg.Balancers)

	return json.MarshalIndent(x, "", "  ")
}

// buildOutbounds renders the egress handlers. Xray routes unmatched traffic to
// the *first* outbound, so we always lead with the freedom "direct" handler —
// regardless of the order outbounds arrive in — to keep the default egress safe
// (never accidentally the blackhole). The well-known "direct" and "blocked" tags
// are guaranteed to exist so routing rules and the API can always reference them.
func buildOutbounds(outs []domain.Outbound) ([]outbound, error) {
	built := map[string]outbound{}
	var order []string
	for i := range outs {
		o := outs[i]
		if !o.Enabled {
			continue
		}
		if err := o.Validate(); err != nil {
			return nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		b, err := buildOutbound(o)
		if err != nil {
			return nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		if _, dup := built[o.Tag]; !dup {
			order = append(order, o.Tag)
		}
		built[o.Tag] = b
	}
	if _, ok := built["direct"]; !ok {
		built["direct"] = outbound{Protocol: "freedom", Tag: "direct"}
		order = append(order, "direct")
	}
	if _, ok := built["blocked"]; !ok {
		built["blocked"] = outbound{Protocol: "blackhole", Tag: "blocked"}
		order = append(order, "blocked")
	}
	// Lead with "direct" so it is Xray's default outbound; preserve the rest.
	result := []outbound{built["direct"]}
	for _, tag := range order {
		if tag == "direct" {
			continue
		}
		result = append(result, built[tag])
	}
	return result, nil
}

func buildOutbound(o domain.Outbound) (outbound, error) {
	ob := outbound{Protocol: string(o.Protocol), Tag: o.Tag}
	switch o.Protocol {
	case domain.OutFreedom, domain.OutBlackhole, domain.OutDNS:
		// No settings needed; freedom/blackhole/dns dispatch locally.
	case domain.OutVLESS:
		ob.Settings = mustRaw(map[string]any{"vnext": []map[string]any{{
			"address": o.Address, "port": o.Port,
			"users": []map[string]any{{"id": o.UUID, "encryption": "none", "flow": o.Flow}},
		}}})
	case domain.OutVMess:
		ob.Settings = mustRaw(map[string]any{"vnext": []map[string]any{{
			"address": o.Address, "port": o.Port,
			"users": []map[string]any{{"id": o.UUID}},
		}}})
	case domain.OutTrojan:
		ob.Settings = mustRaw(map[string]any{"servers": []map[string]any{{
			"address": o.Address, "port": o.Port, "password": o.Password,
		}}})
	case domain.OutShadowsocks:
		ob.Settings = mustRaw(map[string]any{"servers": []map[string]any{{
			"address": o.Address, "port": o.Port, "password": o.Password, "method": orDefault(o.Method, "aes-128-gcm"),
		}}})
	case domain.OutSocks, domain.OutHTTP:
		server := map[string]any{"address": o.Address, "port": o.Port}
		if o.Username != "" {
			server["users"] = []map[string]any{{"user": o.Username, "pass": o.Password}}
		}
		ob.Settings = mustRaw(map[string]any{"servers": []map[string]any{server}})
	default:
		return outbound{}, fmt.Errorf("unsupported outbound protocol %q", o.Protocol)
	}
	if o.Protocol.NeedsEndpoint() {
		ob.StreamSettings = outboundStream(o)
	}
	return ob, nil
}

// outboundStream renders the transport/TLS for a proxy outbound, mirroring the
// inbound streamSettings shape.
func outboundStream(o domain.Outbound) json.RawMessage {
	if o.Raw != nil {
		if raw, ok := o.Raw["streamSettings"]; ok {
			return mustRaw(raw)
		}
	}
	ss := map[string]any{
		"network":  orDefault(o.Network, "tcp"),
		"security": string(orSecurity(o.Security)),
	}
	if o.Security == domain.SecurityTLS && o.SNI != "" {
		ss["tlsSettings"] = map[string]any{"serverName": o.SNI}
	}
	switch o.Network {
	case "ws":
		ws := map[string]any{}
		if o.Path != "" {
			ws["path"] = o.Path
		}
		if o.Host != "" {
			ws["headers"] = map[string]any{"Host": o.Host}
		}
		ss["wsSettings"] = ws
	case "grpc":
		ss["grpcSettings"] = map[string]any{"serviceName": o.Path}
	}
	return mustRaw(ss)
}

func orSecurity(s domain.Security) domain.Security {
	if s == "" {
		return domain.SecurityNone
	}
	return s
}

// buildRouting assembles the routing block. The API inbound is always routed to
// the auto-created API outbound (whose tag equals api.tag = APIInboundTag — per
// Xray docs the api feature builds an outbound named after its tag), then the
// operator's enabled rules follow in priority order.
func buildRouting(rules []domain.RoutingRule, balancers []domain.Balancer) routingConf {
	rc := routingConf{
		Rules: []routingRule{{
			Type: "field", InboundTag: []string{APIInboundTag}, OutboundTag: APIInboundTag,
		}},
	}
	for _, b := range balancers {
		if !b.Enabled {
			continue
		}
		bc := balancerConf{Tag: b.Tag, Selector: b.Selectors}
		if b.Strategy != "" {
			bc.Strategy = balancerStrategy{Type: string(b.Strategy)}
		}
		rc.Balancers = append(rc.Balancers, bc)
	}
	ordered := append([]domain.RoutingRule(nil), rules...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Priority < ordered[j].Priority })
	for _, r := range ordered {
		if !r.Enabled {
			continue
		}
		rule := routingRule{
			Type:        "field",
			InboundTag:  r.InboundTags,
			Domain:      r.Domains,
			IP:          r.IP,
			Port:        r.Port,
			Network:     r.Network,
			Protocol:    r.Protocols,
			OutboundTag: r.OutboundTag,
			BalancerTag: r.BalancerTag,
		}
		rc.Rules = append(rc.Rules, rule)
	}
	return rc
}

// buildObservatory emits a single observatory probing every balancer that wants
// health data. Xray supports one top-level observatory, so the subject selectors
// are unioned; probe URL/interval come from the first balancer that sets them.
func buildObservatory(balancers []domain.Balancer) *observatoryConf {
	var subjects []string
	probeURL, probeInterval := "", ""
	for i := range balancers {
		b := balancers[i]
		if !b.Enabled || !b.WantsObservatory() {
			continue
		}
		subjects = append(subjects, b.Selectors...)
		if probeURL == "" && b.ProbeURL != "" {
			probeURL = b.ProbeURL
		}
		if probeInterval == "" && b.ProbeInterval != "" {
			probeInterval = b.ProbeInterval
		}
	}
	if len(subjects) == 0 {
		return nil
	}
	return &observatoryConf{
		SubjectSelector: subjects,
		ProbeURL:        orDefault(probeURL, "https://www.google.com/generate_204"),
		ProbeInterval:   orDefault(probeInterval, "10s"),
	}
}

func (b Builder) buildInbound(in domain.Inbound, users []*domain.User) (inbound, error) {
	settings, err := protocolSettings(in, users)
	if err != nil {
		return inbound{}, err
	}
	out := inbound{
		Tag:            in.Tag,
		Listen:         orDefault(in.Listen, "0.0.0.0"),
		Port:           in.Port,
		Protocol:       string(in.Protocol),
		Settings:       settings,
		StreamSettings: streamSettings(in),
	}
	return out, nil
}

// protocolSettings renders the protocol-specific "settings" block, including the
// per-user client list keyed off the shared credentials.
func protocolSettings(in domain.Inbound, users []*domain.User) (json.RawMessage, error) {
	switch in.Protocol {
	case domain.ProtoVLESS:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{
				"id": u.Proxies.VLESSUUID.String(), "email": u.ID.String(), "flow": in.Flow,
			})
		}
		return mustRaw(map[string]any{"clients": clients, "decryption": "none"}), nil

	case domain.ProtoVMess:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"id": u.Proxies.VMessUUID.String(), "email": u.ID.String()})
		}
		return mustRaw(map[string]any{"clients": clients}), nil

	case domain.ProtoTrojan:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"password": u.Proxies.TrojanPass, "email": u.ID.String()})
		}
		return mustRaw(map[string]any{"clients": clients}), nil

	case domain.ProtoShadowsocks:
		// Standard Shadowsocks ciphers (aes-*-gcm, chacha20) are single-user in
		// xray: a "clients" array is only valid for SS-2022 methods, and emitting
		// one crashes the inbound. Render a valid single-user listener from the
		// first bound user. (Multi-user SS needs SS-2022, tracked separately.)
		method := "aes-128-gcm"
		password := ""
		if len(users) > 0 {
			method = orDefault(users[0].Proxies.SSMethod, method)
			password = users[0].Proxies.ShadowsocksP
		}
		return mustRaw(map[string]any{"method": method, "password": password, "network": "tcp,udp"}), nil

	default:
		return nil, fmt.Errorf("unsupported protocol %q for xray", in.Protocol)
	}
}

// streamSettings renders transport + security. Operators can override the whole
// block via Inbound.Raw["streamSettings"] for fields the abstraction omits.
func streamSettings(in domain.Inbound) json.RawMessage {
	if raw, ok := in.Raw["streamSettings"]; ok {
		return mustRaw(raw)
	}
	ss := map[string]any{
		"network":  orDefault(in.Network, "tcp"),
		"security": string(in.Security),
	}
	switch in.Security {
	case domain.SecurityTLS:
		tls := map[string]any{}
		if len(in.SNI) > 0 {
			tls["serverName"] = in.SNI[0]
		}
		if cert := tlsCertificate(in.Raw["tls"]); cert != nil {
			tls["certificates"] = []any{cert}
		}
		ss["tlsSettings"] = tls
	case domain.SecurityReality:
		ss["realitySettings"] = realitySettings(in)
	}
	switch in.Network {
	case "ws":
		ws := map[string]any{}
		if in.Path != "" {
			ws["path"] = in.Path
		}
		if len(in.Host) > 0 {
			ws["headers"] = map[string]any{"Host": in.Host[0]}
		}
		ss["wsSettings"] = ws
	case "grpc":
		ss["grpcSettings"] = map[string]any{"serviceName": in.Path}
	}
	return mustRaw(ss)
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// realitySettings renders the server-side REALITY block from the engine-neutral
// params stored in Inbound.Raw["reality"]. serverNames and a default dest fall
// back to the inbound SNI so a freshly generated profile is usable.
func realitySettings(in domain.Inbound) map[string]any {
	p := reality.ParseParams(in.Raw["reality"])
	out := map[string]any{}
	names := p.ServerNames
	if len(names) == 0 {
		names = in.SNI
	}
	if len(names) > 0 {
		out["serverNames"] = names
	}
	if p.PrivateKey != "" {
		out["privateKey"] = p.PrivateKey
	}
	if len(p.ShortIDs) > 0 {
		out["shortIds"] = p.ShortIDs
	}
	dest := p.Dest
	if dest == "" && len(names) > 0 {
		dest = names[0] + ":443"
	}
	if dest != "" {
		out["dest"] = dest
	}
	return out
}

// tlsCertificate renders an inline xray certificate object from the PEM strings
// stored in Inbound.Raw["tls"] ({certificate, key}). xray accepts the PEM as an
// array of lines. Returns nil when no certificate is present.
func tlsCertificate(v any) map[string]any {
	m, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	cert, _ := m["certificate"].(string)
	key, _ := m["key"].(string)
	if cert == "" || key == "" {
		return nil
	}
	return map[string]any{
		"certificate": strings.Split(strings.TrimRight(cert, "\n"), "\n"),
		"key":         strings.Split(strings.TrimRight(key, "\n"), "\n"),
	}
}

// inboundUsable reports whether an inbound's security block has the material the
// core needs to start. A misconfigured inbound (REALITY without a private key,
// TLS without a certificate) would crash the whole engine, so the builder skips
// it instead — the rest of the config still loads. A full streamSettings
// override is trusted as-is.
func inboundUsable(in domain.Inbound) bool {
	if _, ok := in.Raw["streamSettings"]; ok {
		return true
	}
	switch in.Security {
	case domain.SecurityReality:
		return reality.ParseParams(in.Raw["reality"]).PrivateKey != ""
	case domain.SecurityTLS:
		return tlsCertificate(in.Raw["tls"]) != nil
	default:
		return true
	}
}

func mustRaw(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil { // only unmarshalable types (chan/func) hit this; never with our inputs
		panic(fmt.Sprintf("xray: marshal settings: %v", err))
	}
	return b
}
