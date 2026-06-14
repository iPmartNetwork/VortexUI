package subscription

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/vortexui/vortexui/internal/domain"
)

func sampleProxies() []Proxy {
	return []Proxy{
		{Name: "vless1", Protocol: domain.ProtoVLESS, Host: "1.2.3.4", Port: 443, Network: "ws", Security: "tls", SNI: "ex.com", Path: "/v", Flow: "xtls-rprx-vision", UUID: "11111111-1111-1111-1111-111111111111"},
		{Name: "tro1", Protocol: domain.ProtoTrojan, Host: "1.2.3.4", Port: 8443, Network: "tcp", Security: "tls", SNI: "ex.com", Password: "trojpw"},
		{Name: "ss1", Protocol: domain.ProtoShadowsocks, Host: "1.2.3.4", Port: 8388, Password: "sspw", SSMethod: "aes-128-gcm"},
	}
}

func TestShareLinks(t *testing.T) {
	ps := sampleProxies()

	vless := ShareLink(ps[0])
	if !strings.HasPrefix(vless, "vless://11111111-1111-1111-1111-111111111111@1.2.3.4:443") {
		t.Errorf("vless link malformed: %s", vless)
	}
	for _, want := range []string{"type=ws", "security=tls", "sni=ex.com", "flow=xtls-rprx-vision", "#vless1"} {
		if !strings.Contains(vless, want) {
			t.Errorf("vless link missing %q: %s", want, vless)
		}
	}

	if tro := ShareLink(ps[1]); !strings.HasPrefix(tro, "trojan://trojpw@1.2.3.4:8443") {
		t.Errorf("trojan link malformed: %s", tro)
	}

	ss := ShareLink(ps[2])
	if !strings.HasPrefix(ss, "ss://") || !strings.Contains(ss, "@1.2.3.4:8388") {
		t.Errorf("ss link malformed: %s", ss)
	}
	// SIP002 userinfo decodes to method:password.
	raw := ss[len("ss://"):strings.Index(ss, "@")]
	dec, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || string(dec) != "aes-128-gcm:sspw" {
		t.Errorf("ss userinfo = %q (err %v), want aes-128-gcm:sspw", dec, err)
	}
}

func TestDetectFormat(t *testing.T) {
	cases := map[string]Format{
		"clash-verge/1.0":     FormatClash,
		"mihomo":              FormatClash,
		"sing-box 1.8":        FormatSingbox,
		"HiddifyNext/2":       FormatSingbox,
		"v2rayNG/1.8":         FormatBase64,
		"":                    FormatBase64,
	}
	for ua, want := range cases {
		if got := Detect(ua); got != want {
			t.Errorf("Detect(%q) = %s, want %s", ua, got, want)
		}
	}
}

func TestRenderBase64RoundTrips(t *testing.T) {
	body, err := Render(FormatBase64, sampleProxies(), "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	dec, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		t.Fatalf("not valid base64: %v", err)
	}
	lines := strings.Count(strings.TrimSpace(string(dec)), "\n") + 1
	if lines != 3 {
		t.Errorf("want 3 links, got %d:\n%s", lines, dec)
	}
}

func TestRenderClashIsValidYAML(t *testing.T) {
	body, err := Render(FormatClash, sampleProxies(), "MyProfile")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Proxies     []map[string]any `yaml:"proxies"`
		ProxyGroups []struct {
			Name    string   `yaml:"name"`
			Proxies []string `yaml:"proxies"`
		} `yaml:"proxy-groups"`
	}
	if err := yaml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid clash yaml: %v\n%s", err, body)
	}
	if len(parsed.Proxies) != 3 {
		t.Errorf("want 3 clash proxies, got %d", len(parsed.Proxies))
	}
	if len(parsed.ProxyGroups) != 1 || parsed.ProxyGroups[0].Name != "MyProfile" {
		t.Errorf("proxy-group malformed: %+v", parsed.ProxyGroups)
	}
}

func TestRenderSingboxIsValidJSON(t *testing.T) {
	body, err := Render(FormatSingbox, sampleProxies(), "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid singbox json: %v\n%s", err, body)
	}
	// 1 selector + 3 proxies + 1 direct = 5 outbounds.
	if len(parsed.Outbounds) != 5 {
		t.Errorf("want 5 outbounds, got %d", len(parsed.Outbounds))
	}
	if parsed.Outbounds[0]["type"] != "selector" {
		t.Errorf("first outbound should be selector, got %v", parsed.Outbounds[0]["type"])
	}
}
