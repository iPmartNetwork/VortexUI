package service

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// ErrQuotaExceeded is returned when an admin has hit their user or traffic quota.
var ErrQuotaExceeded = errors.New("admin quota exceeded")

// QuotaChecker verifies that an admin hasn't exceeded their assigned quotas
// before allowing user creation. Sudo admins bypass all quotas.
type QuotaChecker struct {
	admins port.AdminRepository
	users  port.UserRepository
}

// NewQuotaChecker builds a checker.
func NewQuotaChecker(admins port.AdminRepository, users port.UserRepository) *QuotaChecker {
	return &QuotaChecker{admins: admins, users: users}
}

// CheckUserQuota verifies the admin hasn't exceeded their user creation quota.
// adminID is extracted from the JWT claims. Returns nil if quota allows creation.
func (q *QuotaChecker) CheckUserQuota(ctx context.Context, adminID uuid.UUID) error {
	admin, err := q.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	// Sudo admins and admins with quota=0 (unlimited) bypass.
	if admin.Sudo || admin.UserQuota == 0 {
		return nil
	}
	// Count users created by this admin (approximation: total users / for
	// reseller model we'd need a creator_id FK — for now, use total as a
	// shared pool check which is the simplest initial approach).
	_, total, err := q.users.List(ctx, port.UserFilter{Limit: 0})
	if err != nil {
		return err
	}
	if total >= admin.UserQuota {
		return ErrQuotaExceeded
	}
	return nil
}

// CheckTrafficQuota verifies the admin hasn't exceeded their traffic allocation.
func (q *QuotaChecker) CheckTrafficQuota(ctx context.Context, adminID uuid.UUID, requestedLimit int64) error {
	admin, err := q.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if admin.Sudo || admin.TrafficQuota == 0 {
		return nil
	}
	if requestedLimit > admin.TrafficQuota {
		return ErrQuotaExceeded
	}
	return nil
}

// OwnerFilter returns the filter an admin should use when listing users. A sudo
// admin sees all; a quota-limited admin may eventually be scoped to their own
// created users (requires creator_id FK — future enhancement).
func (q *QuotaChecker) OwnerFilter(admin *domain.Admin) port.UserFilter {
	// For now all admins see all users (shared pool). A future release will add
	// creator_id to users and scope non-sudo admins to their own created users.
	return port.UserFilter{}
}
