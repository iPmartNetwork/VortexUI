package security

import "testing"

func TestParseLineDetectsRejectedIP(t *testing.T) {
	ev, ok := parseLine(`2026/01/01 [Warning] rejected connection from 203.0.113.10:443 invalid user`)
	if !ok {
		t.Fatal("expected probe parse")
	}
	if ev.SourceIP != "203.0.113.10" || ev.Method != "tls_probe" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestParseLineIgnoresNormalTraffic(t *testing.T) {
	if _, ok := parseLine(`accepted connection for user@example.com`); ok {
		t.Fatal("should ignore normal line")
	}
}
