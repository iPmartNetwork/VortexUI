package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/vortexui/vortexui/internal/domain"
)

// ShareLink renders a Proxy as its protocol's standard share URI (vless://,
// vmess://, trojan://, ss://). An unsupported protocol yields an empty string so
// callers can simply skip it.
func ShareLink(p Proxy) string {
	switch p.Protocol {
	case domain.ProtoVLESS:
		return vlessLink(p)
	case domain.ProtoVMess:
		return vmessLink(p)
	case domain.ProtoTrojan:
		return trojanLink(p)
	case domain.ProtoShadowsocks:
		return ssLink(p)
	default:
		return ""
	}
}

// transportQuery builds the shared query parameters used by VLESS/Trojan URIs.
func transportQuery(p Proxy) url.Values {
	q := url.Values{}
	q.Set("type", orDefault(p.Network, "tcp"))
	q.Set("security", orDefault(p.Security, "none"))
	if p.SNI != "" {
		q.Set("sni", p.SNI)
	}
	switch p.Network {
	case "ws":
		if p.Path != "" {
			q.Set("path", p.Path)
		}
		if p.HostHeader != "" {
			q.Set("host", p.HostHeader)
		}
	case "grpc":
		if p.Path != "" {
			q.Set("serviceName", p.Path)
		}
	}
	return q
}

func vlessLink(p Proxy) string {
	q := transportQuery(p)
	if p.Flow != "" {
		q.Set("flow", p.Flow)
	}
	if p.Security == "reality" {
		if p.PublicKey != "" {
			q.Set("pbk", p.PublicKey)
		}
		if p.ShortID != "" {
			q.Set("sid", p.ShortID)
		}
		q.Set("fp", orDefault(p.Fingerprint, "chrome"))
	}
	u := url.URL{
		Scheme:   "vless",
		User:     url.User(p.UUID),
		Host:     hostPort(p.Host, p.Port),
		RawQuery: q.Encode(),
		Fragment: p.Name,
	}
	return u.String()
}

func trojanLink(p Proxy) string {
	u := url.URL{
		Scheme:   "trojan",
		User:     url.User(p.Password),
		Host:     hostPort(p.Host, p.Port),
		RawQuery: transportQuery(p).Encode(),
		Fragment: p.Name,
	}
	return u.String()
}

// vmessLink uses the de-facto "v2rayN" JSON-over-base64 format.
func vmessLink(p Proxy) string {
	m := map[string]string{
		"v": "2", "ps": p.Name, "add": p.Host, "port": strconv.Itoa(p.Port),
		"id": p.UUID, "aid": "0", "scy": "auto",
		"net": orDefault(p.Network, "tcp"), "type": "none",
		"host": p.HostHeader, "path": p.Path,
		"tls": tlsField(p.Security), "sni": p.SNI,
	}
	raw, _ := json.Marshal(m)
	return "vmess://" + base64.StdEncoding.EncodeToString(raw)
}

// ssLink uses SIP002: ss://base64url(method:password)@host:port#name
func ssLink(p Proxy) string {
	userinfo := base64.RawURLEncoding.EncodeToString([]byte(p.SSMethod + ":" + p.Password))
	return fmt.Sprintf("ss://%s@%s#%s", userinfo, hostPort(p.Host, p.Port), url.PathEscape(p.Name))
}

func hostPort(host string, port int) string {
	return net.JoinHostPort(host, strconv.Itoa(port))
}

func tlsField(security string) string {
	if security == "none" || security == "" {
		return ""
	}
	return security // "tls" or "reality"
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
