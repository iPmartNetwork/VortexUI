package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// --- fakes ---

type fakeAdminRepo struct {
	admin *domain.Admin
	role  *domain.Role
}

func (f *fakeAdminRepo) Create(context.Context, *domain.Admin) error { return nil }
func (f *fakeAdminRepo) GetByUsername(_ context.Context, u string) (*domain.Admin, error) {
	if f.admin != nil && f.admin.Username == u {
		return f.admin, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeAdminRepo) GetByID(context.Context, uuid.UUID) (*domain.Admin, error) {
	return f.admin, nil
}
func (f *fakeAdminRepo) Update(context.Context, *domain.Admin) error { return nil }
func (f *fakeAdminRepo) GetRole(context.Context, uuid.UUID) (*domain.Role, error) {
	return f.role, nil
}

type fakeUserRepo struct {
	created   *domain.User
	bound     []uuid.UUID
	inbounds  []domain.Inbound
}

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error { f.created = u; return nil }
func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (*domain.User, error) {
	if f.created == nil {
		return nil, domain.ErrNotFound
	}
	return f.created, nil
}
func (f *fakeUserRepo) GetBySubToken(context.Context, string) (*domain.User, error) {
	return f.created, nil
}
func (f *fakeUserRepo) Update(context.Context, *domain.User) error          { return nil }
func (f *fakeUserRepo) Delete(context.Context, uuid.UUID) error             { return nil }
func (f *fakeUserRepo) List(context.Context, port.UserFilter) ([]*domain.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) AddUsedTraffic(context.Context, uuid.UUID, int64) error          { return nil }
func (f *fakeUserRepo) AddUsedTrafficBatch(context.Context, map[uuid.UUID]int64) error { return nil }
func (f *fakeUserRepo) SetInbounds(_ context.Context, _ uuid.UUID, ids []uuid.UUID) error {
	f.bound = ids
	return nil
}
func (f *fakeUserRepo) InboundsFor(context.Context, uuid.UUID) ([]domain.Inbound, error) {
	return f.inbounds, nil
}

type fakeNodeOps struct {
	added   []string
	removed []string
}

func (f *fakeNodeOps) AddUser(_ context.Context, nodeID uuid.UUID, tag string, _ *domain.User) error {
	f.added = append(f.added, nodeID.String()+"/"+tag)
	return nil
}
func (f *fakeNodeOps) RemoveUser(_ context.Context, nodeID uuid.UUID, tag string, _ uuid.UUID) error {
	f.removed = append(f.removed, nodeID.String()+"/"+tag)
	return nil
}

// --- tests ---

func newAdmin(t *testing.T, username, password string, totp bool) *domain.Admin {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	a := &domain.Admin{ID: uuid.New(), Username: username, PasswordHash: hash, Sudo: true}
	if totp {
		secret, _, _ := auth.GenerateTOTP("VortexUI", username)
		a.TOTPSecret = secret
		a.TOTPEnabled = true
	}
	return a
}

func TestAuthLogin(t *testing.T) {
	admin := newAdmin(t, "root", "hunter2", false)
	repo := &fakeAdminRepo{admin: admin}
	iss := auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour)
	svc := NewAuthService(repo, iss)
	ctx := context.Background()

	token, err := svc.Login(ctx, LoginInput{Username: "root", Password: "hunter2"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if claims, err := iss.Verify(token); err != nil || claims.AdminID != admin.ID {
		t.Fatalf("issued token invalid: %v", err)
	}

	if _, err := svc.Login(ctx, LoginInput{Username: "root", Password: "wrong"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("wrong password err = %v, want ErrInvalidCredentials", err)
	}
	if _, err := svc.Login(ctx, LoginInput{Username: "ghost", Password: "x"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("unknown user err = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthLoginRequiresTOTPWhenEnabled(t *testing.T) {
	admin := newAdmin(t, "root", "pw", true)
	svc := NewAuthService(&fakeAdminRepo{admin: admin}, auth.NewIssuer([]byte("0123456789abcdef0123456789abcdef"), time.Hour))
	ctx := context.Background()

	if _, err := svc.Login(ctx, LoginInput{Username: "root", Password: "pw"}); !errors.Is(err, ErrInvalidCredentials) {
		t.Error("login without TOTP should fail when 2FA enabled")
	}
}

func TestUserUpdateMetadataAndDisableDeprovisions(t *testing.T) {
	nodeID := uuid.New()
	uid := uuid.New()
	repo := &fakeUserRepo{
		created:  &domain.User{ID: uid, Username: "alice", Status: domain.UserStatusActive, DataLimit: 100},
		inbounds: []domain.Inbound{{NodeID: nodeID, Tag: "vless-ws"}},
	}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)
	exp := time.Now().Add(24 * time.Hour)

	// Plain metadata edit while staying active: no de-provision.
	u, err := svc.Update(context.Background(), uid, UpdateUserInput{
		Status: domain.UserStatusActive, DataLimit: 500, ExpireAt: &exp, Note: "vip",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if u.DataLimit != 500 || u.Note != "vip" || u.ExpireAt == nil {
		t.Errorf("fields not applied: %+v", u)
	}
	if len(ops.removed) != 0 {
		t.Errorf("active edit should not de-provision, got %v", ops.removed)
	}

	// Flip to disabled: must de-provision from every bound node.
	if _, err := svc.Update(context.Background(), uid, UpdateUserInput{Status: domain.UserStatusDisabled}); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if len(ops.removed) != 1 || ops.removed[0] != nodeID.String()+"/vless-ws" {
		t.Errorf("disable did not de-provision: %v", ops.removed)
	}
}

func TestUserServiceCreateGeneratesCredsAndProvisions(t *testing.T) {
	nodeID := uuid.New()
	inID := uuid.New()
	repo := &fakeUserRepo{inbounds: []domain.Inbound{{ID: inID, NodeID: nodeID, Tag: "vless-ws"}}}
	ops := &fakeNodeOps{}
	svc := NewUserService(repo, ops)

	u, err := svc.Create(context.Background(), CreateUserInput{
		Username: "alice", DataLimit: 1 << 30, InboundIDs: []uuid.UUID{inID},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Server-generated identity.
	if u.Proxies.VMessUUID == uuid.Nil || u.Proxies.VLESSUUID == uuid.Nil {
		t.Error("expected generated vmess/vless uuids")
	}
	if u.SubToken == "" || u.Proxies.TrojanPass == "" {
		t.Error("expected generated sub token and trojan password")
	}
	if repo.created == nil || len(repo.bound) != 1 {
		t.Error("user not persisted/bound")
	}
	// Provisioned on the node behind its inbound.
	if len(ops.added) != 1 || ops.added[0] != nodeID.String()+"/vless-ws" {
		t.Errorf("provisioning = %v, want [%s/vless-ws]", ops.added, nodeID)
	}
}
