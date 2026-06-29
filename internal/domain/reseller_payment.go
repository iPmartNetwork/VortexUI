package domain

import "github.com/google/uuid"

// ResellerPaymentConfig holds per-reseller payment gateway configuration.
// Each reseller can independently configure their card-to-card, crypto, and
// ZarinPal payment details for their shop page.
type ResellerPaymentConfig struct {
	AdminID            uuid.UUID         `json:"admin_id"`
	CardNumber         string            `json:"card_number"`
	CardHolder         string            `json:"card_holder"`
	CardBank           string            `json:"card_bank"`
	CryptoAddresses    map[string]string `json:"crypto_addresses"`    // coin -> address
	ZarinpalMerchantID string            `json:"zarinpal_merchant_id"`
	ManualInstructions string            `json:"manual_instructions"`
	EnabledMethods     []string          `json:"enabled_methods"` // e.g. ["zarinpal","card_to_card","crypto"]
}

// KnownPaymentMethods lists valid values for EnabledMethods.
var KnownPaymentMethods = map[string]bool{
	"zarinpal":     true,
	"card_to_card": true,
	"crypto":       true,
	"manual":       true,
}
