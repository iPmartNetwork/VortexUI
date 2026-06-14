package subscription

import (
	"encoding/base64"
	"strings"
)

// Format is a client configuration dialect.
type Format string

const (
	FormatBase64  Format = "base64"  // newline-joined share links, base64-encoded
	FormatClash   Format = "clash"   // Clash / Clash.Meta YAML
	FormatSingbox Format = "singbox" // sing-box JSON
)

// Detect picks the best format for a client from its User-Agent. Defaulting to
// base64 is safe: every generic subscription client understands it.
func Detect(userAgent string) Format {
	ua := strings.ToLower(userAgent)
	switch {
	case strings.Contains(ua, "clash") || strings.Contains(ua, "mihomo") || strings.Contains(ua, "stash"):
		return FormatClash
	case strings.Contains(ua, "sing-box") || strings.Contains(ua, "singbox") || strings.Contains(ua, "hiddify"):
		return FormatSingbox
	default:
		return FormatBase64
	}
}

// ContentType returns the HTTP Content-Type a given format should be served as.
func (f Format) ContentType() string {
	switch f {
	case FormatClash:
		return "text/yaml; charset=utf-8"
	case FormatSingbox:
		return "application/json; charset=utf-8"
	default:
		return "text/plain; charset=utf-8"
	}
}

// Render produces the subscription body for the chosen format.
func Render(f Format, proxies []Proxy, title string) ([]byte, error) {
	switch f {
	case FormatClash:
		return renderClash(proxies, title)
	case FormatSingbox:
		return renderSingbox(proxies, title)
	default:
		return renderBase64(proxies), nil
	}
}

// renderBase64 joins every share link with newlines and base64-encodes the blob,
// the universal subscription format.
func renderBase64(proxies []Proxy) []byte {
	var b strings.Builder
	for _, p := range proxies {
		if link := ShareLink(p); link != "" {
			b.WriteString(link)
			b.WriteString("\n")
		}
	}
	enc := base64.StdEncoding.EncodeToString([]byte(b.String()))
	return []byte(enc)
}
