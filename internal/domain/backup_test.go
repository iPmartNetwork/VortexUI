package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestBuildUsageSummary(t *testing.T) {
	adminID := uuid.New()
	limit := int64(1_000)
	users := []*User{
		{UsedTraffic: 200, DataLimit: limit, AdminID: &adminID},
		{UsedTraffic: 1100, DataLimit: limit, AdminID: &adminID},
		{UsedTraffic: 50, DataLimit: 0},
	}
	admins := []*Admin{{ID: adminID, Username: "res1", WalletTrafficBytes: 5000, WalletUserCredits: 3}}
	u := BuildUsageSummary(users, admins)
	if u.TotalUsers != 3 {
		t.Fatalf("users=%d", u.TotalUsers)
	}
	if u.TotalUsedTraffic != 1350 {
		t.Fatalf("used=%d", u.TotalUsedTraffic)
	}
	if u.TotalRemainingTraffic != 800 {
		t.Fatalf("remaining=%d", u.TotalRemainingTraffic)
	}
	if u.UsersOverLimit != 1 {
		t.Fatalf("over=%d", u.UsersOverLimit)
	}
	if len(u.Resellers) != 1 || u.Resellers[0].UserCount != 2 {
		t.Fatalf("resellers=%+v", u.Resellers)
	}
}

func TestIsSupportedBackupVersion(t *testing.T) {
	for _, v := range []int{BackupVersion, BackupVersionV2, BackupVersionLegacy} {
		if !IsSupportedBackupVersion(v) {
			t.Fatalf("version %d should be supported", v)
		}
	}
	if IsSupportedBackupVersion(999) {
		t.Fatal("999 should not be supported")
	}
}

func TestBuildBackupManifestWarnings(t *testing.T) {
	b := &Backup{
		Version:    BackupVersionV2,
		ExportedAt: time.Now(),
		Admins:     []*Admin{{ID: uuid.New(), Username: "a"}},
		Users:      []*User{{Username: "u"}},
	}
	m := BuildBackupManifest(b, BackupFormatJSON)
	if len(m.Warnings) == 0 {
		t.Fatal("expected warnings for legacy version and missing credentials")
	}
}
