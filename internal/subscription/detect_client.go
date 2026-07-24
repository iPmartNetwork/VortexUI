package subscription

import "strings"

// DetectClientApp extracts the proxy client app name from a User-Agent string.
func DetectClientApp(ua string) string {
	ua = strings.ToLower(ua)
	switch {
	case strings.Contains(ua, "clash"), strings.Contains(ua, "mihomo"):
		return "clash"
	case strings.Contains(ua, "singbox"), strings.Contains(ua, "sing-box"):
		return "sing-box"
	case strings.Contains(ua, "v2rayn"), strings.Contains(ua, "v2rayng"), strings.Contains(ua, "v2ray"):
		return "v2ray"
	case strings.Contains(ua, "shadowrocket"):
		return "shadowrocket"
	case strings.Contains(ua, "nekobox"), strings.Contains(ua, "nekoray"):
		return "nekobox"
	case strings.Contains(ua, "hiddify"):
		return "hiddify"
	case strings.Contains(ua, "streisand"):
		return "streisand"
	case strings.Contains(ua, "foxray"):
		return "foxray"
	case strings.Contains(ua, "karing"):
		return "karing"
	case strings.Contains(ua, "quantumult"), strings.Contains(ua, "surge"):
		return "quantumult"
	default:
		return "unknown"
	}
}
