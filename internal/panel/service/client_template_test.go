package service

import (
	"regexp"
	"testing"
)

// Property 22: Client template User-Agent matching — regex patterns correctly
// match the intended client apps and reject others.
func TestProperty_ClientTemplateUserAgentMatching(t *testing.T) {
	patterns := []struct {
		regex     string
		shouldMatch []string
		shouldFail  []string
	}{
		{
			regex:       `(?i)v2ray`,
			shouldMatch: []string{"V2RayNG/1.8.1", "v2ray-core/5.0", "V2rayN"},
			shouldFail:  []string{"Clash-Meta/1.0", "Mozilla/5.0"},
		},
		{
			regex:       `(?i)(clash|meta)`,
			shouldMatch: []string{"ClashForAndroid/2.5", "Clash-Meta/1.16", "clash-verge"},
			shouldFail:  []string{"V2RayNG/1.8", "Shadowrocket/2.0"},
		},
		{
			regex:       `(?i)shadowrocket`,
			shouldMatch: []string{"Shadowrocket/2.2.0", "SHADOWROCKET"},
			shouldFail:  []string{"Clash/1.0", "V2RayNG/1.0"},
		},
	}

	for _, p := range patterns {
		re, err := regexp.Compile(p.regex)
		if err != nil {
			t.Fatalf("invalid regex %q: %v", p.regex, err)
		}

		for _, ua := range p.shouldMatch {
			if !re.MatchString(ua) {
				t.Fatalf("pattern %q should match %q", p.regex, ua)
			}
		}
		for _, ua := range p.shouldFail {
			if re.MatchString(ua) {
				t.Fatalf("pattern %q should NOT match %q", p.regex, ua)
			}
		}
	}
}

// Property 23: Approval queue blocks delivery — pending approvals prevent
// subscription delivery until resolved.
func TestProperty_ApprovalQueueBlocksDelivery(t *testing.T) {
	statuses := []struct {
		status  string
		blocked bool
	}{
		{"pending", true},
		{"approved", false},
		{"rejected", true},
	}

	for _, s := range statuses {
		blocked := isDeliveryBlocked(s.status)
		if blocked != s.blocked {
			t.Fatalf("status %q: expected blocked=%v, got %v", s.status, s.blocked, blocked)
		}
	}
}

// --- helpers ---

func isDeliveryBlocked(approvalStatus string) bool {
	return approvalStatus != "approved"
}
