package geoip

import "testing"

// A disabled resolver (empty path) must open without error and yield "" for any
// lookup, so geo features degrade gracefully when no DB is configured.
func TestDisabledResolverDegradesGracefully(t *testing.T) {
	r, err := Open("")
	if err != nil {
		t.Fatalf("Open(\"\") returned error: %v", err)
	}
	if r.Enabled() {
		t.Fatal("resolver with empty path should be disabled")
	}
	for _, ip := range []string{"8.8.8.8", "2001:4860:4860::8888", "not-an-ip", ""} {
		if got := r.Country(ip); got != "" {
			t.Errorf("Country(%q) = %q, want \"\"", ip, got)
		}
	}
	if err := r.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// A non-empty path that cannot be opened must surface an error so the caller can
// log it and fall back to a disabled resolver.
func TestOpenMissingFileReturnsError(t *testing.T) {
	if _, err := Open("does-not-exist.mmdb"); err == nil {
		t.Fatal("Open of a missing file should return an error")
	}
}

// The nil pointer is a valid disabled resolver.
func TestNilResolverIsSafe(t *testing.T) {
	var r *Resolver
	if r.Enabled() {
		t.Fatal("nil resolver should report disabled")
	}
	if got := r.Country("8.8.8.8"); got != "" {
		t.Errorf("nil resolver Country = %q, want \"\"", got)
	}
	if err := r.Close(); err != nil {
		t.Errorf("nil resolver Close = %v, want nil", err)
	}
}

// Disabled() is equivalent to an empty-path Open.
func TestDisabledConstructor(t *testing.T) {
	if Disabled().Enabled() {
		t.Fatal("Disabled() must not be enabled")
	}
}
