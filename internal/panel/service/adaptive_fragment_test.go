package service

import (
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestAdaptiveFragment_BestFragment_SeededDefaults(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)

	// MCI should return the 85% success profile
	best := svc.BestFragment(domain.ISPHamrahAval)
	if best == nil {
		t.Fatal("expected non-nil best fragment for MCI")
	}
	if best.Size != "10-50" {
		t.Errorf("expected size 10-50, got %s", best.Size)
	}
	if best.Success != 0.85 {
		t.Errorf("expected 0.85 success, got %f", best.Success)
	}
}

func TestAdaptiveFragment_BestFragment_UnknownISP(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)
	best := svc.BestFragment("unknown_isp")
	if best != nil {
		t.Error("expected nil for unknown ISP")
	}
}

func TestAdaptiveFragment_RecordResult_Success(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)

	// Record successes for MCI's first profile
	initial := svc.BestFragment(domain.ISPHamrahAval).Success
	svc.RecordResult(domain.ISPHamrahAval, "10-50", "10-20", true)

	updated := svc.BestFragment(domain.ISPHamrahAval).Success
	// Success should increase (EMA: 0.85*0.95 + 0.05 = 0.8575)
	if updated <= initial {
		t.Errorf("expected success to increase after positive result, was %f now %f", initial, updated)
	}
}

func TestAdaptiveFragment_RecordResult_Failure(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)

	initial := svc.BestFragment(domain.ISPHamrahAval).Success
	svc.RecordResult(domain.ISPHamrahAval, "10-50", "10-20", false)

	updated := svc.BestFragment(domain.ISPHamrahAval).Success
	// Success should decrease (EMA: 0.85*0.95 = 0.8075)
	if updated >= initial {
		t.Errorf("expected success to decrease after failure, was %f now %f", initial, updated)
	}
}

func TestAdaptiveFragment_RecordResult_NewProfile(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)

	// Record for a completely new size/interval combination
	svc.RecordResult(domain.ISPHamrahAval, "200-300", "30-50", true)

	// Should still return the best (seeded profile with 0.85 > new with 1.0? no, new is 1.0)
	best := svc.BestFragment(domain.ISPHamrahAval)
	// New profile has 1.0 success (but only 1 sample)
	if best.Size != "200-300" {
		// The new profile has Success=1.0 which is > 0.85
		t.Errorf("expected new profile (1.0 success) to be best, got size=%s success=%f", best.Size, best.Success)
	}
}

func TestAdaptiveFragment_FormatFragment(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)

	frag := svc.FormatFragment(domain.ISPHamrahAval)
	expected := "10-50,10-20,tlshello"
	if frag != expected {
		t.Errorf("expected %q, got %q", expected, frag)
	}
}

func TestAdaptiveFragment_FormatFragment_UnknownISP(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)
	frag := svc.FormatFragment("nonexistent")
	if frag != "" {
		t.Errorf("expected empty string for unknown ISP, got %q", frag)
	}
}

func TestAdaptiveFragment_Irancell_BestProfile(t *testing.T) {
	svc := NewAdaptiveFragmentService(nil)
	best := svc.BestFragment(domain.ISPIrancell)
	if best == nil {
		t.Fatal("expected non-nil best fragment for Irancell")
	}
	if best.Success != 0.90 {
		t.Errorf("expected 0.90 success for best Irancell profile, got %f", best.Success)
	}
}
