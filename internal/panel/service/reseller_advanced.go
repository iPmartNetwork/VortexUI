package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/auth"
	"github.com/vortexui/vortexui/internal/domain"
)

var (
	ErrWalletUserCreditsExceeded    = errors.New("insufficient wallet user credits")
	ErrWalletTrafficExceeded        = errors.New("insufficient wallet traffic credits")
	ErrNotParentAdmin               = errors.New("not authorized for this sub-admin")
	ErrCannotImpersonateSudo        = errors.New("cannot impersonate sudo admin")
)

// Wallet ledger + branding + webhook extensions on AdminStore.
type AdminWalletStore interface {
	ApplyWalletDelta(ctx context.Context, adminID uuid.UUID, actorID *uuid.UUID, deltaTraffic int64, deltaUsers int, reason string) error
	ListWalletLedger(ctx context.Context, adminID uuid.UUID, limit int) ([]domain.WalletLedgerEntry, error)
	ListChildren(ctx context.Context, parentID uuid.UUID) ([]*domain.Admin, error)
	UpdateWebhook(ctx context.Context, adminID uuid.UUID, url, secret string, enabled bool) error
	ListWebhookAdmins(ctx context.Context) ([]domain.Admin, error)
	GetPortalBranding(ctx context.Context, adminID uuid.UUID) (*domain.PortalBranding, error)
	GetPortalBrandingBySlug(ctx context.Context, slug string) (*domain.PortalBranding, error)
	UpsertPortalBranding(ctx context.Context, b *domain.PortalBranding) error
}

// WalletView returns prepaid balances and recent ledger entries.
func (s *AdminService) WalletView(ctx context.Context, adminID uuid.UUID) (*domain.AdminWallet, []domain.WalletLedgerEntry, error) {
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return nil, nil, err
	}
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return &domain.AdminWallet{AdminID: adminID, TrafficBytes: admin.WalletTrafficBytes, UserCredits: admin.WalletUserCredits}, nil, nil
	}
	ledger, _ := ws.ListWalletLedger(ctx, adminID, 30)
	return &domain.AdminWallet{
		AdminID:      adminID,
		TrafficBytes: admin.WalletTrafficBytes,
		UserCredits:  admin.WalletUserCredits,
	}, ledger, nil
}

// TopUpWallet credits a reseller wallet (sudo or parent admin).
func (s *AdminService) TopUpWallet(ctx context.Context, actorID, targetID uuid.UUID, trafficBytes int64, userCredits int, reason string) error {
	if err := s.assertCanManageWallet(ctx, actorID, targetID); err != nil {
		return err
	}
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return errors.New("wallet not available")
	}
	if trafficBytes == 0 && userCredits == 0 {
		return errors.New("nothing to top up")
	}
	return ws.ApplyWalletDelta(ctx, targetID, &actorID, trafficBytes, userCredits, reason)
}

// DebitWallet spends prepaid credits after quota checks pass.
func (s *AdminService) DebitWallet(ctx context.Context, adminID uuid.UUID, trafficBytes int64, userCredits int, reason string) error {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	if userCredits > 0 && admin.WalletUserCredits > 0 {
		if userCredits > admin.WalletUserCredits {
			return ErrWalletUserCreditsExceeded
		}
	}
	if trafficBytes > 0 && admin.WalletTrafficBytes > 0 {
		if trafficBytes > admin.WalletTrafficBytes {
			return ErrWalletTrafficExceeded
		}
	}
	if userCredits == 0 && trafficBytes == 0 {
		return nil
	}
	return ws.ApplyWalletDelta(ctx, adminID, nil, -trafficBytes, -userCredits, reason)
}

func (s *AdminService) assertCanManageWallet(ctx context.Context, actorID, targetID uuid.UUID) error {
	if actorID == targetID {
		return nil
	}
	actor, err := s.admins.GetByID(ctx, actorID)
	if err != nil {
		return err
	}
	if actor.Sudo {
		return nil
	}
	target, err := s.admins.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target.ParentAdminID != nil && *target.ParentAdminID == actorID {
		return nil
	}
	return ErrNotParentAdmin
}

// ListSubAdmins returns child resellers for a parent admin.
func (s *AdminService) ListSubAdmins(ctx context.Context, parentID uuid.UUID) ([]*domain.Admin, error) {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return nil, nil
	}
	return ws.ListChildren(ctx, parentID)
}

// CreateSubAdminInput creates a child reseller under a parent.
type CreateSubAdminInput struct {
	ParentID         uuid.UUID
	Username         string
	Password         string
	RoleID           uuid.UUID
	UserQuota        int
	TrafficQuota     int64
	TrafficQuotaMode domain.TrafficQuotaMode
}

// CreateSubAdmin provisions a child reseller capped by the parent's remaining pool.
func (s *AdminService) CreateSubAdmin(ctx context.Context, in CreateSubAdminInput) (*domain.Admin, error) {
	parent, err := s.admins.GetByID(ctx, in.ParentID)
	if err != nil {
		return nil, err
	}
	if parent.Sudo {
		return nil, errors.New("sudo cannot be sub-admin parent")
	}
	admin, _, err := s.Create(ctx, CreateAdminInput{
		Username: in.Username, Password: in.Password, Sudo: false, RoleID: &in.RoleID,
		UserQuota: in.UserQuota, TrafficQuota: in.TrafficQuota, TrafficQuotaMode: in.TrafficQuotaMode,
		ParentAdminID: &in.ParentID,
	})
	if err != nil {
		return nil, err
	}
	return admin, nil
}

// UpdateWebhookConfig sets outbound webhook for a reseller.
func (s *AdminService) UpdateWebhookConfig(ctx context.Context, adminID uuid.UUID, url, secret string, enabled bool) error {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return errors.New("webhooks not available")
	}
	return ws.UpdateWebhook(ctx, adminID, url, secret, enabled)
}

// GetWebhookConfig returns webhook settings (secret masked in handler).
func (s *AdminService) GetWebhookConfig(ctx context.Context, adminID uuid.UUID) (*domain.Admin, error) {
	return s.admins.GetByID(ctx, adminID)
}

// ListWebhookTargets returns admins with webhooks enabled (for dispatcher).
func (s *AdminService) ListWebhookTargets(ctx context.Context) ([]domain.Admin, error) {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return nil, nil
	}
	return ws.ListWebhookAdmins(ctx)
}

// GetPortalBranding loads branding for an admin.
func (s *AdminService) GetPortalBranding(ctx context.Context, adminID uuid.UUID) (*domain.PortalBranding, error) {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return &domain.PortalBranding{AdminID: adminID}, nil
	}
	b, err := ws.GetPortalBranding(ctx, adminID)
	if errors.Is(err, domain.ErrNotFound) {
		return &domain.PortalBranding{AdminID: adminID}, nil
	}
	return b, err
}

// SavePortalBranding upserts portal branding for a reseller.
func (s *AdminService) SavePortalBranding(ctx context.Context, b *domain.PortalBranding) error {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return errors.New("branding not available")
	}
	return ws.UpsertPortalBranding(ctx, b)
}

// GetPortalBrandingBySlug resolves public portal branding.
func (s *AdminService) GetPortalBrandingBySlug(ctx context.Context, slug string) (*domain.PortalBranding, error) {
	ws, ok := s.admins.(AdminWalletStore)
	if !ok {
		return nil, domain.ErrNotFound
	}
	return ws.GetPortalBrandingBySlug(ctx, slug)
}

// Impersonate issues a short-lived token for sudo support.
func (s *AdminService) Impersonate(ctx context.Context, actorID, targetID uuid.UUID, issuer *auth.Issuer) (string, error) {
	actor, err := s.admins.GetByID(ctx, actorID)
	if err != nil {
		return "", err
	}
	if !actor.Sudo {
		return "", errors.New("impersonation requires sudo")
	}
	target, err := s.admins.GetByID(ctx, targetID)
	if err != nil {
		return "", err
	}
	if target.Sudo {
		return "", ErrCannotImpersonateSudo
	}
	return issuer.IssueImpersonation(target.ID, target.Sudo, target.RoleID, actorID)
}

// StopImpersonation re-issues a normal token for the impersonator.
func (s *AdminService) StopImpersonation(ctx context.Context, impersonatorID uuid.UUID, issuer *auth.Issuer) (string, error) {
	admin, err := s.admins.GetByID(ctx, impersonatorID)
	if err != nil {
		return "", err
	}
	return issuer.Issue(admin.ID, admin.Sudo, admin.RoleID)
}

// enrich wallet checks in AssertCanAddUsers
func (s *AdminService) checkWalletForAdd(ctx context.Context, admin *domain.Admin, addCount int, dataLimitPerUser int64) error {
	if admin.WalletUserCredits <= 0 && admin.WalletTrafficBytes <= 0 {
		return nil
	}
	if admin.WalletUserCredits > 0 && addCount > admin.WalletUserCredits {
		return ErrWalletUserCreditsExceeded
	}
	need := dataLimitPerUser * int64(addCount)
	if admin.WalletTrafficBytes > 0 && need > 0 && need > admin.WalletTrafficBytes {
		return ErrWalletTrafficExceeded
	}
	return nil
}

// WalletDebitAfterUserCreate debits wallet after successful provisioning.
func (s *AdminService) WalletDebitAfterUserCreate(ctx context.Context, adminID uuid.UUID, count int, dataLimit int64) error {
	if adminID == uuid.Nil {
		return nil
	}
	admin, err := s.admins.GetByID(ctx, adminID)
	if err != nil {
		return err
	}
	var users int
	var trafficBytes int64
	if admin.WalletUserCredits > 0 {
		users = count
	}
	if admin.WalletTrafficBytes > 0 && dataLimit > 0 {
		trafficBytes = dataLimit * int64(count)
	}
	if users == 0 && trafficBytes == 0 {
		return nil
	}
	return s.DebitWallet(ctx, adminID, trafficBytes, users, fmt.Sprintf("provision %d user(s)", count))
}
