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
	if len(parsed.ProxyGroups) != 2 || parsed.ProxyGroups[0].Name != "MyProfile" {
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
	// 1 selector + 1 urltest + 3 proxies + 1 direct = 6 outbounds.
	if len(parsed.Outbounds) != 6 {
		t.Errorf("want 6 outbounds, got %d", len(parsed.Outbounds))
	}
	if parsed.Outbounds[0]["type"] != "selector" {
		t.Errorf("first outbound should be selector, got %v", parsed.Outbounds[0]["type"])
	}
}

func realityProxy() Proxy {
	return Proxy{
		Name: "reality1", Protocol: domain.ProtoVLESS, Host: "1.2.3.4", Port: 443,
		Network: "tcp", Security: "reality", SNI: "www.microsoft.com", Flow: "xtls-rprx-vision",
		UUID: "22222222-2222-2222-2222-222222222222",
		PublicKey: "PUBKEY123", ShortID: "abcd1234", Fingerprint: "chrome",
	}
}

func TestRealityShareLinkCarriesKeyMaterial(t *testing.T) {
	link := ShareLink(realityProxy())
	for _, want := range []string{"security=reality", "pbk=PUBKEY123", "sid=abcd1234", "fp=chrome", "flow=xtls-rprx-vision", "sni=www.microsoft.com"} {
		if !strings.Contains(link, want) {
			t.Errorf("reality vless link missing %q: %s", want, link)
		}
	}
}

func TestRealityClashCarriesRealityOpts(t *testing.T) {
	body, err := Render(FormatClash, []Proxy{realityProxy()}, "P")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid yaml: %v", err)
	}
	if len(parsed.Proxies) != 1 {
		t.Fatalf("want 1 proxy, got %d", len(parsed.Proxies))
	}
	ro, ok := parsed.Proxies[0]["reality-opts"].(map[string]any)
	if !ok || ro["public-key"] != "PUBKEY123" || ro["short-id"] != "abcd1234" {
		t.Errorf("reality-opts missing/wrong: %v", parsed.Proxies[0]["reality-opts"])
	}
	if parsed.Proxies[0]["client-fingerprint"] != "chrome" {
		t.Errorf("client-fingerprint = %v", parsed.Proxies[0]["client-fingerprint"])
	}
}

func TestRealitySingboxCarriesRealityBlock(t *testing.T) {
	body, err := Render(FormatSingbox, []Proxy{realityProxy()}, "P")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	// outbounds[0] is the selector; the proxy outbound follows.
	var vless map[string]any
	for _, o := range parsed.Outbounds {
		if o["type"] == "vless" {
			vless = o
		}
	}
	if vless == nil {
		t.Fatal("vless outbound missing")
	}
	tls, _ := vless["tls"].(map[string]any)
	rl, _ := tls["reality"].(map[string]any)
	if rl == nil || rl["public_key"] != "PUBKEY123" || rl["short_id"] != "abcd1234" {
		t.Errorf("reality block missing/wrong: %v", tls["reality"])
	}
	utls, _ := tls["utls"].(map[string]any)
	if utls == nil || utls["fingerprint"] != "chrome" {
		t.Errorf("utls fingerprint missing: %v", tls["utls"])
	}
}

func httpUpgradeProxy() Proxy {
	return Proxy{
		Name: "hu1", Protocol: domain.ProtoVLESS, Host: "1.2.3.4", Port: 443,
		Network: "httpupgrade", Security: "tls", SNI: "ex.com", Path: "/hu", HostHeader: "ex.com",
		UUID: "33333333-3333-3333-3333-333333333333",
	}
}

func TestHTTPUpgradeClashRendersAsWSUpgrade(t *testing.T) {
	body, err := Render(FormatClash, []Proxy{httpUpgradeProxy()}, "P")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Proxies []map[string]any `yaml:"proxies"`
	}
	if err := yaml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid yaml: %v", err)
	}
	if len(parsed.Proxies) != 1 {
		t.Fatalf("want 1 proxy, got %d", len(parsed.Proxies))
	}
	if parsed.Proxies[0]["network"] != "ws" {
		t.Errorf("network = %v, want ws", parsed.Proxies[0]["network"])
	}
	opts, ok := parsed.Proxies[0]["ws-opts"].(map[string]any)
	if !ok {
		t.Fatalf("ws-opts missing: %v", parsed.Proxies[0]["ws-opts"])
	}
	if opts["v2ray-http-upgrade"] != true {
		t.Errorf("v2ray-http-upgrade = %v, want true", opts["v2ray-http-upgrade"])
	}
	if opts["path"] != "/hu" {
		t.Errorf("path = %v, want /hu", opts["path"])
	}
}

func TestHTTPUpgradeSingboxRendersHTTPUpgradeTransport(t *testing.T) {
	body, err := Render(FormatSingbox, []Proxy{httpUpgradeProxy()}, "P")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	var vless map[string]any
	for _, o := range parsed.Outbounds {
		if o["type"] == "vless" {
			vless = o
		}
	}
	if vless == nil {
		t.Fatal("vless outbound missing")
	}
	tr, ok := vless["transport"].(map[string]any)
	if !ok {
		t.Fatalf("transport missing: %v", vless["transport"])
	}
	if tr["type"] != "httpupgrade" {
		t.Errorf("transport type = %v, want httpupgrade", tr["type"])
	}
	if tr["path"] != "/hu" {
		t.Errorf("path = %v, want /hu", tr["path"])
	}
	if tr["host"] != "ex.com" {
		t.Errorf("host = %v, want ex.com", tr["host"])
	}
}
