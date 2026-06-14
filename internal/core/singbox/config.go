// Package singbox implements core.CoreDriver for sing-box. Unlike Xray, sing-box
// has no live "alter inbound" gRPC for adding users, so the driver keeps the
// user set in memory and rebuilds + reloads the config on every change. This
// file renders the engine-neutral core.GeneratedConfig into sing-box JSON.
package singbox

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/reality"
	"github.com/vortexui/vortexui/internal/domain"
)

// Builder renders core.GeneratedConfig into sing-box native JSON. APIPort is the
// loopback port for sing-box's V2Ray API, which exposes per-user traffic stats.
type Builder struct {
	APIPort int
}

// Build implements core.Builder.
func (b Builder) Build(cfg *core.GeneratedConfig) ([]byte, error) {
	level := cfg.LogLevel
	if level == "" {
		level = "warn"
	}

	var inbounds []map[string]any
	emailSet := map[string]struct{}{}
	for _, in := range cfg.Inbounds {
		users := cfg.UsersByInbound[in.Tag]
		built, err := buildInbound(in, users)
		if err != nil {
			return nil, fmt.Errorf("inbound %q: %w", in.Tag, err)
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

	outbounds, outboundTags, err := buildOutbounds(cfg.Outbounds, cfg.Balancers)
	if err != nil {
		return nil, err
	}

	conf := map[string]any{
		"log":       map[string]any{"level": level},
		"inbounds":  inbounds,
		"outbounds": outbounds,
		// V2Ray API exposes per-user up/down counters; stats.users must list the
		// users to track, which is why adding a user requires a config rebuild.
		"experimental": map[string]any{
			"v2ray_api": map[string]any{
				"listen": fmt.Sprintf("127.0.0.1:%d", b.APIPort),
				"stats":  map[string]any{"enabled": true, "users": emails},
			},
		},
	}
	conf["route"] = buildRoute(cfg.Routing, outboundTags)
	return json.MarshalIndent(conf, "", "  ")
}

// buildOutbounds renders egress handlers plus the balancer outbound-groups
// (sing-box has no routing balancer; the analog is a urltest/selector group).
// It always guarantees a "direct" and "block" outbound exist so route targets
// and the implicit final route resolve. It returns the rendered outbounds and
// the set of all outbound/group tags, used to expand balancer selector prefixes
// and to choose the route's final outbound.
func buildOutbounds(outs []domain.Outbound, balancers []domain.Balancer) ([]map[string]any, []string, error) {
	var result []map[string]any
	var tags []string
	seen := map[string]bool{}
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
			return nil, nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		built, err := buildOutbound(o)
		if err != nil {
			return nil, nil, fmt.Errorf("outbound %q: %w", o.Tag, err)
		}
		add(built, o.Tag)
	}
	if !seen["direct"] {
		add(map[string]any{"type": "direct", "tag": "direct"}, "direct")
	}
	if !seen["block"] {
		add(map[string]any{"type": "block", "tag": "block"}, "block")
	}
	// Balancer groups reference concrete member tags, so build them after the
	// plain outbounds. Members are the outbounds whose tag has a selector prefix.
	memberTags := append([]string(nil), tags...)
	for i := range balancers {
		bl := balancers[i]
		if !bl.Enabled {
			continue
		}
		if err := bl.Validate(); err != nil {
			return nil, nil, fmt.Errorf("balancer %q: %w", bl.Tag, err)
		}
		members := matchByPrefix(memberTags, bl.Selectors)
		if len(members) == 0 {
			return nil, nil, fmt.Errorf("balancer %q: no outbound matches its selectors", bl.Tag)
		}
		add(buildBalancerGroup(bl, members), bl.Tag)
	}
	return result, tags, nil
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
	default:
		return nil
	}
}

// buildRoute renders the sing-box route block from neutral routing rules. Rules
// are emitted in ascending priority; "final" points at "direct" so unmatched
// traffic egresses directly (the deterministic default, independent of outbound
// order). A route block is always emitted — even with no rules — so the default
// egress never falls back to whichever outbound happens to be first.
func buildRoute(rules []domain.RoutingRule, tags []string) map[string]any {
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
		// In sing-box the balancer group is itself an outbound, so both targets
		// collapse to a single "outbound" field.
		target := r.OutboundTag
		if target == "" {
			target = r.BalancerTag
		}
		rule["outbound"] = target
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
	out := make([]map[string]any, 0, len(users))
	for _, u := range users {
		out = append(out, map[string]any{"name": u.ID.String(), "uuid": u.Proxies.VLESSUUID.String(), "flow": in.Flow})
	}
	return out
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
	if in.Security != domain.SecurityTLS && in.Security != domain.SecurityReality {
		return nil
	}
	tls := map[string]any{"enabled": true}
	if len(in.SNI) > 0 {
		tls["server_name"] = in.SNI[0]
	}
	if in.Security == domain.SecurityReality {
		tls["reality"] = realityBlock(in)
	}
	return tls
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
