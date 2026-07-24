package subscription

import (
	"testing"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

func TestResolveProfileTitle_Basic(t *testing.T) {
	resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	ctx := VarContext{
		User:          &domain.User{Username: "alice"},
		AdminUsername: "admin1",
		NodeName:      "DE-1",
	}
	config := RenderConfig{
		ProfileTitleTemplate: "{USERNAME} - VortexUI",
	}

	got := ResolveProfileTitle(config, resolver, ctx)
	if got != "alice - VortexUI" {
		t.Errorf("ResolveProfileTitle = %q, want %q", got, "alice - VortexUI")
	}
}

func TestResolveProfileTitle_Empty(t *testing.T) {
	resolver := &VarResolver{}
	ctx := VarContext{User: &domain.User{Username: "bob"}}
	config := RenderConfig{ProfileTitleTemplate: ""}

	got := ResolveProfileTitle(config, resolver, ctx)
	if got != "" {
		t.Errorf("ResolveProfileTitle with empty template = %q, want empty", got)
	}
}

func TestApplyTemplates_RemarkAndAddress(t *testing.T) {
	resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	ctx := VarContext{
		User:     &domain.User{Username: "alice"},
		NodeName: "DE-1",
		NodeIP:   "1.2.3.4",
	}
	config := RenderConfig{
		RemarkTemplate:  "{PROTOCOL} - {NODE_NAME}",
		AddressTemplate: "{SERVER_IP}",
	}
	proxies := []Proxy{
		{Name: "old-name-1", Protocol: domain.ProtoVLESS, Host: "original.host", Port: 443, Network: "ws"},
		{Name: "old-name-2", Protocol: domain.ProtoTrojan, Host: "original.host", Port: 8443, Network: "tcp"},
	}

	result := ApplyTemplates(config, resolver, ctx, proxies)

	// Check that original slice is not mutated.
	if proxies[0].Name != "old-name-1" {
		t.Errorf("original proxies[0].Name mutated to %q", proxies[0].Name)
	}

	// Check remark resolution per proxy.
	if result[0].Name != "vless - DE-1" {
		t.Errorf("result[0].Name = %q, want %q", result[0].Name, "vless - DE-1")
	}
	if result[1].Name != "trojan - DE-1" {
		t.Errorf("result[1].Name = %q, want %q", result[1].Name, "trojan - DE-1")
	}

	// Check address resolution.
	if result[0].Host != "1.2.3.4" {
		t.Errorf("result[0].Host = %q, want %q", result[0].Host, "1.2.3.4")
	}
	if result[1].Host != "1.2.3.4" {
		t.Errorf("result[1].Host = %q, want %q", result[1].Host, "1.2.3.4")
	}
}

func TestApplyTemplates_EmptyAddressTemplate_NoChange(t *testing.T) {
	resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	ctx := VarContext{
		User:     &domain.User{Username: "alice"},
		NodeName: "US-2",
		NodeIP:   "5.6.7.8",
	}
	config := RenderConfig{
		RemarkTemplate:  "{USERNAME} @ {NODE_NAME}",
		AddressTemplate: "", // empty = no address override
	}
	proxies := []Proxy{
		{Name: "old", Protocol: domain.ProtoVLESS, Host: "keep-this.com", Port: 443, Network: "ws"},
	}

	result := ApplyTemplates(config, resolver, ctx, proxies)

	if result[0].Host != "keep-this.com" {
		t.Errorf("Host should be unchanged when AddressTemplate is empty, got %q", result[0].Host)
	}
	if result[0].Name != "alice @ US-2" {
		t.Errorf("Name = %q, want %q", result[0].Name, "alice @ US-2")
	}
}

func TestApplyTemplates_EmptyRemarkTemplate_NoChange(t *testing.T) {
	resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	ctx := VarContext{
		User:     &domain.User{Username: "bob"},
		NodeName: "FR-1",
		NodeIP:   "10.0.0.1",
	}
	config := RenderConfig{
		RemarkTemplate:  "", // empty = no remark override
		AddressTemplate: "{SERVER_IP}",
	}
	proxies := []Proxy{
		{Name: "keep-this-name", Protocol: domain.ProtoTrojan, Host: "old.host", Port: 443, Network: "tcp"},
	}

	result := ApplyTemplates(config, resolver, ctx, proxies)

	if result[0].Name != "keep-this-name" {
		t.Errorf("Name should be unchanged when RemarkTemplate is empty, got %q", result[0].Name)
	}
	if result[0].Host != "10.0.0.1" {
		t.Errorf("Host = %q, want %q", result[0].Host, "10.0.0.1")
	}
}

func TestApplyTemplates_EmptyProxies(t *testing.T) {
	resolver := &VarResolver{}
	ctx := VarContext{User: &domain.User{Username: "alice"}}
	config := RenderConfig{
		RemarkTemplate:  "{USERNAME}",
		AddressTemplate: "{SERVER_IP}",
	}

	result := ApplyTemplates(config, resolver, ctx, nil)
	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}

	result = ApplyTemplates(config, resolver, ctx, []Proxy{})
	if len(result) != 0 {
		t.Errorf("expected empty slice for empty input, got %v", result)
	}
}

func TestApplyTemplates_TransportResolvesPerProxy(t *testing.T) {
	resolver := &VarResolver{Now: func() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }}
	ctx := VarContext{
		User:     &domain.User{Username: "charlie"},
		NodeName: "JP-1",
		NodeIP:   "192.168.1.1",
	}
	config := RenderConfig{
		RemarkTemplate: "{PROTOCOL}/{TRANSPORT} @ {NODE_NAME}",
	}
	proxies := []Proxy{
		{Name: "x", Protocol: domain.ProtoVLESS, Host: "h", Port: 443, Network: "ws"},
		{Name: "y", Protocol: domain.ProtoVMess, Host: "h", Port: 80, Network: "grpc"},
		{Name: "z", Protocol: domain.ProtoTrojan, Host: "h", Port: 8443, Network: "tcp"},
	}

	result := ApplyTemplates(config, resolver, ctx, proxies)

	expected := []string{
		"vless/ws @ JP-1",
		"vmess/grpc @ JP-1",
		"trojan/tcp @ JP-1",
	}
	for i, want := range expected {
		if result[i].Name != want {
			t.Errorf("result[%d].Name = %q, want %q", i, result[i].Name, want)
		}
	}
}
