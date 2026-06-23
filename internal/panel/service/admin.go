package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ErrAdminExists is returned when creating an admin whose username is taken.
var ErrAdminExists = errors.New("admin already exists")

// ErrLastSudo guards against locking everyone out by removing or demoting the
// final full-privilege admin.
var ErrLastSudo = errors.New("cannot remove the last sudo admin")

// ErrWrongPassword is returned when a self password change supplies the wrong
// current password.
var ErrWrongPassword = errors.New("current password is incorrect")

// TOTP self-enrollment errors.
var (
	ErrTOTPAlreadyEnabled = errors.New("2fa already enabled")
	ErrTOTPNotEnrolled    = errors.New("no 2fa enrollment in progress")
	ErrTOTPNotEnabled     = errors.New("2fa is not enabled")
	ErrTOTPInvalidCode    = errors.New("invalid 2fa code")
)

// AdminStore is the richer admin persistence the management endpoints need. It
// embeds the minimal port (used by auth) and adds listing, deletion, the
// last-sudo guard count, and role management. *postgres.AdminRepo satisfies it.
type AdminStore interface {
	port.AdminRepository
	List(ctx context.Context) ([]*domain.Admin, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountSudo(ctx context.Context) (int, error)
	CreateRole(ctx context.Context, role *domain.Role) error
	ListRoles(ctx context.Context) ([]*domain.Role, error)
}

// AdminService manages panel operators: bootstrap, CRUD, and role management.
type AdminService struct {
	admins AdminStore
	now    func() time.Time
}

// NewAdminService wires the service.
func NewAdminService(admins AdminStore) *AdminService {
	return &AdminService{admins: admins, now: time.Now}
}

// CreateAdminInput describes a new operator. The password is hashed here and the
// plaintext is never stored or returned.
type CreateAdminInput struct {
	Username     string
	Password     string
	Sudo         bool
	EnableTOTP   bool
	RoleID       *uuid.UUID // required when Sudo is false
	UserQuota    int
	TrafficQuota int64
}

// Create provisions an admin, refusing to clobber an existing username. When
// EnableTOTP is set it returns the otpauth:// enrollment URL exactly once — it
// is derived from the secret, which is never exposed again afterwards.
func (s *AdminService) Create(ctx context.Context, in CreateAdminInput) (admin *domain.Admin, totpURL string, err error) {
	if in.Username == "" || in.Password == "" {
		return nil, "", errors.New("username and password are required")
	}
	if !in.Sudo {
		if in.RoleID == nil {
			return nil, "", errors.New("role is required for non-sudo admins")
		}
		if _, err := s.admins.GetRole(ctx, *in.RoleID); errors.Is(err, domain.ErrNotFound) {
			return nil, "", errors.New("role not found")
		} else if err != nil {
			return nil, "", fmt.Errorf("load role: %w", err)
		}
	}
	if _, err := s.admins.GetByUsername(ctx, in.Username); err == nil {
		return nil, "", ErrAdminExists
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, "", fmt.Errorf("check existing: %w", err)
	}

	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, "", err
	}
	a := &domain.Admin{
		ID:           uuid.New(),
		Username:     in.Username,
		PasswordHash: hash,
		Sudo:         in.Sudo,
		UserQuota:    in.UserQuota,
		TrafficQuota: in.TrafficQuota,
		CreatedAt:    s.now(),
	}
	if !in.Sudo {
		a.RoleID = in.RoleID
	}
	if in.EnableTOTP {
		secret, url, err := auth.GenerateTOTP("VortexUI", in.Username)
		if err != nil {
			return nil, "", err
		}
		a.TOTPSecret = secret
		a.TOTPEnabled = true
		totpURL = url
	}
	if err := s.admins.Create(ctx, a); err != nil {
		return nil, "", fmt.Errorf("persist admin: %w", err)
	}
	return a, totpURL, nil
}

// List returns all admins.
func (s *AdminService) List(ctx context.Context) ([]*domain.Admin, error) {
	return s.admins.List(ctx)
}

// Get returns one admin.
func (s *AdminService) Get(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	return s.admins.GetByID(ctx, id)
}

// UpdateAdminInput is the mutable subset of an admin. Password is re-hashed only
// when non-empty; 2FA can be turned off here (turning it on is a self-service
// enrollment flow, not an admin override).
type UpdateAdminInput struct {
	Password     string
	Sudo         bool
	UserQuota    int
	TrafficQuota int64
	RoleID       *uuid.UUID
	DisableTOTP  bool
}

// Update applies changes to an admin. Demoting the last sudo admin is refused so
// the panel can never be locked out.
func (s *AdminService) Update(ctx context.Context, id uuid.UUID, in UpdateAdminInput) (*domain.Admin, error) {
	a, err := s.admins.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Sudo && !in.Sudo {
		if err := s.ensureNotLastSudo(ctx); err != nil {
			return nil, err
		}
	}
	if in.Password != "" {
		hash, err := auth.HashPassword(in.Password)
		if err != nil {
			return nil, err
		}
		a.PasswordHash = hash
	}
	a.Sudo = in.Sudo
	a.UserQuota = in.UserQuota
	a.TrafficQuota = in.TrafficQuota
	a.RoleID = in.RoleID
	if in.DisableTOTP {
		a.TOTPEnabled = false
		a.TOTPSecret = ""
	}
	if err := s.admins.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Delete removes an admin, refusing to remove the last sudo admin.
func (s *AdminService) Delete(ctx context.Context, id uuid.UUID) error {
	a, err := s.admins.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if a.Sudo {
		if err := s.ensureNotLastSudo(ctx); err != nil {
			return err
		}
	}
	return s.admins.Delete(ctx, id)
}

func (s *AdminService) ensureNotLastSudo(ctx context.Context) error {
	n, err := s.admins.CountSudo(ctx)
	if err != nil {
		return err
	}
	if n <= 1 {
		return ErrLastSudo
	}
	return nil
}

// CreateRole defines a named permission bundle for non-sudo admins.
func (s *AdminService) CreateRole(ctx context.Context, name string, perms []domain.Permission) (*domain.Role, error) {
	if name == "" {
		return nil, errors.New("role name is required")
	}
	role := &domain.Role{ID: uuid.New(), Name: name, Permissions: perms}
	if err := s.admins.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// ListRoles returns all roles.
func (s *AdminService) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.admins.ListRoles(ctx)
}

// Permissions returns the effective permission set for an admin. Sudo admins
// receive every known permission; role-based admins receive their role's bundle.
func (s *AdminService) Permissions(ctx context.Context, id uuid.UUID) ([]domain.Permission, error) {
	admin, err := s.admins.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if admin.Sudo {
		return []domain.Permission{
			domain.PermUserRead, domain.PermUserWrite,
			domain.PermNodeRead, domain.PermNodeWrite,
			domain.PermInboundRead, domain.PermInboundWrite,
			domain.PermAdminManage, domain.PermSystemRead,
		}, nil
	}
	if admin.RoleID == nil {
		return nil, nil
	}
	role, err := s.admins.GetRole(ctx, *admin.RoleID)
	if err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

// ChangePassword lets an admin change their own password after proving the
// current one — so a hijacked session cannot silently lock the owner out.
func (s *AdminService) ChangePassword(ctx context.Context, adminID uuid.UUID, current, next string) error {
	if len(next) < 6 {
		return errors.New("new password must be at least 6 characters")
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if !auth.CheckPassword(admin.PasswordHash, current) {
		return ErrWrongPassword
	}
	hash, err := auth.HashPassword(next)
	if err != nil {
		return err
	}
	admin.PasswordHash = hash
	return s.admins.Update(ctx, admin)
}

// --- TOTP self-enrollment (an admin enabling 2FA on their own account) ---

// BeginTOTP starts enrollment: it generates and stores a secret but leaves 2FA
// disabled until ConfirmTOTP succeeds, so a mistyped secret can't lock the admin
// out. Returns the otpauth URL to render as a QR code.
func (s *AdminService) BeginTOTP(ctx context.Context, adminID uuid.UUID) (secret, url string, err error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return "", "", err
	}
	if admin.TOTPEnabled {
		return "", "", ErrTOTPAlreadyEnabled
	}
	secret, url, err = auth.GenerateTOTP("VortexUI", admin.Username)
	if err != nil {
		return "", "", err
	}
	admin.TOTPSecret = secret
	admin.TOTPEnabled = false
	if err := s.admins.Update(ctx, admin); err != nil {
		return "", "", err
	}
	return secret, url, nil
}

// ConfirmTOTP activates 2FA once the admin proves they can produce a valid code
// from the enrolled secret.
func (s *AdminService) ConfirmTOTP(ctx context.Context, adminID uuid.UUID, code string) error {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.TOTPSecret == "" {
		return ErrTOTPNotEnrolled
	}
	if !auth.VerifyTOTP(admin.TOTPSecret, code) {
		return ErrTOTPInvalidCode
	}
	admin.TOTPEnabled = true
	return s.admins.Update(ctx, admin)
}

// DisableTOTP turns 2FA off, requiring a valid current code so a hijacked session
// alone cannot remove the second factor. (A locked-out admin is recovered by a
// sudo admin via UpdateAdmin{DisableTOTP}.)
func (s *AdminService) DisableTOTP(ctx context.Context, adminID uuid.UUID, code string) error {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if !admin.TOTPEnabled {
		return ErrTOTPNotEnabled
	}
	if !auth.VerifyTOTP(admin.TOTPSecret, code) {
		return ErrTOTPInvalidCode
	}
	admin.TOTPEnabled = false
	admin.TOTPSecret = ""
	return s.admins.Update(ctx, admin)
}
