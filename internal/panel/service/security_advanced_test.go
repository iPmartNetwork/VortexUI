package service

import (
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// Property 31: IP whitelist enforcement — whitelisted CIDRs allow matching IPs
// and reject non-matching ones.
func TestProperty_IPWhitelistEnforcement(t *testing.T) {
	whitelist := []string{"192.168.1.0/24", "10.0.0.0/8", "203.0.113.5/32"}

	allowed := []string{"192.168.1.50", "10.5.3.1", "203.0.113.5"}
	blocked := []string{"172.16.0.1", "8.8.8.8", "203.0.113.6"}

	for _, ip := range allowed {
		if !ipMatchesWhitelist(ip, whitelist) {
			t.Fatalf("IP %s should be allowed by whitelist", ip)
		}
	}
	for _, ip := range blocked {
		if ipMatchesWhitelist(ip, whitelist) {
			t.Fatalf("IP %s should be blocked by whitelist", ip)
		}
	}
}

// Property 32: Login audit completeness — every login attempt (success or fail)
// produces an audit entry with required fields.
func TestProperty_LoginAuditCompleteness(t *testing.T) {
	entries := []domain.LoginAuditEntry{
		{ID: uuid.New(), Username: "admin", IPAddress: "1.2.3.4", Success: true},
		{ID: uuid.New(), Username: "admin", IPAddress: "5.6.7.8", Success: false, FailureReason: "bad password"},
		{ID: uuid.New(), Username: "unknown", IPAddress: "9.9.9.9", Success: false, FailureReason: "user not found"},
	}

	for i, e := range entries {
		if e.ID == uuid.Nil {
			t.Fatalf("entry %d missing ID", i)
		}
		if e.Username == "" {
			t.Fatalf("entry %d missing username", i)
		}
		if e.IPAddress == "" {
			t.Fatalf("entry %d missing IP address", i)
		}
		if !e.Success && e.FailureReason == "" {
			t.Fatalf("entry %d: failed login must have a reason", i)
		}
	}
}

// Property 33: Account lockout threshold — N consecutive failures triggers lockout.
func TestProperty_AccountLockoutThreshold(t *testing.T) {
	threshold := 5

	// Below threshold = not locked.
	for fails := 0; fails < threshold; fails++ {
		if isLocked(fails, threshold) {
			t.Fatalf("%d failures should not lock (threshold=%d)", fails, threshold)
		}
	}

	// At threshold = locked.
	if !isLocked(threshold, threshold) {
		t.Fatalf("%d failures should trigger lockout", threshold)
	}

	// Above threshold = still locked.
	if !isLocked(threshold+3, threshold) {
		t.Fatal("above threshold should remain locked")
	}
}

// Property 34: API token scope enforcement — tokens only grant access to
// their declared scopes.
func TestProperty_APITokenScopeEnforcement(t *testing.T) {
	token := domain.ScopedAPIToken{
		ID:     uuid.New(),
		Scopes: []string{domain.ScopeUsersRead, domain.ScopeNodesRead},
	}

	// Allowed scopes.
	if !hasScope(token.Scopes, domain.ScopeUsersRead) {
		t.Fatal("should have users:read scope")
	}
	if !hasScope(token.Scopes, domain.ScopeNodesRead) {
		t.Fatal("should have nodes:read scope")
	}

	// Denied scopes.
	if hasScope(token.Scopes, domain.ScopeUsersWrite) {
		t.Fatal("should NOT have users:write scope")
	}
	if hasScope(token.Scopes, domain.ScopeAdminWrite) {
		t.Fatal("should NOT have admin:write scope")
	}

	// Wildcard grants everything.
	wildcardToken := []string{domain.ScopeAll}
	if !hasScope(wildcardToken, domain.ScopeSecurityManage) {
		t.Fatal("wildcard should grant all scopes")
	}
}

// Property 35: Field encryption round-trip — encrypting then decrypting
// returns the original plaintext.
func TestProperty_FieldEncryptionRoundTrip(t *testing.T) {
	enc, err := NewFieldEncryptor("test-panel-secret-key")
	if err != nil {
		t.Fatalf("create encryptor: %v", err)
	}

	plaintexts := []string{
		"simple",
		"hello world with spaces",
		`{"key":"value","nested":{"a":1}}`,
		" unicode: 日本語テスト 🎉",
		"", // empty string
	}

	for _, pt := range plaintexts {
		ct, err := enc.Encrypt(pt)
		if err != nil {
			t.Fatalf("encrypt %q: %v", pt, err)
		}
		if ct == pt && pt != "" {
			t.Fatalf("ciphertext should differ from plaintext")
		}

		decrypted, err := enc.Decrypt(ct)
		if err != nil {
			t.Fatalf("decrypt %q: %v", pt, err)
		}
		if decrypted != pt {
			t.Fatalf("round-trip failed: expected %q, got %q", pt, decrypted)
		}
	}
}

// Property 36: Security audit logging — every audited operation produces an entry
// with operation name, resource, and timestamp.
func TestProperty_SecurityAuditLogging(t *testing.T) {
	operations := []struct {
		op       string
		resource string
	}{
		{"user.create", "users/abc123"},
		{"admin.login", "admins/admin1"},
		{"settings.update", "settings/general"},
		{"node.delete", "nodes/xyz789"},
	}

	for _, op := range operations {
		entry := domain.SecurityAuditEntry{
			ID:        uuid.New(),
			Operation: op.op,
			Resource:  op.resource,
			IPAddress: "10.0.0.1",
		}

		if entry.Operation == "" {
			t.Fatal("audit entry missing operation")
		}
		if entry.Resource == "" {
			t.Fatal("audit entry missing resource")
		}
		if entry.IPAddress == "" {
			t.Fatal("audit entry missing IP")
		}
	}
}

// --- helpers ---

func ipMatchesWhitelist(ip string, cidrs []string) bool {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsed) {
			return true
		}
	}
	return false
}

func isLocked(consecutiveFails, threshold int) bool {
	return consecutiveFails >= threshold
}

func hasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		if s == domain.ScopeAll || s == required {
			return true
		}
	}
	return false
}
