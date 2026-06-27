package domain

import (
	"time"

	"github.com/google/uuid"
)

// WalletPackage is a prepaid top-up offer for reseller wallets.
type WalletPackage struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	TrafficBytes int64     `json:"traffic_bytes"`
	UserCredits  int       `json:"user_credits"`
	PriceAmount  int64     `json:"price_amount"`
	Currency     string    `json:"currency"`
	Methods      []string  `json:"methods"`
	Enabled      bool      `json:"enabled"`
	SortOrder    int       `json:"sort_order"`
	CreatedAt    time.Time `json:"created_at"`
}

// BillingSettings holds manual payment instructions (card-to-card, crypto).
type BillingSettings struct {
	CardNumber         string            `json:"card_number"`
	CardHolder         string            `json:"card_holder"`
	CardBank           string            `json:"card_bank"`
	CryptoAddresses    map[string]string `json:"crypto_addresses"`
	ManualInstructions string            `json:"manual_instructions"`
}

// WalletDepositMethod identifies how a reseller pays.
type WalletDepositMethod string

const (
	WalletDepositZarinPal    WalletDepositMethod = "zarinpal"
	WalletDepositCardToCard  WalletDepositMethod = "card_to_card"
	WalletDepositCrypto      WalletDepositMethod = "crypto"
	WalletDepositNowPayments WalletDepositMethod = "nowpayments"
)

// WalletDepositStatus tracks deposit lifecycle.
type WalletDepositStatus string

const (
	WalletDepositPending  WalletDepositStatus = "pending"
	WalletDepositPaid     WalletDepositStatus = "paid"
	WalletDepositApproved WalletDepositStatus = "approved"
	WalletDepositRejected WalletDepositStatus = "rejected"
	WalletDepositFailed   WalletDepositStatus = "failed"
)

// WalletDeposit is a reseller wallet top-up request or online payment.
type WalletDeposit struct {
	ID            uuid.UUID           `json:"id"`
	AdminID       uuid.UUID           `json:"admin_id"`
	AdminUsername string              `json:"admin_username,omitempty"`
	PackageID     *uuid.UUID          `json:"package_id,omitempty"`
	PackageName   string              `json:"package_name,omitempty"`
	Method        WalletDepositMethod `json:"method"`
	Status        WalletDepositStatus `json:"status"`
	Amount        int64               `json:"amount"`
	Currency      string              `json:"currency"`
	TrafficBytes  int64               `json:"traffic_bytes"`
	UserCredits   int               `json:"user_credits"`
	GatewayID     string              `json:"gateway_id,omitempty"`
	TxID          string              `json:"tx_id,omitempty"`
	ProofImage    string              `json:"proof_image,omitempty"`
	ResellerNote  string              `json:"reseller_note,omitempty"`
	AdminNote     string              `json:"admin_note,omitempty"`
	ReviewerID    *uuid.UUID          `json:"reviewer_id,omitempty"`
	ReviewedAt    *time.Time          `json:"reviewed_at,omitempty"`
	PaidAt        *time.Time          `json:"paid_at,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
}
