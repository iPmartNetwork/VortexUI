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

// Reseller quota errors.
var (
	ErrUserQuotaExceeded    = errors.New("user quota exceeded")
	ErrTrafficQuotaExceeded = errors.New("traffic quota exceeded")
)

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
	UpdateRole(ctx context.Context, role *domain.Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error
	SetInbounds(ctx context.Context, adminID uuid.UUID, inboundIDs []uuid.UUID) error
	ListInboundIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	CountInboundAccess(ctx context.Context, adminID uuid.UUID, inboundIDs []uuid.UUID) (int64, error)
	SetPlans(ctx context.Context, adminID uuid.UUID, planIDs []uuid.UUID) error
	ListPlanIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	CountPlanAccess(ctx context.Context, adminID uuid.UUID, planIDs []uuid.UUID) (int64, error)
	SetNodes(ctx context.Context, adminID uuid.UUID, nodeIDs []uuid.UUID) error
	ListNodeIDs(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error)
	CountNodeAccess(ctx context.Context, adminID uuid.UUID, nodeIDs []uuid.UUID) (int64, error)
}

// AdminUserStatsReader loads aggregated usage for reseller quota checks.
type AdminUserStatsReader interface {
	StatsForAdmin(ctx context.Context, adminID uuid.UUID) (domain.AdminUserStats, error)
	StatsByStatusForAdmin(ctx context.Context, adminID uuid.UUID) (map[string]int64, error)
	TopUsersForAdmin(ctx context.Context, adminID uuid.UUID, limit int32) ([]domain.ResellerTopUser, error)
	CountExpiringSoonForAdmin(ctx context.Context, adminID uuid.UUID) (int64, error)
	CountCreatedSinceForAdmin(ctx context.Context, adminID uuid.UUID, since time.Time) (int64, error)
}

// AdminService manages panel operators: bootstrap, CRUD, and role management.
type AdminService struct {
	admins AdminStore
	users  AdminUserStatsReader
	now    func() time.Time
}

// NewAdminService wires the service. users may be nil (quota features disabled).
func NewAdminService(admins AdminStore, users AdminUserStatsReader) *AdminService {
	return &AdminService{admins: admins, users: users, now: time.Now}
}

// CreateAdminInput describes a new operator. The password is hashed here and the
// plaintext is never stored or returned.
type CreateAdminInput struct {
	Username     string
	Password     string
	Sudo         bool
	EnableTOTP   bool
	RoleID       *uuid.UUID // required when Sudo is false
	UserQuota          int
	TrafficQuota       int64
	TrafficQuotaMode   domain.TrafficQuotaMode
	InboundIDs         []uuid.UUID // optional allowlist for resellers
	NodeIDs            []uuid.UUID
	PlanIDs            []uuid.UUID
	ParentAdminID      *uuid.UUID
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
		UserQuota:        in.UserQuota,
		TrafficQuota:     in.TrafficQuota,
		TrafficQuotaMode: in.TrafficQuotaMode,
		CreatedAt:        s.now(),
	}
	if in.ParentAdminID != nil {
		a.ParentAdminID = in.ParentAdminID
	}
	if a.TrafficQuotaMode == "" {
		a.TrafficQuotaMode = domain.TrafficQuotaAllocated
	}
	if !in.Sudo {
		a.RoleID = in.RoleID
		a.PolicyAllowBulkDelete = true
		a.PolicyAllowBulkCreate = true
		a.AutoSuspendEnabled = true
		a.SuspendGraceMinutes = 60
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
	if !in.Sudo && len(in.InboundIDs) > 0 {
		if err := s.admins.SetInbounds(ctx, a.ID, in.InboundIDs); err != nil {
			return nil, "", fmt.Errorf("set inbounds: %w", err)
		}
	}
	if !in.Sudo && len(in.NodeIDs) > 0 {
		if err := s.admins.SetNodes(ctx, a.ID, in.NodeIDs); err != nil {
			return nil, "", fmt.Errorf("set nodes: %w", err)
		}
	}
	if !in.Sudo && len(in.PlanIDs) > 0 {
		if err := s.admins.SetPlans(ctx, a.ID, in.PlanIDs); err != nil {
			return nil, "", fmt.Errorf("set plans: %w", err)
		}
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
	UserQuota          int
	TrafficQuota       int64
	TrafficQuotaMode   *domain.TrafficQuotaMode
	RoleID             *uuid.UUID
	DisableTOTP          bool
	InboundIDs           *[]uuid.UUID // nil = leave allowlist unchanged
	NodeIDs              *[]uuid.UUID
	PlanIDs              *[]uuid.UUID
	PolicyMaxDataLimit          *int64
	PolicyMaxExpireDays         *int
	PolicyAllowBulkDelete       *bool
	PolicyAllowBulkCreate       *bool
	AutoSuspendEnabled          *bool
	IPViolationSuspendThreshold *int
	SuspendGraceMinutes         *int
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
	if in.TrafficQuotaMode != nil {
		a.TrafficQuotaMode = *in.TrafficQuotaMode
		if a.TrafficQuotaMode == "" {
			a.TrafficQuotaMode = domain.TrafficQuotaAllocated
		}
	}
	a.RoleID = in.RoleID
	if in.DisableTOTP {
		a.TOTPEnabled = false
		a.TOTPSecret = ""
	}
	if in.PolicyMaxDataLimit != nil {
		a.PolicyMaxDataLimit = *in.PolicyMaxDataLimit
	}
	if in.PolicyMaxExpireDays != nil {
		a.PolicyMaxExpireDays = *in.PolicyMaxExpireDays
	}
	if in.PolicyAllowBulkDelete != nil {
		a.PolicyAllowBulkDelete = *in.PolicyAllowBulkDelete
	}
	if in.PolicyAllowBulkCreate != nil {
		a.PolicyAllowBulkCreate = *in.PolicyAllowBulkCreate
	}
	if in.AutoSuspendEnabled != nil {
		a.AutoSuspendEnabled = *in.AutoSuspendEnabled
	}
	if in.IPViolationSuspendThreshold != nil {
		a.IPViolationSuspendThreshold = *in.IPViolationSuspendThreshold
	}
	if in.SuspendGraceMinutes != nil {
		a.SuspendGraceMinutes = *in.SuspendGraceMinutes
	}
	if err := s.admins.Update(ctx, a); err != nil {
		return nil, err
	}
	if in.InboundIDs != nil && !a.Sudo {
		if err := s.admins.SetInbounds(ctx, a.ID, *in.InboundIDs); err != nil {
			return nil, fmt.Errorf("set inbounds: %w", err)
		}
	}
	if in.NodeIDs != nil && !a.Sudo {
		if err := s.admins.SetNodes(ctx, a.ID, *in.NodeIDs); err != nil {
			return nil, fmt.Errorf("set nodes: %w", err)
		}
	}
	if in.PlanIDs != nil && !a.Sudo {
		if err := s.admins.SetPlans(ctx, a.ID, *in.PlanIDs); err != nil {
			return nil, fmt.Errorf("set plans: %w", err)
		}
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

// UpdateRole replaces a role's name and permission bundle.
func (s *AdminService) UpdateRole(ctx context.Context, id uuid.UUID, name string, perms []domain.Permission) (*domain.Role, error) {
	role, err := s.admins.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != "" {
		role.Name = name
	}
	if perms != nil {
		role.Permissions = perms
	}
	if err := s.admins.UpdateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

// DeleteRole removes a role. Admins bound to it keep working with role_id cleared
// (ON DELETE SET NULL).
func (s *AdminService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	if _, err := s.admins.GetRole(ctx, id); err != nil {
		return err
	}
	return s.admins.DeleteRole(ctx, id)
}

// InboundIDsForAdmin returns inbound IDs a reseller may assign to users.
func (s *AdminService) InboundIDsForAdmin(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return s.admins.ListInboundIDs(ctx, adminID)
}

// ValidateInboundAccess ensures every inbound is on the admin's allowlist.
func (s *AdminService) ValidateInboundAccess(ctx context.Context, adminID uuid.UUID, inboundIDs []uuid.UUID) error {
	if len(inboundIDs) == 0 {
		return nil
	}
	n, err := s.admins.CountInboundAccess(ctx, adminID, inboundIDs)
	if err != nil {
		return err
	}
	if int(n) != len(inboundIDs) {
		return errors.New("inbound not allowed for this admin")
	}
	return nil
}

// PlanIDsForAdmin returns plan IDs a reseller may sell.
func (s *AdminService) PlanIDsForAdmin(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return s.admins.ListPlanIDs(ctx, adminID)
}

// NodeIDsForAdmin returns node IDs a reseller may use.
func (s *AdminService) NodeIDsForAdmin(ctx context.Context, adminID uuid.UUID) ([]uuid.UUID, error) {
	return s.admins.ListNodeIDs(ctx, adminID)
}

// ValidatePlanAccess ensures every plan is on the admin's allowlist.
func (s *AdminService) ValidatePlanAccess(ctx context.Context, adminID uuid.UUID, planID uuid.UUID) error {
	n, err := s.admins.CountPlanAccess(ctx, adminID, []uuid.UUID{planID})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.New("plan not allowed for this admin")
	}
	return nil
}

// ValidateNodeAccess ensures every node is on the admin's allowlist.
func (s *AdminService) ValidateNodeAccess(ctx context.Context, adminID uuid.UUID, nodeIDs []uuid.UUID) error {
	if len(nodeIDs) == 0 {
		return nil
	}
	n, err := s.admins.CountNodeAccess(ctx, adminID, nodeIDs)
	if err != nil {
		return err
	}
	if int(n) != len(nodeIDs) {
		return errors.New("node not allowed for this admin")
	}
	return nil
}

// FilterPlanIDs returns the subset of planIDs the admin may access.
func (s *AdminService) FilterPlanIDs(ctx context.Context, adminID uuid.UUID, all []uuid.UUID) ([]uuid.UUID, error) {
	allowed, err := s.admins.ListPlanIDs(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	allow := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allow[id] = struct{}{}
	}
	out := make([]uuid.UUID, 0)
	for _, id := range all {
		if _, ok := allow[id]; ok {
			out = append(out, id)
		}
	}
	return out, nil
}

// FilterPlans returns plans a reseller may sell (empty allowlist = all).
func (s *AdminService) FilterPlans(ctx context.Context, adminID uuid.UUID, plans []*domain.Plan) ([]*domain.Plan, error) {
	allowed, err := s.admins.ListPlanIDs(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if len(allowed) == 0 {
		return plans, nil
	}
	allow := make(map[uuid.UUID]struct{}, len(allowed))
	for _, id := range allowed {
		allow[id] = struct{}{}
	}
	out := make([]*domain.Plan, 0, len(plans))
	for _, p := range plans {
		if _, ok := allow[p.ID]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

// AdjustQuota applies deltas to a reseller's pool limits (sudo quick-adjust).
func (s *AdminService) AdjustQuota(ctx context.Context, id uuid.UUID, userDelta int, trafficDelta int64) (*domain.Admin, error) {
	a, err := s.admins.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Sudo {
		return nil, errors.New("cannot adjust sudo admin quota")
	}
	if userDelta != 0 {
		a.UserQuota += userDelta
		if a.UserQuota < 0 {
			a.UserQuota = 0
		}
	}
	if trafficDelta != 0 {
		a.TrafficQuota += trafficDelta
		if a.TrafficQuota < 0 {
			a.TrafficQuota = 0
		}
	}
	if err := s.admins.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// ResellerDashboard builds the reseller home summary.
func (s *AdminService) ResellerDashboard(ctx context.Context, adminID uuid.UUID) (*domain.ResellerDashboard, error) {
	quota, err := s.QuotaUsage(ctx, adminID)
	if err != nil {
		return nil, err
	}
	if s.users == nil {
		return &domain.ResellerDashboard{Quota: *quota}, nil
	}
	byStatus, err := s.users.StatsByStatusForAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	top, err := s.users.TopUsersForAdmin(ctx, adminID, 5)
	if err != nil {
		return nil, err
	}
	expiring, err := s.users.CountExpiringSoonForAdmin(ctx, adminID)
	if err != nil {
		return nil, err
	}
	now := s.now()
	n7, err := s.users.CountCreatedSinceForAdmin(ctx, adminID, now.Add(-7*24*time.Hour))
	if err != nil {
		return nil, err
	}
	n30, err := s.users.CountCreatedSinceForAdmin(ctx, adminID, now.Add(-30*24*time.Hour))
	if err != nil {
		return nil, err
	}
	return &domain.ResellerDashboard{
		Quota:         *quota,
		UsersByStatus: byStatus,
		TopUsers:      top,
		ExpiringSoon:  expiring,
		NewUsers7d:    n7,
		NewUsers30d:   n30,
	}, nil
}

// QuotaUsage returns limits and live usage for one admin.
func (s *AdminService) QuotaUsage(ctx context.Context, adminID uuid.UUID) (*domain.AdminQuotaUsage, error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	return s.quotaUsageFor(ctx, admin)
}

// ListResellerQuotaUsage returns usage for every non-sudo admin.
func (s *AdminService) ListResellerQuotaUsage(ctx context.Context) ([]*domain.AdminQuotaUsage, error) {
	admins, err := s.admins.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AdminQuotaUsage, 0)
	for _, a := range admins {
		if a.Sudo {
			continue
		}
		u, err := s.quotaUsageFor(ctx, a)
		if err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, nil
}

func (s *AdminService) quotaUsageFor(ctx context.Context, admin *domain.Admin) (*domain.AdminQuotaUsage, error) {
	u := &domain.AdminQuotaUsage{
		AdminID:          admin.ID,
		Username:         admin.Username,
		UserQuota:        admin.UserQuota,
		TrafficQuota:     admin.TrafficQuota,
		TrafficQuotaMode: string(admin.TrafficQuotaMode),
	}
	if s.users != nil {
		stats, err := s.users.StatsForAdmin(ctx, admin.ID)
		if err != nil {
			return nil, err
		}
		u.UserCount = stats.UserCount
		u.TrafficUsed = stats.TrafficUsed
		u.TrafficAllocated = stats.TrafficAllocated
	}
	if admin.UserQuota > 0 {
		rem := int64(admin.UserQuota) - u.UserCount
		if rem < 0 {
			rem = 0
		}
		u.UsersRemaining = &rem
	}
	if admin.TrafficQuota > 0 {
		var used int64
		if admin.TrafficQuotaMode == domain.TrafficQuotaConsumed {
			used = u.TrafficUsed
		} else {
			used = u.TrafficAllocated
		}
		rem := admin.TrafficQuota - used
		if rem < 0 {
			rem = 0
		}
		u.TrafficRemaining = &rem
	}
	return u, nil
}

// AssertCanAddUsers checks reseller user/traffic limits before provisioning.
func (s *AdminService) AssertCanAddUsers(ctx context.Context, adminID uuid.UUID, addCount int, dataLimitPerUser int64) error {
	if adminID == uuid.Nil || s.users == nil {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Sudo {
		return nil
	}
	stats, err := s.users.StatsForAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.UserQuota > 0 && stats.UserCount+int64(addCount) > int64(admin.UserQuota) {
		return ErrUserQuotaExceeded
	}
	if admin.TrafficQuota > 0 && admin.TrafficQuotaMode != domain.TrafficQuotaConsumed {
		addAlloc := dataLimitPerUser * int64(addCount)
		if stats.TrafficAllocated+addAlloc > admin.TrafficQuota {
			return ErrTrafficQuotaExceeded
		}
	}
	return s.checkWalletForAdd(ctx, admin, addCount, dataLimitPerUser)
}

// AssertCanSetDataLimit checks traffic pool when changing a user's data cap.
func (s *AdminService) AssertCanSetDataLimit(ctx context.Context, adminID uuid.UUID, currentLimit, newLimit int64) error {
	if adminID == uuid.Nil || s.users == nil || newLimit <= 0 {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Sudo || admin.TrafficQuota <= 0 {
		return nil
	}
	if admin.TrafficQuotaMode == domain.TrafficQuotaConsumed {
		return nil
	}
	stats, err := s.users.StatsForAdmin(ctx, adminID)
	if err != nil {
		return err
	}
	next := stats.TrafficAllocated - currentLimit + newLimit
	if next > admin.TrafficQuota {
		return ErrTrafficQuotaExceeded
	}
	return nil
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
