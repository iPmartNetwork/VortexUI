package domain

import (
	"time"

	"github.com/google/uuid"
)

// RewardType defines what the referrer gets.
type RewardType string

const (
	RewardData     RewardType = "data"     // extra data bytes
	RewardDays     RewardType = "days"     // extra days of subscription
	RewardDiscount RewardType = "discount" // % off next purchase
)

// ReferralConfig stores the global referral program settings.
type ReferralConfig struct {
	Enabled       bool       `json:"enabled"`
	RewardType    RewardType `json:"reward_type"`
	RewardAmount  int64      `json:"reward_amount"`  // bytes for data, days for time, % for discount
	MaxReferrals  int        `json:"max_referrals"`  // max referrals per user; 0 = unlimited
	RequirePaid   bool       `json:"require_paid"`   // referred user must complete a purchase before reward
}

// DefaultReferralConfig returns sensible defaults.
func DefaultReferralConfig() ReferralConfig {
	return ReferralConfig{
		Enabled:      false,
		RewardType:   RewardData,
		RewardAmount: 1024 * 1024 * 1024, // 1 GB
		MaxReferrals: 0,
		RequirePaid:  false,
	}
}

// ReferralCode is a unique invite code linked to a user.
type ReferralCode struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Code      string    `json:"code"`
	Uses      int       `json:"uses"`
	MaxUses   int       `json:"max_uses"` // 0 = unlimited
	CreatedAt time.Time `json:"created_at"`
}

// ReferralEvent records a successful referral.
type ReferralEvent struct {
	ID             uuid.UUID  `json:"id"`
	ReferrerID     uuid.UUID  `json:"referrer_id"`
	ReferrerName   string     `json:"referrer_name,omitempty"`
	ReferredID     uuid.UUID  `json:"referred_id"`
	ReferredName   string     `json:"referred_name,omitempty"`
	CodeUsed       string     `json:"code_used"`
	RewardType     RewardType `json:"reward_type"`
	RewardAmount   int64      `json:"reward_amount"`
	RewardApplied  bool       `json:"reward_applied"`
	CreatedAt      time.Time  `json:"created_at"`
}
