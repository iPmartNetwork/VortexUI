package config

import (
	"os"
	"testing"
)

func TestDefaultCoreBin(t *testing.T) {
	tests := []struct {
		core string
		want string
	}{
		{"xray", "/usr/local/bin/xray"},
		{"singbox", "/usr/local/bin/sing-box"},
		{"custom", "custom"},
	}
	for _, tc := range tests {
		if got := DefaultCoreBin(tc.core); got != tc.want {
			t.Errorf("DefaultCoreBin(%q) = %q, want %q", tc.core, got, tc.want)
		}
	}
}

func TestSingboxV2RayAPIDefaults(t *testing.T) {
	t.Setenv("VORTEX_SINGBOX_V2RAY_API", "")
	if got := singboxV2RayAPIFromEnv("singbox"); got {
		t.Fatal("singbox should default V2Ray API to false when env unset")
	}
	if got := singboxV2RayAPIFromEnv("xray"); !got {
		t.Fatal("xray should default V2Ray API to true when env unset")
	}
}

func TestSingboxV2RayAPIExplicitEnv(t *testing.T) {
	t.Setenv("VORTEX_SINGBOX_V2RAY_API", "true")
	if !singboxV2RayAPIFromEnv("singbox") {
		t.Fatal("explicit true should win for singbox")
	}
	t.Setenv("VORTEX_SINGBOX_V2RAY_API", "false")
	if singboxV2RayAPIFromEnv("xray") {
		t.Fatal("explicit false should win for xray")
	}
}

func TestLoadNodeSingboxDefaults(t *testing.T) {
	os.Clearenv()
	t.Setenv("VORTEX_TLS_CERT", "/c.crt")
	t.Setenv("VORTEX_TLS_KEY", "/c.key")
	t.Setenv("VORTEX_TLS_CA", "/ca.crt")
	t.Setenv("VORTEX_CORE", "singbox")

	n, err := LoadNode()
	if err != nil {
		t.Fatalf("LoadNode: %v", err)
	}
	if n.CoreBin != "/usr/local/bin/sing-box" {
		t.Fatalf("CoreBin = %q", n.CoreBin)
	}
	if n.SingboxV2RayAPI {
		t.Fatal("stock singbox node should default SingboxV2RayAPI to false")
	}
}

func TestLoadPanelSingboxDefaults(t *testing.T) {
	os.Clearenv()
	t.Setenv("VORTEX_DATABASE_URL", "postgres://u:p@localhost/vortex?sslmode=disable")
	t.Setenv("VORTEX_JWT_SECRET", "01234567890123456789012345678901")
	t.Setenv("VORTEX_CORE", "singbox")

	p, err := LoadPanel()
	if err != nil {
		t.Fatalf("LoadPanel: %v", err)
	}
	if p.CoreBin != "/usr/local/bin/sing-box" {
		t.Fatalf("CoreBin = %q", p.CoreBin)
	}
	if p.SingboxV2RayAPI {
		t.Fatal("stock singbox panel local node should default SingboxV2RayAPI to false")
	}
}
