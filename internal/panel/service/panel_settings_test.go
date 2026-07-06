package service

import (
	"context"
	"testing"

	"github.com/vortexui/vortexui/internal/domain"
)

type memPanelSettingsRepo struct {
	s *domain.PanelSettings
}

func (m *memPanelSettingsRepo) Get(_ context.Context) (*domain.PanelSettings, error) {
	if m.s == nil {
		return &domain.PanelSettings{}, nil
	}
	cp := *m.s
	return &cp, nil
}

func (m *memPanelSettingsRepo) Save(_ context.Context, s *domain.PanelSettings) error {
	cp := *s
	m.s = &cp
	return nil
}

func TestPanelSettingsService_UpdateAppliesIPHook(t *testing.T) {
	var wl, bl string
	repo := &memPanelSettingsRepo{}
	svc := NewPanelSettingsService(repo, PanelSettingsHooks{
		OnIPGuard: func(w, b string) {
			wl, bl = w, b
		},
	})

	_, err := svc.Update(context.Background(), domain.PanelSettings{
		PanelName:   "TestPanel",
		IPWhitelist: "203.0.113.0/24",
		IPBlacklist: "10.0.0.0/8",
	})
	if err != nil {
		t.Fatal(err)
	}
	if wl != "203.0.113.0/24" || bl != "10.0.0.0/8" {
		t.Fatalf("ip hook not called: wl=%q bl=%q", wl, bl)
	}
}

func TestPanelSettingsService_AutoBackupSnapshot(t *testing.T) {
	repo := &memPanelSettingsRepo{s: &domain.PanelSettings{
		AutoBackupEnabled:       true,
		AutoBackupIntervalHours: 12,
	}}
	svc := NewPanelSettingsService(repo, PanelSettingsHooks{})
	snap := svc.AutoBackupSnapshot(context.Background())
	if !snap.Enabled || snap.IntervalHours != 12 {
		t.Fatalf("unexpected snapshot: %+v", snap)
	}
}

func TestPanelSettingsService_UpdateNormalizesValues(t *testing.T) {
	repo := &memPanelSettingsRepo{}
	svc := NewPanelSettingsService(repo, PanelSettingsHooks{})

	out, err := svc.Update(context.Background(), domain.PanelSettings{
		PanelName:      "  Test Panel  ",
		SubURLTemplate: "  https://{domain}/sub/{token}  ",
		AccentColor:    "  #123456  ",
		FooterText:     "   ",
		AutoBackupIntervalHours: 200,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.PanelName != "Test Panel" {
		t.Fatalf("unexpected panel name: %q", out.PanelName)
	}
	if out.SubURLTemplate != "https://{domain}/sub/{token}" {
		t.Fatalf("unexpected sub URL template: %q", out.SubURLTemplate)
	}
	if out.AccentColor != "#123456" {
		t.Fatalf("unexpected accent color: %q", out.AccentColor)
	}
	if out.FooterText != domain.DefaultPanelSettings().FooterText {
		t.Fatalf("unexpected footer text: %q", out.FooterText)
	}
	if out.AutoBackupIntervalHours != 168 {
		t.Fatalf("unexpected backup interval: %d", out.AutoBackupIntervalHours)
	}
}
