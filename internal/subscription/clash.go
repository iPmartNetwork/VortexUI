package subscription

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderClash builds a minimal but valid Clash.Meta config: every proxy, a
// single selector group containing them all, and a rules section. When rules is
// empty the rules section is exactly the historical catch-all (`MATCH,<title>`)
// so output stays byte-identical; when a pack supplies rules they are translated
// to Clash rule strings followed by a MATCH fallback to the selector group.
func renderClash(proxies []Proxy, title string, rules []domain.RoutingRule, groups []ProtocolGroupRender) ([]byte, error) {
	var clashProxies []map[string]any
	var names []string
	for _, p := range proxies {
		m := clashProxy(p)
		if m == nil {
			continue // protocol Clash can't express; skip rather than emit garbage
		}
		clashProxies = append(clashProxies, m)
		names = append(names, p.Name)
	}
	if title == "" {
		title = "VortexUI"
	}

	clashRuleStrings := []string{"MATCH," + title}
	if len(rules) > 0 {
		clashRuleStrings = append(clashRules(rules, title), "MATCH,"+title)
	}

	// Per-group url-test proxy-groups for auto-protocol switching. Each group
	// gets its own url-test block with probe settings from the ProtocolGroup.
	var groupProxyGroups []map[string]any
	var groupNames []string
	for _, g := range groups {
		if len(g.ProxyNames) == 0 {
			continue
		}
		interval := 90
		if g.ProbeInterval > 0 {
			interval = g.ProbeInterval
		}
		probeURL := g.ProbeURL
		if probeURL == "" {
			probeURL = "https://www.gstatic.com/generate_204"
		}
		groupName := fmt.Sprintf("🔄 %s", g.Name)
		groupProxyGroups = append(groupProxyGroups, map[string]any{
			"name":      groupName,
			"type":      "url-test",
			"proxies":   g.ProxyNames,
			"url":       probeURL,
			"interval":  interval,
			"tolerance": 150,
		})
		groupNames = append(groupNames, groupName)
	}

	// Build the selector's proxy list: groups first, then global Auto/Fallback,
	// then DIRECT, then individual proxies.
	selectorProxies := make([]string, 0, len(groupNames)+3+len(names))
	selectorProxies = append(selectorProxies, groupNames...)
	selectorProxies = append(selectorProxies, "♻️ Auto", "♻️ Fallback", "DIRECT")
	selectorProxies = append(selectorProxies, names...)

	proxyGroups := []map[string]any{
		{
			"name":    title,
			"type":    "select",
			"proxies": selectorProxies,
		},
		{
			"name":      "♻️ Auto",
			"type":      "url-test",
			"proxies":   names,
			"url":       "https://www.gstatic.com/generate_204",
			"interval":  90,
			"tolerance": 150,
		},
		{
			"name":      "♻️ Fallback",
			"type":      "fallback",
			"proxies":   names,
			"url":       "https://www.gstatic.com/generate_204",
			"interval":  90,
		},
	}
	proxyGroups = append(proxyGroups, groupProxyGroups...)

	cfg := map[string]any{
		"proxies":      clashProxies,
		"proxy-groups": proxyGroups,
		"rules":        clashRuleStrings,
	}
	return yaml.Marshal(cfg)
}

func clashProxy(p Proxy) map[string]any {
	base := map[string]any{
		"name":   p.Name,
		"server": p.Host,
		"port":   p.Port,
		"udp":    true,
	}
	// Port hopping: Clash Meta supports "ports" field for port-range rotation.
	if p.PortEnd > 0 && p.PortEnd > p.Port {
		base["ports"] = fmt.Sprintf("%d-%d", p.Port, p.PortEnd)
		if p.HopInterval > 0 {
			base["hop-interval"] = p.HopInterval
		}
	}
	tls := p.Security == "tls" || p.Security == "reality"

	switch p.Protocol {
	case domain.ProtoVMess:
		base["type"] = "vmess"
		base["uuid"] = p.UUID
		base["alterId"] = 0
		base["cipher"] = "auto"
		base["tls"] = tls
		applyNetwork(base, p)
	case domain.ProtoVLESS:
		base["type"] = "vless"
		base["uuid"] = p.UUID
		base["tls"] = tls
		if p.Flow != "" {
			base["flow"] = p.Flow
		}
		applyNetwork(base, p)
	case domain.ProtoTrojan:
		base["type"] = "trojan"
		base["password"] = p.Password
		applyNetwork(base, p)
	case domain.ProtoShadowsocks:
		base["type"] = "ss"
		base["cipher"] = p.SSMethod
		base["password"] = p.Password
	default:
		return nil
	}
	if tls && p.SNI != "" {
		base["servername"] = p.SNI
		base["sni"] = p.SNI
	}
	if p.AllowInsecure {
		base["skip-cert-verify"] = true
	}
	// uTLS fingerprint for all TLS proxies (not just REALITY) to avoid Go-TLS detection.
	if tls && p.Security != "reality" {
		base["client-fingerprint"] = orDefault(p.Fingerprint, "chrome")
	}
	if p.Security == "reality" {
		base["reality-opts"] = map[string]any{"public-key": p.PublicKey, "short-id": p.ShortID}
		base["client-fingerprint"] = orDefault(p.Fingerprint, "chrome")
	}
	// Additive host overrides: only set when present so existing proxies render
	// identically.
	if len(p.ALPN) > 0 {
		base["alpn"] = p.ALPN
	}
	if p.Mux {
		base["smux"] = map[string]any{"enabled": true}
	}
	// TLS fragment: Clash Meta supports the `tls-fragment` option for anti-DPI.
	if p.Fragment != "" && tls {
		parts := strings.Split(p.Fragment, ",")
		if len(parts) >= 2 {
			base["tls-fragment"] = map[string]any{
				"enabled": true,
				"size":    parts[0],
				"sleep":   parts[1],
			}
		}
	}
	// Connection resilience: TCP Fast Open reduces handshake latency.
	base["tfo"] = true
	return base
}

func applyNetwork(base map[string]any, p Proxy) {
	switch p.Network {
	case "ws":
		base["network"] = "ws"
		opts := map[string]any{}
		if p.Path != "" {
			opts["path"] = p.Path
		}
		if p.HostHeader != "" {
			opts["headers"] = map[string]any{"Host": p.HostHeader}
		}
		base["ws-opts"] = opts
	case "grpc":
		base["network"] = "grpc"
		base["grpc-opts"] = map[string]any{"grpc-service-name": p.Path}
	case "httpupgrade":
		// mihomo expresses HTTPUpgrade as a ws transport with v2ray-http-upgrade.
		base["network"] = "ws"
		opts := map[string]any{"v2ray-http-upgrade": true}
		if p.Path != "" {
			opts["path"] = p.Path
		}
		if p.HostHeader != "" {
			opts["headers"] = map[string]any{"Host": p.HostHeader}
		}
		base["ws-opts"] = opts
	case "http", "h2":
		base["network"] = "h2"
		opts := map[string]any{}
		if p.Path != "" {
			opts["path"] = p.Path
		}
		if p.HostHeader != "" {
			opts["host"] = []string{p.HostHeader}
		}
		base["h2-opts"] = opts
	case "xhttp":
		// XHTTP/SplitHTTP: Clash Meta expresses this as an h2 transport with
		// custom path. Some mihomo forks support xhttp natively.
		base["network"] = "h2"
		opts := map[string]any{}
		if p.Path != "" {
			opts["path"] = p.Path
		}
		if p.HostHeader != "" {
			opts["host"] = []string{p.HostHeader}
		}
		base["h2-opts"] = opts
	case "quic":
		base["network"] = "quic"
		base["quic-opts"] = map[string]any{"security": "none", "header": map[string]any{"type": "none"}}
	}
}
