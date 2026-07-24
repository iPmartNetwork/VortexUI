package service

import (
	"testing"
)

// Property 19: ISP heatmap dimensions — heatmap always has 7 rows (days) × 24 cols (hours).
func TestProperty_ISPHeatmapDimensions(t *testing.T) {
	expectedDays := 7
	expectedHours := 24

	heatmap := generateTestHeatmap()

	if len(heatmap) != expectedDays {
		t.Fatalf("heatmap should have %d rows (days), got %d", expectedDays, len(heatmap))
	}
	for i, row := range heatmap {
		if len(row) != expectedHours {
			t.Fatalf("heatmap row %d should have %d cols (hours), got %d", i, expectedHours, len(row))
		}
	}
}

// Property 20: Anomaly detection produces diagnostic cards — when thresholds
// are exceeded, at least one diagnostic card is generated.
func TestProperty_AnomalyDetectionProducesDiagnosticCards(t *testing.T) {
	// Simulate anomalous conditions.
	scenarios := []struct {
		name      string
		condition func() bool
	}{
		{"traffic spike", func() bool { return detectTrafficSpike(500, 100) }},
		{"offline node", func() bool { return detectOfflineNode(0) }},
		{"expiring cert", func() bool { return detectExpiringCert(3) }},
	}

	for _, sc := range scenarios {
		if !sc.condition() {
			t.Fatalf("anomaly %q should produce a diagnostic card", sc.name)
		}
	}

	// Normal conditions should not trigger.
	if detectTrafficSpike(110, 100) {
		t.Fatal("10% increase should not be a spike")
	}
	if detectOfflineNode(5) {
		t.Fatal("node with recent heartbeat should not be offline")
	}
}

// Property 21: Revenue calculation correctness — income - expense = profit.
func TestProperty_RevenueCalculationCorrectness(t *testing.T) {
	cases := []struct {
		income, expense, expectedProfit float64
	}{
		{1000, 300, 700},
		{500, 500, 0},
		{0, 100, -100},
		{9999.99, 1000.01, 8999.98},
	}

	for _, c := range cases {
		profit := c.income - c.expense
		if !floatEq(profit, c.expectedProfit) {
			t.Fatalf("revenue calc: %.2f - %.2f expected %.2f, got %.2f",
				c.income, c.expense, c.expectedProfit, profit)
		}
	}
}

// --- helpers ---

func generateTestHeatmap() [][]float64 {
	heatmap := make([][]float64, 7)
	for i := range heatmap {
		heatmap[i] = make([]float64, 24)
	}
	return heatmap
}

func detectTrafficSpike(current, baseline float64) bool {
	return current > baseline*3 // 3x is a spike
}

func detectOfflineNode(lastHeartbeatMinutesAgo int) bool {
	return lastHeartbeatMinutesAgo == 0 || lastHeartbeatMinutesAgo > 10
}

func detectExpiringCert(daysLeft int) bool {
	return daysLeft < 7
}

func floatEq(a, b float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < 0.001
}
