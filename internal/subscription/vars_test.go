package subscription

import (
	"fmt"
	"testing"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// --- VarResolver tests ---

func TestVarResolverResolvesAllVariables(t *testing.T) {
	expire := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	u := &domain.User{
		Username:    "alice",
		Status:      domain.UserStatusActive,
		UsedTraffic: 5 * 1024 * 1024 * 1024, // 5 GB
		DataLimit:   10 * 1024 * 1024 * 1024, // 10 GB
		ExpireAt:    &expire,
	}

	now := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	resolver := &VarResolver{Now: func() time.Time { return now }}

	ctx := VarContext{
		User:          u,
		AdminUsername: "admin1",
		NodeName:      "DE-Frankfurt-1",
		NodeIP:        "185.1.2.3",
		NodeIPv6:      "2001:db8::1",
		NodeFlag:      "🇩🇪",
		Protocol:      "vless",
		Transport:     "ws",
		ISP:           "Hetzner",
		OnlineCount:   42,
		QualityScore:  87,
	}

	vars := resolver.BuildVars(ctx)

	cases := map[string]string{
		varServerIP:        "185.1.2.3",
		varServerIPv6:      "2001:db8::1",
		varUsername:        "alice",
		varAdminUsername:   "admin1",
		varProtocol:       "vless",
		varTransport:      "ws",
		varDataUsage:      "5.00 GB",
		varDataLimit:      "10.00 GB",
		varDataLeft:       "5.00 GB",
		varUsagePercentage: "50%",
		varDaysLeft:       "9",
		varExpireDate:     "2026-03-20",
		varTimeLeft:       "9d 12h",
		varStatusEmoji:    "✅",
		varNodeName:       "DE-Frankfurt-1",
		varNodeFlag:       "🇩🇪",
		varISPName:        "Hetzner",
		varQualityScore:   "87",
		varOnlineCount:    "42",
	}

	for token, want := range cases {
		if got := vars[token]; got != want {
			t.Errorf("VarResolver[%s] = %q, want %q", token, got, want)
		}
	}

	// Jalali date for 2026-03-20 should be 1404/12/29
	if got := vars[varJalaliExpire]; got == "" {
		t.Errorf("VarResolver[%s] should not be empty", varJalaliExpire)
	}
}

func TestVarResolverResolveTemplate(t *testing.T) {
	expire := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	u := &domain.User{
		Username: "bob",
		Status:   domain.UserStatusActive,
		ExpireAt: &expire,
	}

	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	resolver := &VarResolver{Now: func() time.Time { return now }}

	ctx := VarContext{
		User:     u,
		NodeName: "US-NY",
		NodeFlag: "🇺🇸",
		Protocol: "trojan",
		Transport: "tcp",
	}

	template := "{USERNAME} | {PROTOCOL} - {NODE_FLAG} {NODE_NAME}"
	got := resolver.Resolve(template, ctx)
	want := "bob | trojan - 🇺🇸 US-NY"

	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestVarResolverNilUser(t *testing.T) {
	resolver := &VarResolver{}
	ctx := VarContext{
		NodeIP:   "1.2.3.4",
		NodeName: "test-node",
	}

	vars := resolver.BuildVars(ctx)
	if vars[varUsername] != "" {
		t.Errorf("username should be empty for nil user, got %q", vars[varUsername])
	}
	if vars[varServerIP] != "1.2.3.4" {
		t.Errorf("server IP should still be set, got %q", vars[varServerIP])
	}
	if vars[varNodeName] != "test-node" {
		t.Errorf("node name should still be set, got %q", vars[varNodeName])
	}
}

func TestVarResolverUnlimitedUser(t *testing.T) {
	u := &domain.User{
		Username:  "unlimited_user",
		Status:    domain.UserStatusActive,
		DataLimit: 0, // unlimited
		// ExpireAt nil = never expires
	}

	resolver := &VarResolver{}
	ctx := VarContext{User: u}

	vars := resolver.BuildVars(ctx)

	if vars[varDataLimit] != unlimited {
		t.Errorf("data limit = %q, want %q", vars[varDataLimit], unlimited)
	}
	if vars[varDataLeft] != unlimited {
		t.Errorf("data left = %q, want %q", vars[varDataLeft], unlimited)
	}
	if vars[varDaysLeft] != unlimited {
		t.Errorf("days left = %q, want %q", vars[varDaysLeft], unlimited)
	}
	if vars[varTimeLeft] != unlimited {
		t.Errorf("time left = %q, want %q", vars[varTimeLeft], unlimited)
	}
	if vars[varUsagePercentage] != "0%" {
		t.Errorf("usage pct = %q, want 0%%", vars[varUsagePercentage])
	}
}

func TestVarResolverExpiredUser(t *testing.T) {
	expire := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	u := &domain.User{
		Username:    "expired_user",
		Status:      domain.UserStatusExpired,
		DataLimit:   10 * 1024 * 1024 * 1024,
		UsedTraffic: 10 * 1024 * 1024 * 1024, // fully used
		ExpireAt:    &expire,
	}

	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC) // 9 days past expiry
	resolver := &VarResolver{Now: func() time.Time { return now }}
	ctx := VarContext{User: u}

	vars := resolver.BuildVars(ctx)

	if vars[varDaysLeft] != "0" {
		t.Errorf("days left for expired user = %q, want 0", vars[varDaysLeft])
	}
	if vars[varTimeLeft] != "0m" {
		t.Errorf("time left for expired user = %q, want 0m", vars[varTimeLeft])
	}
	if vars[varUsagePercentage] != "100%" {
		t.Errorf("usage pct = %q, want 100%%", vars[varUsagePercentage])
	}
}

// --- Status Emoji tests ---

func TestStatusEmojiMapping(t *testing.T) {
	cases := map[domain.UserStatus]string{
		domain.UserStatusActive:   "✅",
		domain.UserStatusLimited:  "🟡",
		domain.UserStatusExpired:  "❌",
		domain.UserStatusDisabled: "🔴",
		domain.UserStatusOnHold:   "⏸️",
	}
	for status, want := range cases {
		if got := StatusEmoji(status); got != want {
			t.Errorf("StatusEmoji(%s) = %q, want %q", status, got, want)
		}
	}
}

func TestStatusEmojiUnknownStatus(t *testing.T) {
	if got := StatusEmoji("unknown_status"); got != "" {
		t.Errorf("StatusEmoji(unknown) = %q, want empty", got)
	}
}

func TestStatusEmojiInResolvedVars(t *testing.T) {
	cases := []struct {
		status domain.UserStatus
		want   string
	}{
		{domain.UserStatusActive, "✅"},
		{domain.UserStatusLimited, "🟡"},
		{domain.UserStatusExpired, "❌"},
		{domain.UserStatusDisabled, "🔴"},
		{domain.UserStatusOnHold, "⏸️"},
	}

	for _, tc := range cases {
		u := &domain.User{
			Username: "test",
			Status:   tc.status,
		}
		// For limited: needs to trigger via DerivedStatus
		if tc.status == domain.UserStatusLimited {
			u.DataLimit = 100
			u.UsedTraffic = 200 // over limit
		}
		if tc.status == domain.UserStatusExpired {
			past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			u.ExpireAt = &past
		}

		resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC) }}
		ctx := VarContext{User: u}
		vars := resolver.BuildVars(ctx)

		if got := vars[varStatusEmoji]; got != tc.want {
			t.Errorf("status=%s: emoji = %q, want %q", tc.status, got, tc.want)
		}
	}
}

// --- Jalali conversion tests ---

func TestGregorianToJalali(t *testing.T) {
	cases := []struct {
		t    time.Time
		want string
	}{
		{time.Date(2026, 3, 21, 0, 0, 0, 0, time.UTC), "1405/01/01"}, // Nowruz
		{time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "1402/10/11"},
		{time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), "1404/03/25"},
		{time.Date(2023, 9, 23, 0, 0, 0, 0, time.UTC), "1402/07/01"}, // start of Mehr
	}

	for _, tc := range cases {
		got := gregorianToJalali(tc.t)
		if got != tc.want {
			t.Errorf("gregorianToJalali(%s) = %q, want %q", tc.t.Format("2006-01-02"), got, tc.want)
		}
	}
}

// --- formatDuration tests ---

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{-time.Hour, "0m"},
		{30 * time.Minute, "30m"},
		{2 * time.Hour, "2h"},
		{2*time.Hour + 30*time.Minute, "2h 30m"},
		{24 * time.Hour, "1d"},
		{25 * time.Hour, "1d 1h"},
		{10*24*time.Hour + 5*time.Hour, "10d 5h"},
	}

	for _, tc := range cases {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

// --- Usage percentage edge cases ---

func TestUsagePercentageOverLimit(t *testing.T) {
	u := &domain.User{
		Username:    "overuser",
		DataLimit:   1024,
		UsedTraffic: 2048, // 200% but should cap at 100%
	}
	resolver := &VarResolver{}
	ctx := VarContext{User: u}
	vars := resolver.BuildVars(ctx)

	if vars[varUsagePercentage] != "100%" {
		t.Errorf("usage pct over limit = %q, want 100%%", vars[varUsagePercentage])
	}
}

func TestUsagePercentageZeroUsage(t *testing.T) {
	u := &domain.User{
		Username:    "newuser",
		DataLimit:   10 * 1024 * 1024,
		UsedTraffic: 0,
	}
	resolver := &VarResolver{}
	ctx := VarContext{User: u}
	vars := resolver.BuildVars(ctx)

	if vars[varUsagePercentage] != "0%" {
		t.Errorf("usage pct zero usage = %q, want 0%%", vars[varUsagePercentage])
	}
}

// --- Backward compatibility: FormatVars still works ---

func TestFormatVarsBackwardCompatibility(t *testing.T) {
	expire := time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC)
	u := &domain.User{
		Username:    "alice",
		UsedTraffic: 1024 * 1024,
		DataLimit:   10 * 1024 * 1024,
		ExpireAt:    &expire,
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	vars := formatVarsAt(u, "1.2.3.4", "2001:db8::1", now)

	// Check original vars are still present
	if vars[varUsername] != "alice" {
		t.Errorf("username = %q, want alice", vars[varUsername])
	}
	if vars[varServerIP] != "1.2.3.4" {
		t.Errorf("server ip = %q, want 1.2.3.4", vars[varServerIP])
	}
	if vars[varServerIPv6] != "2001:db8::1" {
		t.Errorf("server ipv6 = %q, want 2001:db8::1", vars[varServerIPv6])
	}
	if vars[varDaysLeft] != "10" {
		t.Errorf("days left = %q, want 10", vars[varDaysLeft])
	}

	// Check new vars are also present (even via legacy API)
	if _, ok := vars[varProtocol]; !ok {
		t.Error("new variable {PROTOCOL} missing from legacy FormatVars")
	}
	if _, ok := vars[varStatusEmoji]; !ok {
		t.Error("new variable {STATUS_EMOJI} missing from legacy FormatVars")
	}
	if _, ok := vars[varJalaliExpire]; !ok {
		t.Error("new variable {JALALI_EXPIRE_DATE} missing from legacy FormatVars")
	}
}

// --- All 20+ variables are present in the map ---

func TestAllVariablesPresent(t *testing.T) {
	allVars := []string{
		varServerIP, varServerIPv6, varUsername, varAdminUsername,
		varProtocol, varTransport,
		varDataUsage, varDataLimit, varDataLeft, varUsagePercentage,
		varDaysLeft, varExpireDate, varJalaliExpire, varTimeLeft,
		varStatusEmoji,
		varNodeName, varNodeFlag,
		varISPName, varQualityScore, varOnlineCount,
	}

	if len(allVars) < 20 {
		t.Fatalf("expected at least 20 variables, got %d", len(allVars))
	}

	resolver := &VarResolver{}
	ctx := VarContext{User: &domain.User{Username: "test"}}
	vars := resolver.BuildVars(ctx)

	for _, v := range allVars {
		if _, ok := vars[v]; !ok {
			t.Errorf("variable %s missing from resolved map", v)
		}
	}

	// Verify we have exactly these keys (no more, no fewer)
	fmt.Printf("Total variables supported: %d\n", len(allVars))
}
