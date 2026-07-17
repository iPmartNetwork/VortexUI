package subscription

import (
	"encoding/json"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderSingbox builds a sing-box config fragment: one outbound per proxy plus a
// selector that fronts them. Clients merge this with their own inbounds/route.
// When rules is empty no route section is emitted, so output stays byte-identical
// to before; when a pack supplies rules they are translated into route.rules with
// a route.final pointing at the selector, and a block outbound is appended only
// if a rule rejects traffic.
func renderSingbox(proxies []Proxy, title string, rules []domain.RoutingRule) ([]byte, error) {
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
	direct := map[string]any{"type": "direct", "tag": "direct"}
	all := append([]map[string]any{selector, autoTest, fallback}, outbounds...)
	all = append(all, direct)

	cfg := map[string]any{"outbounds": all}
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
	o["dial_timeout"] = "5s"
	o["tcp_fast_open"] = true
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
	default:
		return nil
	}
}
