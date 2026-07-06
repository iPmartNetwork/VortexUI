package service

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// walletAdminRepo extends stubAdminRepo with the AdminWalletStore surface so
// TopUpWallet/ApplyWalletDelta work in tests. It records wallet deltas applied
// and mutates the in-memory admin's prepaid balances.
type walletAdminRepo struct {
	*stubAdminRepo
	walletCalls   int
	lastDeltaTrf  int64
	lastDeltaUser int
}

func newWalletAdminRepo() *walletAdminRepo {
	return &walletAdminRepo{stubAdminRepo: newStubAdminRepo()}
}

func (w *walletAdminRepo) ApplyWalletDelta(_ context.Context, adminID uuid.UUID, _ *uuid.UUID, deltaTraffic int64, deltaUsers int, _ string) error {
	w.walletCalls++
	w.lastDeltaTrf = deltaTraffic
	w.lastDeltaUser = deltaUsers
	for _, a := range w.byName {
		if a.ID == adminID {
			a.WalletTrafficBytes += deltaTraffic
			a.WalletUserCredits += deltaUsers
		}
	}
	return nil
}

func (w *walletAdminRepo) ListWalletLedger(context.Context, uuid.UUID, int) ([]domain.WalletLedgerEntry, error) {
	return nil, nil
}
func (w *walletAdminRepo) ListChildren(context.Context, uuid.UUID) ([]*domain.Admin, error) {
	return nil, nil
}
func (w *walletAdminRepo) UpdateWebhook(context.Context, uuid.UUID, string, string, bool) error {
	return nil
}
func (w *walletAdminRepo) ListWebhookAdmins(context.Context) ([]domain.Admin, error) {
	return nil, nil
}
func (w *walletAdminRepo) GetPortalBranding(context.Context, uuid.UUID) (*domain.PortalBranding, error) {
	return nil, domain.ErrNotFound
}
func (w *walletAdminRepo) GetPortalBrandingBySlug(context.Context, string) (*domain.PortalBranding, error) {
	return nil, domain.ErrNotFound
}
func (w *walletAdminRepo) UpsertPortalBranding(context.Context, *domain.PortalBranding) error {
	return nil
}

// fakeBillingRepo is a minimal in-memory WalletBillingRepository storing only
// the deposits needed for review/completion tests.
type fakeBillingRepo struct {
	deposits map[uuid.UUID]*domain.WalletDeposit
}

func newFakeBillingRepo() *fakeBillingRepo {
	return &fakeBillingRepo{deposits: map[uuid.UUID]*domain.WalletDeposit{}}
}

func (f *fakeBillingRepo) CreatePackage(context.Context, *domain.WalletPackage) error { return nil }
func (f *fakeBillingRepo) UpdatePackage(context.Context, *domain.WalletPackage) error { return nil }
func (f *fakeBillingRepo) DeletePackage(context.Context, uuid.UUID) error             { return nil }
func (f *fakeBillingRepo) GetPackage(context.Context, uuid.UUID) (*domain.WalletPackage, error) {
	return nil, domain.ErrNotFound
}
func (f *fakeBillingRepo) ListPackages(context.Context, bool) ([]*domain.WalletPackage, error) {
	return nil, nil
}
func (f *fakeBillingRepo) GetBillingSettings(context.Context) (*domain.BillingSettings, error) {
	return &domain.BillingSettings{}, nil
}
func (f *fakeBillingRepo) UpdateBillingSettings(context.Context, *domain.BillingSettings) error {
	return nil
}
func (f *fakeBillingRepo) CreateDeposit(_ context.Context, d *domain.WalletDeposit) error {
	f.deposits[d.ID] = d
	return nil
}
func (f *fakeBillingRepo) UpdateDeposit(_ context.Context, d *domain.WalletDeposit) error {
	f.deposits[d.ID] = d
	return nil
}
func (f *fakeBillingRepo) GetDeposit(_ context.Context, id uuid.UUID) (*domain.WalletDeposit, error) {
	if d, ok := f.deposits[id]; ok {
		return d, nil
	}
	return nil, domain.ErrNotFound
}
func (f *fakeBillingRepo) GetDepositByGatewayID(_ context.Context, gwID string) (*domain.WalletDeposit, error) {
	for _, d := range f.deposits {
		if d.GatewayID == gwID {
			return d, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeBillingRepo) ListDeposits(context.Context, *uuid.UUID, *domain.WalletDepositStatus, int) ([]*domain.WalletDeposit, error) {
	return nil, nil
}

func TestReviewDepositStacksPurchaseOntoResellerMainQuota(t *testing.T) {
	repo := newWalletAdminRepo()
	roleID := uuid.New()
	repo.roles[roleID] = &domain.Role{ID: roleID, Name: "reseller"}
	adminSvc := NewAdminService(repo, nil)
	ctx := context.Background()

	// A sudo reviewer and a non-sudo reseller with an existing allowance.
	reviewer, _, err := adminSvc.Create(ctx, CreateAdminInput{Username: "sudo", Password: "pw", Sudo: true})
	if err != nil {
		t.Fatalf("create sudo: %v", err)
	}
	const startTraffic int64 = 1000
	const startUsers = 5
	reseller, _, err := adminSvc.Create(ctx, CreateAdminInput{
		Username: "reseller", Password: "pw", RoleID: &roleID,
		TrafficQuota: startTraffic, UserQuota: startUsers,
	})
	if err != nil {
		t.Fatalf("create reseller: %v", err)
	}

	billing := newFakeBillingRepo()
	svc := NewWalletBillingService(billing, adminSvc, nil, nil)

	const pkgTraffic int64 = 300
	const pkgUsers = 2
	deposit := &domain.WalletDeposit{
		ID:           uuid.New(),
		AdminID:      reseller.ID,
		PackageName:  "Starter",
		Method:       domain.WalletDepositCardToCard,
		Status:       domain.WalletDepositPending,
		TrafficBytes: pkgTraffic,
		UserCredits:  pkgUsers,
	}
	if err := billing.CreateDeposit(ctx, deposit); err != nil {
		t.Fatalf("seed deposit: %v", err)
	}

	if _, err := svc.ReviewDeposit(ctx, reviewer.ID, deposit.ID, true, "ok"); err != nil {
		t.Fatalf("review approve: %v", err)
	}

	// Main quota stacked additively.
	if reseller.TrafficQuota != startTraffic+pkgTraffic {
		t.Errorf("TrafficQuota = %d, want %d", reseller.TrafficQuota, startTraffic+pkgTraffic)
	}
	if reseller.UserQuota != startUsers+pkgUsers {
		t.Errorf("UserQuota = %d, want %d", reseller.UserQuota, startUsers+pkgUsers)
	}
	// Wallet was still credited (top-up kept as-is).
	if repo.walletCalls != 1 {
		t.Errorf("wallet credit calls = %d, want 1", repo.walletCalls)
	}
	if reseller.WalletTrafficBytes != pkgTraffic || reseller.WalletUserCredits != pkgUsers {
		t.Errorf("wallet balances = (%d,%d), want (%d,%d)",
			reseller.WalletTrafficBytes, reseller.WalletUserCredits, pkgTraffic, pkgUsers)
	}
}

func TestReviewDepositSkipsMainQuotaForSudoTarget(t *testing.T) {
	repo := newWalletAdminRepo()
	adminSvc := NewAdminService(repo, nil)
	ctx := context.Background()

	reviewer, _, err := adminSvc.Create(ctx, CreateAdminInput{Username: "sudo", Password: "pw", Sudo: true})
	if err != nil {
		t.Fatalf("create reviewer: %v", err)
	}
	// Target is itself sudo: AdjustQuota refuses, but the approval must not fail.
	target, _, err := adminSvc.Create(ctx, CreateAdminInput{Username: "sudo2", Password: "pw", Sudo: true})
	if err != nil {
		t.Fatalf("create target: %v", err)
	}

	billing := newFakeBillingRepo()
	svc := NewWalletBillingService(billing, adminSvc, nil, nil)

	deposit := &domain.WalletDeposit{
		ID:           uuid.New(),
		AdminID:      target.ID,
		Method:       domain.WalletDepositCardToCard,
		Status:       domain.WalletDepositPending,
		TrafficBytes: 500,
		UserCredits:  3,
	}
	if err := billing.CreateDeposit(ctx, deposit); err != nil {
		t.Fatalf("seed deposit: %v", err)
	}

	got, err := svc.ReviewDeposit(ctx, reviewer.ID, deposit.ID, true, "ok")
	if err != nil {
		t.Fatalf("review approve should not fail for sudo target: %v", err)
	}
	if got.Status != domain.WalletDepositApproved {
		t.Errorf("status = %v, want approved", got.Status)
	}
	// Sudo target quota left untouched (skipped gracefully).
	if target.TrafficQuota != 0 || target.UserQuota != 0 {
		t.Errorf("sudo target quota changed: traffic=%d users=%d, want 0/0", target.TrafficQuota, target.UserQuota)
	}
	// Wallet credit still happened.
	if repo.walletCalls != 1 {
		t.Errorf("wallet credit calls = %d, want 1", repo.walletCalls)
	}
}

func TestShortDepositID(t *testing.T) {
	id := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	if got := shortDepositID(id); got != "550e8400" {
		t.Errorf("shortDepositID = %q, want 550e8400", got)
	}
}
