// Package warp provides Cloudflare WARP/WARP+ integration as an outbound
// option. When enabled, traffic exits through Cloudflare's network, which
// bypasses many censorship systems and provides a clean IP for services that
// block datacenter IPs.
//
// Usage: add a WARP outbound to a node, then route specific traffic through it
// via routing rules (e.g. "geosite:openai" → warp, or all traffic → warp).
package warp

import (
	"encoding/json"
	"fmt"
)

// Config generates the Xray/sing-box outbound configuration for WARP.
// For Xray: uses WireGuard outbound with Cloudflare's WARP keys.
// For sing-box: native wireguard outbound.
type Config struct {
	// PrivateKey is the WireGuard private key from warp-cli registration.
	PrivateKey string `json:"private_key"`
	// PublicKey is Cloudflare's WARP public key.
	PublicKey string `json:"public_key"`
	// Address is the assigned IPv4/IPv6 from WARP registration.
	Address []string `json:"address"` // e.g. ["172.16.0.2/32", "fd01:...::1/128"]
	// Reserved is the client ID bytes (3 bytes as ints).
	Reserved []int `json:"reserved,omitempty"` // e.g. [171, 48, 225]
	// Endpoint is the WARP server to connect to.
	Endpoint string `json:"endpoint"` // e.g. "engage.cloudflareclient.com:2408"
	// LicenseKey is the WARP+ license (optional, for premium routing).
	LicenseKey string `json:"license_key,omitempty"`
}

// DefaultEndpoint is Cloudflare's WARP engagement server.
const DefaultEndpoint = "engage.cloudflareclient.com:2408"

// DefaultPublicKey is Cloudflare's WARP server public key.
const DefaultPublicKey = "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo="

// XrayOutbound generates an Xray WireGuard outbound JSON config for WARP.
func (c Config) XrayOutbound(tag string) map[string]any {
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}
	if c.PublicKey == "" {
		c.PublicKey = DefaultPublicKey
	}
	peers := []map[string]any{{
		"publicKey": c.PublicKey,
		"endpoint":  c.Endpoint,
	}}
	settings := map[string]any{
		"secretKey": c.PrivateKey,
		"address":   c.Address,
		"peers":     peers,
		"mtu":       1280,
	}
	if len(c.Reserved) == 3 {
		settings["reserved"] = c.Reserved
	}
	return map[string]any{
		"tag":      tag,
		"protocol": "wireguard",
		"settings": settings,
	}
}

// SingboxOutbound generates a sing-box WireGuard outbound config for WARP.
func (c Config) SingboxOutbound(tag string) map[string]any {
	if c.Endpoint == "" {
		c.Endpoint = DefaultEndpoint
	}
	if c.PublicKey == "" {
		c.PublicKey = DefaultPublicKey
	}
	out := map[string]any{
		"type":        "wireguard",
		"tag":         tag,
		"server":      "engage.cloudflareclient.com",
		"server_port": 2408,
		"private_key": c.PrivateKey,
		"peer_public_key": c.PublicKey,
		"local_address":   c.Address,
		"mtu":             1280,
	}
	if len(c.Reserved) == 3 {
		out["reserved"] = fmt.Sprintf("%d,%d,%d", c.Reserved[0], c.Reserved[1], c.Reserved[2])
	}
	return out
}

// ToJSON serializes the config for storage.
func (c Config) ToJSON() []byte {
	data, _ := json.Marshal(c)
	return data
}

// FromJSON deserializes a stored WARP config.
func FromJSON(data []byte) (*Config, error) {
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}
