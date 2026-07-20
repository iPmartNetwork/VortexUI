package subscription

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderSingbox builds a sing-box config fragment: one outbound per proxy plus a
// selector that fronts them. Clients merge this with their own inbounds/route.
// When rules is empty no route section is emitted, so output stays byte-identical
// to before; when a pack supplies rules they are translated into route.rules with
// a route.final pointing at the selector, and a block outbound is appended only
// if a rule rejects traffic.
func renderSingbox(proxies []Proxy, title string, rules []domain.RoutingRule, groups []ProtocolGroupRender) ([]byte, error) {
	var outbounds []map[string]any
	var tags []string
	for _, p := range proxies {
		o := singboxOutbound(p)
		if o == nil {
			continue
		}
		outbounds = append(outbounds, o)
		tags = append(tags, p.Name)
	}
	if title == "" {
		title = "VortexUI"
	}

	selector := map[string]any{
		"type":      "selector",
		"tag":       title,
		"outbounds": append([]string{"♻️ Auto", "♻️ Fallback"}, tags...),
	}
	autoTest := map[string]any{
		"type":      "urltest",
		"tag":       "♻️ Auto",
		"outbounds": tags,
		"url":       "https://www.gstatic.com/generate_204",
		"interval":  "90s",
		"tolerance": 150,
	}
	fallback := map[string]any{
		"type":      "urltest",
		"tag":       "♻️ Fallback",
		"outbounds": tags,
		"url":       "https://www.gstatic.com/generate_204",
		"interval":  "90s",
		"tolerance": 0,
	}

	// Per-group urltest outbounds for auto-protocol switching. Each group gets
	// its own urltest block with probe settings from the ProtocolGroup definition.
	// Group tags are prepended to the selector so clients see them as top-level
	// switching options alongside the global Auto/Fallback.
	var groupOutbounds []map[string]any
	var groupTags []string
	for _, g := range groups {
		if len(g.ProxyNames) == 0 {
			continue
		}
		interval := "90s"
		if g.ProbeInterval > 0 {
			interval = fmt.Sprintf("%ds", g.ProbeInterval)
		}
		probeURL := g.ProbeURL
		if probeURL == "" {
			probeURL = "https://www.gstatic.com/generate_204"
		}
		groupTag := "🔄 " + g.Name
		groupOutbounds = append(groupOutbounds, map[string]any{
			"type":      "urltest",
			"tag":       groupTag,
			"outbounds": g.ProxyNames,
			"url":       probeURL,
			"interval":  interval,
			"tolerance": 150,
		})
		groupTags = append(groupTags, groupTag)
	}

	// Rebuild selector outbounds: groups first, then global Auto/Fallback, then
	// individual proxies.
	if len(groupTags) > 0 {
		selectorOutbounds := make([]string, 0, len(groupTags)+2+len(tags))
		selectorOutbounds = append(selectorOutbounds, groupTags...)
		selectorOutbounds = append(selectorOutbounds, "♻️ Auto", "♻️ Fallback")
		selectorOutbounds = append(selectorOutbounds, tags...)
		selector["outbounds"] = selectorOutbounds
	}

	direct := map[string]any{"type": "direct", "tag": "direct"}
	all := append([]map[string]any{selector, autoTest, fallback}, groupOutbounds...)
	all = append(all, outbounds...)
	all = append(all, direct)

	cfg := map[string]any{"outbounds": all}

	// DNS-over-HTTPS: when no routing rules are provided (simple subscription),
	// inject a DoH DNS section so clients don't leak DNS queries to ISP resolvers.
	if len(rules) == 0 {
		cfg["dns"] = map[string]any{
			"servers": []map[string]any{
				{
					"tag":     "dns-remote",
					"address": "https://1.1.1.1/dns-query",
					"detour":  title,
				},
				{
					"tag":     "dns-direct",
					"address": "https://8.8.8.8/dns-query",
					"detour":  "direct",
				},
			},
			"rules": []map[string]any{
				{"outbound": []string{"any"}, "server": "dns-remote"},
			},
		}
	}

	if len(rules) > 0 {
		routeRules := singboxRules(rules, title)
		if singboxNeedsBlock(rules) {
			all = append(all, map[string]any{"type": "block", "tag": "block"})
			cfg["outbounds"] = all
		}
		cfg["route"] = map[string]any{
			"rules": routeRules,
			"final": title,
		}
	}

	return json.MarshalIndent(cfg, "", "  ")
}

func singboxOutbound(p Proxy) map[string]any {
	o := map[string]any{
		"tag":         p.Name,
		"server":      p.Host,
		"server_port": p.Port,
	}
	// Port hopping: when a port range is defined, emit the hop fields so
	// sing-box clients can rotate ports within the range.
	if p.PortEnd > 0 && p.PortEnd > p.Port {
		o["server_port_range"] = fmt.Sprintf("%d:%d", p.Port, p.PortEnd)
		if p.HopInterval > 0 {
			o["hop_interval"] = fmt.Sprintf("%ds", p.HopInterval)
		}
	}
	o["dial_timeout"] = "5s"
	o["tcp_fast_open"] = true
	o["tcp_keep_alive_interval"] = "30s"
	o["domain_strategy"] = "prefer_ipv4"
	if p.Security == "tls" || p.Security == "reality" {
		tls := map[string]any{"enabled": true}
		if p.SNI != "" {
			tls["server_name"] = p.SNI
		}
		if p.AllowInsecure {
			tls["insecure"] = true
		}
		// Additive: only emit ALPN when the host override supplies it.
		if len(p.ALPN) > 0 {
			tls["alpn"] = p.ALPN
		}
		if p.Security == "reality" {
			tls["reality"] = map[string]any{"enabled": true, "public_key": p.PublicKey, "short_id": p.ShortID}
			tls["utls"] = map[string]any{"enabled": true, "fingerprint": orDefault(p.Fingerprint, "chrome")}
		} else {
			// uTLS for regular TLS too — prevents Go-TLS fingerprint detection.
			tls["utls"] = map[string]any{"enabled": true, "fingerprint": orDefault(p.Fingerprint, "chrome")}
		}
		// TLS fragment for anti-DPI bypass.
		if p.Fragment != "" {
			parts := strings.Split(p.Fragment, ",")
			if len(parts) >= 2 {
				tls["fragment"] = map[string]any{
					"enabled": true,
					"size":    parts[0],
					"sleep":   parts[1],
				}
			}
		}
		// ECH (Encrypted Client Hello) for hiding SNI from DPI.
		if p.ECH {
			tls["ech"] = map[string]any{
				"enabled": true,
			}
		}
		// Random padding to defeat length-based DPI fingerprinting.
		if p.Padding != "" {
			tls["padding_size"] = p.Padding
		}
		o["tls"] = tls
	}
	if tr := singboxTransport(p); tr != nil {
		o["transport"] = tr
	}
	// Additive: client-side multiplexing only when the host enables it.
	if p.Mux {
		o["multiplex"] = map[string]any{"enabled": true, "protocol": "smux", "idle_timeout": "30s"}
	}

	switch p.Protocol {
	case domain.ProtoVMess:
		o["type"] = "vmess"
		o["uuid"] = p.UUID
		o["security"] = "auto"
	case domain.ProtoVLESS:
		o["type"] = "vless"
		o["uuid"] = p.UUID
		if p.Flow != "" {
			o["flow"] = p.Flow
		}
	case domain.ProtoTrojan:
		o["type"] = "trojan"
		o["password"] = p.Password
	case domain.ProtoShadowsocks:
		o["type"] = "shadowsocks"
		o["method"] = p.SSMethod
		o["password"] = p.Password
	case domain.ProtoHysteria2:
		o["type"] = "hysteria2"
		o["password"] = p.Password
		// Hysteria2-specific fields from Hy2Opts
		if p.Hy2Obfs != "" {
			o["obfs"] = map[string]any{"type": "salamander", "password": p.Hy2Obfs}
		}
		if p.Hy2Up > 0 {
			o["up_mbps"] = p.Hy2Up
		}
		if p.Hy2Down > 0 {
			o["down_mbps"] = p.Hy2Down
		}
		// Hysteria2 is TLS-mandatory. Force a TLS block even if Security field
		// was not explicitly set to "tls" (e.g. older records or edge cases).
		if _, hasTLS := o["tls"]; !hasTLS {
			tls := map[string]any{"enabled": true, "insecure": true, "alpn": []string{"h3"}}
			if p.SNI != "" {
				tls["server_name"] = p.SNI
			} else if p.Host != "" {
				tls["server_name"] = p.Host
			}
			o["tls"] = tls
		} else {
			// Existing TLS block — ensure ALPN h3 is present for QUIC mandate.
			if tlsObj, ok := o["tls"].(map[string]any); ok {
				if _, hasALPN := tlsObj["alpn"]; !hasALPN {
					tlsObj["alpn"] = []string{"h3"}
				}
			}
		}
	case domain.ProtoTUIC:
		o["type"] = "tuic"
		o["uuid"] = p.UUID
		o["password"] = p.Password
		o["congestion_control"] = "bbr"
	default:
		return nil
	}
	return o
}

func singboxTransport(p Proxy) map[string]any {
	switch p.Network {
	case "ws":
		t := map[string]any{"type": "ws"}
		if p.Path != "" {
			t["path"] = p.Path
		}
		if p.HostHeader != "" {
			t["headers"] = map[string]any{"Host": p.HostHeader}
		}
		// Early-data (0-RTT) reduces connection latency; safe behind CDN.
		t["max_early_data"] = 2048
		t["early_data_header_name"] = "Sec-WebSocket-Protocol"
		return t
	case "grpc":
		return map[string]any{"type": "grpc", "service_name": p.Path}
	case "httpupgrade":
		t := map[string]any{"type": "httpupgrade"}
		if p.Path != "" {
			t["path"] = p.Path
		}
		if p.HostHeader != "" {
			t["host"] = p.HostHeader
		}
		return t
	case "http", "h2":
		t := map[string]any{"type": "http"}
		if p.Path != "" {
			t["path"] = p.Path
		}
		if p.HostHeader != "" {
			t["host"] = []string{p.HostHeader}
		}
		return t
	case "xhttp":
		// XHTTP/SplitHTTP: sing-box expresses this as http transport.
		t := map[string]any{"type": "http"}
		if p.Path != "" {
			t["path"] = p.Path
		}
		if p.HostHeader != "" {
			t["host"] = []string{p.HostHeader}
		}
		return t
	case "quic":
		t := map[string]any{"type": "quic"}
		return t
	default:
		return nil
	}
}
