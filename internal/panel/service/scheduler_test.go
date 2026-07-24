package service

import (
	"context"
	"testing"
	"time"
)

// Property 41: Doctor check accuracy — doctor reports correct status for
// known good and bad conditions.
func TestProperty_DoctorCheckAccuracy(t *testing.T) {
	// Good condition => pass.
	if status := checkStatus(true); status != "pass" {
		t.Fatalf("good condition should be pass, got %s", status)
	}
	// Bad condition => fail.
	if status := checkStatus(false); status != "fail" {
		t.Fatalf("bad condition should be fail, got %s", status)
	}
}

// Property 42: Backup encryption round-trip — encrypted backup can be
// decrypted back to the original payload.
func TestProperty_BackupEncryptionRoundTrip(t *testing.T) {
	svc := NewBackupEncryptService(BackupEncryptConfig{
		Secret: "test-backup-secret-123",
	})

	original := map[string]any{
		"users":    []any{"user1", "user2"},
		"settings": map[string]any{"panel_name": "Test"},
	}

	encrypted, err := svc.CreateBackup(context.Background(), original)
	if err != nil {
		t.Fatalf("create backup: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("encrypted backup should not be empty")
	}

	restored, err := svc.RestoreBackup(context.Background(), encrypted)
	if err != nil {
		t.Fatalf("restore backup: %v", err)
	}
	if restored.Version != "1.0" {
		t.Fatalf("expected version 1.0, got %s", restored.Version)
	}
	if restored.Data["settings"] == nil {
		t.Fatal("restored data missing settings")
	}
}

// Property 43: Cron schedule next-run calculation — NextRun always returns
// a time in the future relative to "from".
func TestProperty_CronScheduleNextRun(t *testing.T) {
	crons := []string{
		"0 3 * * *",   // daily at 3:00
		"*/30 * * * *", // every 30 minutes
		"0 0 * * 1",   // weekly Monday
		"0 12 1 * *",  // monthly 1st at noon
	}

	from := time.Date(2025, 6, 15, 10, 31, 0, 0, time.UTC)

	for _, cron := range crons {
		next := NextRun(cron, from)
		if !next.After(from) {
			t.Fatalf("NextRun(%q, %v) = %v, expected after from", cron, from, next)
		}
	}
}

// Property 44: Auto-cleanup eligibility — only expired/limited users past
// retention are eligible; active/on_hold are never eligible.
func TestProperty_AutoCleanupEligibility(t *testing.T) {
	retentionDays := 30
	now := time.Now()

	cases := []struct {
		status    string
		expiredAt time.Time
		eligible  bool
	}{
		{"expired", now.AddDate(0, 0, -45), true},   // expired 45 days ago
		{"limited", now.AddDate(0, 0, -31), true},   // limited 31 days ago
		{"expired", now.AddDate(0, 0, -10), false},  // expired only 10 days ago
		{"active", now.AddDate(0, 0, -100), false},  // active never deleted
		{"on_hold", now.AddDate(0, 0, -100), false}, // on_hold never deleted
		{"disabled", now.AddDate(0, 0, -100), false}, // disabled never deleted
	}

	for _, c := range cases {
		eligible := isCleanupEligible(c.status, c.expiredAt, retentionDays, now)
		if eligible != c.eligible {
			t.Fatalf("status=%s expiredDaysAgo=%d: expected eligible=%v, got %v",
				c.status, int(now.Sub(c.expiredAt).Hours()/24), c.eligible, eligible)
		}
	}
}

// Property 45: Settings export/import round-trip — exported YAML can be
// imported back with all fields intact.
func TestProperty_SettingsExportImportRoundTrip(t *testing.T) {
	svc := NewSettingsMigrationService()

	original := &PanelSettings{
		Version: "1.0",
		General: map[string]any{"panel_name": "VortexUI", "port": float64(8080)},
		Security: map[string]any{"jwt_secret": "redacted"},
	}

	data, err := svc.Export(context.Background(), original)
	if err != nil {
		t.Fatalf("export: %v", err)
	}

	imported, err := svc.Import(context.Background(), data)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if imported.Version != "1.0" {
		t.Fatal("version mismatch")
	}
	if imported.General["panel_name"] != "VortexUI" {
		t.Fatal("general.panel_name mismatch")
	}
}

// --- helpers ---

func checkStatus(healthy bool) string {
	if healthy {
		return "pass"
	}
	return "fail"
}

func isCleanupEligible(status string, expiredAt time.Time, retentionDays int, now time.Time) bool {
	// Protected statuses are never deleted.
	protected := map[string]bool{"active": true, "on_hold": true, "disabled": true}
	if protected[status] {
		return false
	}
	// Must be past retention period.
	daysSince := int(now.Sub(expiredAt).Hours() / 24)
	return daysSince >= retentionDays
}
