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
	PrivateKey string `json:"private_key"` //nolint:gosec // not a hardcoded secret, user-provided
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
	// MTU is the WireGuard interface MTU; 0 means use the WARP default (1280).
	MTU int `json:"mtu,omitempty"`
}

// defaultMTU is the WARP WireGuard interface MTU used when none is configured.
const defaultMTU = 1280

// DefaultEndpoint is Cloudflare's WARP engagement server.
const DefaultEndpoint = "engage.cloudflareclient.com:2408"

// DefaultPublicKey is Cloudflare's WARP server public key.
const DefaultPublicKey = "bmXOC+F1FxEMF9dyiK2H5/1SUtzH0JuVo51h2wPfgyo="

// XrayOutbound generates an Xray WireGuard outbound JSON config for WARP.
func (c Config) XrayOutbound(tag string) map[string]any {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = DefaultEndpoint
	}
	pubKey := c.PublicKey
	if pubKey == "" {
		pubKey = DefaultPublicKey
	}
	peers := []map[string]any{{
		"publicKey": pubKey,
		"endpoint":  endpoint,
	}}
	settings := map[string]any{
		"secretKey": c.PrivateKey,
		"address":   c.Address,
		"peers":     peers,
		"mtu":       c.mtu(),
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
	pubKey := c.PublicKey
	if pubKey == "" {
		pubKey = DefaultPublicKey
	}
	server := "engage.cloudflareclient.com"
	port := 2408
	if c.Endpoint != "" {
		server = c.Endpoint
		// If endpoint contains a port, it's handled by the client
	}
	out := map[string]any{
		"type":            "wireguard",
		"tag":             tag,
		"server":          server,
		"server_port":     port,
		"private_key":     c.PrivateKey,
		"peer_public_key": pubKey,
		"local_address":   c.Address,
		"mtu":             c.mtu(),
	}
	if len(c.Reserved) == 3 {
		out["reserved"] = fmt.Sprintf("%d,%d,%d", c.Reserved[0], c.Reserved[1], c.Reserved[2])
	}
	return out
}

// ConfigFromMap builds a Config from a generic decoded-JSON map, as produced by
// unmarshalling Outbound.Raw["wireguard"] (where JSON numbers decode as float64
// and arrays as []any). It is tolerant of missing keys and mixed numeric types
// so renderers can construct a Config directly from persisted jsonb without an
// intermediate typed round-trip.
func ConfigFromMap(m map[string]any) Config {
	c := Config{
		PrivateKey: mapString(m["private_key"]),
		PublicKey:  mapString(m["public_key"]),
		Endpoint:   mapString(m["endpoint"]),
		LicenseKey: mapString(m["license_key"]),
		Address:    mapStringSlice(m["address"]),
	}
	if reserved := mapIntSlice(m["reserved"]); len(reserved) == 3 {
		c.Reserved = reserved
	}
	if mtu, ok := mapInt(m["mtu"]); ok {
		c.MTU = mtu
	}
	return c
}

// mtu returns the configured WireGuard MTU, falling back to the WARP default
// when unset (0).
func (c Config) mtu() int {
	if c.MTU > 0 {
		return c.MTU
	}
	return defaultMTU
}

// mapString coerces a decoded-JSON value into a string, returning "" for any
// non-string (including nil).
func mapString(v any) string {
	s, _ := v.(string)
	return s
}

// mapStringSlice coerces a decoded-JSON value into a []string. It accepts both a
// native []string and a []any whose elements are strings (the JSON-decoded
// shape), dropping non-string entries.
func mapStringSlice(v any) []string {
	switch s := v.(type) {
	case []string:
		return s
	case []any:
		out := make([]string, 0, len(s))
		for _, item := range s {
			if str, ok := item.(string); ok {
				out = append(out, str)
			}
		}
		return out
	default:
		return nil
	}
}

// mapIntSlice coerces a decoded-JSON value into a []int. JSON numbers decode as
// float64, so float64/int/int64 elements are all accepted; non-numeric entries
// are dropped.
func mapIntSlice(v any) []int {
	switch s := v.(type) {
	case []int:
		return s
	case []any:
		out := make([]int, 0, len(s))
		for _, item := range s {
			if n, ok := mapInt(item); ok {
				out = append(out, n)
			}
		}
		return out
	default:
		return nil
	}
}

// mapInt coerces a decoded-JSON numeric value into an int, reporting whether a
// usable number was present.
func mapInt(v any) (int, bool) {
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	default:
		return 0, false
	}
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
