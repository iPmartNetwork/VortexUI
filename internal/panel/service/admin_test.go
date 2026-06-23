package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
)

// stubAdminRepo records created admins and answers lookups from that map.
type stubAdminRepo struct {
	byName map[string]*domain.Admin
	roles  map[uuid.UUID]*domain.Role
}

func newStubAdminRepo() *stubAdminRepo {
	return &stubAdminRepo{byName: map[string]*domain.Admin{}, roles: map[uuid.UUID]*domain.Role{}}
}

func (s *stubAdminRepo) Create(_ context.Context, a *domain.Admin) error {
	s.byName[a.Username] = a
	return nil
}
func (s *stubAdminRepo) GetByUsername(_ context.Context, u string) (*domain.Admin, error) {
	if a, ok := s.byName[u]; ok {
		return a, nil
	}
	return nil, domain.ErrNotFound
}
func (s *stubAdminRepo) GetByID(_ context.Context, id uuid.UUID) (*domain.Admin, error) {
	for _, a := range s.byName {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (s *stubAdminRepo) Update(context.Context, *domain.Admin) error              { return nil }
func (s *stubAdminRepo) GetRole(_ context.Context, id uuid.UUID) (*domain.Role, error) {
	if r, ok := s.roles[id]; ok {
		return r, nil
	}
	return nil, domain.ErrNotFound
}
func (s *stubAdminRepo) List(context.Context) ([]*domain.Admin, error) {
	out := make([]*domain.Admin, 0, len(s.byName))
	for _, a := range s.byName {
		out = append(out, a)
	}
	return out, nil
}
func (s *stubAdminRepo) Delete(_ context.Context, id uuid.UUID) error {
	for name, a := range s.byName {
		if a.ID == id {
			delete(s.byName, name)
		}
	}
	return nil
}
func (s *stubAdminRepo) CountSudo(context.Context) (int, error) {
	n := 0
	for _, a := range s.byName {
		if a.Sudo {
			n++
		}
	}
	return n, nil
}
func (s *stubAdminRepo) CreateRole(context.Context, *domain.Role) error    { return nil }
func (s *stubAdminRepo) ListRoles(context.Context) ([]*domain.Role, error) { return nil, nil }
func (s *stubAdminRepo) UpdateRole(context.Context, *domain.Role) error    { return nil }
func (s *stubAdminRepo) DeleteRole(context.Context, uuid.UUID) error       { return nil }
func (s *stubAdminRepo) SetInbounds(context.Context, uuid.UUID, []uuid.UUID) error { return nil }
func (s *stubAdminRepo) ListInboundIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *stubAdminRepo) CountInboundAccess(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}

func TestAdminCreateRequiresRoleForReseller(t *testing.T) {
	repo := newStubAdminRepo()
	roleID := uuid.New()
	repo.roles[roleID] = &domain.Role{ID: roleID, Name: "reseller", Permissions: []domain.Permission{domain.PermUserRead}}
	svc := NewAdminService(repo, nil)
	ctx := context.Background()

	if _, _, err := svc.Create(ctx, CreateAdminInput{Username: "r1", Password: "pw"}); err == nil {
		t.Fatal("expected error when non-sudo admin has no role")
	}
	a, _, err := svc.Create(ctx, CreateAdminInput{Username: "r1", Password: "pw", RoleID: &roleID, UserQuota: 10})
	if err != nil {
		t.Fatalf("create reseller: %v", err)
	}
	if a.RoleID == nil || *a.RoleID != roleID || a.UserQuota != 10 {
		t.Fatalf("reseller fields not saved: %+v", a)
	}
}

func TestAdminCreate(t *testing.T) {
	repo := newStubAdminRepo()
	svc := NewAdminService(repo, nil)
	ctx := context.Background()

	a, url, err := svc.Create(ctx, CreateAdminInput{Username: "root", Password: "pw", Sudo: true})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !a.Sudo || a.PasswordHash == "pw" || a.PasswordHash == "" {
		t.Errorf("admin not set up correctly: %+v", a)
	}
	if !auth.CheckPassword(a.PasswordHash, "pw") {
		t.Error("stored hash does not verify against password")
	}
	if url != "" {
		t.Error("no TOTP url expected when 2FA disabled")
	}
}

func TestAdminCreateRejectsDuplicate(t *testing.T) {
	repo := newStubAdminRepo()
	svc := NewAdminService(repo, nil)
	ctx := context.Background()

	if _, _, err := svc.Create(ctx, CreateAdminInput{Username: "root", Password: "pw", Sudo: true}); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if _, _, err := svc.Create(ctx, CreateAdminInput{Username: "root", Password: "other", Sudo: true}); !errors.Is(err, ErrAdminExists) {
		t.Errorf("duplicate err = %v, want ErrAdminExists", err)
	}
}

func TestAdminUpdateGuardsLastSudo(t *testing.T) {
	repo := newStubAdminRepo()
	svc := NewAdminService(repo, nil)
	ctx := context.Background()

	root, _, _ := svc.Create(ctx, CreateAdminInput{Username: "root", Password: "pw", Sudo: true})

	// Demoting the only sudo admin must be refused.
	if _, err := svc.Update(ctx, root.ID, UpdateAdminInput{Sudo: false}); !errors.Is(err, ErrLastSudo) {
		t.Fatalf("demote last sudo err = %v, want ErrLastSudo", err)
	}
	// Deleting the only sudo admin must be refused too.
	if err := svc.Delete(ctx, root.ID); !errors.Is(err, ErrLastSudo) {
		t.Fatalf("delete last sudo err = %v, want ErrLastSudo", err)
	}

	// With a second sudo present, demotion is allowed.
	if _, _, err := svc.Create(ctx, CreateAdminInput{Username: "root2", Password: "pw", Sudo: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Update(ctx, root.ID, UpdateAdminInput{Sudo: false}); err != nil {
		t.Errorf("demote with another sudo present should succeed, got %v", err)
	}
}

func TestAdminUpdateRehashesPassword(t *testing.T) {
	repo := newStubAdminRepo()
	svc := NewAdminService(repo, nil)
	ctx := context.Background()
	a, _, _ := svc.Create(ctx, CreateAdminInput{Username: "u", Password: "old", Sudo: true})

	updated, err := svc.Update(ctx, a.ID, UpdateAdminInput{Password: "new", Sudo: true})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if !auth.CheckPassword(updated.PasswordHash, "new") || auth.CheckPassword(updated.PasswordHash, "old") {
		t.Error("password was not re-hashed to the new value")
	}
}

func TestTOTPSelfEnrollmentFlow(t *testing.T) {
	repo := newStubAdminRepo()
	svc := NewAdminService(repo, nil)
	ctx := context.Background()
	admin, _, _ := svc.Create(ctx, CreateAdminInput{Username: "u", Password: "pw", Sudo: true})

	// Begin: secret is stored but 2FA stays OFF until confirmed.
	secret, url, err := svc.BeginTOTP(ctx, admin.ID)
	if err != nil || secret == "" || url == "" {
		t.Fatalf("begin: secret=%q url=%q err=%v", secret, url, err)
	}
	if admin.TOTPEnabled {
		t.Fatal("2fa must not be enabled before confirmation")
	}

	// A wrong code does not enable it.
	if err := svc.ConfirmTOTP(ctx, admin.ID, "000000"); !errors.Is(err, ErrTOTPInvalidCode) {
		t.Fatalf("confirm wrong code err = %v, want ErrTOTPInvalidCode", err)
	}

	// The right code enables it.
	code, _ := totp.GenerateCode(secret, time.Now())
	if err := svc.ConfirmTOTP(ctx, admin.ID, code); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !admin.TOTPEnabled {
		t.Fatal("2fa should be enabled after a valid confirmation")
	}

	// Re-enrolling while enabled is refused.
	if _, _, err := svc.BeginTOTP(ctx, admin.ID); !errors.Is(err, ErrTOTPAlreadyEnabled) {
		t.Errorf("re-enroll err = %v, want ErrTOTPAlreadyEnabled", err)
	}

	// Disable requires a valid code; then the secret is cleared.
	code2, _ := totp.GenerateCode(secret, time.Now())
	if err := svc.DisableTOTP(ctx, admin.ID, code2); err != nil {
		t.Fatalf("disable: %v", err)
	}
	if admin.TOTPEnabled || admin.TOTPSecret != "" {
		t.Errorf("after disable: enabled=%v secret=%q, want false/empty", admin.TOTPEnabled, admin.TOTPSecret)
	}
}

func TestAdminCreateWithTOTP(t *testing.T) {
	svc := NewAdminService(newStubAdminRepo(), nil)
	a, url, err := svc.Create(context.Background(), CreateAdminInput{Username: "2fa", Password: "pw", Sudo: true, EnableTOTP: true})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if !a.TOTPEnabled || a.TOTPSecret == "" {
		t.Error("TOTP not enabled on admin")
	}
	if url == "" {
		t.Error("expected otpauth enrollment URL")
	}
}
