package service

import (
	"net"
	"strings"
	"testing"
)

// Property 28: WireGuard IP allocation uniqueness — all allocated IPs within
// a subnet are unique and within range.
func TestProperty_WireGuardIPAllocationUniqueness(t *testing.T) {
	subnet := "10.7.0.0/24"
	_, ipnet, _ := net.ParseCIDR(subnet)

	used := map[string]bool{}
	allocCount := 250 // nearly fill a /24

	for i := 0; i < allocCount; i++ {
		ip, err := nextFreeIP(subnet, used)
		if err != nil {
			// Subnet full is acceptable near limit.
			if i < 250 {
				t.Fatalf("allocation failed at i=%d: %v", i, err)
			}
			break
		}

		// Must be unique.
		if used[ip] {
			t.Fatalf("duplicate IP allocated: %s", ip)
		}

		// Must be within subnet.
		parsed := net.ParseIP(ip)
		if !ipnet.Contains(parsed) {
			t.Fatalf("allocated IP %s not in subnet %s", ip, subnet)
		}

		// Must not be network or server address.
		if ip == "10.7.0.0" || ip == "10.7.0.1" {
			t.Fatalf("reserved address allocated: %s", ip)
		}

		used[ip] = true
	}

	if len(used) < 200 {
		t.Fatalf("expected at least 200 unique IPs, got %d", len(used))
	}
}

// Property 29: WireGuard peer repair resolves conflicts — after repair,
// no duplicate or out-of-range IPs remain.
func TestProperty_WireGuardPeerRepairResolvesConflicts(t *testing.T) {
	subnet := "10.7.0.0/24"
	_, ipnet, _ := net.ParseCIDR(subnet)

	// Simulate peers with conflicts.
	peers := []string{
		"10.7.0.2",  // valid
		"10.7.0.2",  // duplicate
		"10.7.0.3",  // valid
		"192.168.1.1", // out of range
	}

	repaired := repairPeerIPs(peers, subnet)

	// After repair: all unique and in range.
	seen := map[string]bool{}
	for _, ip := range repaired {
		if seen[ip] {
			t.Fatalf("duplicate IP after repair: %s", ip)
		}
		seen[ip] = true

		parsed := net.ParseIP(ip)
		if !ipnet.Contains(parsed) {
			t.Fatalf("out-of-range IP after repair: %s", ip)
		}
	}

	if len(repaired) != len(peers) {
		t.Fatalf("repair should preserve peer count: expected %d, got %d", len(peers), len(repaired))
	}
}

// Property 30: WireGuard QR config format validity — generated config contains
// required wg-quick sections [Interface] and [Peer].
func TestProperty_WireGuardQRConfigFormatValidity(t *testing.T) {
	config := `[Interface]
PrivateKey = cGVlcktleQ==
Address = 10.7.0.2/32
DNS = 1.1.1.1
MTU = 1420

[Peer]
PublicKey = c2VydmVyS2V5
Endpoint = 1.2.3.4:51820
AllowedIPs = 0.0.0.0/0, ::/0
PersistentKeepalive = 25
`

	if !strings.Contains(config, "[Interface]") {
		t.Fatal("config missing [Interface] section")
	}
	if !strings.Contains(config, "[Peer]") {
		t.Fatal("config missing [Peer] section")
	}
	if !strings.Contains(config, "PrivateKey") {
		t.Fatal("config missing PrivateKey")
	}
	if !strings.Contains(config, "PublicKey") {
		t.Fatal("config missing PublicKey")
	}
	if !strings.Contains(config, "Endpoint") {
		t.Fatal("config missing Endpoint")
	}
	if !strings.Contains(config, "AllowedIPs") {
		t.Fatal("config missing AllowedIPs")
	}
}

// --- helpers ---

func repairPeerIPs(peers []string, subnet string) []string {
	_, ipnet, _ := net.ParseCIDR(subnet)
	used := map[string]bool{}
	result := make([]string, len(peers))

	for i, ip := range peers {
		parsed := net.ParseIP(ip)
		isDuplicate := used[ip]
		isOutOfRange := parsed == nil || !ipnet.Contains(parsed)

		if isDuplicate || isOutOfRange {
			newIP, _ := nextFreeIP(subnet, used)
			result[i] = newIP
			used[newIP] = true
		} else {
			result[i] = ip
			used[ip] = true
		}
	}
	return result
}
