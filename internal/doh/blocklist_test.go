package doh

import "testing"

func TestBlockMatcherAdsAndMalware(t *testing.T) {
	m := NewBlockMatcher(nil, true, true)
	if !m.Blocked("ads.doubleclick.net") {
		t.Fatal("expected ads block")
	}
	if !m.Blocked("login.phishing.com") {
		t.Fatal("expected malware block")
	}
	if m.Blocked("example.com") {
		t.Fatal("expected allow")
	}
}

func TestBlockMatcherCustom(t *testing.T) {
	m := NewBlockMatcher([]string{"blocked.test"}, false, false)
	if !m.Blocked("sub.blocked.test") {
		t.Fatal("expected custom block")
	}
}
