package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// fakeRoutingPackRepo is an in-memory RoutingPackRepository for unit tests.
type fakeRoutingPackRepo struct {
	packs    map[uuid.UUID]*domain.RoutingPack
	global   string
	userPack map[uuid.UUID]string
}

func newFakeRoutingPackRepo() *fakeRoutingPackRepo {
	return &fakeRoutingPackRepo{
		packs:    map[uuid.UUID]*domain.RoutingPack{},
		userPack: map[uuid.UUID]string{},
	}
}

func (r *fakeRoutingPackRepo) Create(_ context.Context, p *domain.RoutingPack) error {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return err
	}
	cp := *p
	r.packs[id] = &cp
	return nil
}

func (r *fakeRoutingPackRepo) Update(_ context.Context, p *domain.RoutingPack) error {
	id, err := uuid.Parse(p.ID)
	if err != nil {
		return err
	}
	if _, ok := r.packs[id]; !ok {
		return domain.ErrNotFound
	}
	cp := *p
	r.packs[id] = &cp
	return nil
}

func (r *fakeRoutingPackRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.packs, id)
	return nil
}

func (r *fakeRoutingPackRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.RoutingPack, error) {
	p, ok := r.packs[id]
	if !ok {
		return nil, nil
	}
	cp := *p
	return &cp, nil
}

func (r *fakeRoutingPackRepo) List(_ context.Context) ([]*domain.RoutingPack, error) {
	var out []*domain.RoutingPack
	for _, p := range r.packs {
		cp := *p
		out = append(out, &cp)
	}
	return out, nil
}

func (r *fakeRoutingPackRepo) GetGlobalDefault(_ context.Context) (string, error) {
	return r.global, nil
}

func (r *fakeRoutingPackRepo) SetGlobalDefault(_ context.Context, packID string) error {
	r.global = packID
	return nil
}

func (r *fakeRoutingPackRepo) GetUserPack(_ context.Context, userID uuid.UUID) (string, error) {
	return r.userPack[userID], nil
}

func (r *fakeRoutingPackRepo) SetUserPack(_ context.Context, userID uuid.UUID, packID string) error {
	r.userPack[userID] = packID
	return nil
}

// stubRuleApplier records the rules it was asked to create. When resyncErr is
// set it mimics RoutingService's "saved but resync failed" contract: it returns
// a non-nil rule together with the error.
type stubRuleApplier struct {
	calls     []RoutingRuleInput
	nodeIDs   []uuid.UUID
	resyncErr error
}

func (s *stubRuleApplier) Create(_ context.Context, nodeID uuid.UUID, in RoutingRuleInput) (*domain.RoutingRule, error) {
	s.calls = append(s.calls, in)
	s.nodeIDs = append(s.nodeIDs, nodeID)
	rule := &domain.RoutingRule{ID: uuid.New(), NodeID: nodeID, Name: in.Name, OutboundTag: in.OutboundTag}
	if s.resyncErr != nil {
		return rule, s.resyncErr // saved, but resync failed
	}
	return rule, nil
}

// stubOutboundApplier records outbound creations.
type stubOutboundApplier struct {
	calls []CreateOutboundInput
}

func (s *stubOutboundApplier) Create(_ context.Context, in CreateOutboundInput) (*domain.Outbound, error) {
	s.calls = append(s.calls, in)
	return &domain.Outbound{ID: uuid.New(), NodeID: in.NodeID, Tag: in.Tag}, nil
}

func TestRoutingPackListMergesBuiltinAndCustom(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	rules := &stubRuleApplier{}
	outs := &stubOutboundApplier{}
	svc := NewRoutingPackService(repo, rules, outs)

	custom, err := svc.Create(context.Background(), RoutingPackInput{Name: "My Custom Pack", Category: "custom"})
	if err != nil {
		t.Fatalf("create custom: %v", err)
	}

	packs, err := svc.ListPacks(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}

	builtinCount := len(domain.BuiltinRoutingPacks())
	if len(packs) != builtinCount+1 {
		t.Fatalf("ListPacks returned %d packs, want %d builtin + 1 custom", len(packs), builtinCount)
	}
	// Built-ins come first and are flagged builtin with name-IDs.
	if !packs[0].Builtin || packs[0].ID != packs[0].Name {
		t.Errorf("first pack not a name-ID builtin: %+v", packs[0])
	}
	last := packs[len(packs)-1]
	if last.Builtin || last.ID != custom.ID || last.Name != "My Custom Pack" {
		t.Errorf("last pack not the custom pack: %+v", last)
	}
}

func TestRoutingPackGetResolvesBuiltinAndCustom(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	svc := NewRoutingPackService(repo, &stubRuleApplier{}, &stubOutboundApplier{})

	// Built-in resolved by name.
	got, err := svc.GetPack(context.Background(), "Iran Direct")
	if err != nil {
		t.Fatalf("get builtin: %v", err)
	}
	if got.Name != "Iran Direct" || !got.Builtin {
		t.Errorf("builtin lookup wrong: %+v", got)
	}

	// Unknown ID is ErrNotFound.
	if _, err := svc.GetPack(context.Background(), "does-not-exist"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown pack err = %v, want ErrNotFound", err)
	}
}

func TestRoutingPackApplyToNodeCreatesRulePerRule(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	rules := &stubRuleApplier{}
	outs := &stubOutboundApplier{}
	svc := NewRoutingPackService(repo, rules, outs)

	nodeID := uuid.New()
	// "Iran Direct" has exactly two rules and no outbounds.
	if err := svc.ApplyToNode(context.Background(), nodeID, "Iran Direct"); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(rules.calls) != 2 {
		t.Fatalf("RuleApplier.Create called %d times, want 2", len(rules.calls))
	}
	for _, n := range rules.nodeIDs {
		if n != nodeID {
			t.Errorf("rule created for node %v, want %v", n, nodeID)
		}
	}
	if rules.calls[0].OutboundTag != "direct" {
		t.Errorf("first rule outbound_tag = %q, want direct", rules.calls[0].OutboundTag)
	}
}

func TestRoutingPackApplyToNodeCreatesOutbounds(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	rules := &stubRuleApplier{}
	outs := &stubOutboundApplier{}
	svc := NewRoutingPackService(repo, rules, outs)

	pack, err := svc.Create(context.Background(), RoutingPackInput{
		Name: "with-outbound",
		Rules: []domain.RoutingRule{
			{Name: "via-warp", Domains: []string{"openai.com"}, OutboundTag: "warp", Enabled: true},
		},
		Outbounds: []domain.Outbound{
			{Tag: "warp", Protocol: domain.OutFreedom},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	nodeID := uuid.New()
	if err := svc.ApplyToNode(context.Background(), nodeID, pack.ID); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(outs.calls) != 1 || outs.calls[0].Tag != "warp" || outs.calls[0].NodeID != nodeID {
		t.Errorf("outbound creation wrong: %+v", outs.calls)
	}
	if len(rules.calls) != 1 {
		t.Errorf("rule creations = %d, want 1", len(rules.calls))
	}
}

func TestRoutingPackApplyToNodeResyncFailureIsNonFatal(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	rules := &stubRuleApplier{resyncErr: errors.New("node unreachable")}
	outs := &stubOutboundApplier{}
	svc := NewRoutingPackService(repo, rules, outs)

	err := svc.ApplyToNode(context.Background(), uuid.New(), "Iran Direct")
	if !errors.Is(err, ErrPackResyncFailed) {
		t.Fatalf("apply err = %v, want ErrPackResyncFailed", err)
	}
	// Both rules were still attempted (saved) despite the resync failures.
	if len(rules.calls) != 2 {
		t.Errorf("RuleApplier.Create called %d times, want 2 (saved despite resync failure)", len(rules.calls))
	}
}

func TestRoutingPackSelectionRoundTrip(t *testing.T) {
	repo := newFakeRoutingPackRepo()
	svc := NewRoutingPackService(repo, &stubRuleApplier{}, &stubOutboundApplier{})
	ctx := context.Background()

	if err := svc.SetGlobalDefault(ctx, "Block Ads & Malware"); err != nil {
		t.Fatalf("set global: %v", err)
	}
	got, err := svc.GetGlobalDefault(ctx)
	if err != nil {
		t.Fatalf("get global: %v", err)
	}
	if got != "Block Ads & Malware" {
		t.Errorf("global default = %q, want Block Ads & Malware", got)
	}

	userID := uuid.New()
	if err := svc.SetUserPack(ctx, userID, "Iran Direct"); err != nil {
		t.Fatalf("set user pack: %v", err)
	}
	up, err := svc.GetUserPack(ctx, userID)
	if err != nil {
		t.Fatalf("get user pack: %v", err)
	}
	if up != "Iran Direct" {
		t.Errorf("user pack = %q, want Iran Direct", up)
	}
}
