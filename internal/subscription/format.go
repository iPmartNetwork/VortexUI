package subscription

import (
	"encoding/base64"
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// Format is a client configuration dialect.
type Format string

const (
	FormatBase64  Format = "base64"  // newline-joined share links, base64-encoded
	FormatClash   Format = "clash"   // Clash / Clash.Meta YAML
	FormatSingbox Format = "singbox" // sing-box JSON
	FormatXray    Format = "xray"    // raw Xray/V2Ray outbounds JSON
	FormatOutline Format = "outline" // ss:// list for Outline
	FormatLinks   Format = "links"   // plain newline-joined share links (no base64)
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
	case strings.Contains(ua, "outline"):
		return FormatOutline
	case strings.Contains(ua, "v2rayng") || strings.Contains(ua, "v2rayn"):
		return FormatLinks
	default:
		return FormatBase64
	}
}

// ContentType returns the HTTP Content-Type a given format should be served as.
func (f Format) ContentType() string {
	switch f {
	case FormatClash:
		return "text/yaml; charset=utf-8"
	case FormatSingbox, FormatXray:
		return "application/json; charset=utf-8"
	default:
		// base64, outline and links are all plain text.
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
	case FormatXray:
		return renderXrayJSON(proxies)
	case FormatOutline:
		return renderOutline(proxies), nil
	case FormatLinks:
		return renderLinks(proxies), nil
	default:
		return renderBase64(proxies), nil
	}
}

// renderBase64 joins every share link with newlines and base64-encodes the blob,
// the universal subscription format.
func renderBase64(proxies []Proxy) []byte {
	enc := base64.StdEncoding.EncodeToString(renderLinks(proxies))
	return []byte(enc)
}

// renderLinks joins every share link with newlines WITHOUT base64 encoding —
// the V2rayN-friendly plain-text variant. renderBase64 is exactly this output
// run through base64, so the two stay byte-consistent by construction.
func renderLinks(proxies []Proxy) []byte {
	var b strings.Builder
	for _, p := range proxies {
		if link := ShareLink(p); link != "" {
			b.WriteString(link)
			b.WriteString("\n")
		}
	}
	return []byte(b.String())
}

// renderOutline emits an ss:// link per shadowsocks-capable proxy and skips
// everything else, which is what the Outline client can import. With no
// shadowsocks proxies the body is empty but still valid.
func renderOutline(proxies []Proxy) []byte {
	var b strings.Builder
	for _, p := range proxies {
		if p.Protocol != domain.ProtoShadowsocks {
			continue
		}
		if link := ssLink(p); link != "" {
			b.WriteString(link)
			b.WriteString("\n")
		}
	}
	return []byte(b.String())
}
