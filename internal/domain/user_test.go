package domain

import (
	"testing"
	"time"
)

func TestUserDerivedStatus(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name string
		user User
		want UserStatus
	}{
		{"disabled wins over everything", User{Status: UserStatusDisabled, DataLimit: 1, UsedTraffic: 5}, UserStatusDisabled},
		{"on hold with no usage stays on hold", User{Status: UserStatusOnHold, UsedTraffic: 0}, UserStatusOnHold},
		{"data limit reached -> limited", User{Status: UserStatusActive, DataLimit: 100, UsedTraffic: 100}, UserStatusLimited},
		{"expired -> expired", User{Status: UserStatusActive, ExpireAt: &past}, UserStatusExpired},
		{"healthy -> active", User{Status: UserStatusActive, DataLimit: 100, UsedTraffic: 10, ExpireAt: &future}, UserStatusActive},
		{"unlimited never limited", User{Status: UserStatusActive, DataLimit: 0, UsedTraffic: 1 << 40}, UserStatusActive},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.user.DerivedStatus(now); got != tc.want {
				t.Fatalf("DerivedStatus = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestUserIsActive(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)

	if (&User{Status: UserStatusActive, DataLimit: 10, UsedTraffic: 10}).IsActive(now) {
		t.Fatal("user at data limit should be inactive")
	}
	if (&User{Status: UserStatusActive, ExpireAt: &past}).IsActive(now) {
		t.Fatal("expired user should be inactive")
	}
	if !(&User{Status: UserStatusActive}).IsActive(now) {
		t.Fatal("fresh unlimited user should be active")
	}
}
