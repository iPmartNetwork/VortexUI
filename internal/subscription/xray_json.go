package subscription

import (
	"encoding/json"

	"github.com/vortexui/vortexui/internal/domain"
)

// renderXrayJSON builds a raw Xray/V2Ray client configuration document holding
// one outbound per proxy: {"outbounds": [ ... ]}. V2Ray-core clients import this
// directly. It mirrors the per-protocol + stream/TLS field mapping the share
// links and sing-box renderer already use, expressed in Xray's outbound dialect.
func renderXrayJSON(proxies []Proxy) ([]byte, error) {
	outbounds := make([]map[string]any, 0, len(proxies)+1)
	for _, p := range proxies {
		if o := xrayOutbound(p); o != nil {
			outbounds = append(outbounds, o)
		}
	}
	// A freedom outbound is the conventional direct/default tail in client configs.
	outbounds = append(outbounds, map[string]any{"protocol": "freedom", "tag": "direct"})
	return json.MarshalIndent(map[string]any{"outbounds": outbounds}, "", "  ")
}

// xrayOutbound maps one Proxy to a single Xray outbound object, or nil for an
// unsupported protocol so the caller can skip it.
func xrayOutbound(p Proxy) map[string]any {
	o := map[string]any{"tag": p.Name}

	switch p.Protocol {
	case domain.ProtoVLESS:
		user := map[string]any{"id": p.UUID, "encryption": "none"}
		if p.Flow != "" {
			user["flow"] = p.Flow
		}
		o["protocol"] = "vless"
		o["settings"] = map[string]any{
			"vnext": []map[string]any{{
				"address": p.Host,
				"port":    p.Port,
				"users":   []map[string]any{user},
			}},
		}
	case domain.ProtoVMess:
		o["protocol"] = "vmess"
		o["settings"] = map[string]any{
			"vnext": []map[string]any{{
				"address": p.Host,
				"port":    p.Port,
				"users":   []map[string]any{{"id": p.UUID, "alterId": 0, "security": "auto"}},
			}},
		}
	case domain.ProtoTrojan:
		o["protocol"] = "trojan"
		o["settings"] = map[string]any{
			"servers": []map[string]any{{
				"address":  p.Host,
				"port":     p.Port,
				"password": p.Password,
			}},
		}
	case domain.ProtoShadowsocks:
		o["protocol"] = "shadowsocks"
		o["settings"] = map[string]any{
			"servers": []map[string]any{{
				"address":  p.Host,
				"port":     p.Port,
				"method":   p.SSMethod,
				"password": p.Password,
			}},
		}
	default:
		return nil
	}

	if ss := xrayStreamSettings(p); ss != nil {
		o["streamSettings"] = ss
	}
	if p.Mux {
		o["mux"] = map[string]any{"enabled": true}
	}
	return o
}

// xrayStreamSettings builds the streamSettings block (network + TLS/REALITY)
// using the same transport/TLS semantics as the other renderers. Shadowsocks
// outbounds carry no stream settings (returns nil) when plain tcp/none.
func xrayStreamSettings(p Proxy) map[string]any {
	network := orDefault(p.Network, "tcp")
	security := orDefault(p.Security, "none")

	ss := map[string]any{
		"network":  network,
		"security": security,
	}

	switch security {
	case "tls":
		tls := map[string]any{}
		if p.SNI != "" {
			tls["serverName"] = p.SNI
		}
		if p.AllowInsecure {
			tls["allowInsecure"] = true
		}
		if len(p.ALPN) > 0 {
			tls["alpn"] = p.ALPN
		}
		if p.Fingerprint != "" {
			tls["fingerprint"] = p.Fingerprint
		}
		ss["tlsSettings"] = tls
	case "reality":
		ss["realitySettings"] = map[string]any{
			"serverName":  p.SNI,
			"publicKey":   p.PublicKey,
			"shortId":     p.ShortID,
			"fingerprint": orDefault(p.Fingerprint, "chrome"),
		}
	}

	switch network {
	case "ws":
		w := map[string]any{}
		if p.Path != "" {
			w["path"] = p.Path
		}
		if p.HostHeader != "" {
			w["headers"] = map[string]any{"Host": p.HostHeader}
		}
		ss["wsSettings"] = w
	case "grpc":
		ss["grpcSettings"] = map[string]any{"serviceName": p.Path}
	case "httpupgrade":
		h := map[string]any{}
		if p.Path != "" {
			h["path"] = p.Path
		}
		if p.HostHeader != "" {
			h["host"] = p.HostHeader
		}
		ss["httpupgradeSettings"] = h
	case "http", "h2":
		h := map[string]any{}
		if p.Path != "" {
			h["path"] = p.Path
		}
		if p.HostHeader != "" {
			h["host"] = []string{p.HostHeader}
		}
		ss["httpSettings"] = h
	}

	return ss
}
