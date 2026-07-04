package postgres

import "strings"

// ProtocolLabel builds a human-readable protocol label from inbound fields.
func ProtocolLabel(protocol, transport, security string) string {
	p := strings.ToLower(protocol)
	t := strings.ToLower(transport)
	s := strings.ToLower(security)
	switch {
	case p == "vless" && strings.Contains(s, "reality"):
		return "VLESS+Reality"
	case p == "vmess" && strings.Contains(t, "ws"):
		return "VMess+WS+CDN"
	case p == "hysteria2" || p == "hysteria":
		return "Hysteria2 UDP"
	case p == "trojan" || p == "shadowsocks":
		return "Trojan / SS"
	default:
		if t != "" {
			return strings.ToUpper(p) + "+" + strings.ToUpper(t)
		}
		if p != "" {
			return strings.ToUpper(p)
		}
		return "—"
	}
}
