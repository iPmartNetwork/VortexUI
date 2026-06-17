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
	return c.JSON(http.StatusOK, echo.Map{"plans": plans})
}

func (h *Handlers) DeletePlan(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid id")
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
	return c.JSON(http.StatusOK, echo.Map{"orders": orders})
}

// --- Payment (public purchase flow) ---

type purchaseRequest struct {
	PlanID   string `json:"plan_id"`
	Username string `json:"username"` // for new users
	Gateway  string `json:"gateway"`  // "zarinpal" | "nowpayments"
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

	// Select gateway
	var gw payment.Gateway
	var amount int64
	var currency string
	switch req.Gateway {
	case "zarinpal":
		if h.ZarinPal == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "zarinpal not configured")
		}
		gw = h.ZarinPal
		amount = plan.PriceToman
		currency = "IRR"
	case "nowpayments", "crypto":
		if h.NowPayments == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "crypto payments not configured")
		}
		gw = h.NowPayments
		amount = int64(plan.PriceUSD * 100) // cents
		currency = "USD"
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "unknown gateway (zarinpal or nowpayments)")
	}

	// Create order
	order := &domain.Order{
		ID:        uuid.New(),
		PlanID:    planID,
		Username:  req.Username,
		Status:    domain.OrderPending,
		Gateway:   req.Gateway,
		Amount:    amount,
		Currency:  currency,
		CreatedAt: time.Now(),
	}
	if err := h.Plans.CreateOrder(c.Request().Context(), order); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "create order failed")
	}

	// Initiate payment
	callbackURL := c.Scheme() + "://" + c.Request().Host + "/api/payment/callback?order_id=" + order.ID.String()
	resp, err := gw.CreatePayment(c.Request().Context(), payment.PaymentRequest{
		Amount:      amount,
		Description: "VortexUI Plan: " + plan.Name,
		CallbackURL: callbackURL,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusBadGateway, "payment gateway error: "+err.Error())
	}

	// Save gateway authority
	order.GatewayID = resp.Authority
	_ = h.Plans.UpdateOrder(c.Request().Context(), order)

	return c.JSON(http.StatusOK, echo.Map{
		"order_id":     order.ID,
		"redirect_url": resp.RedirectURL,
		"gateway_id":   resp.Authority,
	})
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
	// Filter to enabled only
	var enabled []*domain.Plan
	for _, p := range plans {
		if p.Enabled {
			enabled = append(enabled, p)
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
	if h.Plans == nil {
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
		// Find order by gateway_id and fulfill
		// This is best-effort — if the redirect already fulfilled it, this is a no-op
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
