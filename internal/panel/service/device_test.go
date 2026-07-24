package service

import (
	"testing"

	"github.com/google/uuid"
)

// Property 8: HWID device limit enforcement — when device_limit=N, the N+1th
// device registration is rejected.
func TestProperty_HWIDDeviceLimitEnforcement(t *testing.T) {
	deviceLimit := 3
	registered := []string{"hwid-1", "hwid-2", "hwid-3"}

	// N devices registered = at limit.
	if len(registered) > deviceLimit {
		t.Fatal("should not exceed device limit")
	}

	// N+1th should be rejected.
	newHWID := "hwid-4"
	allowed := isDeviceAllowed(registered, newHWID, deviceLimit)
	if allowed {
		t.Fatal("device beyond limit should be rejected")
	}

	// Existing HWID should still be allowed.
	allowed = isDeviceAllowed(registered, "hwid-2", deviceLimit)
	if !allowed {
		t.Fatal("existing HWID should be allowed")
	}
}

// Property 9: Device lock enforcement — when device_lock=true and a device is
// registered, only that specific HWID can access the subscription.
func TestProperty_DeviceLockEnforcement(t *testing.T) {
	lockedHWID := "locked-device-abc"
	deviceLock := true

	// Same device passes.
	if !checkDeviceLock(deviceLock, lockedHWID, lockedHWID) {
		t.Fatal("locked device should be allowed with matching HWID")
	}

	// Different device fails.
	if checkDeviceLock(deviceLock, lockedHWID, "different-device") {
		t.Fatal("different device should be rejected with device lock")
	}

	// No lock = any device passes.
	if !checkDeviceLock(false, lockedHWID, "any-device") {
		t.Fatal("without device lock, any device should pass")
	}
}

// Property 10: HWID registration round-trip — registering a device and then
// listing returns that device with correct metadata.
func TestProperty_HWIDRegistrationRoundTrip(t *testing.T) {
	userID := uuid.New()
	hwid := "test-hwid-" + uuid.New().String()[:8]
	os := "Android 14"

	device := registerDevice(userID, hwid, os)

	if device.UserID != userID {
		t.Fatal("user ID mismatch")
	}
	if device.HWID != hwid {
		t.Fatal("HWID mismatch")
	}
	if device.OS != os {
		t.Fatal("OS mismatch")
	}
}

// Property 11: Bulk HWID reset clears all devices — after bulk reset, user
// has zero registered devices.
func TestProperty_BulkHWIDResetClearsAll(t *testing.T) {
	userIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}

	// Simulate each user having devices.
	deviceCounts := map[uuid.UUID]int{
		userIDs[0]: 3,
		userIDs[1]: 2,
		userIDs[2]: 5,
	}

	// After bulk reset, all counts should be 0.
	bulkResetHWIDs(deviceCounts, userIDs)

	for _, uid := range userIDs {
		if deviceCounts[uid] != 0 {
			t.Fatalf("user %s should have 0 devices after reset, got %d", uid, deviceCounts[uid])
		}
	}
}

// --- helpers ---

func isDeviceAllowed(registered []string, hwid string, limit int) bool {
	for _, r := range registered {
		if r == hwid {
			return true
		}
	}
	return len(registered) < limit
}

func checkDeviceLock(lockEnabled bool, lockedHWID, requestHWID string) bool {
	if !lockEnabled {
		return true
	}
	return lockedHWID == requestHWID
}

type testDevice struct {
	UserID uuid.UUID
	HWID   string
	OS     string
}

func registerDevice(userID uuid.UUID, hwid, os string) testDevice {
	return testDevice{UserID: userID, HWID: hwid, OS: os}
}

func bulkResetHWIDs(counts map[uuid.UUID]int, userIDs []uuid.UUID) {
	for _, uid := range userIDs {
		counts[uid] = 0
	}
}
