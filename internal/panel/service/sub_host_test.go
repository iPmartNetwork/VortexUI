package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// fakeSubHostRepo is an in-memory SubHostRepository for service unit tests.
type fakeSubHostRepo struct {
	hosts map[uuid.UUID]*domain.SubHost
}

func newFakeSubHostRepo() *fakeSubHostRepo {
	return &fakeSubHostRepo{hosts: map[uuid.UUID]*domain.SubHost{}}
}

func (r *fakeSubHostRepo) Create(_ context.Context, h *domain.SubHost) error {
	cp := *h
	r.hosts[h.ID] = &cp
	return nil
}

func (r *fakeSubHostRepo) Update(_ context.Context, h *domain.SubHost) error {
	if _, ok := r.hosts[h.ID]; !ok {
		return domain.ErrNotFound
	}
	cp := *h
	r.hosts[h.ID] = &cp
	return nil
}

func (r *fakeSubHostRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(r.hosts, id)
	return nil
}

func (r *fakeSubHostRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.SubHost, error) {
	h, ok := r.hosts[id]
	if !ok {
		return nil, nil
	}
	cp := *h
	return &cp, nil
}

func (r *fakeSubHostRepo) ListByInbound(_ context.Context, inboundID uuid.UUID) ([]*domain.SubHost, error) {
	var out []*domain.SubHost
	for _, h := range r.hosts {
		if h.InboundID == inboundID {
			cp := *h
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (r *fakeSubHostRepo) ListByInbounds(_ context.Context, inboundIDs []uuid.UUID) ([]*domain.SubHost, error) {
	set := map[uuid.UUID]bool{}
	for _, id := range inboundIDs {
		set[id] = true
	}
	var out []*domain.SubHost
	for _, h := range r.hosts {
		if set[h.InboundID] {
			cp := *h
			out = append(out, &cp)
		}
	}
	return out, nil
}

func TestSubHostValidate(t *testing.T) {
	tests := []struct {
		name    string
		host    domain.SubHost
		wantErr bool
	}{
		{"ok minimal", domain.SubHost{Remark: "r", Security: domain.HostSecurityInboundDefault}, false},
		{"ok fragment", domain.SubHost{Remark: "r", Security: domain.HostSecurityTLS, Fragment: "100-200,10-20,1"}, false},
		{"empty remark", domain.SubHost{Remark: "  ", Security: domain.HostSecurityTLS}, true},
		{"bad security", domain.SubHost{Remark: "r", Security: domain.HostSecurity("bogus")}, true},
		{"bad fragment", domain.SubHost{Remark: "r", Security: domain.HostSecurityTLS, Fragment: "not,a,fragment"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.host.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubHostServiceCreateDefaultsSecurity(t *testing.T) {
	svc := NewSubHostService(newFakeSubHostRepo())
	in := SubHostInput{InboundID: uuid.New(), Remark: "r", Enabled: true}
	h, err := svc.Create(context.Background(), in)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if h.Security != domain.HostSecurityInboundDefault {
		t.Errorf("security default = %q, want inbound_default", h.Security)
	}
}

func TestSubHostServiceCreateRejectsInvalid(t *testing.T) {
	svc := NewSubHostService(newFakeSubHostRepo())
	if _, err := svc.Create(context.Background(), SubHostInput{Remark: "r"}); err == nil {
		t.Error("expected error for missing inbound_id")
	}
	if _, err := svc.Create(context.Background(), SubHostInput{InboundID: uuid.New(), Remark: ""}); err == nil {
		t.Error("expected error for empty remark")
	}
}

func TestSubHostServiceUpdateKeepsInboundAndNotFound(t *testing.T) {
	repo := newFakeSubHostRepo()
	svc := NewSubHostService(repo)
	inbound := uuid.New()
	h, err := svc.Create(context.Background(), SubHostInput{InboundID: inbound, Remark: "r", Enabled: true})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Update attempts to change inbound; service must ignore it.
	updated, err := svc.Update(context.Background(), h.ID, SubHostInput{InboundID: uuid.New(), Remark: "r2"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.InboundID != inbound {
		t.Errorf("inbound changed on update: %v want %v", updated.InboundID, inbound)
	}
	if updated.Remark != "r2" {
		t.Errorf("remark = %q, want r2", updated.Remark)
	}
	// Updating a missing host returns ErrNotFound.
	if _, err := svc.Update(context.Background(), uuid.New(), SubHostInput{Remark: "x"}); err != domain.ErrNotFound {
		t.Errorf("update missing err = %v, want ErrNotFound", err)
	}
}

func TestSubHostServiceReorder(t *testing.T) {
	repo := newFakeSubHostRepo()
	svc := NewSubHostService(repo)
	inbound := uuid.New()
	var ids []uuid.UUID
	for i := 0; i < 3; i++ {
		h, err := svc.Create(context.Background(), SubHostInput{InboundID: inbound, Remark: "r", Priority: 99, Enabled: true})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		ids = append(ids, h.ID)
	}
	// Reorder reversed: ids[2], ids[0], ids[1] => priorities 0,1,2.
	order := []uuid.UUID{ids[2], ids[0], ids[1]}
	if err := svc.Reorder(context.Background(), order); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	for wantPriority, id := range order {
		h, _ := repo.GetByID(context.Background(), id)
		if h.Priority != wantPriority {
			t.Errorf("host %v priority = %d, want %d", id, h.Priority, wantPriority)
		}
	}
	// Reorder with an unknown ID returns ErrNotFound.
	if err := svc.Reorder(context.Background(), []uuid.UUID{uuid.New()}); err != domain.ErrNotFound {
		t.Errorf("reorder missing err = %v, want ErrNotFound", err)
	}
}
