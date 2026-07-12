package core

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestProbeBlockRules(t *testing.T) {
	nodeID := uuid.New()
	rules := ProbeBlockRules(nodeID, []string{"1.2.3.4"}, []string{"in-443"})
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].OutboundTag != "blocked" {
		t.Fatalf("expected blocked outbound, got %q", rules[0].OutboundTag)
	}
}

func TestDecoyFallbackDest(t *testing.T) {
	in := domain.Inbound{
		Raw: map[string]any{
			"decoy": map[string]any{
				"mode":       "proxy",
				"target_url": "https://example.com",
			},
		},
	}
	if got := DecoyFallbackDest(in); got != "example.com:443" {
		t.Fatalf("dest = %q, want example.com:443", got)
	}
}

func TestActiveBlockedIPs(t *testing.T) {
	now := time.Now()
	ips := ActiveBlockedIPs([]domain.BlockedIP{
		{IP: "1.1.1.1", ExpiresAt: now.Add(time.Hour)},
		{IP: "2.2.2.2", ExpiresAt: now.Add(-time.Hour)},
	}, now)
	if len(ips) != 1 || ips[0] != "1.1.1.1" {
		t.Fatalf("unexpected active ips: %#v", ips)
	}
}
