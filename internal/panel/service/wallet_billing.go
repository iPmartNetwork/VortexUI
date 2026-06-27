package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/payment"
)

const maxProofImageBytes = 512_000

// WalletBillingRepository persists packages, deposits, and billing settings.
type WalletBillingRepository interface {
	CreatePackage(ctx context.Context, p *domain.WalletPackage) error
	UpdatePackage(ctx context.Context, p *domain.WalletPackage) error
	DeletePackage(ctx context.Context, id uuid.UUID) error
	GetPackage(ctx context.Context, id uuid.UUID) (*domain.WalletPackage, error)
	ListPackages(ctx context.Context, enabledOnly bool) ([]*domain.WalletPackage, error)

	GetBillingSettings(ctx context.Context) (*domain.BillingSettings, error)
	UpdateBillingSettings(ctx context.Context, s *domain.BillingSettings) error

	CreateDeposit(ctx context.Context, d *domain.WalletDeposit) error
	UpdateDeposit(ctx context.Context, d *domain.WalletDeposit) error
	GetDeposit(ctx context.Context, id uuid.UUID) (*domain.WalletDeposit, error)
	GetDepositByGatewayID(ctx context.Context, gatewayID string) (*domain.WalletDeposit, error)
	ListDeposits(ctx context.Context, adminID *uuid.UUID, status *domain.WalletDepositStatus, limit int) ([]*domain.WalletDeposit, error)
}

// WalletBillingService manages reseller wallet top-ups.
type WalletBillingService struct {
	repo   WalletBillingRepository
	admins *AdminService
	zarin  *payment.ZarinPal
	crypto *payment.NowPayments
	now    func() time.Time
}

// NewWalletBillingService wires the service.
func NewWalletBillingService(repo WalletBillingRepository, admins *AdminService, zarin *payment.ZarinPal, crypto *payment.NowPayments) *WalletBillingService {
	return &WalletBillingService{repo: repo, admins: admins, zarin: zarin, crypto: crypto, now: time.Now}
}

type CreateWalletPackageInput struct {
	Name         string
	Description  string
	TrafficBytes int64
	UserCredits  int
	PriceAmount  int64
	Currency     string
	Methods      []string
	Enabled      bool
	SortOrder    int
}

func (s *WalletBillingService) CreatePackage(ctx context.Context, in CreateWalletPackageInput) (*domain.WalletPackage, error) {
	if in.Name == "" {
		return nil, errors.New("package name is required")
	}
	if in.TrafficBytes <= 0 && in.UserCredits <= 0 {
		return nil, errors.New("package must grant traffic or user credits")
	}
	cur := strings.ToUpper(strings.TrimSpace(in.Currency))
	if cur == "" {
		cur = "IRR"
	}
	p := &domain.WalletPackage{
		ID:           uuid.New(),
		Name:         in.Name,
		Description:  in.Description,
		TrafficBytes: in.TrafficBytes,
		UserCredits:  in.UserCredits,
		PriceAmount:  in.PriceAmount,
		Currency:     cur,
		Methods:      normalizeMethods(in.Methods),
		Enabled:      in.Enabled,
		SortOrder:    in.SortOrder,
		CreatedAt:    s.now(),
	}
	if err := s.repo.CreatePackage(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *WalletBillingService) UpdatePackage(ctx context.Context, p *domain.WalletPackage) error {
	p.Currency = strings.ToUpper(strings.TrimSpace(p.Currency))
	p.Methods = normalizeMethods(p.Methods)
	return s.repo.UpdatePackage(ctx, p)
}

func (s *WalletBillingService) DeletePackage(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePackage(ctx, id)
}

func (s *WalletBillingService) ListPackages(ctx context.Context, enabledOnly bool) ([]*domain.WalletPackage, error) {
	return s.repo.ListPackages(ctx, enabledOnly)
}

func (s *WalletBillingService) GetPackage(ctx context.Context, id uuid.UUID) (*domain.WalletPackage, error) {
	return s.repo.GetPackage(ctx, id)
}

func (s *WalletBillingService) GetBillingSettings(ctx context.Context) (*domain.BillingSettings, error) {
	return s.repo.GetBillingSettings(ctx)
}

func (s *WalletBillingService) UpdateBillingSettings(ctx context.Context, settings *domain.BillingSettings) error {
	if settings.CryptoAddresses == nil {
		settings.CryptoAddresses = map[string]string{}
	}
	return s.repo.UpdateBillingSettings(ctx, settings)
}

type InitWalletDepositInput struct {
	AdminID      uuid.UUID
	PackageID    uuid.UUID
	Method       domain.WalletDepositMethod
	TxID         string
	ProofImage   string
	ResellerNote string
	CallbackBase string // scheme://host for payment redirect
}

type InitWalletDepositResult struct {
	Deposit     *domain.WalletDeposit
	RedirectURL string
}

func (s *WalletBillingService) InitDeposit(ctx context.Context, in InitWalletDepositInput) (*InitWalletDepositResult, error) {
	pkg, err := s.repo.GetPackage(ctx, in.PackageID)
	if err != nil {
		return nil, err
	}
	if !pkg.Enabled {
		return nil, errors.New("package is disabled")
	}
	if !methodAllowed(pkg.Methods, string(in.Method)) {
		return nil, errors.New("payment method not allowed for this package")
	}

	deposit := &domain.WalletDeposit{
		ID:           uuid.New(),
		AdminID:      in.AdminID,
		PackageID:    &pkg.ID,
		PackageName:  pkg.Name,
		Method:       in.Method,
		Status:       domain.WalletDepositPending,
		Amount:       pkg.PriceAmount,
		Currency:     pkg.Currency,
		TrafficBytes: pkg.TrafficBytes,
		UserCredits:  pkg.UserCredits,
		ResellerNote: strings.TrimSpace(in.ResellerNote),
		CreatedAt:    s.now(),
	}

	switch in.Method {
	case domain.WalletDepositZarinPal:
		if s.zarin == nil {
			return nil, errors.New("zarinpal not configured")
		}
		if pkg.Currency != "IRR" {
			return nil, errors.New("zarinpal only supports IRR packages")
		}
		if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
			return nil, err
		}
		callbackURL := strings.TrimRight(in.CallbackBase, "/") + "/api/payment/wallet/callback?deposit_id=" + deposit.ID.String()
		resp, err := s.zarin.CreatePayment(ctx, payment.PaymentRequest{
			Amount:      pkg.PriceAmount,
			Description: "VortexUI Wallet: " + pkg.Name,
			CallbackURL: callbackURL,
		})
		if err != nil {
			deposit.Status = domain.WalletDepositFailed
			_ = s.repo.UpdateDeposit(ctx, deposit)
			return nil, err
		}
		deposit.GatewayID = resp.Authority
		if err := s.repo.UpdateDeposit(ctx, deposit); err != nil {
			return nil, err
		}
		return &InitWalletDepositResult{Deposit: deposit, RedirectURL: resp.RedirectURL}, nil

	case domain.WalletDepositNowPayments:
		if s.crypto == nil {
			return nil, errors.New("crypto payments not configured")
		}
		if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
			return nil, err
		}
		amount := pkg.PriceAmount
		if pkg.Currency == "USD" {
			// price_amount stored as cents for USD packages
		} else if pkg.Currency != "USD" {
			amount = pkg.PriceAmount // best-effort; admin sets USD-priced packages
		}
		callbackURL := strings.TrimRight(in.CallbackBase, "/") + "/api/payment/wallet/callback?deposit_id=" + deposit.ID.String()
		resp, err := s.crypto.CreatePayment(ctx, payment.PaymentRequest{
			Amount:      amount,
			Description: "VortexUI Wallet: " + pkg.Name,
			CallbackURL: callbackURL,
		})
		if err != nil {
			deposit.Status = domain.WalletDepositFailed
			_ = s.repo.UpdateDeposit(ctx, deposit)
			return nil, err
		}
		deposit.GatewayID = resp.Authority
		if err := s.repo.UpdateDeposit(ctx, deposit); err != nil {
			return nil, err
		}
		return &InitWalletDepositResult{Deposit: deposit, RedirectURL: resp.RedirectURL}, nil

	case domain.WalletDepositCardToCard, domain.WalletDepositCrypto:
		txID := strings.TrimSpace(in.TxID)
		proof := strings.TrimSpace(in.ProofImage)
		if txID == "" && proof == "" {
			return nil, errors.New("tx_id or proof_image is required")
		}
		if len(proof) > maxProofImageBytes {
			return nil, errors.New("proof image too large (max 500KB)")
		}
		deposit.TxID = txID
		deposit.ProofImage = proof
		if err := s.repo.CreateDeposit(ctx, deposit); err != nil {
			return nil, err
		}
		return &InitWalletDepositResult{Deposit: deposit}, nil

	default:
		return nil, errors.New("unknown payment method")
	}
}

func (s *WalletBillingService) GetDeposit(ctx context.Context, id uuid.UUID) (*domain.WalletDeposit, error) {
	return s.repo.GetDeposit(ctx, id)
}

func (s *WalletBillingService) ListDeposits(ctx context.Context, adminID *uuid.UUID, status *domain.WalletDepositStatus, limit int) ([]*domain.WalletDeposit, error) {
	return s.repo.ListDeposits(ctx, adminID, status, limit)
}

func (s *WalletBillingService) ReviewDeposit(ctx context.Context, reviewerID, depositID uuid.UUID, approve bool, note string) (*domain.WalletDeposit, error) {
	admin, err := s.admins.Get(ctx, reviewerID)
	if err != nil {
		return nil, err
	}
	if !admin.Sudo {
		return nil, errors.New("review requires sudo")
	}
	deposit, err := s.repo.GetDeposit(ctx, depositID)
	if err != nil {
		return nil, err
	}
	if deposit.Status != domain.WalletDepositPending {
		return nil, errors.New("deposit is not pending")
	}
	now := s.now()
	deposit.ReviewerID = &reviewerID
	deposit.ReviewedAt = &now
	deposit.AdminNote = strings.TrimSpace(note)
	if approve {
		deposit.Status = domain.WalletDepositApproved
		deposit.PaidAt = &now
		reason := fmt.Sprintf("deposit %s approved", deposit.ID.String()[:8])
		if deposit.PackageName != "" {
			reason = deposit.PackageName + " (" + reason + ")"
		}
		if err := s.admins.TopUpWallet(ctx, reviewerID, deposit.AdminID, deposit.TrafficBytes, deposit.UserCredits, reason); err != nil {
			return nil, err
		}
	} else {
		deposit.Status = domain.WalletDepositRejected
	}
	if err := s.repo.UpdateDeposit(ctx, deposit); err != nil {
		return nil, err
	}
	return deposit, nil
}

func (s *WalletBillingService) CompleteOnlineDeposit(ctx context.Context, depositID uuid.UUID, gatewayID string, amount int64) error {
	deposit, err := s.repo.GetDeposit(ctx, depositID)
	if err != nil {
		return err
	}
	if deposit.Status != domain.WalletDepositPending {
		return nil
	}
	if deposit.GatewayID != "" && gatewayID != "" && deposit.GatewayID != gatewayID {
		return errors.New("gateway id mismatch")
	}

	var verifyErr error
	switch deposit.Method {
	case domain.WalletDepositZarinPal:
		if s.zarin == nil {
			return errors.New("zarinpal not configured")
		}
		_, verifyErr = s.zarin.VerifyPayment(ctx, deposit.GatewayID, amount)
	case domain.WalletDepositNowPayments:
		if s.crypto == nil {
			return errors.New("crypto not configured")
		}
		_, verifyErr = s.crypto.VerifyPayment(ctx, deposit.GatewayID, amount)
	default:
		return errors.New("not an online deposit")
	}
	if verifyErr != nil {
		deposit.Status = domain.WalletDepositFailed
		_ = s.repo.UpdateDeposit(ctx, deposit)
		return verifyErr
	}

	now := s.now()
	deposit.Status = domain.WalletDepositPaid
	deposit.PaidAt = &now
	if err := s.repo.UpdateDeposit(ctx, deposit); err != nil {
		return err
	}
	reason := fmt.Sprintf("online deposit %s", deposit.ID.String()[:8])
	if deposit.PackageName != "" {
		reason = deposit.PackageName + " (" + reason + ")"
	}
	return s.admins.TopUpWallet(ctx, deposit.AdminID, deposit.AdminID, deposit.TrafficBytes, deposit.UserCredits, reason)
}

func (s *WalletBillingService) CompleteDepositByGatewayID(ctx context.Context, gatewayID string) error {
	deposit, err := s.repo.GetDepositByGatewayID(ctx, gatewayID)
	if err != nil {
		return err
	}
	return s.CompleteOnlineDeposit(ctx, deposit.ID, gatewayID, deposit.Amount)
}

func normalizeMethods(methods []string) []string {
	if len(methods) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(methods))
	seen := map[string]bool{}
	for _, m := range methods {
		m = strings.ToLower(strings.TrimSpace(m))
		if m == "" || seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out
}

func methodAllowed(methods []string, want string) bool {
	want = strings.ToLower(want)
	for _, m := range methods {
		if strings.ToLower(m) == want {
			return true
		}
	}
	return false
}
