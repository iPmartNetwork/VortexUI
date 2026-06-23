package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

var (
	ErrPolicyDataLimitExceeded = errors.New("data limit exceeds reseller policy maximum")
	ErrPolicyExpireExceeded    = errors.New("expiry exceeds reseller policy maximum")
	ErrPolicyBulkCreateDenied  = errors.New("bulk user creation not allowed for this reseller")
	ErrPolicyBulkDeleteDenied  = errors.New("bulk user deletion not allowed for this reseller")
	ErrAdminSuspended          = errors.New("admin account is suspended")
)

// ValidateUserPolicy enforces per-reseller caps on user provisioning.
func (s *AdminService) ValidateUserPolicy(ctx context.Context, adminID uuid.UUID, dataLimit int64, expireAt *time.Time) error {
	if adminID == uuid.Nil {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Sudo {
		return nil
	}
	if admin.Suspended {
		return ErrAdminSuspended
	}
	if admin.PolicyMaxDataLimit > 0 && dataLimit > admin.PolicyMaxDataLimit {
		return ErrPolicyDataLimitExceeded
	}
	if admin.PolicyMaxExpireDays > 0 && expireAt != nil {
		max := s.now().Add(time.Duration(admin.PolicyMaxExpireDays) * 24 * time.Hour)
		if expireAt.After(max) {
			return ErrPolicyExpireExceeded
		}
	}
	return nil
}

// PolicyAllowsBulkCreate reports whether bulk/import provisioning is allowed.
func (s *AdminService) PolicyAllowsBulkCreate(ctx context.Context, adminID uuid.UUID) (bool, error) {
	if adminID == uuid.Nil {
		return true, nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return false, err
	}
	return admin.Sudo || admin.PolicyAllowBulkCreate, nil
}

// PolicyAllowsBulkDelete reports whether deleting multiple users at once is allowed.
func (s *AdminService) PolicyAllowsBulkDelete(ctx context.Context, adminID uuid.UUID) (bool, error) {
	if adminID == uuid.Nil {
		return true, nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return false, err
	}
	return admin.Sudo || admin.PolicyAllowBulkDelete, nil
}

// EnsureActiveAdmin rejects suspended reseller accounts.
func (s *AdminService) EnsureActiveAdmin(ctx context.Context, adminID uuid.UUID) error {
	if adminID == uuid.Nil {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if !admin.Sudo && admin.Suspended {
		return ErrAdminSuspended
	}
	return nil
}

// AdminSuspendStore persists suspension state.
type AdminSuspendStore interface {
	Suspend(ctx context.Context, id uuid.UUID, at time.Time, reason string) error
	Unsuspend(ctx context.Context, id uuid.UUID) error
	SetQuotaBreachedAt(ctx context.Context, id uuid.UUID, at *time.Time) error
	ListResellerCandidates(ctx context.Context) ([]*domain.Admin, error)
}

// SuspendAdmin disables a reseller and returns the updated record.
func (s *AdminService) SuspendAdmin(ctx context.Context, id uuid.UUID, reason string) (*domain.Admin, error) {
	st, ok := s.admins.(AdminSuspendStore)
	if !ok {
		return nil, errors.New("suspend not available")
	}
	admin, err := s.admins.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if admin.Sudo {
		return nil, errors.New("cannot suspend sudo admin")
	}
	if err := st.Suspend(ctx, id, s.now(), reason); err != nil {
		return nil, err
	}
	admin.Suspended = true
	t := s.now()
	admin.SuspendedAt = &t
	admin.SuspendReason = reason
	return admin, nil
}

// UnsuspendAdmin re-enables a suspended reseller.
func (s *AdminService) UnsuspendAdmin(ctx context.Context, id uuid.UUID) (*domain.Admin, error) {
	st, ok := s.admins.(AdminSuspendStore)
	if !ok {
		return nil, errors.New("suspend not available")
	}
	if err := st.Unsuspend(ctx, id); err != nil {
		return nil, err
	}
	return s.admins.GetByID(ctx, id)
}
