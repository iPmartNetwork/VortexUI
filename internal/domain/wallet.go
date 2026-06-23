package domain

import (
	"time"

	"github.com/google/uuid"
)

// AdminWallet is the spendable prepaid balance for a reseller.
type AdminWallet struct {
	AdminID            uuid.UUID `json:"admin_id"`
	TrafficBytes       int64     `json:"traffic_bytes"`
	UserCredits        int       `json:"user_credits"`
}

// WalletLedgerEntry records a wallet credit or debit.
type WalletLedgerEntry struct {
	ID           uuid.UUID  `json:"id"`
	AdminID      uuid.UUID  `json:"admin_id"`
	DeltaTraffic int64      `json:"delta_traffic"`
	DeltaUsers   int        `json:"delta_users"`
	Reason       string     `json:"reason"`
	ActorAdminID *uuid.UUID `json:"actor_admin_id,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PortalBranding customizes the end-user portal per reseller.
type PortalBranding struct {
	AdminID      uuid.UUID `json:"admin_id"`
	PanelTitle   string    `json:"panel_title"`
	LogoURL      string    `json:"logo_url"`
	AccentColor  string    `json:"accent_color"`
	FooterText   string    `json:"footer_text"`
	PortalSlug   string    `json:"portal_slug,omitempty"`
	CustomDomain string    `json:"custom_domain,omitempty"`
}
