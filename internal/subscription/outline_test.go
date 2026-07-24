package subscription

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
)

func TestRenderOutlineReturnsValidJSON(t *testing.T) {
	proxies := []OutlineProxy{
		{ID: "1", Name: "Server A", Server: "1.2.3.4", Port: 8388, Password: "secret1", Method: "aes-256-gcm"},
		{ID: "2", Name: "Server B", Server: "5.6.7.8", Port: 9090, Password: "secret2", Method: "chacha20-ietf-poly1305"},
	}

	data, err := RenderOutline(proxies)
	if err != nil {
		t.Fatalf("RenderOutline returned error: %v", err)
	}

	var keys []OutlineAccessKey
	if err := json.Unmarshal(data, &keys); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	// Verify first entry.
	if keys[0].ID != "1" || keys[0].Server != "1.2.3.4" || keys[0].Port != 8388 {
		t.Errorf("key[0] mismatch: %+v", keys[0])
	}
	if keys[0].Method != "aes-256-gcm" || keys[0].Password != "secret1" {
		t.Errorf("key[0] credentials mismatch: %+v", keys[0])
	}
	if keys[0].Name != "Server A" {
		t.Errorf("key[0] name: got %q, want %q", keys[0].Name, "Server A")
	}

	// Verify second entry.
	if keys[1].ID != "2" || keys[1].Server != "5.6.7.8" || keys[1].Port != 9090 {
		t.Errorf("key[1] mismatch: %+v", keys[1])
	}
}

func TestRenderOutlineErrorOnEmpty(t *testing.T) {
	_, err := RenderOutline(nil)
	if err == nil {
		t.Fatal("expected error for empty proxies, got nil")
	}
	if err != ErrNoSSProxies {
		t.Fatalf("expected ErrNoSSProxies, got: %v", err)
	}

	_, err = RenderOutline([]OutlineProxy{})
	if err != ErrNoSSProxies {
		t.Fatalf("expected ErrNoSSProxies for empty slice, got: %v", err)
	}
}

func TestRenderOutlineOmitsEmptyName(t *testing.T) {
	proxies := []OutlineProxy{
		{ID: "1", Name: "", Server: "1.2.3.4", Port: 8388, Password: "pw", Method: "aes-256-gcm"},
	}

	data, err := RenderOutline(proxies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The "name" field should be omitted when empty (omitempty tag).
	if strings.Contains(string(data), `"name"`) {
		t.Errorf("expected name field to be omitted, got: %s", data)
	}
}

func TestRenderOutlineURIFormat(t *testing.T) {
	proxy := OutlineProxy{
		ID:       "abc",
		Name:     "My Server",
		Server:   "10.0.0.1",
		Port:     443,
		Password: "hunter2",
		Method:   "aes-256-gcm",
	}

	uri := RenderOutlineURI(proxy)

	// Should start with ss://
	if !strings.HasPrefix(uri, "ss://") {
		t.Fatalf("expected ss:// prefix, got: %s", uri)
	}

	// Parse the URI: ss://userinfo@host:port#fragment
	// Remove "ss://" prefix for parsing.
	rest := strings.TrimPrefix(uri, "ss://")

	// Split on # to get fragment (name).
	parts := strings.SplitN(rest, "#", 2)
	if len(parts) != 2 {
		t.Fatalf("expected fragment in URI, got: %s", uri)
	}

	decodedName, err := url.PathUnescape(parts[1])
	if err != nil {
		t.Fatalf("failed to unescape fragment: %v", err)
	}
	if decodedName != "My Server" {
		t.Errorf("fragment name: got %q, want %q", decodedName, "My Server")
	}

	// Split on @ to get userinfo and host:port.
	hostParts := strings.SplitN(parts[0], "@", 2)
	if len(hostParts) != 2 {
		t.Fatalf("expected @ separator in URI, got: %s", parts[0])
	}

	// Decode the base64 userinfo.
	decoded, err := base64.RawURLEncoding.DecodeString(hostParts[0])
	if err != nil {
		t.Fatalf("failed to decode base64 userinfo: %v", err)
	}
	expected := "aes-256-gcm:hunter2"
	if string(decoded) != expected {
		t.Errorf("userinfo: got %q, want %q", string(decoded), expected)
	}

	// Verify host:port.
	if hostParts[1] != "10.0.0.1:443" {
		t.Errorf("host:port: got %q, want %q", hostParts[1], "10.0.0.1:443")
	}
}

func TestRenderOutlineURISpecialCharsInName(t *testing.T) {
	proxy := OutlineProxy{
		ID:       "1",
		Name:     "DE 🇩🇪 / Frankfurt",
		Server:   "1.2.3.4",
		Port:     8388,
		Password: "pw",
		Method:   "chacha20-ietf-poly1305",
	}

	uri := RenderOutlineURI(proxy)

	// The fragment should be URL-encoded.
	if !strings.HasPrefix(uri, "ss://") {
		t.Fatalf("invalid ss URI: %s", uri)
	}

	// Extract and decode the fragment.
	idx := strings.LastIndex(uri, "#")
	if idx == -1 {
		t.Fatalf("missing fragment in URI: %s", uri)
	}
	decoded, err := url.PathUnescape(uri[idx+1:])
	if err != nil {
		t.Fatalf("failed to unescape: %v", err)
	}
	if decoded != proxy.Name {
		t.Errorf("name round-trip failed: got %q, want %q", decoded, proxy.Name)
	}
}

func TestRenderOutlineJSONFieldNames(t *testing.T) {
	proxies := []OutlineProxy{
		{ID: "x", Name: "test", Server: "host", Port: 1234, Password: "p", Method: "m"},
	}

	data, err := RenderOutline(proxies)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the exact JSON field names match the Outline schema.
	var raw []map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	required := []string{"id", "method", "password", "server", "server_port"}
	for _, field := range required {
		if _, ok := raw[0][field]; !ok {
			t.Errorf("missing required field %q in JSON output", field)
		}
	}

	// server_port should be a number.
	if port, ok := raw[0]["server_port"].(float64); !ok || int(port) != 1234 {
		t.Errorf("server_port: got %v, want 1234", raw[0]["server_port"])
	}
}

func TestRenderOutlineURIConsistency(t *testing.T) {
	// RenderOutlineURI and the ssLink from links.go should produce equivalent
	// formats for the same input data.
	proxy := OutlineProxy{
		ID:       "1",
		Name:     "test",
		Server:   "1.2.3.4",
		Port:     8388,
		Password: "secret",
		Method:   "aes-128-gcm",
	}

	uri := RenderOutlineURI(proxy)
	expected := fmt.Sprintf("ss://%s@%s:%d#%s",
		base64.RawURLEncoding.EncodeToString([]byte(proxy.Method+":"+proxy.Password)),
		proxy.Server,
		proxy.Port,
		url.PathEscape(proxy.Name),
	)

	if uri != expected {
		t.Errorf("URI mismatch:\ngot:  %s\nwant: %s", uri, expected)
	}
}
