package subscription

import (
	"reflect"
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

// Each RoutingRule shape must map to the expected Clash rule string.
func TestClashRulesMapEachShape(t *testing.T) {
	rules := []domain.RoutingRule{
		{Domains: []string{"geosite:category-ir"}, OutboundTag: "direct"},
		{IP: []string{"geoip:ir"}, OutboundTag: "direct"},
		{Domains: []string{"ads.example.com"}, OutboundTag: "blocked"},
		{Domains: []string{"keyword:torrent"}, OutboundTag: "block"},
		{IP: []string{"10.0.0.0/8"}, OutboundTag: "proxy"},
		{IP: []string{"1.2.3.4"}, OutboundTag: "proxy"}, // bare IP gets /32
		{Domains: []string{"openai.com"}, OutboundTag: "warp"},
	}
	got := clashRules(rules, "MyGroup")
	want := []string{
		"GEOSITE,category-ir,DIRECT",
		"GEOIP,ir,DIRECT",
		"DOMAIN-SUFFIX,ads.example.com,REJECT",
		"DOMAIN-KEYWORD,torrent,REJECT",
		"IP-CIDR,10.0.0.0/8,MyGroup",
		"IP-CIDR,1.2.3.4/32,MyGroup",
		"DOMAIN-SUFFIX,openai.com,MyGroup",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("clashRules mismatch:\n got=%v\nwant=%v", got, want)
	}
}

// Each RoutingRule shape must map to the expected sing-box route.rule object.
func TestSingboxRulesMapEachShape(t *testing.T) {
	rules := []domain.RoutingRule{
		{Domains: []string{"geosite:category-ir"}, OutboundTag: "direct"},
		{IP: []string{"geoip:ir"}, OutboundTag: "direct"},
		{Domains: []string{"ads.example.com"}, OutboundTag: "blocked"},
		{Domains: []string{"keyword:torrent"}, OutboundTag: "block"},
		{IP: []string{"10.0.0.0/8"}, OutboundTag: "proxy"},
		{Domains: []string{"openai.com"}, OutboundTag: "warp"},
	}
	got := singboxRules(rules, "MySelector")
	want := []map[string]any{
		{"geosite": []string{"category-ir"}, "outbound": "direct"},
		{"geoip": []string{"ir"}, "outbound": "direct"},
		{"domain_suffix": []string{"ads.example.com"}, "outbound": "block"},
		{"domain_keyword": []string{"torrent"}, "outbound": "block"},
		{"ip_cidr": []string{"10.0.0.0/8"}, "outbound": "MySelector"},
		{"domain_suffix": []string{"openai.com"}, "outbound": "MySelector"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("singboxRules mismatch:\n got=%v\nwant=%v", got, want)
	}
}

func TestClashTargetMapping(t *testing.T) {
	cases := map[string]string{
		"direct":  "DIRECT",
		"block":   "REJECT",
		"blocked": "REJECT",
		"reject":  "REJECT",
		"warp":    "Group",
		"proxy":   "Group",
	}
	for tag, want := range cases {
		if got := clashTarget(tag, "Group"); got != want {
			t.Errorf("clashTarget(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestSingboxTargetMapping(t *testing.T) {
	cases := map[string]string{
		"direct":  "direct",
		"block":   "block",
		"blocked": "block",
		"warp":    "Sel",
	}
	for tag, want := range cases {
		if got := singboxTarget(tag, "Sel"); got != want {
			t.Errorf("singboxTarget(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestNormalizeCIDR(t *testing.T) {
	cases := map[string]string{
		"1.2.3.4":       "1.2.3.4/32",
		"10.0.0.0/8":    "10.0.0.0/8",
		"2001:db8::1":   "2001:db8::1/128",
		"2001:db8::/32": "2001:db8::/32",
	}
	for in, want := range cases {
		if got := normalizeCIDR(in); got != want {
			t.Errorf("normalizeCIDR(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSingboxNeedsBlock(t *testing.T) {
	with := []domain.RoutingRule{{Domains: []string{"x.com"}, OutboundTag: "blocked"}}
	if !singboxNeedsBlock(with) {
		t.Error("expected block needed for a blocked rule")
	}
	without := []domain.RoutingRule{{Domains: []string{"x.com"}, OutboundTag: "direct"}}
	if singboxNeedsBlock(without) {
		t.Error("did not expect block needed for a direct-only rule set")
	}
}
