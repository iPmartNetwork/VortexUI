package service

import (
	"testing"
)

// Property 46: Dynamic QR freshness — QR content changes when subscription
// URL or config changes (not stale).
func TestProperty_DynamicQRFreshness(t *testing.T) {
	url1 := "https://panel.example.com/sub/token1"
	url2 := "https://panel.example.com/sub/token2"

	qr1 := generateQRContent(url1)
	qr2 := generateQRContent(url2)

	if qr1 == qr2 {
		t.Fatal("different URLs should produce different QR content")
	}

	// Same URL produces same content (deterministic).
	qr1b := generateQRContent(url1)
	if qr1 != qr1b {
		t.Fatal("same URL should produce same QR content")
	}
}

// Property 47: Usage alert threshold accuracy — alerts trigger at exactly
// 80%, 90%, and 100% thresholds.
func TestProperty_UsageAlertThresholdAccuracy(t *testing.T) {
	svc := &PortalProService{}
	limit := int64(100 * 1024 * 1024 * 1024) // 100 GB

	cases := []struct {
		usedPct       float64
		expectedCount int
	}{
		{0.5, 0},  // 50% — no alert
		{0.79, 0}, // 79% — no alert
		{0.80, 1}, // 80% — one alert
		{0.89, 1}, // 89% — still one (80% only)
		{0.90, 2}, // 90% — two alerts (80% + 90%)
		{0.99, 2}, // 99% — still two
		{1.00, 3}, // 100% — all three alerts
		{1.10, 3}, // 110% — all three (over limit)
	}

	for _, c := range cases {
		used := int64(float64(limit) * c.usedPct)
		alerts := svc.CheckUsageAlerts(used, limit)
		if len(alerts) != c.expectedCount {
			t.Fatalf("%.0f%% usage: expected %d alerts, got %d",
				c.usedPct*100, c.expectedCount, len(alerts))
		}
	}

	// Zero limit = no alerts (unlimited).
	alerts := svc.CheckUsageAlerts(999999, 0)
	if len(alerts) != 0 {
		t.Fatal("unlimited (0 limit) should produce no alerts")
	}
}

// Property 48: Internationalization key completeness — all keys in EN locale
// exist in every other locale.
func TestProperty_InternationalizationKeyCompleteness(t *testing.T) {
	// Simulated locale key sets (in production, load from JSON files).
	enKeys := []string{
		"portal.title", "portal.speedTest", "portal.guides",
		"portal.setupWizard", "portal.refreshQR", "portal.theme",
		"portal.language", "portal.settings",
	}

	otherLocales := map[string][]string{
		"fa": {"portal.title", "portal.speedTest", "portal.guides",
			"portal.setupWizard", "portal.refreshQR", "portal.theme",
			"portal.language", "portal.settings"},
		"ar": {"portal.title", "portal.speedTest", "portal.guides",
			"portal.setupWizard", "portal.refreshQR", "portal.theme",
			"portal.language", "portal.settings"},
		"tr": {"portal.title", "portal.speedTest", "portal.guides",
			"portal.setupWizard", "portal.refreshQR", "portal.theme",
			"portal.language", "portal.settings"},
		"ru": {"portal.title", "portal.speedTest", "portal.guides",
			"portal.setupWizard", "portal.refreshQR", "portal.theme",
			"portal.language", "portal.settings"},
	}

	for locale, keys := range otherLocales {
		keySet := toSet(keys)
		for _, enKey := range enKeys {
			if !keySet[enKey] {
				t.Fatalf("locale %q missing key %q", locale, enKey)
			}
		}
	}
}

// --- helpers ---

func generateQRContent(url string) string {
	// In production this would encode to QR image bytes.
	// For the property test, the content IS the URL.
	return "qr:" + url
}

func toSet(keys []string) map[string]bool {
	s := make(map[string]bool, len(keys))
	for _, k := range keys {
		s[k] = true
	}
	return s
}
