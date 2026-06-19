package wireguard

// This file holds the higher-level server/peer/client config types used by the
// WireGuard subscription delivery (later stages). Key generation lives in
// keys.go.

import (
	"fmt"
)

// ServerConfig holds the WireGuard server (inbound) configuration.
type ServerConfig struct {
	PrivateKey string   `json:"private_key"`
	PublicKey  string   `json:"public_key"`
	ListenPort int      `json:"listen_port"`
	Address    []string `json:"address"` // server tunnel addresses, e.g. ["10.0.0.1/24"]
	DNS        []string `json:"dns"`
	MTU        int      `json:"mtu"`
}

// PeerConfig is a single client (user) peer entry.
type PeerConfig struct {
	PublicKey    string   `json:"public_key"`
	AllowedIPs  []string `json:"allowed_ips"`  // e.g. ["10.0.0.2/32"]
	PresharedKey string  `json:"preshared_key,omitempty"`
}

// ClientConfig is what gets delivered to the user in their subscription.
type ClientConfig struct {
	PrivateKey  string   `json:"private_key"`
	Address     []string `json:"address"`     // client tunnel IP
	DNS         []string `json:"dns"`
	ServerPubKey string  `json:"server_public_key"`
	Endpoint    string   `json:"endpoint"`    // server:port
	AllowedIPs  []string `json:"allowed_ips"` // usually ["0.0.0.0/0", "::/0"]
	MTU         int      `json:"mtu"`
}

// RenderClientINI produces a standard WireGuard .conf file content.
func (c ClientConfig) RenderClientINI() string {
	mtu := c.MTU
	if mtu == 0 {
		mtu = 1280
	}
	conf := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = %s
MTU = %d

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = %s
PersistentKeepalive = 25
`, c.PrivateKey, joinComma(c.Address), joinComma(c.DNS), mtu,
		c.ServerPubKey, c.Endpoint, joinComma(c.AllowedIPs))
	return conf
}

func joinComma(ss []string) string {
	if len(ss) == 0 {
		return ""
	}
	out := ss[0]
	for _, s := range ss[1:] {
		out += ", " + s
	}
	return out
}
