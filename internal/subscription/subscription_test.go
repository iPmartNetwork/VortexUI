package subscription

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

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
		"clash-verge/1.0": FormatClash,
		"mihomo":          FormatClash,
		"sing-box 1.8":    FormatSingbox,
		"HiddifyNext/2":   FormatSingbox,
		"v2rayNG/1.8":     FormatLinks,
		"":                FormatBase64,
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
	if len(parsed.ProxyGroups) != 3 || parsed.ProxyGroups[0].Name != "MyProfile" {
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
	// 1 selector + 1 urltest + 1 fallback + 3 proxies + 1 direct = 7 outbounds.
	if len(parsed.Outbounds) != 7 {
		t.Errorf("want 7 outbounds, got %d", len(parsed.Outbounds))
	}
	if parsed.Outbounds[0]["type"] != "selector" {
		t.Errorf("first outbound should be selector, got %v", parsed.Outbounds[0]["type"])
	}
}

func realityProxy() Proxy {
	return Proxy{
		Name: "reality1", Protocol: domain.ProtoVLESS, Host: "1.2.3.4", Port: 443,
		Network: "tcp", Security: "reality", SNI: "www.microsoft.com", Flow: "xtls-rprx-vision",
		UUID:      "22222222-2222-2222-2222-222222222222",
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

// --- Task 1.4: template variables (FormatVars / Expand) ---

func TestFormatVarsKnownTokens(t *testing.T) {
	expire := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	u := &domain.User{
		Username:    "alice",
		UsedTraffic: 1024 * 1024,      // 1.00 MB
		DataLimit:   10 * 1024 * 1024, // 10.00 MB
		ExpireAt:    &expire,
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) // 10 days before expiry
	vars := formatVarsAt(u, "1.2.3.4", "2001:db8::1", now)

	cases := map[string]string{
		"{USERNAME}":       "alice",
		"{SERVER_IP}":      "1.2.3.4",
		"{SERVER_IPV6}":    "2001:db8::1",
		"{DATA_USAGE}":     "1.00 MB",
		"{DATA_LIMIT}":     "10.00 MB",
		"{DATA_LEFT}":      "9.00 MB",
		"{DAYS_LEFT}":      "10",
		"{EXPIRE_DATE}":    "2026-01-11",
		"{ADMIN_USERNAME}": "",
	}
	for token, want := range cases {
		if got := vars[token]; got != want {
			t.Errorf("FormatVars[%s] = %q, want %q", token, got, want)
		}
	}
}

func TestFormatVarsUnlimitedAndNoExpiry(t *testing.T) {
	u := &domain.User{Username: "bob"} // DataLimit 0, ExpireAt nil
	vars := formatVarsAt(u, "1.1.1.1", "", time.Now())
	for _, token := range []string{"{DATA_LIMIT}", "{DATA_LEFT}", "{DAYS_LEFT}"} {
		if vars[token] != unlimited {
			t.Errorf("FormatVars[%s] = %q, want %q", token, vars[token], unlimited)
		}
	}
	if vars["{EXPIRE_DATE}"] != "" {
		t.Errorf("expire date for never-expiring user = %q, want empty", vars["{EXPIRE_DATE}"])
	}
}

func TestExpandSubstitutesKnownTokens(t *testing.T) {
	vars := map[string]string{"{USERNAME}": "alice", "{SERVER_IP}": "1.2.3.4"}
	got := Expand("{USERNAME} @ {SERVER_IP}", vars)
	if want := "alice @ 1.2.3.4"; got != want {
		t.Errorf("Expand = %q, want %q", got, want)
	}
}

func TestExpandLeavesUnknownTokensLiteral(t *testing.T) {
	vars := map[string]string{"{USERNAME}": "alice"}
	got := Expand("{USERNAME}-{MYSTERY}-{UNCLOSED", vars)
	if want := "alice-{MYSTERY}-{UNCLOSED"; got != want {
		t.Errorf("Expand = %q, want %q (unknown/unclosed tokens must stay literal)", got, want)
	}
}

func TestExpandNoTokenUnchanged(t *testing.T) {
	const s = "de-1 plain label"
	if got := Expand(s, map[string]string{"{USERNAME}": "alice"}); got != s {
		t.Errorf("Expand = %q, want unchanged %q", got, s)
	}
}

// --- Task 1.4: additive new Proxy fields must not change zero-value output ---

func TestNewProxyFieldsAreAdditive(t *testing.T) {
	// A proxy with empty ALPN / Mux=false / empty Fragment must render exactly as
	// it did before those fields existed, across every format and share link.
	for _, p := range append(sampleProxies(), realityProxy(), httpUpgradeProxy()) {
		if link := ShareLink(p); strings.Contains(link, "alpn=") || strings.Contains(link, "fragment=") {
			t.Errorf("zero-value proxy leaked host fields into link: %s", link)
		}
	}

	clash, err := Render(FormatClash, sampleProxies(), "P")
	if err != nil {
		t.Fatal(err)
	}
	if s := string(clash); strings.Contains(s, "alpn") || strings.Contains(s, "smux") {
		t.Errorf("zero-value proxies leaked alpn/smux into clash output:\n%s", s)
	}

	singbox, err := Render(FormatSingbox, sampleProxies(), "P")
	if err != nil {
		t.Fatal(err)
	}
	if s := string(singbox); strings.Contains(s, "alpn") || strings.Contains(s, "multiplex") {
		t.Errorf("zero-value proxies leaked alpn/multiplex into singbox output:\n%s", s)
	}
}

func TestHostFieldsRenderWhenSet(t *testing.T) {
	p := Proxy{
		Name: "h1", Protocol: domain.ProtoVLESS, Host: "cdn.example.com", Port: 443,
		Network: "ws", Security: "tls", SNI: "cdn.example.com", Path: "/v",
		UUID: "44444444-4444-4444-4444-444444444444",
		ALPN: []string{"h2", "http/1.1"}, Mux: true, Fragment: "100-200,10-20,tlshello",
	}
	link := ShareLink(p)
	for _, want := range []string{"alpn=h2", "fragment="} {
		if !strings.Contains(link, want) {
			t.Errorf("vless link missing %q: %s", want, link)
		}
	}

	clash, err := Render(FormatClash, []Proxy{p}, "P")
	if err != nil {
		t.Fatal(err)
	}
	if s := string(clash); !strings.Contains(s, "alpn") || !strings.Contains(s, "smux") {
		t.Errorf("clash output missing alpn/smux:\n%s", s)
	}

	singbox, err := Render(FormatSingbox, []Proxy{p}, "P")
	if err != nil {
		t.Fatal(err)
	}
	if s := string(singbox); !strings.Contains(s, "alpn") || !strings.Contains(s, "multiplex") {
		t.Errorf("singbox output missing alpn/multiplex:\n%s", s)
	}
}

// --- Phase 2: additional output formats (xray / outline / links) ---

func vmessWSProxy() Proxy {
	return Proxy{
		Name: "vmessws", Protocol: domain.ProtoVMess, Host: "5.6.7.8", Port: 80,
		Network: "ws", Security: "none", Path: "/vm", HostHeader: "vm.example.com",
		UUID: "55555555-5555-5555-5555-555555555555",
	}
}

func TestRenderXrayJSONOutbounds(t *testing.T) {
	// A vless+tls proxy and a vmess+ws proxy exercise both vnext shapes and the
	// TLS + websocket stream settings.
	ps := []Proxy{sampleProxies()[0], vmessWSProxy()}
	body, err := Render(FormatXray, ps, "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid xray json: %v\n%s", err, body)
	}
	// 2 proxies + 1 freedom tail.
	if len(parsed.Outbounds) != 3 {
		t.Fatalf("want 3 outbounds, got %d:\n%s", len(parsed.Outbounds), body)
	}

	vless := parsed.Outbounds[0]
	if vless["protocol"] != "vless" {
		t.Errorf("outbound[0] protocol = %v, want vless", vless["protocol"])
	}
	settings, _ := vless["settings"].(map[string]any)
	vnext, _ := settings["vnext"].([]any)
	if len(vnext) != 1 {
		t.Fatalf("vless vnext malformed: %v", settings["vnext"])
	}
	first, _ := vnext[0].(map[string]any)
	if first["address"] != "1.2.3.4" {
		t.Errorf("vless address = %v, want 1.2.3.4", first["address"])
	}
	users, _ := first["users"].([]any)
	u0, _ := users[0].(map[string]any)
	if u0["id"] != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("vless user id = %v", u0["id"])
	}
	if u0["flow"] != "xtls-rprx-vision" {
		t.Errorf("vless flow = %v, want xtls-rprx-vision", u0["flow"])
	}
	stream, _ := vless["streamSettings"].(map[string]any)
	if stream["security"] != "tls" || stream["network"] != "ws" {
		t.Errorf("vless streamSettings = %v", stream)
	}
	if _, ok := stream["tlsSettings"].(map[string]any); !ok {
		t.Errorf("vless tlsSettings missing: %v", stream["tlsSettings"])
	}

	vmess := parsed.Outbounds[1]
	if vmess["protocol"] != "vmess" {
		t.Errorf("outbound[1] protocol = %v, want vmess", vmess["protocol"])
	}
	vstream, _ := vmess["streamSettings"].(map[string]any)
	if vstream["network"] != "ws" {
		t.Errorf("vmess network = %v, want ws", vstream["network"])
	}
	ws, _ := vstream["wsSettings"].(map[string]any)
	if ws == nil || ws["path"] != "/vm" {
		t.Errorf("vmess wsSettings path = %v", vstream["wsSettings"])
	}

	if parsed.Outbounds[2]["protocol"] != "freedom" {
		t.Errorf("tail outbound = %v, want freedom", parsed.Outbounds[2]["protocol"])
	}
}

func TestRenderOutlineOnlyShadowsocks(t *testing.T) {
	// sampleProxies has one shadowsocks proxy (index 2) and two non-ss proxies.
	body, err := Render(FormatOutline, sampleProxies(), "P")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	if len(lines) != 1 {
		t.Fatalf("want exactly 1 ss line, got %d:\n%s", len(lines), body)
	}
	if !strings.HasPrefix(lines[0], "ss://") {
		t.Errorf("outline line not ss://: %s", lines[0])
	}
	if strings.Contains(string(body), "vless://") || strings.Contains(string(body), "trojan://") {
		t.Errorf("outline leaked non-ss links:\n%s", body)
	}
}

func TestRenderOutlineEmptyWhenNoShadowsocks(t *testing.T) {
	body, err := Render(FormatOutline, []Proxy{sampleProxies()[0]}, "P") // vless only
	if err != nil {
		t.Fatal(err)
	}
	if len(strings.TrimSpace(string(body))) != 0 {
		t.Errorf("want empty body, got:\n%s", body)
	}
}

func TestRenderLinksEqualsDecodedBase64(t *testing.T) {
	ps := sampleProxies()
	links, err := Render(FormatLinks, ps, "P")
	if err != nil {
		t.Fatal(err)
	}
	b64, err := Render(FormatBase64, ps, "P")
	if err != nil {
		t.Fatal(err)
	}
	dec, err := base64.StdEncoding.DecodeString(string(b64))
	if err != nil {
		t.Fatalf("base64 output not decodable: %v", err)
	}
	if string(links) != string(dec) {
		t.Errorf("renderLinks != base64-decoded renderBase64:\nlinks=%q\ndec=%q", links, dec)
	}
	// And it must not itself be base64-wrapped.
	if strings.Contains(string(links), "vless://") == false {
		t.Errorf("links output should contain plain share links:\n%s", links)
	}
}

func TestDetectNewFormats(t *testing.T) {
	cases := map[string]Format{
		"Outline/1.2 (client)": FormatOutline,
		"v2rayNG/1.8":          FormatLinks,
		"v2rayN/6.0":           FormatLinks,
		"some-unknown-client":  FormatBase64,
		"clash-verge":          FormatClash,
		"sing-box":             FormatSingbox,
	}
	for ua, want := range cases {
		if got := Detect(ua); got != want {
			t.Errorf("Detect(%q) = %s, want %s", ua, got, want)
		}
	}
}

func TestContentTypeNewFormats(t *testing.T) {
	cases := map[Format]string{
		FormatXray:    "application/json; charset=utf-8",
		FormatOutline: "text/plain; charset=utf-8",
		FormatLinks:   "text/plain; charset=utf-8",
	}
	for f, want := range cases {
		if got := f.ContentType(); got != want {
			t.Errorf("%s.ContentType() = %q, want %q", f, got, want)
		}
	}
}

// Regression: the existing formats must produce byte-identical output to a
// direct call of their renderer (Requirement 2.5).
func TestExistingFormatsUnchanged(t *testing.T) {
	ps := append(sampleProxies(), realityProxy(), httpUpgradeProxy())

	b64, err := Render(FormatBase64, ps, "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	if string(b64) != string(renderBase64(ps)) {
		t.Errorf("FormatBase64 output diverged from renderBase64")
	}

	clashDirect, err := renderClash(ps, "VortexUI", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	clashRender, err := Render(FormatClash, ps, "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	if string(clashRender) != string(clashDirect) {
		t.Errorf("FormatClash output diverged from renderClash")
	}

	sbDirect, err := renderSingbox(ps, "VortexUI", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	sbRender, err := Render(FormatSingbox, ps, "VortexUI")
	if err != nil {
		t.Fatal(err)
	}
	if string(sbRender) != string(sbDirect) {
		t.Errorf("FormatSingbox output diverged from renderSingbox")
	}
}

// Task 2.3: the new formats take []Proxy exactly like the others, so a
// host-projected slice flows through unchanged with one entry per proxy.
func TestNewFormatsConsumeProjectedProxiesUnchanged(t *testing.T) {
	// Simulate a host-projected slice: two entries derived from one inbound.
	projected := []Proxy{
		{Name: "alice @ cdn-a", Protocol: domain.ProtoShadowsocks, Host: "a.cdn", Port: 8388, Password: "pw", SSMethod: "aes-128-gcm"},
		{Name: "alice @ cdn-b", Protocol: domain.ProtoShadowsocks, Host: "b.cdn", Port: 8388, Password: "pw", SSMethod: "aes-128-gcm"},
	}
	out, err := Render(FormatOutline, projected, "P")
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 ss entries for 2 projected proxies, got %d:\n%s", len(lines), out)
	}

	links, err := Render(FormatLinks, projected, "P")
	if err != nil {
		t.Fatal(err)
	}
	if n := strings.Count(strings.TrimSpace(string(links)), "\n") + 1; n != 2 {
		t.Errorf("want 2 plain links, got %d:\n%s", n, links)
	}
}

// --- Phase 3.3: routing-pack embedding into Clash / sing-box ---

// samplePackRules is a 2-rule pack: Iran domains direct + ads blocked, covering
// a direct target, a reject target, a geosite token and a plain domain.
func samplePackRules() []domain.RoutingRule {
	return []domain.RoutingRule{
		{Name: "iran-direct", Domains: []string{"geosite:category-ir"}, OutboundTag: "direct", Priority: 1, Enabled: true},
		{Name: "block-ads", Domains: []string{"ads.example.com"}, OutboundTag: "blocked", Priority: 2, Enabled: true},
	}
}

// With no rules, RenderWith must produce byte-identical output to today's Render
// for every format (Requirement 3.3.3 — no regression).
func TestRenderWithEmptyRulesMatchesRender(t *testing.T) {
	ps := append(sampleProxies(), realityProxy(), httpUpgradeProxy())
	for _, f := range []Format{FormatBase64, FormatClash, FormatSingbox, FormatXray, FormatOutline, FormatLinks} {
		legacy, err := Render(f, ps, "VortexUI")
		if err != nil {
			t.Fatalf("Render(%s): %v", f, err)
		}
		with, err := RenderWith(f, ps, RenderOpts{Title: "VortexUI"})
		if err != nil {
			t.Fatalf("RenderWith(%s): %v", f, err)
		}
		if string(legacy) != string(with) {
			t.Errorf("format %s: RenderWith(empty rules) diverged from Render", f)
		}
	}
}

// Non-routing formats must ignore rules entirely (Requirement 3.3.2): output is
// identical whether or not a pack is supplied.
func TestRenderWithRulesIgnoredByNonRoutingFormats(t *testing.T) {
	ps := sampleProxies()
	for _, f := range []Format{FormatBase64, FormatLinks, FormatXray, FormatOutline} {
		none, err := RenderWith(f, ps, RenderOpts{Title: "P"})
		if err != nil {
			t.Fatalf("RenderWith(%s, none): %v", f, err)
		}
		withRules, err := RenderWith(f, ps, RenderOpts{Title: "P", Rules: samplePackRules()})
		if err != nil {
			t.Fatalf("RenderWith(%s, rules): %v", f, err)
		}
		if string(none) != string(withRules) {
			t.Errorf("format %s changed when rules supplied; non-routing formats must ignore rules", f)
		}
	}
}

// Clash output with a 2-rule pack embeds the mapped rules and ends with the
// MATCH fallback to the selector group (Requirement 3.3.1).
func TestRenderClashEmbedsPackRules(t *testing.T) {
	body, err := RenderWith(FormatClash, sampleProxies(), RenderOpts{Title: "MyProfile", Rules: samplePackRules()})
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Rules []string `yaml:"rules"`
	}
	if err := yaml.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid clash yaml: %v\n%s", err, body)
	}
	want := []string{
		"GEOSITE,category-ir,DIRECT",
		"DOMAIN-SUFFIX,ads.example.com,REJECT",
		"MATCH,MyProfile",
	}
	if len(parsed.Rules) != len(want) {
		t.Fatalf("clash rules = %v, want %v", parsed.Rules, want)
	}
	for i := range want {
		if parsed.Rules[i] != want[i] {
			t.Errorf("clash rule[%d] = %q, want %q", i, parsed.Rules[i], want[i])
		}
	}
}

// sing-box output with a 2-rule pack embeds route.rules + a route.final selector
// and appends a block outbound the reject rule can reference (Requirement 3.3.1).
func TestRenderSingboxEmbedsPackRules(t *testing.T) {
	body, err := RenderWith(FormatSingbox, sampleProxies(), RenderOpts{Title: "MyProfile", Rules: samplePackRules()})
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
		Route     struct {
			Rules []map[string]any `json:"rules"`
			Final string           `json:"final"`
		} `json:"route"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid singbox json: %v\n%s", err, body)
	}
	if parsed.Route.Final != "MyProfile" {
		t.Errorf("route.final = %q, want MyProfile", parsed.Route.Final)
	}
	if len(parsed.Route.Rules) != 2 {
		t.Fatalf("want 2 route rules, got %d: %v", len(parsed.Route.Rules), parsed.Route.Rules)
	}
	// First rule: geosite → direct.
	if parsed.Route.Rules[0]["outbound"] != "direct" {
		t.Errorf("rule[0] outbound = %v, want direct", parsed.Route.Rules[0]["outbound"])
	}
	if _, ok := parsed.Route.Rules[0]["geosite"]; !ok {
		t.Errorf("rule[0] missing geosite matcher: %v", parsed.Route.Rules[0])
	}
	// Second rule: domain_suffix → block.
	if parsed.Route.Rules[1]["outbound"] != "block" {
		t.Errorf("rule[1] outbound = %v, want block", parsed.Route.Rules[1]["outbound"])
	}
	// A block outbound must exist so the reject rule references a real outbound.
	var hasBlock bool
	for _, o := range parsed.Outbounds {
		if o["type"] == "block" && o["tag"] == "block" {
			hasBlock = true
		}
	}
	if !hasBlock {
		t.Errorf("block outbound missing; reject rule would dangle:\n%s", body)
	}
}

// --- ECH (Encrypted Client Hello) support ---

func echProxy() Proxy {
	return Proxy{
		Name: "ech1", Protocol: domain.ProtoVLESS, Host: "1.2.3.4", Port: 443,
		Network: "tcp", Security: "tls", SNI: "ex.com",
		UUID: "66666666-6666-6666-6666-666666666666",
		ECH:  true,
	}
}

func TestECHSingboxRendersECHBlock(t *testing.T) {
	body, err := Render(FormatSingbox, []Proxy{echProxy()}, "P")
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Outbounds []map[string]any `json:"outbounds"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("invalid json: %v\n%s", err, body)
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
	tls, _ := vless["tls"].(map[string]any)
	if tls == nil {
		t.Fatal("tls block missing")
	}
	ech, _ := tls["ech"].(map[string]any)
	if ech == nil {
		t.Fatal("ech block missing in tls")
	}
	if ech["enabled"] != true {
		t.Errorf("ech.enabled = %v, want true", ech["enabled"])
	}
}

func TestECHFalseDoesNotRenderECHBlock(t *testing.T) {
	p := echProxy()
	p.ECH = false
	body, err := Render(FormatSingbox, []Proxy{p}, "P")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), `"ech"`) {
		t.Errorf("ECH=false should not emit ech block:\n%s", body)
	}
}
