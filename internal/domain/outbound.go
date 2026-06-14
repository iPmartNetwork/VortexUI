package domain

import "github.com/google/uuid"

// OutboundProtocol enumerates the egress protocols an outbound can speak. Unlike
// inbound Protocol these include the local-dispatch protocols (freedom/blackhole/
// dns) used by routing as well as the proxy protocols used for chaining/balancing.
type OutboundProtocol string

const (
	OutFreedom     OutboundProtocol = "freedom"   // direct egress
	OutBlackhole   OutboundProtocol = "blackhole" // drop (block)
	OutDNS         OutboundProtocol = "dns"       // dns hijack outbound
	OutVMess       OutboundProtocol = "vmess"
	OutVLESS       OutboundProtocol = "vless"
	OutTrojan      OutboundProtocol = "trojan"
	OutShadowsocks OutboundProtocol = "shadowsocks"
	OutSocks       OutboundProtocol = "socks"
	OutHTTP        OutboundProtocol = "http"
)

// Valid reports whether the protocol is one VortexUI knows how to render.
func (p OutboundProtocol) Valid() bool {
	switch p {
	case OutFreedom, OutBlackhole, OutDNS, OutVMess, OutVLESS, OutTrojan, OutShadowsocks, OutSocks, OutHTTP:
		return true
	default:
		return false
	}
}

// NeedsEndpoint reports whether the protocol dials an upstream server and so
// requires an address and port. Local-dispatch protocols (freedom/blackhole/dns)
// do not.
func (p OutboundProtocol) NeedsEndpoint() bool {
	switch p {
	case OutVMess, OutVLESS, OutTrojan, OutShadowsocks, OutSocks, OutHTTP:
		return true
	default:
		return false
	}
}

// Outbound is an egress endpoint configured on a node. It is engine-neutral: each
// core builder renders it into its own JSON. freedom/blackhole/dns need no
// endpoint; proxy protocols carry the upstream address, credentials, and
// transport so they can be used for chaining and as balancer members.
type Outbound struct {
	ID       uuid.UUID        `json:"id"`
	NodeID   uuid.UUID        `json:"node_id"`
	Tag      string           `json:"tag"` // unique per node; referenced by routing/balancers
	Protocol OutboundProtocol `json:"protocol"`

	// Upstream endpoint (proxy protocols only).
	Address string `json:"address,omitempty"`
	Port    int    `json:"port,omitempty"`

	// Credentials. Only the field relevant to Protocol is used.
	UUID     string `json:"uuid,omitempty"`     // vmess / vless
	Password string `json:"password,omitempty"` // trojan / shadowsocks / socks / http
	Username string `json:"username,omitempty"` // socks / http
	Method   string `json:"method,omitempty"`   // shadowsocks cipher
	Flow     string `json:"flow,omitempty"`     // vless flow (e.g. xtls-rprx-vision)

	// Transport / TLS layer for proxy outbounds.
	Network  string   `json:"network,omitempty"` // tcp, ws, grpc
	Security Security `json:"security,omitempty"`
	SNI      string   `json:"sni,omitempty"`
	Path     string   `json:"path,omitempty"`
	Host     string   `json:"host,omitempty"`

	// Raw is an engine-native override merged over the generated settings, for
	// fields the abstraction does not model.
	Raw map[string]any `json:"raw,omitempty"`

	Enabled bool `json:"enabled"`
}

// Validate checks the outbound is internally consistent before persistence.
func (o *Outbound) Validate() error {
	if o.Tag == "" {
		return errInvalid("outbound tag is required")
	}
	if !o.Protocol.Valid() {
		return errInvalid("unknown outbound protocol %q", string(o.Protocol))
	}
	if o.Protocol.NeedsEndpoint() && (o.Address == "" || o.Port == 0) {
		return errInvalid("outbound %q (%s) requires address and port", o.Tag, o.Protocol)
	}
	return nil
}
