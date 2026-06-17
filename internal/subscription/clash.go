package subscription

import (
	"gopkg.in/yaml.v3"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderClash builds a minimal but valid Clash.Meta config: every proxy, a
// single selector group containing them all, and a default catch-all rule.
func renderClash(proxies []Proxy, title string) ([]byte, error) {
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

	cfg := map[string]any{
		"proxies": clashProxies,
		"proxy-groups": []map[string]any{
			{
				"name":    title,
				"type":    "select",
				"proxies": append([]string{"♻️ Auto", "DIRECT"}, names...),
			},
			{
				"name":     "♻️ Auto",
				"type":     "url-test",
				"proxies":  names,
				"url":      "https://www.gstatic.com/generate_204",
				"interval": 300,
				"tolerance": 50,
			},
		},
		"rules": []string{"MATCH," + title},
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
	if p.Security == "reality" {
		base["reality-opts"] = map[string]any{"public-key": p.PublicKey, "short-id": p.ShortID}
		base["client-fingerprint"] = orDefault(p.Fingerprint, "chrome")
	}
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
	}
}
