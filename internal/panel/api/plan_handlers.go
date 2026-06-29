package api

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/service"
	"github.com/vortexui/vortexui/internal/payment"
)

// --- Plans CRUD ---

type createPlanRequest struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DataLimit     int64     `json:"data_limit"`
	DurationDays  int       `json:"duration_days"`
	DeviceLimit   int       `json:"device_limit"`
	ResetStrategy string    `json:"reset_strategy"`
	InboundIDs    []string  `json:"inbound_ids"`
	PriceToman    int64     `json:"price_toman"`
	PriceUSD      float64   `json:"price_usd"`
	MaxUsers      int       `json:"max_users"`
}

func (h *Handlers) CreatePlan(c echo.Context) error {
	var req createPlanRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	ids, _ := parseUUIDs(req.InboundIDs)
	p, err := h.Plans.CreatePlan(c.Request().Context(), service.CreatePlanInput{
		Name:          req.Name,
		Description:   req.Description,
		DataLimit:     req.DataLimit,
		DurationDays:  req.DurationDays,
		DeviceLimit:   req.DeviceLimit,
		ResetStrategy: domain.ResetStrategy(req.ResetStrategy),
		InboundIDs:    ids,
		PriceToman:    req.PriceToman,
		PriceUSD:      req.PriceUSD,
		MaxUsers:      req.MaxUsers,
		AdminID:       &claims.AdminID,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, errString(err))
	}
	return c.JSON(http.StatusCreated, echo.Map{"plan": p})
}

func (h *Handlers) ListPlans(c echo.Context) error {
	plans, err := h.Plans.ListPlans(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		filtered := plans[:0]
		for _, p := range plans {
			if p.AdminID != nil && *p.AdminID == claims.AdminID {
				filtered = append(filtered, p)
			}
		}
		plans = filtered
	}
	return c.JSON(http.StatusOK, echo.Map{"plans": plans})
}

func (h *Handlers) DeletePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
	}
	plan, err := h.Plans.GetPlan(c.Request().Context(), id)
	if err != nil || plan == nil {
		return echo.NewHTTPError(http.StatusNotFound, "plan not found")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		if plan.AdminID == nil || *plan.AdminID != claims.AdminID {
			return echo.NewHTTPError(http.StatusForbidden, "you do not own this plan")
		}
	}
	if err := h.Plans.DeletePlan(c.Request().Context(), id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "delete failed")
	}
	return c.NoContent(http.StatusNoContent)
}

// --- Orders ---

func (h *Handlers) ListOrders(c echo.Context) error {
	var userID *uuid.UUID
	if uid := c.QueryParam("user_id"); uid != "" {
		id, err := uuid.Parse(uid)
		if err == nil {
			userID = &id
		}
	}
	orders, err := h.Plans.ListOrders(c.Request().Context(), userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	if claims := claimsFrom(c); claims != nil && !claims.Sudo {
		filtered := orders[:0]
		for _, o := range orders {
			if o.AdminID != nil && *o.AdminID == claims.AdminID {
				filtered = append(filtered, o)
			}
		}
		orders = filtered
	}
	return c.JSON(http.StatusOK, echo.Map{"orders": orders})
}

// --- Payment (public purchase flow) ---

type purchaseRequest struct {
	PlanID     string `json:"plan_id"`
	Username   string `json:"username"`   // for new users
	SubToken   string `json:"sub_token"`  // for existing user renewal (optional)
	Gateway    string `json:"gateway"`    // "zarinpal" | "nowpayments" | "card_to_card" | "crypto"
	TxID       string `json:"tx_id"`      // manual methods: transaction ID
	ProofImage string `json:"proof_image"` // manual methods: base64 image of receipt
	CryptoCoin string `json:"crypto_coin"` // crypto method: which coin (btc, usdt, etc.)
}

// InitPurchase creates an order and returns the payment redirect URL.
// Public endpoint (no auth) — anyone can buy a plan.
func (h *Handlers) InitPurchase(c echo.Context) error {
	var req purchaseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid plan_id")
	}
	plan, err := h.Plans.GetPlan(c.Request().Context(), planID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "plan not found")
	}

	// Resolve existing user for renewal (sub_token) or validate new-user input.
	var userID *uuid.UUID
	var adminID *uuid.UUID
	if req.SubToken != "" {
		u, err := h.Repo.GetBySubToken(c.Request().Context(), req.SubToken)
		if err != nil || u == nil {
			return echo.NewHTTPError(http.StatusNotFound, "user not found for this token")
		}
		userID = &u.ID
		adminID = u.AdminID
	} else if req.Username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "must provide username (new user) or sub_token (renewal)")
	}

	// Load per-reseller payment config (if user belongs to a reseller).
	var resellerCfg *domain.ResellerPaymentConfig
	if adminID != nil && h.ResellerPayment != nil {
		cfg, err := h.ResellerPayment.GetPaymentConfig(c.Request().Context(), *adminID)
		if err == nil && cfg != nil {
			resellerCfg = cfg
		}
	}

	// Guard against cross-reseller purchases: a user may only buy plans owned by
	// their own reseller.
	if plan.AdminID != nil && adminID != nil && *plan.AdminID != *adminID {
		return echo.NewHTTPError(http.StatusForbidden, "this plan is not available for your account")
	}

	// Handle gateway selection.
	switch req.Gateway {
	case "card_to_card":
		// Manual card-to-card: create order as pending, no gateway contact.
		txID := req.TxID
		if txID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "tx_id is required for card_to_card")
		}
		order := &domain.Order{
			ID:        uuid.New(),
			PlanID:    planID,
			UserID:    userID,
			AdminID:   adminID,
			Username:  req.Username,
			Status:    domain.OrderPending,
			Gateway:   "card_to_card",
			GatewayID: txID,
			Amount:    plan.PriceToman,
			Currency:  "IRR",
			CreatedAt: time.Now(),
		}
		if err := h.Plans.CreateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "create order failed")
		}
		return c.JSON(http.StatusOK, echo.Map{
			"order_id": order.ID,
			"status":   "pending_review",
		})

	case "crypto":
		// Manual crypto: create order as pending, store coin:txid in GatewayID.
		txID := req.TxID
		coin := req.CryptoCoin
		if txID == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "tx_id is required for crypto")
		}
		if coin == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "crypto_coin is required")
		}
		gatewayID := coin + ":" + txID
		order := &domain.Order{
			ID:        uuid.New(),
			PlanID:    planID,
			UserID:    userID,
			AdminID:   adminID,
			Username:  req.Username,
			Status:    domain.OrderPending,
			Gateway:   "crypto",
			GatewayID: gatewayID,
			Amount:    int64(plan.PriceUSD * 100),
			Currency:  "USD",
			CreatedAt: time.Now(),
		}
		if err := h.Plans.CreateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "create order failed")
		}
		return c.JSON(http.StatusOK, echo.Map{
			"order_id": order.ID,
			"status":   "pending_review",
		})

	case "zarinpal":
		// Online ZarinPal: use per-reseller merchant if configured.
		var gw payment.Gateway
		if resellerCfg != nil && resellerCfg.ZarinpalMerchantID != "" {
			gw = payment.NewZarinPal(resellerCfg.ZarinpalMerchantID)
		} else if h.ZarinPal != nil {
			gw = h.ZarinPal
		} else {
			return echo.NewHTTPError(http.StatusBadRequest, "zarinpal not configured")
		}
		amount := plan.PriceToman
		order := &domain.Order{
			ID:        uuid.New(),
			PlanID:    planID,
			UserID:    userID,
			AdminID:   adminID,
			Username:  req.Username,
			Status:    domain.OrderPending,
			Gateway:   "zarinpal",
			Amount:    amount,
			Currency:  "IRR",
			CreatedAt: time.Now(),
		}
		if err := h.Plans.CreateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "create order failed")
		}
		callbackURL := c.Scheme() + "://" + c.Request().Host + "/api/payment/callback?order_id=" + order.ID.String()
		resp, err := gw.CreatePayment(c.Request().Context(), payment.PaymentRequest{
			Amount:      amount,
			Description: "VortexUI Plan: " + plan.Name,
			CallbackURL: callbackURL,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "payment gateway error: "+err.Error())
		}
		order.GatewayID = resp.Authority
		_ = h.Plans.UpdateOrder(c.Request().Context(), order)
		return c.JSON(http.StatusOK, echo.Map{
			"order_id":     order.ID,
			"redirect_url": resp.RedirectURL,
			"gateway_id":   resp.Authority,
		})

	case "nowpayments":
		// Online NowPayments (crypto gateway).
		if h.NowPayments == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "crypto payments not configured")
		}
		amount := int64(plan.PriceUSD * 100)
		order := &domain.Order{
			ID:        uuid.New(),
			PlanID:    planID,
			UserID:    userID,
			AdminID:   adminID,
			Username:  req.Username,
			Status:    domain.OrderPending,
			Gateway:   "nowpayments",
			Amount:    amount,
			Currency:  "USD",
			CreatedAt: time.Now(),
		}
		if err := h.Plans.CreateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "create order failed")
		}
		callbackURL := c.Scheme() + "://" + c.Request().Host + "/api/payment/callback?order_id=" + order.ID.String()
		resp, err := h.NowPayments.CreatePayment(c.Request().Context(), payment.PaymentRequest{
			Amount:      amount,
			Description: "VortexUI Plan: " + plan.Name,
			CallbackURL: callbackURL,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, "payment gateway error: "+err.Error())
		}
		order.GatewayID = resp.Authority
		_ = h.Plans.UpdateOrder(c.Request().Context(), order)
		return c.JSON(http.StatusOK, echo.Map{
			"order_id":     order.ID,
			"redirect_url": resp.RedirectURL,
			"gateway_id":   resp.Authority,
		})

	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unknown gateway (zarinpal, nowpayments, card_to_card, crypto)")
	}
}

// PaymentCallback handles the gateway redirect after payment.
func (h *Handlers) PaymentCallback(c echo.Context) error {
	orderID, err := uuid.Parse(c.QueryParam("order_id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order_id")
	}
	order, err := h.Plans.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "order not found")
	}
	if order.Status != domain.OrderPending {
		return c.Redirect(http.StatusFound, "/payment/result?status="+string(order.Status))
	}

	// Verify with gateway
	var gw payment.Gateway
	switch order.Gateway {
	case "zarinpal":
		gw = h.ZarinPal
	case "nowpayments", "crypto":
		gw = h.NowPayments
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unknown gateway")
	}

	_, verifyErr := gw.VerifyPayment(c.Request().Context(), order.GatewayID, order.Amount)
	if verifyErr != nil {
		order.Status = domain.OrderFailed
		_ = h.Plans.UpdateOrder(c.Request().Context(), order)
		return c.Redirect(http.StatusFound, "/payment/result?status=failed")
	}

	// Payment successful — fulfill
	now := time.Now()
	order.Status = domain.OrderPaid
	order.PaidAt = &now
	_ = h.Plans.UpdateOrder(c.Request().Context(), order)

	if err := h.Plans.FulfillOrder(c.Request().Context(), order.ID); err != nil {
		return c.Redirect(http.StatusFound, "/payment/result?status=error&msg=fulfillment+failed")
	}

	return c.Redirect(http.StatusFound, "/payment/result?status=success")
}

// PublicPlans lists enabled plans (public, no auth needed for purchase page).
func (h *Handlers) PublicPlans(c echo.Context) error {
	plans, err := h.Plans.ListPlans(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	// Build the set of sudo admin IDs so the public storefront only exposes the
	// main panel's plans, not reseller-owned ones.
	sudoSet := map[uuid.UUID]struct{}{}
	if h.Admins != nil {
		if admins, aerr := h.Admins.List(c.Request().Context()); aerr == nil {
			for _, a := range admins {
				if a.Sudo {
					sudoSet[a.ID] = struct{}{}
				}
			}
		}
	}
	// Filter to enabled plans owned by a sudo admin. If we couldn't resolve any
	// sudo admins, fall back to all enabled plans so nothing breaks.
	var enabled []*domain.Plan
	for _, p := range plans {
		if !p.Enabled {
			continue
		}
		if len(sudoSet) == 0 {
			enabled = append(enabled, p)
			continue
		}
		if p.AdminID != nil {
			if _, ok := sudoSet[*p.AdminID]; ok {
				enabled = append(enabled, p)
			}
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"plans": enabled})
}

// PlanServiceInterface is what the handlers need from the plan service.
type PlanServiceInterface interface {
	CreatePlan(ctx context.Context, in service.CreatePlanInput) (*domain.Plan, error)
	ListPlans(ctx context.Context) ([]*domain.Plan, error)
	GetPlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error)
	DeletePlan(ctx context.Context, id uuid.UUID) error
	CreateOrder(ctx context.Context, o *domain.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateOrder(ctx context.Context, o *domain.Order) error
	ListOrders(ctx context.Context, userID *uuid.UUID) ([]*domain.Order, error)
	FulfillOrder(ctx context.Context, orderID uuid.UUID) error
}

// NowPaymentsIPN handles async IPN webhook from NowPayments.
func (h *Handlers) NowPaymentsIPN(c echo.Context) error {
	if h.Plans == nil && h.WalletBilling == nil {
		return c.NoContent(http.StatusOK)
	}
	// Read and parse IPN body
	var payload struct {
		PaymentID     string `json:"payment_id"`
		PaymentStatus string `json:"payment_status"`
	}
	if err := c.Bind(&payload); err != nil {
		return c.NoContent(http.StatusOK)
	}
	if payload.PaymentStatus == "finished" || payload.PaymentStatus == "confirmed" {
		// Wallet deposit via NowPayments
		if h.WalletBilling != nil {
			if err := h.WalletBilling.CompleteDepositByGatewayID(c.Request().Context(), payload.PaymentID); err == nil {
				return c.NoContent(http.StatusOK)
			}
		}
		// Plan order via NowPayments
		if h.Plans == nil {
			return c.NoContent(http.StatusOK)
		}
		orders, _ := h.Plans.ListOrders(c.Request().Context(), nil)
		for _, o := range orders {
			if o.GatewayID == payload.PaymentID && o.Status == domain.OrderPending {
				now := time.Now()
				o.Status = domain.OrderPaid
				o.PaidAt = &now
				_ = h.Plans.UpdateOrder(c.Request().Context(), o)
				_ = h.Plans.FulfillOrder(c.Request().Context(), o.ID)
				break
			}
		}
	}
	return c.NoContent(http.StatusOK)
}

// --- Order Review (reseller approves/rejects manual payments) ---

type reviewOrderRequest struct {
	Action string `json:"action"` // "approve" | "reject"
	Note   string `json:"note"`
}

// ReviewOrder lets a reseller approve or reject a manual payment order.
// POST /api/orders/:id/review
func (h *Handlers) ReviewOrder(c echo.Context) error {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
	}
	var req reviewOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid body")
	}
	if req.Action != "approve" && req.Action != "reject" {
		return echo.NewHTTPError(http.StatusBadRequest, "action must be approve or reject")
	}

	order, err := h.Plans.GetOrder(c.Request().Context(), orderID)
	if err != nil || order == nil {
		return echo.NewHTTPError(http.StatusNotFound, "order not found")
	}
	if order.Status != domain.OrderPending {
		return echo.NewHTTPError(http.StatusBadRequest, "order is not pending")
	}

	// Gate: caller must be the order's admin (reseller) or sudo.
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}
	if !claims.Sudo {
		if order.AdminID == nil || *order.AdminID != claims.AdminID {
			return echo.NewHTTPError(http.StatusForbidden, "you do not own this order")
		}
	}

	switch req.Action {
	case "approve":
		now := time.Now()
		order.Status = domain.OrderPaid
		order.PaidAt = &now
		if err := h.Plans.UpdateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "update order failed")
		}
		if err := h.Plans.FulfillOrder(c.Request().Context(), order.ID); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "fulfillment failed: "+err.Error())
		}
		return c.JSON(http.StatusOK, echo.Map{"status": "approved", "order_id": order.ID})

	case "reject":
		order.Status = domain.OrderCancelled
		if err := h.Plans.UpdateOrder(c.Request().Context(), order); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "update order failed")
		}
		return c.JSON(http.StatusOK, echo.Map{"status": "rejected", "order_id": order.ID})
	}

	return echo.NewHTTPError(http.StatusBadRequest, "invalid action")
}

// ListPendingOrders returns orders pending review, scoped to the calling admin.
// GET /api/orders/pending
func (h *Handlers) ListPendingOrders(c echo.Context) error {
	orders, err := h.Plans.ListOrders(c.Request().Context(), nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "list failed")
	}
	claims := claimsFrom(c)
	if claims == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "unauthorized")
	}

	var pending []*domain.Order
	for _, o := range orders {
		if o.Status != domain.OrderPending {
			continue
		}
		if claims.Sudo {
			pending = append(pending, o)
		} else if o.AdminID != nil && *o.AdminID == claims.AdminID {
			pending = append(pending, o)
		}
	}
	return c.JSON(http.StatusOK, echo.Map{"orders": pending})
}
