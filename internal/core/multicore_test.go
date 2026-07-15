package core

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

type stubDriver struct {
	coreType domain.CoreType
	mu       sync.Mutex
	started  []*GeneratedConfig
}

func (s *stubDriver) Type() domain.CoreType { return s.coreType }
func (s *stubDriver) Start(_ context.Context, cfg *GeneratedConfig) error {
	s.mu.Lock()
	s.started = append(s.started, cfg)
	s.mu.Unlock()
	return nil
}
func (s *stubDriver) Stop(context.Context) error                          { return nil }
func (s *stubDriver) Reload(context.Context, *GeneratedConfig) error      { return nil }
func (s *stubDriver) AddUser(context.Context, string, *domain.User) error { return nil }
func (s *stubDriver) RemoveUser(context.Context, string, string) error    { return nil }
func (s *stubDriver) StreamTraffic(context.Context) (<-chan domain.TrafficDelta, error) {
	ch := make(chan domain.TrafficDelta)
	close(ch)
	return ch, nil
}
func (s *stubDriver) OnlineStats(context.Context) (map[string]int, error) { return nil, nil }
func (s *stubDriver) OnlineIPList(context.Context, string) (map[string]int64, error) {
	return nil, nil
}
func (s *stubDriver) UpdateGeoAssets(context.Context, string, string) (int64, int64, error) {
	return 0, 0, nil
}
func (s *stubDriver) Logs(context.Context, int) ([]string, error) { return nil, nil }
func (s *stubDriver) Health(context.Context) (domain.NodeHealth, error) {
	return domain.NodeHealth{CoreRunning: true}, nil
}
func (s *stubDriver) Version(context.Context) (string, error) { return "test", nil }

func TestSplitConfigByCore(t *testing.T) {
	cfg := &GeneratedConfig{
		LogLevel: "warning",
		Inbounds: []domain.Inbound{
			{Tag: "x-in", Core: domain.CoreXray, Port: 443},
			{Tag: "s-in", Core: domain.CoreSingbox, Port: 8443},
		},
		UsersByInbound: map[string][]*domain.User{
			"x-in": {{ID: uuid.New()}},
			"s-in": {{ID: uuid.New()}},
		},
		Outbounds: []domain.Outbound{{Tag: "direct"}},
	}
	split, tags := splitConfigByCore(cfg)
	if len(split) != 2 {
		t.Fatalf("split len = %d, want 2", len(split))
	}
	if tags["x-in"] != domain.CoreXray || tags["s-in"] != domain.CoreSingbox {
		t.Fatalf("unexpected tag map: %v", tags)
	}
	if len(split[domain.CoreXray].Inbounds) != 1 || split[domain.CoreXray].Inbounds[0].Tag != "x-in" {
		t.Fatalf("xray inbounds: %+v", split[domain.CoreXray].Inbounds)
	}
	if len(split[domain.CoreSingbox].UsersByInbound["s-in"]) != 1 {
		t.Fatalf("singbox users missing")
	}
	if len(split[domain.CoreXray].Outbounds) != 1 {
		t.Fatalf("outbounds not shared")
	}
}

func TestCompositeDriverStartRoutesByTag(t *testing.T) {
	xrayStub := &stubDriver{coreType: domain.CoreXray}
	singStub := &stubDriver{coreType: domain.CoreSingbox}
	comp, err := NewCompositeDriver(map[domain.CoreType]CoreDriver{
		domain.CoreXray:    xrayStub,
		domain.CoreSingbox: singStub,
	})
	if err != nil {
		t.Fatal(err)
	}
	cfg := &GeneratedConfig{
		Inbounds: []domain.Inbound{
			{Tag: "a", Core: domain.CoreXray},
			{Tag: "b", Core: domain.CoreSingbox},
		},
	}
	if err := comp.Start(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	xrayStub.mu.Lock()
	defer xrayStub.mu.Unlock()
	singStub.mu.Lock()
	defer singStub.mu.Unlock()
	if len(xrayStub.started) != 1 || len(xrayStub.started[0].Inbounds) != 1 || xrayStub.started[0].Inbounds[0].Tag != "a" {
		t.Fatalf("xray config: %+v", xrayStub.started)
	}
	if len(singStub.started) != 1 || len(singStub.started[0].Inbounds) != 1 || singStub.started[0].Inbounds[0].Tag != "b" {
		t.Fatalf("singbox config: %+v", singStub.started)
	}
}

func TestCompositeDriverType(t *testing.T) {
	comp, err := NewCompositeDriver(map[domain.CoreType]CoreDriver{
		domain.CoreXray: &stubDriver{coreType: domain.CoreXray},
	})
	if err != nil {
		t.Fatal(err)
	}
	if comp.Type() != domain.CoreMulti {
		t.Fatalf("Type() = %q, want multi", comp.Type())
	}
}
