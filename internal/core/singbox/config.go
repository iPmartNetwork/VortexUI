// Package singbox implements core.CoreDriver for sing-box. Unlike Xray, sing-box
// has no live "alter inbound" gRPC for adding users, so the driver keeps the
// user set in memory and rebuilds + reloads the config on every change. This
// file renders the engine-neutral core.GeneratedConfig into sing-box JSON.
package singbox

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/warp"
)

// Builder renders core.GeneratedConfig into sing-box native JSON. APIPort is the
// loopback port for sing-box's V2Ray API, which exposes per-user traffic stats.
type Builder struct {
	APIPort      int
	OmitV2RayAPI bool // skip experimental.v2ray_api (for binaries built without with_v2ray_api)
}

// Build implements core.Builder.
func (b Builder) Build(cfg *core.GeneratedConfig) ([]byte, error) {
	level := cfg.LogLevel
	if level == "" {
		level = "warn"
	}

	var inbounds []map[string]any
	var endpoints []map[string]any
	emailSet := map[string]struct{}{}
	for _, in := range cfg.Inbounds {
		users := cfg.UsersByInbound[in.Tag]
		// sing-box >= 1.11 renders WireGuard as a top-level `endpoints` entry, not
		// a regular inbound, so route it to the dedicated builder.
		if in.Protocol == domain.ProtoWireGuard {
			endpoints = append(endpoints, buildWireGuardEndpoint(in, cfg.WireGuardPeers[in.Tag]))
			for _, u := range users {
				emailSet[u.ID.String()] = struct{}{}
			}
			continue
		}
		// Skip inbounds whose security material is missing rather than emitting an
		// unusable block (e.g. reality.enabled:true with an empty private_key),
		// which would make sing-box reject the entire config. Mirrors xray.
		if !inboundUsable(in) {
			continue
		}
		built, err := buildInbound(in, users)
		if err != nil {
			// Skip misconfigured inbounds — one bad entry must not crash the core.
			continue
		}
		inbounds = append(inbounds, built)
		for _, u := range users {
			emailSet[u.ID.String()] = struct{}{}
		}
	}

	emails := make([]string, 0, len(emailSet))
	for e := range emailSet {
		emails = append(emails, e)
	}

	outbounds, outboundTags, actions, err := buildOutbounds(cfg.Outbounds, cfg.Balancers)
	if err != nil {
		return nil, err
	}

	conf := map[string]any{
		"log":       map[string]any{"level": level},
		"inbounds":  inbounds,
		"outbounds": outbounds,
	}
	if !b.OmitV2RayAPI {
		// V2Ray API exposes per-user up/down counters; stats.users lists the
		// users to track, which is why adding a user requires a config rebuild.
		// Requires a sing-box binary built with the with_v2ray_api tag.
		conf["experimental"] = map[string]any{
			"v2ray_api": map[string]any{
				"listen": fmt.Sprintf("127.0.0.1:%d", b.APIPort),
				"stats":  map[string]any{"enabled": true, "users": emails},
			},
		}
	}
	conf["route"] = buildRoute(cfg.Routing, outboundTags, actions)
	if len(endpoints) > 0 {
		conf["endpoints"] = endpoints
	}
	return json.MarshalIndent(conf, "", "  ")
}

// buildOutbounds renders egress handlers plus balancer outbound-groups. sing-box
// >= 1.12 removed the legacy `block` and `dns` outbounds, so blackhole/dns
// targets are not emitted as outbounds; instead the returned `actions` map tells
// buildRoute to translate any rule pointing at them into a modern rule action
// (`reject` / `hijack-dns`). A `direct` outbound is always guaranteed so route
// targets and the implicit final route resolve.
func buildOutbounds(outs []domain.Outbound, balancers []domain.Balancer) ([]map[string]any, []string, map[string]string, error) {
	var result []map[string]any
	var tags []string
	seen := map[string]bool{}
	// Route targets that map to a modern rule action instead of an outbound.
	actions := map[string]string{"block": "reject", "blocked": "reject"}
	add := func(m map[string]any, tag string) {
		result = append(result, m)
		tags = append(tags, tag)
		seen[tag] = true
	}
	for i := range outs {
		o := outs[i]
		if !o.Enabled {
			continue
		}
		if err := o.Validate(); err != nil {
			return nil, nil, nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		// Legacy special outbounds no longer exist in sing-box -> rule actions.
		switch o.Protocol {
		case domain.OutBlackhole:
			actions[o.Tag] = "reject"
			continue
		case domain.OutDNS:
			actions[o.Tag] = "hijack-dns"
			continue
		}
		built, err := buildOutbound(o)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		add(built, o.Tag)
	}
	if !seen["direct"] {
		add(map[string]any{"type": "direct", "tag": "direct"}, "direct")
	}
	// Balancer groups reference concrete member tags, so build them last.
	memberTags := append([]string(nil), tags...)
	for i := range balancers {
		bl := balancers[i]
		if !bl.Enabled {
			continue
		}
		if err := bl.Validate(); err != nil {
			return nil, nil, nil, fmt.Errorf("balancer %q: %w", bl.Tag, err)
		}
		members := matchByPrefix(memberTags, bl.Selectors)
		if len(members) == 0 {
			return nil, nil, nil, fmt.Errorf("balancer %q: no outbound matches its selectors", bl.Tag)
		}
		add(buildBalancerGroup(bl, members), bl.Tag)
	}
	return result, tags, actions, nil
}

// matchByPrefix returns every tag that starts with one of the given prefixes,
// preserving the tag order and de-duplicating.
func matchByPrefix(tags, prefixes []string) []string {
	var out []string
	for _, t := range tags {
		for _, p := range prefixes {
			if strings.HasPrefix(t, p) {
				out = append(out, t)
				break
			}
		}
	}
	return out
}

// buildBalancerGroup maps a domain balancer onto a sing-box outbound group:
// latency/observe strategies become a urltest group (auto lowest-latency),
// everything else a selector group (manual/default selection).
func buildBalancerGroup(bl domain.Balancer, members []string) map[string]any {
	if bl.WantsObservatory() {
		return map[string]any{
			"type":      "urltest",
			"tag":       bl.Tag,
			"outbounds": members,
			"url":       orDefault(bl.ProbeURL, "https://www.gstatic.com/generate_204"),
			"interval":  orDefault(bl.ProbeInterval, "10s"),
		}
	}
	return map[string]any{
		"type":      "selector",
		"tag":       bl.Tag,
		"outbounds": members,
		"default":   members[0],
	}
}

func buildOutbound(o domain.Outbound) (map[string]any, error) {
	m := map[string]any{"tag": o.Tag}
	switch o.Protocol {
	case domain.OutFreedom:
		m["type"] = "direct"
	case domain.OutBlackhole:
		m["type"] = "block"
	case domain.OutDNS:
		m["type"] = "dns"
	case domain.OutVLESS:
		m["type"] = "vless"
		m["server"] = o.Address
		m["server_port"] = o.Port
		m["uuid"] = o.UUID
		if o.Flow != "" {
			m["flow"] = o.Flow
		}
	case domain.OutVMess:
		m["type"] = "vmess"
		m["server"] = o.Address
		m["server_port"] = o.Port
		m["uuid"] = o.UUID
	case domain.OutTrojan:
		m["type"] = "trojan"
		m["server"] = o.Address
		m["server_port"] = o.Port
		m["password"] = o.Password
	case domain.OutShadowsocks:
		m["type"] = "shadowsocks"
		m["server"] = o.Address
		m["server_port"] = o.Port
		m["password"] = o.Password
		m["method"] = orDefault(o.Method, "aes-128-gcm")
	case domain.OutSocks:
		m["type"] = "socks"
		m["server"] = o.Address
		m["server_port"] = o.Port
		if o.Username != "" {
			m["username"] = o.Username
			m["password"] = o.Password
		}
	case domain.OutHTTP:
		m["type"] = "http"
		m["server"] = o.Address
		m["server_port"] = o.Port
		if o.Username != "" {
			m["username"] = o.Username
			m["password"] = o.Password
		}
	case domain.OutWireguard:
		// WireGuard/WARP: build from the params persisted in Raw["wireguard"].
		// SingboxOutbound returns the complete {type:"wireguard",tag,...} map
		// (peer_public_key/endpoint default to Cloudflare's WARP values when
		// unset). NOTE: sing-box's WireGuard shape is version-dependent (>=1.11
		// prefers a top-level `endpoints` entry); rendering it as a
		// `type:"wireguard"` outbound per internal/warp is acceptable for now.
		return warp.ConfigFromMap(asWireguardMap(o.Raw["wireguard"])).SingboxOutbound(o.Tag), nil
	default:
		return nil, fmt.Errorf("unsupported outbound protocol %q for sing-box", o.Protocol)
	}
	if o.Protocol.NeedsEndpoint() {
		if o.Security == domain.SecurityTLS || o.Security == domain.SecurityReality {
			tls := map[string]any{"enabled": true}
			if o.SNI != "" {
				tls["server_name"] = o.SNI
			}
			m["tls"] = tls
		}
		if tr := outboundTransport(o); tr != nil {
			m["transport"] = tr
		}
	}
	return m, nil
}

func outboundTransport(o domain.Outbound) map[string]any {
	switch o.Network {
	case "ws":
		t := map[string]any{"type": "ws"}
		if o.Path != "" {
			t["path"] = o.Path
		}
		if o.Host != "" {
			t["headers"] = map[string]any{"Host": o.Host}
		}
		return t
	case "grpc":
		return map[string]any{"type": "grpc", "service_name": o.Path}
	case "httpupgrade":
		t := map[string]any{"type": "httpupgrade"}
		if o.Path != "" {
			t["path"] = o.Path
		}
		if o.Host != "" {
			t["host"] = o.Host
		}
		return t
	case "http", "h2":
		t := map[string]any{"type": "http"}
		if o.Path != "" {
			t["path"] = o.Path
		}
		if o.Host != "" {
			t["host"] = []string{o.Host}
		}
		return t
	case "quic":
		return map[string]any{"type": "quic"}
	default:
		return nil
	}
}

// buildRoute renders the sing-box route block. Rules are emitted in ascending
// priority. A target listed in `actions` (a legacy block/dns outbound) becomes a
// modern rule action (reject/hijack-dns); every other target is a normal route
// to that outbound. `final` points at `direct` so unmatched traffic egresses
// directly regardless of outbound order.
func buildRoute(rules []domain.RoutingRule, tags []string, actions map[string]string) map[string]any {
	ordered := append([]domain.RoutingRule(nil), rules...)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Priority < ordered[j].Priority })

	routeRules := []map[string]any{}
	for _, r := range ordered {
		if !r.Enabled {
			continue
		}
		rule := map[string]any{}
		if len(r.InboundTags) > 0 {
			rule["inbound"] = r.InboundTags
		}
		if len(r.Domains) > 0 {
			rule["domain"] = r.Domains
		}
		if len(r.IP) > 0 {
			rule["ip_cidr"] = r.IP
		}
		if len(r.Protocols) > 0 {
			rule["protocol"] = r.Protocols
		}
		if n := singleNetwork(r.Network); n != "" {
			rule["network"] = n
		}
		applyPort(rule, r.Port)
		target := r.OutboundTag
		if target == "" {
			target = r.BalancerTag
		}
		switch {
		case actions[target] != "":
			rule["action"] = actions[target]
		case target != "":
			rule["outbound"] = target
		default:
			continue // no target -> skip rather than emit an invalid rule
		}
		routeRules = append(routeRules, rule)
	}
	final := "direct"
	if !containsString(tags, "direct") && len(tags) > 0 {
		final = tags[0]
	}
	route := map[string]any{"final": final}
	if len(routeRules) > 0 {
		route["rules"] = routeRules
	}
	return route
}

// singleNetwork returns a single network token sing-box accepts, or "" when the
// rule matches both tcp and udp (sing-box treats an absent network as "any").
func singleNetwork(n string) string {
	switch n {
	case "tcp", "udp":
		return n
	default:
		return ""
	}
}

// applyPort translates a neutral port spec ("443", "80,443", "1000-2000") into
// sing-box's `port` (discrete) and `port_range` ("1000:2000") fields.
func applyPort(rule map[string]any, spec string) {
	if spec == "" {
		return
	}
	var ports []int
	var ranges []string
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if lo, hi, ok := strings.Cut(part, "-"); ok {
			ranges = append(ranges, strings.TrimSpace(lo)+":"+strings.TrimSpace(hi))
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			ports = append(ports, n)
		}
	}
	if len(ports) > 0 {
		rule["port"] = ports
	}
	if len(ranges) > 0 {
		rule["port_range"] = ranges
	}
}

func containsString(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

func buildInbound(in domain.Inbound, users []*domain.User) (map[string]any, error) {
	m := map[string]any{
		"tag":         in.Tag,
		"listen":      orDefault(in.Listen, "::"),
		"listen_port": in.Port,
	}

	switch in.Protocol {
	case domain.ProtoVLESS:
		m["type"] = "vless"
		m["users"] = vlessUsers(in, users)
	case domain.ProtoVMess:
		m["type"] = "vmess"
		m["users"] = vmessUsers(users)
	case domain.ProtoTrojan:
		m["type"] = "trojan"
		m["users"] = trojanUsers(users)
	case domain.ProtoShadowsocks:
		m["type"] = "shadowsocks"
		m["method"] = ssMethod(users)
		m["users"] = ssUsers(users)
	case domain.ProtoHysteria2:
		// hysteria2: QUIC-native, mandatory TLS. Bandwidth/obfs/masquerade are
		// inbound-level knobs sourced from Raw["hysteria2"] (no migration).
		m["type"] = "hysteria2"
		applyHysteria2Tuning(m, in)
		m["users"] = hysteria2Users(users)
	case domain.ProtoTUIC:
		// tuic: QUIC-native, mandatory TLS. congestion_control defaults to bbr;
		// zero_rtt_handshake/auth_timeout are optional knobs from Raw["tuic"].
		m["type"] = "tuic"
		applyTUICTuning(m, in)
		m["users"] = tuicUsers(users)
	case domain.ProtoHysteria:
		// Hysteria v1: UDP-native, mandatory TLS. Bandwidth/obfs come from Raw
		// since they are inbound-level knobs we do not model as columns.
		m["type"] = "hysteria"
		up, down := hysteriaBandwidth(in)
		m["up_mbps"] = up
		m["down_mbps"] = down
		if obfs := hysteriaObfs(in); obfs != "" {
			m["obfs"] = obfs
		}
		m["users"] = hysteriaUsers(users)
	case domain.ProtoShadowTLS:
		// ShadowTLS v3 fronts a real TLS handshake; it carries no tls block of
		// its own. Version/handshake/strict_mode come from Raw["shadowtls"].
		m["type"] = "shadowtls"
		m["version"] = shadowTLSVersion(in)
		if hs := shadowTLSHandshake(in); hs != nil {
			m["handshake"] = hs
		}
		m["users"] = shadowTLSUsers(in, users)
		m["strict_mode"] = shadowTLSStrict(in)
	case domain.ProtoAnyTLS:
		// AnyTLS: TCP-based, mandatory TLS (rendered below via tlsBlock).
		m["type"] = "anytls"
		m["users"] = anyTLSUsers(users)
		if pad := anyTLSPadding(in); len(pad) > 0 {
			m["padding_scheme"] = pad
		}
	case domain.ProtoSocks:
		// SOCKS5 utility proxy. Plain (security none), no stream transport. An
		// empty users list means no auth in sing-box.
		m["type"] = "socks"
		m["users"] = usernamePasswordUsers(users)
	case domain.ProtoHTTP:
		// HTTP CONNECT utility proxy. Plain (security none), no stream transport.
		m["type"] = "http"
		m["users"] = usernamePasswordUsers(users)
	case domain.ProtoNaive:
		// NaiveProxy: mandates TLS (the tls block is added by tlsBlock below when
		// security==tls); carries no stream transport.
		m["type"] = "naive"
		m["users"] = usernamePasswordUsers(users)
	default:
		return nil, fmt.Errorf("unsupported protocol %q for sing-box", in.Protocol)
	}

	if tls := tlsBlock(in); tls != nil {
		m["tls"] = tls
	}
	if tr := transportBlock(in); tr != nil {
		m["transport"] = tr
	}
	if in.Raw != nil {
		if extra, ok := in.Raw["singbox"].(map[string]any); ok {
			for k, v := range extra {
				m[k] = v // operator override/escape hatch
			}
		}
	}
	return m, nil
}

func vlessUsers(in domain.Inbound, users []*domain.User) []map[string]any {
	flow := effectiveFlow(in)
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		user := map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VLESSUUID.String()}
		// xtls-rprx-vision is only valid for VLESS over raw TCP with TLS or
		// REALITY; omit the key entirely otherwise or sing-box rejects the config.
		if flow != "" {
			user["flow"] = flow
		}
		out = append(out, user)
	}
	return out
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

func vmessUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VMessUUID.String(), "alterId": 0})
	}
	return out
}

func trojanUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.TrojanPass})
	}
	return out
}

// applyHysteria2Tuning sets the optional hysteria2 inbound knobs from
// Raw["hysteria2"], matching sing-box's native JSON shape:
//   - up_mbps / down_mbps (ints) — congestion-control bandwidth hints
//   - obfs {type:"salamander", password:<pw>} — Salamander obfuscation; the Raw
//     value may be the password string directly or an object carrying "password"
//   - masquerade (string) — HTTP masquerade target
//
// Each field is emitted only when present so a bare hysteria2 inbound stays
// minimal (just users + the mandatory TLS block added later).
func applyHysteria2Tuning(m map[string]any, in domain.Inbound) {
	h, _ := in.Raw["hysteria2"].(map[string]any)
	if h == nil {
		return
	}
	if v, ok := rawInt(h["up_mbps"]); ok {
		m["up_mbps"] = v
	}
	if v, ok := rawInt(h["down_mbps"]); ok {
		m["down_mbps"] = v
	}
	if pw := hysteria2ObfsPassword(h["obfs"]); pw != "" {
		m["obfs"] = map[string]any{"type": "salamander", "password": pw}
	}
	if s, ok := rawString(h["masquerade"]); ok && s != "" {
		m["masquerade"] = s
	}
}

// hysteria2ObfsPassword extracts the Salamander obfs password from a Raw value
// that may be either the password string directly or an object {password:<pw>}.
func hysteria2ObfsPassword(v any) string {
	switch o := v.(type) {
	case string:
		return o
	case map[string]any:
		s, _ := o["password"].(string)
		return s
	default:
		return ""
	}
}

// applyTUICTuning sets the tuic inbound knobs from Raw["tuic"]:
//   - congestion_control (default "bbr")
//   - zero_rtt_handshake (bool) — emitted only when explicitly set
//   - auth_timeout (string) — emitted only when a non-empty string
func applyTUICTuning(m map[string]any, in domain.Inbound) {
	t, _ := in.Raw["tuic"].(map[string]any)
	cc := "bbr"
	if s, ok := rawString(t["congestion_control"]); ok && s != "" {
		cc = s
	}
	m["congestion_control"] = cc
	if b, ok := rawBool(t["zero_rtt_handshake"]); ok {
		m["zero_rtt_handshake"] = b
	}
	if s, ok := rawString(t["auth_timeout"]); ok && s != "" {
		m["auth_timeout"] = s
	}
}

// hysteria2Users renders the per-user password list. Reuses the trojan password.
func hysteria2Users(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.TrojanPass})
	}
	return out
}

// tuicUsers renders per-user uuid+password. Reuses the vless uuid + trojan pass.
func tuicUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VLESSUUID.String(), "password": u.Proxies.TrojanPass})
	}
	return out
}

// hysteriaUsers renders the per-user list for Hysteria v1. sing-box expects each
// user's authentication string in `auth_str`; we reuse the trojan password as
// the shared per-user secret (same approach as hysteria2/tuic).
func hysteriaUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "auth_str": u.Proxies.TrojanPass})
	}
	return out
}

// hysteriaBandwidth reads the inbound up/down rate limits (Mbps) from
// Raw["hysteria"], falling back to a sane non-zero default so the inbound is
// valid even when the operator left them unset.
func hysteriaBandwidth(in domain.Inbound) (up, down int) {
	up, down = 100, 100
	h, _ := in.Raw["hysteria"].(map[string]any)
	if v, ok := rawInt(h["up_mbps"]); ok {
		up = v
	}
	if v, ok := rawInt(h["down_mbps"]); ok {
		down = v
	}
	return up, down
}

// hysteriaObfs returns the optional obfuscation password from Raw["hysteria"].
func hysteriaObfs(in domain.Inbound) string {
	h, _ := in.Raw["hysteria"].(map[string]any)
	s, _ := h["obfs"].(string)
	return s
}

// usernamePasswordUsers renders the {username, password} user list shared by the
// socks/http/naive utility proxies. The username reuses the existing credential
// (User.Username, falling back to the user ID when empty) and the password is
// the shared trojan password. sing-box accepts an empty list (no auth).
func usernamePasswordUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{
			"username": proxyUsername(u),
			"password": u.Proxies.TrojanPass,
		})
	}
	return out
}

// proxyUsername returns the per-user proxy username: User.Username when set,
// otherwise the user ID string.
func proxyUsername(u *domain.User) string {
	if u.Username != "" {
		return u.Username
	}
	return u.ID.String()
}

// anyTLSUsers renders the per-user password list for AnyTLS, reusing the trojan
// password as the per-user secret.
func anyTLSUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.TrojanPass})
	}
	return out
}

// anyTLSPadding returns the optional padding_scheme lines from Raw["anytls"].
func anyTLSPadding(in domain.Inbound) []string {
	a, _ := in.Raw["anytls"].(map[string]any)
	return rawStrings(a["padding_scheme"])
}

// shadowTLSUsers renders the ShadowTLS v3 user list (name+password), reusing the
// trojan password. When no users are bound it falls back to the single
// server-level password from Raw["shadowtls"] so a detour-only setup still works.
func shadowTLSUsers(in domain.Inbound, users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.TrojanPass})
	}
	if len(out) == 0 {
		s, _ := in.Raw["shadowtls"].(map[string]any)
		if pw, _ := s["password"].(string); pw != "" {
			out = append(out, map[string]any{"password": pw})
		}
	}
	return out
}

// shadowTLSVersion returns the protocol version from Raw["shadowtls"], defaulting
// to 3 (the only version with per-user authentication).
func shadowTLSVersion(in domain.Inbound) int {
	s, _ := in.Raw["shadowtls"].(map[string]any)
	if v, ok := rawInt(s["version"]); ok && v > 0 {
		return v
	}
	return 3
}

// shadowTLSStrict returns the strict_mode flag from Raw["shadowtls"].
func shadowTLSStrict(in domain.Inbound) bool {
	s, _ := in.Raw["shadowtls"].(map[string]any)
	b, _ := s["strict_mode"].(bool)
	return b
}

// shadowTLSHandshake builds the {server, server_port} handshake target ShadowTLS
// fronts. Returns nil when no handshake server is configured (the inbound is
// then skipped by inboundUsable). The server comes from Raw["shadowtls"], with
// the inbound SNI as a fallback; the port defaults to 443.
func shadowTLSHandshake(in domain.Inbound) map[string]any {
	server := shadowTLSHandshakeServer(in)
	if server == "" {
		return nil
	}
	port := 443
	s, _ := in.Raw["shadowtls"].(map[string]any)
	if v, ok := rawInt(s["handshake_port"]); ok && v > 0 {
		port = v
	}
	return map[string]any{"server": server, "server_port": port}
}

// shadowTLSHandshakeServer resolves the handshake server name, preferring the
// explicit Raw value and falling back to the inbound's first SNI entry.
func shadowTLSHandshakeServer(in domain.Inbound) string {
	s, _ := in.Raw["shadowtls"].(map[string]any)
	if server, _ := s["handshake_server"].(string); server != "" {
		return server
	}
	if len(in.SNI) > 0 {
		return in.SNI[0]
	}
	return ""
}

// rawInt coerces a JSON-decoded number (float64) or a native int into an int.
func rawInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

// rawString coerces a value into a string. Returns ok=false when not a string.
func rawString(v any) (string, bool) {
	s, ok := v.(string)
	return s, ok
}

// rawBool coerces a value into a bool. Returns ok=false when not a bool.
func rawBool(v any) (bool, bool) {
	b, ok := v.(bool)
	return b, ok
}

// asWireguardMap safely coerces Outbound.Raw["wireguard"] (an any) into a
// map[string]any for warp.ConfigFromMap, returning nil when the value is absent
// or the wrong shape (ConfigFromMap tolerates a nil map).
func asWireguardMap(v any) map[string]any {
	m, _ := v.(map[string]any)
	return m
}

// rawStrings coerces a JSON-decoded []any (or []string) of strings into []string.
func rawStrings(v any) []string {
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

func ssUsers(users []*domain.User) []map[string]any {
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "password": u.Proxies.ShadowsocksP})
	}
	return out
}

// ssMethod takes the cipher from the first user; all users on an SS inbound share
// one method in sing-box.
func ssMethod(users []*domain.User) string {
	if len(users) > 0 && users[0].Proxies.SSMethod != "" {
		return users[0].Proxies.SSMethod
	}
	return "aes-128-gcm"
}

func tlsBlock(in domain.Inbound) map[string]any {
	// ShadowTLS supplies its own TLS handshake via the handshake target; it must
	// not carry a separate inbound tls block.
	if in.Protocol == domain.ProtoShadowTLS {
		return nil
	}
	if in.Security != domain.SecurityTLS && in.Security != domain.SecurityReality {
		return nil
	}
	tls := map[string]any{"enabled": true}
	serverName := ""
	if len(in.SNI) > 0 {
		serverName = in.SNI[0]
	}
	if in.Security == domain.SecurityReality {
		// sing-box REQUIRES tls.server_name for REALITY. The inbound SNI field is
		// often left empty (provisionSecurity only fills Raw["reality"].server_names),
		// so fall back to the reality profile's server name (xray parity) — otherwise
		// the core rejects the config and never starts.
		rp := reality.ParseParams(in.Raw["reality"])
		if serverName == "" && len(rp.ServerNames) > 0 {
			serverName = rp.ServerNames[0]
		}
		tls["reality"] = realityBlock(in)
	} else if cert, key := tlsCertLines(in.Raw["tls"]); cert != nil {
		// Inline the auto-generated (or operator-supplied) certificate so TLS
		// inbounds — including the mandatory-TLS hysteria2/tuic — actually start.
		tls["certificate"] = cert
		tls["key"] = key
	}
	// ALPN applies to plain TLS only (not REALITY, which negotiates its own).
	// Operators supply it through Raw["tls"].alpn, e.g. ["h2","http/1.1"].
	if in.Security == domain.SecurityTLS {
		if t, ok := in.Raw["tls"].(map[string]any); ok {
			if alpn := rawStrings(t["alpn"]); len(alpn) > 0 {
				tls["alpn"] = alpn
			}
		}
	}
	if serverName != "" {
		tls["server_name"] = serverName
	}
	return tls
}

// tlsCertLines splits the PEM strings in Inbound.Raw["tls"] ({certificate, key})
// into the line arrays sing-box expects. Returns nil when absent.
func tlsCertLines(v any) (cert, key []string) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, nil
	}
	c, _ := m["certificate"].(string)
	k, _ := m["key"].(string)
	if c == "" || k == "" {
		return nil, nil
	}
	return strings.Split(strings.TrimRight(c, "\n"), "\n"), strings.Split(strings.TrimRight(k, "\n"), "\n")
}

// inboundUsable reports whether an inbound has the security material sing-box
// needs to accept it. A REALITY inbound without a private key would render
// reality.enabled:true with an empty private_key, which makes sing-box reject
// the whole config — so the builder skips that one inbound instead (xray
// parity, see xray.inboundUsable). A full operator override is trusted as-is.
func inboundUsable(in domain.Inbound) bool {
	if in.Raw != nil {
		if _, ok := in.Raw["singbox"].(map[string]any); ok {
			return true
		}
	}
	if in.Security == domain.SecurityReality && reality.ParseParams(in.Raw["reality"]).PrivateKey == "" {
		return false
	}
	switch in.Protocol {
	case domain.ProtoShadowTLS:
		// ShadowTLS cannot start without a real TLS handshake target to front.
		return shadowTLSHandshake(in) != nil
	case domain.ProtoAnyTLS, domain.ProtoHysteria:
		// Both mandate a TLS layer; without one sing-box rejects the inbound, so
		// skip it rather than emit a broken block.
		return in.Security == domain.SecurityTLS || in.Security == domain.SecurityReality
	case domain.ProtoNaive:
		// NaiveProxy mandates TLS; without it sing-box rejects the inbound.
		return in.Security == domain.SecurityTLS
	}
	return true
}

// realityBlock renders the sing-box server REALITY config from the engine-neutral
// params in Inbound.Raw["reality"], translating the neutral keys to sing-box's
// shape (private_key, short_id, handshake.server/server_port).
func realityBlock(in domain.Inbound) map[string]any {
	p := reality.ParseParams(in.Raw["reality"])
	r := map[string]any{"enabled": true}
	if p.PrivateKey != "" {
		r["private_key"] = p.PrivateKey
	}
	if len(p.ShortIDs) > 0 {
		r["short_id"] = p.ShortIDs
	}
	dest := p.Dest
	if dest == "" && len(in.SNI) > 0 {
		dest = in.SNI[0] + ":443"
	}
	if host, port, ok := splitDest(dest); ok {
		r["handshake"] = map[string]any{"server": host, "server_port": port}
	}
	return r
}

// splitDest parses a "host:port" handshake target into its parts.
func splitDest(dest string) (string, int, bool) {
	host, portStr, ok := strings.Cut(dest, ":")
	if !ok || host == "" {
		return "", 0, false
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, false
	}
	return host, port, true
}

func transportBlock(in domain.Inbound) map[string]any {
	switch in.Network {
	case "ws":
		t := map[string]any{"type": "ws"}
		if in.Path != "" {
			t["path"] = in.Path
		}
		if len(in.Host) > 0 {
			t["headers"] = map[string]any{"Host": in.Host[0]}
		}
		return t
	case "grpc":
		return map[string]any{"type": "grpc", "service_name": in.Path}
	case "httpupgrade":
		t := map[string]any{"type": "httpupgrade"}
		if in.Path != "" {
			t["path"] = in.Path
		}
		if len(in.Host) > 0 {
			t["host"] = in.Host[0]
		}
		return t
	case "http", "h2":
		t := map[string]any{"type": "http"}
		if in.Path != "" {
			t["path"] = in.Path
		}
		if len(in.Host) > 0 {
			t["host"] = in.Host
		}
		return t
	case "quic":
		return map[string]any{"type": "quic"}
	default:
		return nil
	}
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

// buildWireGuardEndpoint renders a WireGuard server inbound as a sing-box
// top-level `endpoints` entry (sing-box >= 1.11). Each bound user is a peer with
// its assigned /32 tunnel IP; egress is handled by the route's `direct` final.
func buildWireGuardEndpoint(in domain.Inbound, peers []domain.WireGuardPeer) map[string]any {
	wg, _ := in.Raw["wireguard"].(map[string]any)
	privateKey, _ := wg["private_key"].(string)
	subnet, _ := wg["subnet"].(string)
	if subnet == "" {
		subnet = "10.7.0.0/24"
	}
	// server address = .1 of the subnet
	serverAddr := serverAddress(subnet)
	listenPort := in.Port
	if lp, ok := wg["listen_port"].(int); ok && lp != 0 {
		listenPort = lp
	} else if lpf, ok := wg["listen_port"].(float64); ok && lpf != 0 {
		listenPort = int(lpf)
	}
	ep := map[string]any{
		"type":        "wireguard",
		"tag":         in.Tag,
		"system":      false,
		"address":     []string{serverAddr},
		"private_key": privateKey,
		"listen_port": listenPort,
	}
	var ps []map[string]any
	for _, p := range peers {
		ps = append(ps, map[string]any{
			"public_key":  p.PublicKey,
			"allowed_ips": []string{p.Address + "/32"},
		})
	}
	if len(ps) > 0 {
		ep["peers"] = ps
	}
	return ep
}

// serverAddress returns the server's tunnel address (".1/24") for the subnet.
func serverAddress(subnet string) string {
	ip, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "10.7.0.1/24"
	}
	base := ip.Mask(ipnet.Mask).To4()
	if base == nil {
		return "10.7.0.1/24"
	}
	base[3] = 1
	ones, _ := ipnet.Mask.Size()
	return fmt.Sprintf("%s/%d", base.String(), ones)
}
