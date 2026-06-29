package domain

import (
	"time"

	"github.com/google/uuid"
)

// Plan is a predefined subscription template that admins create and users
// purchase. It bundles data limit, duration, price, and optionally inbound
// bindings into a one-click purchase.
type Plan struct {
	ID          uuid.UUID  `json:"id"`
	AdminID     *uuid.UUID `json:"admin_id,omitempty"` // owning admin/reseller (creator)
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`

	// Subscription parameters applied when a user purchases this plan.
	DataLimit     int64         `json:"data_limit"`     // bytes; 0 = unlimited
	Duration      int           `json:"duration_days"`  // days of validity
	DeviceLimit   int           `json:"device_limit"`   // 0 = unlimited
	ResetStrategy ResetStrategy `json:"reset_strategy"`
	InboundIDs    []uuid.UUID   `json:"inbound_ids,omitempty"` // auto-bind to these

	// Pricing (in smallest unit — toman for ZarinPal, cents for crypto).
	PriceToman int64   `json:"price_toman"` // 0 = free
	PriceUSD   float64 `json:"price_usd"`   // for crypto gateways

	// Limits on this plan.
	MaxUsers int  `json:"max_users,omitempty"` // 0 = unlimited stock
	Enabled  bool `json:"enabled"`

	CreatedAt time.Time `json:"created_at"`
}

// Order represents a purchase of a plan by a user (or by an admin on behalf of
// a user). It tracks payment status through the gateway flow.
type Order struct {
	ID        uuid.UUID   `json:"id"`
	UserID    *uuid.UUID  `json:"user_id,omitempty"` // nil for new-user orders
	AdminID   *uuid.UUID  `json:"admin_id,omitempty"`
	PlanID    uuid.UUID   `json:"plan_id"`
	Username  string      `json:"username,omitempty"` // for new-user creation
	Status    OrderStatus `json:"status"`
	Gateway   string      `json:"gateway"`   // "zarinpal" | "nowpayments" | "manual"
	GatewayID string      `json:"gateway_id"` // external transaction/authority ID
	Amount    int64       `json:"amount"`     // paid amount in gateway's unit
	Currency   string      `json:"currency"`   // "IRR" | "USD" | "USDT"
	ProofImage string      `json:"proof_image,omitempty"` // base64 data URL of payment receipt/screenshot
	CreatedAt  time.Time   `json:"created_at"`
	PaidAt     *time.Time  `json:"paid_at,omitempty"`
}

// OrderStatus tracks the payment lifecycle.
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"
	OrderPaid      OrderStatus = "paid"
	OrderFailed    OrderStatus = "failed"
	OrderCancelled OrderStatus = "cancelled"
	OrderRefunded  OrderStatus = "refunded"
)
