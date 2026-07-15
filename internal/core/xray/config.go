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
	"github.com/vortexui/vortexui/internal/warp"
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
				"0": {StatsUserUplink: true, StatsUserDownlink: true, StatsUserOnline: true},
			},
			System: systemPolicy{StatsInboundUplink: true, StatsInboundDownlink: true},
		},
	}

	// Reserved loopback inbound carrying the Xray API.
	apiSettings, err := tryRaw(map[string]any{"address": "127.0.0.1"})
	if err != nil {
		return nil, err
	}
	x.Inbounds = append(x.Inbounds, inbound{
		Tag: APIInboundTag, Listen: "127.0.0.1", Port: b.APIPort,
		Protocol: "dokodemo-door",
		Settings: apiSettings,
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
			// Log but skip — one bad inbound must not take down the whole core.
			continue
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
	var (
		s   json.RawMessage
		err error
	)
	switch o.Protocol {
	case domain.OutFreedom, domain.OutBlackhole, domain.OutDNS:
		// No settings needed; freedom/blackhole/dns dispatch locally.
	case domain.OutVLESS:
		s, err = tryRaw(map[string]any{"vnext": []map[string]any{{
			"address": o.Address, "port": o.Port,
			"users": []map[string]any{{"id": o.UUID, "encryption": "none", "flow": o.Flow}},
		}}})
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q vless: %w", o.Tag, err)
		}
		ob.Settings = s
	case domain.OutVMess:
		s, err = tryRaw(map[string]any{"vnext": []map[string]any{{
			"address": o.Address, "port": o.Port,
			"users": []map[string]any{{"id": o.UUID}},
		}}})
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q vmess: %w", o.Tag, err)
		}
		ob.Settings = s
	case domain.OutTrojan:
		s, err = tryRaw(map[string]any{"servers": []map[string]any{{
			"address": o.Address, "port": o.Port, "password": o.Password,
		}}})
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q trojan: %w", o.Tag, err)
		}
		ob.Settings = s
	case domain.OutShadowsocks:
		s, err = tryRaw(map[string]any{"servers": []map[string]any{{
			"address": o.Address, "port": o.Port, "password": o.Password, "method": orDefault(o.Method, "aes-128-gcm"),
		}}})
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q shadowsocks: %w", o.Tag, err)
		}
		ob.Settings = s
	case domain.OutSocks, domain.OutHTTP:
		server := map[string]any{"address": o.Address, "port": o.Port}
		if o.Username != "" {
			server["users"] = []map[string]any{{"user": o.Username, "pass": o.Password}}
		}
		s, err = tryRaw(map[string]any{"servers": []map[string]any{server}})
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q socks/http: %w", o.Tag, err)
		}
		ob.Settings = s
	case domain.OutWireguard:
		// WireGuard/WARP: build the native xray wireguard outbound from the params
		// persisted in Raw["wireguard"]. cfg.XrayOutbound returns the full
		// {tag,protocol,settings} object; we take its "settings" sub-map (the
		// peers default to Cloudflare's endpoint/public key when unset). No
		// streamSettings are attached — wireguard carries its own endpoint.
		cfg := warp.ConfigFromMap(asWireguardMap(o.Raw["wireguard"]))
		ob.Protocol = "wireguard"
		s, err = tryRaw(cfg.XrayOutbound(o.Tag)["settings"])
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q wireguard: %w", o.Tag, err)
		}
		ob.Settings = s
		return ob, nil
	default:
		return outbound{}, fmt.Errorf("unsupported outbound protocol %q", o.Protocol)
	}
	if o.Protocol.NeedsEndpoint() {
		var st json.RawMessage
		st, err = outboundStream(o)
		if err != nil {
			return outbound{}, fmt.Errorf("outbound %q stream: %w", o.Tag, err)
		}
		ob.StreamSettings = st
	}
	return ob, nil
}

// outboundStream renders the transport/TLS for a proxy outbound, mirroring the
// inbound streamSettings shape.
func outboundStream(o domain.Outbound) (json.RawMessage, error) {
	if o.Raw != nil {
		if raw, ok := o.Raw["streamSettings"]; ok {
			return tryRaw(raw)
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
	case "httpupgrade":
		hu := map[string]any{}
		if o.Path != "" {
			hu["path"] = o.Path
		}
		if o.Host != "" {
			hu["host"] = o.Host
		}
		ss["httpupgradeSettings"] = hu
	}
	return tryRaw(ss)
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

// xrayProtocolName maps a domain.Protocol to the protocol name xray expects in
// the inbound "protocol" field. Almost all protocols use their value verbatim;
// dokodemo is the exception — its xray wire name is "dokodemo-door".
func xrayProtocolName(p domain.Protocol) string {
	if p == domain.ProtoDokodemo {
		return "dokodemo-door"
	}
	return string(p)
}

func (b Builder) buildInbound(in domain.Inbound, users []*domain.User) (inbound, error) {	// Skip reality inbounds that lack required key material — they would crash
	// xray on startup. The admin must generate keys before the inbound is usable.
	if in.Security == domain.SecurityReality {
		p := reality.ParseParams(in.Raw["reality"])
		if p.PrivateKey == "" {
			return inbound{}, fmt.Errorf("reality inbound %q skipped: private key not configured", in.Tag)
		}
	}
	settings, err := protocolSettings(in, users)
	if err != nil {
		return inbound{}, err
	}
	out := inbound{
		Tag:      in.Tag,
		Listen:   orDefault(in.Listen, "0.0.0.0"),
		Port:     in.Port,
		Protocol: xrayProtocolName(in.Protocol),
		Settings: settings,
	}
	// Protocols that carry no stream transport (socks/http utility proxies) get
	// no streamSettings block at all — their only allowed security is none, so a
	// plain proxy with no transport layer is correct.
	if !core.SkipsTransport(domain.CoreXray, in.Protocol) {
		st, err := streamSettings(in)
		if err != nil {
			return inbound{}, fmt.Errorf("inbound %s stream settings: %w", in.Tag, err)
		}
		out.StreamSettings = st
	}
	return out, nil
}

// protocolSettings renders the protocol-specific "settings" block, including the
// per-user client list keyed off the shared credentials.
func protocolSettings(in domain.Inbound, users []*domain.User) (json.RawMessage, error) {
	switch in.Protocol {
	case domain.ProtoVLESS:
		flow := effectiveFlow(in)
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			client := map[string]any{
				"id": u.Proxies.VLESSUUID.String(), "email": u.ID.String(),
			}
			// xtls-rprx-vision is only valid for VLESS over raw TCP with TLS or
			// REALITY. Emitting it on ws/grpc/http or with security=none makes the
			// core reject the config, so omit the key entirely when not applicable.
			if flow != "" {
				client["flow"] = flow
			}
			clients = append(clients, client)
		}
		settings := map[string]any{"clients": clients, "decryption": "none"}
		if fb := decoyFallbacks(in); len(fb) > 0 {
			settings["fallbacks"] = fb
		}
		raw, err := tryRaw(settings)
		if err != nil {
			return nil, err
		}
		return raw, nil

	case domain.ProtoVMess:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"id": u.Proxies.VMessUUID.String(), "email": u.ID.String()})
		}
		raw, err := tryRaw(map[string]any{"clients": clients})
		if err != nil {
			return nil, err
		}
		return raw, nil

	case domain.ProtoTrojan:
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{"password": u.Proxies.TrojanPass, "email": u.ID.String()})
		}
		settings := map[string]any{"clients": clients}
		if fb := decoyFallbacks(in); len(fb) > 0 {
			settings["fallbacks"] = fb
		}
		raw, err := tryRaw(settings)
		if err != nil {
			return nil, err
		}
		return raw, nil

	case domain.ProtoShadowsocks:
		return shadowsocksSettings(in, users)

	case domain.ProtoDokodemo:
		return dokodemoSettings(in)

	case domain.ProtoSocks:
		// SOCKS5 utility proxy. Per-user auth reuses the existing credentials
		// (username = User.Username, fallback to the user ID; password = the
		// trojan password). With >=1 user we require password auth; with none we
		// fall back to noauth so the inbound still starts.
		accounts := socksHTTPAccounts(users)
		if len(accounts) > 0 {
			raw, err := tryRaw(map[string]any{"auth": "password", "accounts": accounts, "udp": true})
			if err != nil {
				return nil, err
			}
			return raw, nil
		}
		raw, err := tryRaw(map[string]any{"auth": "noauth", "udp": true})
		if err != nil {
			return nil, err
		}
		return raw, nil

	case domain.ProtoHTTP:
		// HTTP CONNECT utility proxy. accounts reuse the same shared credentials;
		// when no users are bound we omit accounts entirely (open proxy).
		accounts := socksHTTPAccounts(users)
		settings := map[string]any{"allowTransparent": false}
		if len(accounts) > 0 {
			settings["accounts"] = accounts
		}
		raw, err := tryRaw(settings)
		if err != nil {
			return nil, err
		}
		return raw, nil

	default:
		return nil, fmt.Errorf("unsupported protocol %q for xray", in.Protocol)
	}
}

// socksHTTPAccounts renders the xray socks/http account list ({user, pass}),
// one entry per bound user, reusing the existing credentials: the username is
// User.Username (falling back to the user ID when empty) and the password is the
// shared trojan password.
func socksHTTPAccounts(users []*domain.User) []map[string]any {
	accounts := make([]map[string]any, 0, len(users))
	for _, u := range users {
		accounts = append(accounts, map[string]any{
			"user": proxyUsername(u),
			"pass": u.Proxies.TrojanPass,
		})
	}
	return accounts
}

// proxyUsername returns the per-user proxy username: User.Username when set,
// otherwise the user ID string. Shared by the socks/http utility proxies.
func proxyUsername(u *domain.User) string {
	if u.Username != "" {
		return u.Username
	}
	return u.ID.String()
}

// shadowsocksSettings renders the Shadowsocks "settings" block. xray models two
// distinct shapes (verified against 3x-ui's golden fixtures):
//
//   - Legacy ciphers (aes-*-gcm, chacha20-ietf-poly1305, ...) are single
//     credential: settings = {method, password, network}. A "clients" array is
//     NOT valid here and makes xray reject the inbound.
//   - Shadowsocks-2022 ciphers (2022-blake3-*) are multi-user: settings carries
//     a server-level PSK in "password" plus a "clients" array, one entry per
//     bound user with that user's PSK as "password" and email = user UUID.
//
// This stops the previous behaviour where only users[0] was ever rendered (a
// silent single-user drop) — with a 2022 method every bound user is now emitted.
func shadowsocksSettings(in domain.Inbound, users []*domain.User) (json.RawMessage, error) {
	method := ssInboundMethod(users)
	const network = "tcp,udp"

	if isSS2022Method(method) {
		// SS-2022 multi-user: server PSK + per-client PSK/email, matching the
		// 3x-ui shape ({method, password, network, clients:[{password,email}]}).
		clients := make([]map[string]any, 0, len(users))
		for _, u := range users {
			clients = append(clients, map[string]any{
				"password": u.Proxies.ShadowsocksP,
				"email":    u.ID.String(),
			})
		}
		return tryRaw(map[string]any{
			"method":   method,
			"password": ssServerPassword(in),
			"network":  network,
			"clients":  clients,
		})
	}

	// Legacy ciphers are single-credential by design in xray. xray/3x-ui model
	// legacy SS as one inbound-level method+password (no per-client passwords),
	// so we render the first bound user's credential as that shared password.
	// Any additional users on a legacy SS inbound necessarily share this single
	// credential — this is an xray protocol limitation, not a silent per-user
	// drop; operators who need true multi-user SS must pick a 2022 method. The
	// placeholder keeps the core booting cleanly when no users are bound yet.
	password := ""
	if len(users) > 0 {
		password = users[0].Proxies.ShadowsocksP
	}
	if password == "" {
		password = "placeholder-no-users-bound"
	}
	return tryRaw(map[string]any{"method": method, "password": password, "network": network})
}

// isSS2022Method reports whether an SS cipher is a Shadowsocks-2022 method
// (2022-blake3-aes-128-gcm / -256-gcm / chacha20-poly1305). Only these support
// the multi-user server-PSK + per-client-PSK shape.
func isSS2022Method(method string) bool {
	return strings.HasPrefix(method, "2022-")
}

// ssInboundMethod picks the SS cipher for the inbound. All users on one SS
// inbound share a single cipher, so the first user's method wins; fall back to a
// safe legacy default when unset.
func ssInboundMethod(users []*domain.User) string {
	if len(users) > 0 && users[0].Proxies.SSMethod != "" {
		return users[0].Proxies.SSMethod
	}
	return "aes-128-gcm"
}

// ssServerPassword sources the server-level PSK for an SS-2022 inbound. The
// domain model has no first-class SS server PSK field, so operators supply it
// through the Raw escape hatch (Raw["ss"].password or Raw["password"]); absent
// that we emit a placeholder so the renderer never produces empty key material.
func ssServerPassword(in domain.Inbound) string {
	if in.Raw != nil {
		if ss, ok := in.Raw["ss"].(map[string]any); ok {
			if p, ok := ss["password"].(string); ok && p != "" {
				return p
			}
		}
		if p, ok := in.Raw["password"].(string); ok && p != "" {
			return p
		}
	}
	return "placeholder-no-server-psk"
}

// effectiveFlow returns the VLESS flow to emit for an inbound, or "" when the
// flow must be omitted. The xtls-rprx-vision flow is only valid for VLESS over
// raw TCP secured by TLS or REALITY; on any other transport (ws/grpc/http) or
// with security=none it would make the core reject the whole config.
//
// For VLESS over raw TCP with REALITY where the operator left Flow blank, this
// defaults to xtls-rprx-vision (the REALITY best-practice flow). An explicitly
// set Flow is always preserved; TLS keeps the emit-only-if-explicit behaviour.
func effectiveFlow(in domain.Inbound) string {
	if in.Protocol != domain.ProtoVLESS {
		return ""
	}
	if in.Network != "tcp" && in.Network != "" {
		return ""
	}
	if in.Security != domain.SecurityTLS && in.Security != domain.SecurityReality {
		return ""
	}
	if in.Flow == "" {
		// Auto-enable vision for REALITY (best practice); for plain TLS keep the
		// historical behaviour of only emitting a flow when explicitly set.
		if in.Security == domain.SecurityReality {
			return "xtls-rprx-vision"
		}
		return ""
	}
	return in.Flow
}

// streamSettings renders transport + security. Operators can override the whole
// block via Inbound.Raw["streamSettings"] for fields the abstraction omits.
func streamSettings(in domain.Inbound) (json.RawMessage, error) {
	if raw, ok := in.Raw["streamSettings"]; ok {
		return tryRaw(raw)
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
		// Operator-supplied ALPN list (Raw["tls"].alpn) — e.g. ["h2","http/1.1"].
		if t, ok := in.Raw["tls"].(map[string]any); ok {
			if alpn := rawStringList(t["alpn"]); len(alpn) > 0 {
				tls["alpn"] = alpn
			}
		}
		ss["tlsSettings"] = tls
	case domain.SecurityReality:
		ss["realitySettings"] = realitySettings(in)
	}
	switch in.Network {
	case "tcp":
		ss["tcpSettings"] = tcpHeaderSettings(in)
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
	case "httpupgrade":
		hu := map[string]any{}
		if in.Path != "" {
			hu["path"] = in.Path
		}
		if len(in.Host) > 0 {
			hu["host"] = in.Host[0]
		}
		ss["httpupgradeSettings"] = hu
	case "http", "h2":
		ss["network"] = "http" // xray's HTTP/2 transport token
		h := map[string]any{}
		if in.Path != "" {
			h["path"] = in.Path
		}
		if len(in.Host) > 0 {
			h["host"] = in.Host
		}
		ss["httpSettings"] = h
	case "xhttp":
		x := map[string]any{"mode": "auto"}
		if xr, ok := in.Raw["xhttp"].(map[string]any); ok {
			if mode, ok := xr["mode"].(string); ok && mode != "" {
				x["mode"] = mode
			}
			// Operator escape hatch: merge arbitrary extra keys into xhttpSettings.
			if extra, ok := xr["extra"].(map[string]any); ok {
				for k, v := range extra {
					x[k] = v
				}
			}
		}
		if in.Path != "" {
			x["path"] = in.Path
		}
		if len(in.Host) > 0 {
			x["host"] = in.Host[0]
		}
		ss["xhttpSettings"] = x
	case "kcp":
		ss["kcpSettings"] = kcpSettings(in)
	}
	return tryRaw(ss)
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

// rawStringList coerces a JSON-decoded []any (or []string) of strings into a
// []string, dropping non-string entries. Used for fields like TLS ALPN that the
// abstraction does not model as typed columns.
func rawStringList(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

// tcpHeaderSettings renders the tcpSettings block for a raw-TCP inbound. By
// default it emits {"header": {"type": "none"}}; operators can override the
// header verbatim through Inbound.Raw["tcp"].header to enable HTTP camouflage
// ({"type":"http","request":{...},"response":{...}}).
func tcpHeaderSettings(in domain.Inbound) map[string]any {
	if t, ok := in.Raw["tcp"].(map[string]any); ok {
		if header, ok := t["header"].(map[string]any); ok {
			return map[string]any{"header": header}
		}
	}
	return map[string]any{"header": map[string]any{"type": "none"}}
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

// dokodemoSettings renders the xray dokodemo-door "settings" block. dokodemo is
// a transparent/redirect inbound with no per-user auth; its target and network
// come from Inbound.Raw["dokodemo"]. Every field is optional and the renderer
// picks sensible defaults so a bare dokodemo (no Raw) still produces a VALID
// xray inbound:
//
//   - When no "address" is supplied we enable followRedirect (transparent-proxy
//     mode), which xray accepts without a target address.
//   - When an "address" IS supplied we emit address+port (port default 443) and
//     do NOT force followRedirect.
//   - "network" always defaults to "tcp,udp".
//   - "timeout" / "userLevel" are passed through only when present.
func dokodemoSettings(in domain.Inbound) (json.RawMessage, error) {
	var raw map[string]any
	if in.Raw != nil {
		raw, _ = in.Raw["dokodemo"].(map[string]any)
	}
	settings := map[string]any{
		"network": orDefault(rawString(raw["network"]), "tcp,udp"),
	}
	if address := rawString(raw["address"]); address != "" {
		settings["address"] = address
		port, ok := rawInt(raw["port"])
		if !ok {
			port = 443
		}
		settings["port"] = port
		// followRedirect is honoured only if explicitly set; an address-targeted
		// dokodemo forwards to that fixed destination.
		if fr, ok := rawBool(raw["followRedirect"]); ok {
			settings["followRedirect"] = fr
		}
	} else {
		// No address: transparent mode. Default followRedirect on (xray accepts a
		// dokodemo with no address only when redirecting), but honour an explicit
		// override if the operator set one.
		fr := true
		if v, ok := rawBool(raw["followRedirect"]); ok {
			fr = v
		}
		settings["followRedirect"] = fr
	}
	if timeout, ok := rawInt(raw["timeout"]); ok {
		settings["timeout"] = timeout
	}
	if userLevel, ok := rawInt(raw["userLevel"]); ok {
		settings["userLevel"] = userLevel
	}
	return tryRaw(settings)
}

// kcpSettings renders the mKCP "kcpSettings" block (network token "kcp"). Only
// the header type and optional obfuscation seed are modelled here; richer mKCP
// knobs (mtu/tti/...) can be supplied via the Raw["streamSettings"] override,
// which short-circuits streamSettings entirely. The header type is passed
// through verbatim (none/srtp/utp/wechat-video/dtls/wireguard), defaulting to
// "none".
func kcpSettings(in domain.Inbound) map[string]any {
	var raw map[string]any
	if in.Raw != nil {
		raw, _ = in.Raw["kcp"].(map[string]any)
	}
	headerType := "none"
	if h, ok := raw["header"].(map[string]any); ok {
		if t := rawString(h["type"]); t != "" {
			headerType = t
		}
	} else if t := rawString(raw["type"]); t != "" {
		headerType = t
	}
	out := map[string]any{
		"header": map[string]any{"type": headerType},
	}
	if seed := rawString(raw["seed"]); seed != "" {
		out["seed"] = seed
	}
	return out
}

// rawString coerces a JSON-decoded value into a string, returning "" for any
// non-string (including nil). Used for optional Raw fields the abstraction does
// not model as typed columns.
func rawString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// rawInt coerces a JSON-decoded numeric value into an int. JSON numbers decode
// as float64, so that case is handled alongside the native int kinds. The bool
// return reports whether a usable number was present.
func rawInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
}

// rawBool coerces a JSON-decoded value into a bool, reporting whether a bool was
// actually present (so callers can distinguish "unset" from "false").
func rawBool(v any) (bool, bool) {
	if b, ok := v.(bool); ok {
		return b, true
	}
	return false, false
}

// asWireguardMap safely coerces Outbound.Raw["wireguard"] (an any) into a
// map[string]any for warp.ConfigFromMap, returning nil when the value is absent
// or the wrong shape (ConfigFromMap tolerates a nil map).
func asWireguardMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

func decoyFallbacks(in domain.Inbound) []map[string]any {
	dest := core.DecoyFallbackDest(in)
	if dest == "" {
		return nil
	}
	return []map[string]any{{"dest": dest, "xver": 0}}
}

func tryRaw(v any) (json.RawMessage, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("xray: marshal settings: %w", err)
	}
	return b, nil
}
