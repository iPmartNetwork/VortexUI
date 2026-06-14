package xray

import (
	"testing"

	"github.com/google/uuid"

	ssproxy "github.com/xtls/xray-core/proxy/shadowsocks"
	trojanproxy "github.com/xtls/xray-core/proxy/trojan"
	vlessproxy "github.com/xtls/xray-core/proxy/vless"
	vmessproxy "github.com/xtls/xray-core/proxy/vmess"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestParseUserStat(t *testing.T) {
	cases := []struct {
		name       string
		email, dir string
		ok         bool
	}{
		{"user>>>alice>>>traffic>>>uplink", "alice", "uplink", true},
		{"user>>>bob@x>>>traffic>>>downlink", "bob@x", "downlink", true},
		{"inbound>>>tag>>>traffic>>>uplink", "", "", false},
		{"user>>>alice>>>traffic", "", "", false},
		{"garbage", "", "", false},
	}
	for _, c := range cases {
		email, dir, ok := parseUserStat(c.name)
		if ok != c.ok || email != c.email || dir != c.dir {
			t.Errorf("parseUserStat(%q) = (%q,%q,%v), want (%q,%q,%v)", c.name, email, dir, ok, c.email, c.dir, c.ok)
		}
	}
}

func TestBuildAccountPerProtocol(t *testing.T) {
	vless, vmess := uuid.New(), uuid.New()
	u := &domain.User{Proxies: domain.UserCredentials{
		VLESSUUID: vless, VMessUUID: vmess,
		TrojanPass: "tpw", ShadowsocksP: "spw", SSMethod: "chacha20-poly1305",
	}}

	if acc, err := buildAccount(domain.Inbound{Protocol: domain.ProtoVLESS, Flow: "xtls-rprx-vision"}, u); err != nil {
		t.Fatal(err)
	} else if a := acc.(*vlessproxy.Account); a.Id != vless.String() || a.Flow != "xtls-rprx-vision" || a.Encryption != "none" {
		t.Errorf("vless account wrong: %+v", a)
	}

	if acc, err := buildAccount(domain.Inbound{Protocol: domain.ProtoVMess}, u); err != nil {
		t.Fatal(err)
	} else if a := acc.(*vmessproxy.Account); a.Id != vmess.String() || a.SecuritySettings == nil {
		t.Errorf("vmess account wrong: %+v", a)
	}

	if acc, err := buildAccount(domain.Inbound{Protocol: domain.ProtoTrojan}, u); err != nil {
		t.Fatal(err)
	} else if a := acc.(*trojanproxy.Account); a.Password != "tpw" {
		t.Errorf("trojan account wrong: %+v", a)
	}

	if acc, err := buildAccount(domain.Inbound{Protocol: domain.ProtoShadowsocks}, u); err != nil {
		t.Fatal(err)
	} else if a := acc.(*ssproxy.Account); a.Password != "spw" || a.CipherType != ssproxy.CipherType_CHACHA20_POLY1305 {
		t.Errorf("ss account wrong: %+v", a)
	}

	if _, err := buildAccount(domain.Inbound{Protocol: domain.ProtoWireGuard}, u); err == nil {
		t.Error("expected error for unsupported protocol")
	}
}

func TestCipherType(t *testing.T) {
	cases := map[string]ssproxy.CipherType{
		"aes-128-gcm":       ssproxy.CipherType_AES_128_GCM,
		"aes-256-gcm":       ssproxy.CipherType_AES_256_GCM,
		"chacha20-poly1305": ssproxy.CipherType_CHACHA20_POLY1305,
		"none":              ssproxy.CipherType_NONE,
		"something-unknown": ssproxy.CipherType_AES_128_GCM, // safe default
	}
	for method, want := range cases {
		if got := cipherType(method); got != want {
			t.Errorf("cipherType(%q) = %v, want %v", method, got, want)
		}
	}
}
