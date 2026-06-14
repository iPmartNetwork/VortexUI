// Package subscription turns a user's bound inbounds into the various client
// configuration formats (share links, Clash, sing-box). It is pure: it takes
// normalized Proxy values and renders bytes, with no knowledge of storage or
// HTTP, so every renderer is unit-tested in isolation.
package subscription

import "github.com/vortexui/vortexui/internal/domain"

// Proxy is a single, fully-resolved outbound endpoint a client can connect to.
// It flattens a user's credentials together with one inbound's transport and the
// node's public host, which is everything each renderer needs.
type Proxy struct {
	Name     string          // display label, e.g. "alice @ de-1"
	Protocol domain.Protocol
	Host     string          // node public host/IP
	Port     int

	// Transport / TLS.
	Network    string // tcp, ws, grpc
	Security   string // none, tls, reality
	SNI        string
	Path       string
	HostHeader string
	Flow       string

	// Credentials (only the field relevant to Protocol is populated).
	UUID     string // vmess / vless
	Password string // trojan / shadowsocks
	SSMethod string // shadowsocks cipher
}
