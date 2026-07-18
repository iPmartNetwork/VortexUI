package service

import (
	"testing"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestAutoConfigEngine_PeakHours_MCI(t *testing.T) {
	engine := &AutoConfigEngine{now: func() time.Time {
		// 20:00 Tehran time — peak hours
		return time.Date(2025, 1, 15, 20, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPHamrahAval, false)
	if rec.Protocol != domain.ProtoVLESS {
		t.Errorf("expected VLESS, got %s", rec.Protocol)
	}
	if rec.Transport != "ws" {
		t.Errorf("expected ws transport, got %s", rec.Transport)
	}
	if !rec.Mux {
		t.Error("expected mux enabled during peak hours")
	}
	if !rec.ECH {
		t.Error("expected ECH enabled during peak hours")
	}
	if rec.Fragment == "" {
		t.Error("expected fragment set during peak hours")
	}
}

func TestAutoConfigEngine_OffPeak_MCI_WithCDN(t *testing.T) {
	engine := &AutoConfigEngine{now: func() time.Time {
		// 10:00 — off-peak
		return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPHamrahAval, true)
	if rec.Transport != "ws" {
		t.Errorf("expected ws transport with CDN, got %s", rec.Transport)
	}
	if rec.Security != "tls" {
		t.Errorf("expected tls security, got %s", rec.Security)
	}
	if rec.Mux {
		t.Error("expected mux disabled off-peak with CDN")
	}
}

func TestAutoConfigEngine_OffPeak_MCI_NoCDN(t *testing.T) {
	engine := &AutoConfigEngine{now: func() time.Time {
		return time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPHamrahAval, false)
	if rec.Security != "reality" {
		t.Errorf("expected reality security for direct MCI off-peak, got %s", rec.Security)
	}
	if rec.Transport != "tcp" {
		t.Errorf("expected tcp transport, got %s", rec.Transport)
	}
}

func TestAutoConfigEngine_Mokhaberat_AlwaysMaxEvasion(t *testing.T) {
	// Even off-peak, Mokhaberat (TCI) needs max evasion
	engine := &AutoConfigEngine{now: func() time.Time {
		return time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPMokhaberat, false)
	if !rec.Mux {
		t.Error("expected mux for Mokhaberat")
	}
	if !rec.ECH {
		t.Error("expected ECH for Mokhaberat")
	}
	if rec.Fragment == "" {
		t.Error("expected fragment for Mokhaberat")
	}
}

func TestAutoConfigEngine_UnknownISP_CDN(t *testing.T) {
	engine := NewAutoConfigEngine()
	rec := engine.Recommend("unknown_isp", true)
	if rec.Transport != "ws" {
		t.Errorf("expected ws for unknown ISP with CDN, got %s", rec.Transport)
	}
	if rec.Security != "tls" {
		t.Errorf("expected tls for unknown ISP with CDN, got %s", rec.Security)
	}
}

func TestAutoConfigEngine_UnknownISP_NoCDN(t *testing.T) {
	engine := NewAutoConfigEngine()
	rec := engine.Recommend("unknown_isp", false)
	if rec.Security != "reality" {
		t.Errorf("expected reality for unknown ISP direct, got %s", rec.Security)
	}
}

func TestAutoConfigEngine_Irancell_Peak(t *testing.T) {
	engine := &AutoConfigEngine{now: func() time.Time {
		return time.Date(2025, 1, 15, 21, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPIrancell, false)
	if rec.Transport != "grpc" {
		t.Errorf("expected grpc for Irancell peak, got %s", rec.Transport)
	}
	if !rec.Mux {
		t.Error("expected mux for Irancell peak")
	}
}

func TestAutoConfigEngine_Irancell_OffPeak(t *testing.T) {
	engine := &AutoConfigEngine{now: func() time.Time {
		return time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	}}
	rec := engine.Recommend(domain.ISPIrancell, false)
	if rec.Security != "reality" {
		t.Errorf("expected reality for Irancell off-peak, got %s", rec.Security)
	}
}
