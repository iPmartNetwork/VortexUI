package subscription

import (
	"encoding/json"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderSingbox builds a sing-box config fragment: one outbound per proxy plus a
// selector that fronts them. Clients merge this with their own inbounds/route.
func renderSingbox(proxies []Proxy, title string) ([]byte, error) {
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
		"outbounds": tags,
	}
	direct := map[string]any{"type": "direct", "tag": "direct"}
	all := append([]map[string]any{selector}, outbounds...)
	all = append(all, direct)

	return json.MarshalIndent(map[string]any{"outbounds": all}, "", "  ")
}

func singboxOutbound(p Proxy) map[string]any {
	o := map[string]any{
		"tag":         p.Name,
		"server":      p.Host,
		"server_port": p.Port,
	}
	if p.Security == "tls" || p.Security == "reality" {
		tls := map[string]any{"enabled": true}
		if p.SNI != "" {
			tls["server_name"] = p.SNI
		}
		if p.AllowInsecure {
			tls["insecure"] = true
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
	default:
		return nil
	}
}
