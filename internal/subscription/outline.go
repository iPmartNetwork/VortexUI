package subscription

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

// OutlineConfig represents the basic Shadowsocks configuration format used by
// the Outline client application.
type OutlineConfig struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

// OutlineAccessKey is the Outline-native access key format returned in the JSON
// array by RenderOutline. It maps to the schema the Outline app expects when
// importing server configurations.
type OutlineAccessKey struct {
	ID       string `json:"id"`
	Method   string `json:"method"`
	Password string `json:"password"`
	Server   string `json:"server"`
	Port     int    `json:"server_port"`
	Name     string `json:"name,omitempty"`
}

// OutlineProxy is the input struct for the Outline renderer. Each entry
// represents a Shadowsocks proxy endpoint that should be rendered for the
// Outline client.
type OutlineProxy struct {
	ID       string
	Name     string
	Server   string
	Port     int
	Password string
	Method   string
}

// ErrNoSSProxies is returned when RenderOutline receives no Shadowsocks proxies.
var ErrNoSSProxies = errors.New("no shadowsocks proxies available for Outline format")

// RenderOutline takes a list of OutlineProxy entries, filters to only
// Shadowsocks protocol entries (all OutlineProxy entries are assumed to be SS),
// and returns a JSON array of OutlineAccessKey entries. Returns an error if no
// proxies are provided.
func RenderOutline(proxies []OutlineProxy) ([]byte, error) {
	if len(proxies) == 0 {
		return nil, ErrNoSSProxies
	}

	keys := make([]OutlineAccessKey, 0, len(proxies))
	for _, p := range proxies {
		keys = append(keys, OutlineAccessKey{
			ID:       p.ID,
			Method:   p.Method,
			Password: p.Password,
			Server:   p.Server,
			Port:     p.Port,
			Name:     p.Name,
		})
	}

	return json.Marshal(keys)
}

// RenderOutlineURI returns the ss:// URI format for a single proxy:
// ss://base64(method:password)@server:port#name
// This is the standard SIP002 share link format used by Outline for importing
// individual access keys.
func RenderOutlineURI(proxy OutlineProxy) string {
	userinfo := base64.RawURLEncoding.EncodeToString(
		[]byte(proxy.Method + ":" + proxy.Password),
	)
	fragment := url.PathEscape(proxy.Name)
	return fmt.Sprintf("ss://%s@%s:%d#%s", userinfo, proxy.Server, proxy.Port, fragment)
}
