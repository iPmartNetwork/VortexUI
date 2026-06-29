package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// PlanRepository persists plans and orders.
type PlanRepository interface {
	CreatePlan(ctx context.Context, p *domain.Plan) error
	GetPlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error)
	ListPlans(ctx context.Context) ([]*domain.Plan, error)
	UpdatePlan(ctx context.Context, p *domain.Plan) error
	DeletePlan(ctx context.Context, id uuid.UUID) error

	CreateOrder(ctx context.Context, o *domain.Order) error
	GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error)
	UpdateOrder(ctx context.Context, o *domain.Order) error
	ListOrders(ctx context.Context, userID *uuid.UUID, limit int) ([]*domain.Order, error)
}

// PlanService manages subscription plans and purchase orders.
type PlanService struct {
	repo  PlanRepository
	users *UserService
	now   func() time.Time
}

// NewPlanService wires the service.
func NewPlanService(repo PlanRepository, users *UserService) *PlanService {
	return &PlanService{repo: repo, users: users, now: time.Now}
}

// CreatePlanInput describes a new plan.
type CreatePlanInput struct {
	Name          string
	Description   string
	DataLimit     int64
	DurationDays  int
	DeviceLimit   int
	ResetStrategy domain.ResetStrategy
	InboundIDs    []uuid.UUID
	PriceToman    int64
	PriceUSD      float64
	MaxUsers      int
}

// CreatePlan persists a new subscription plan.
func (s *PlanService) CreatePlan(ctx context.Context, in CreatePlanInput) (*domain.Plan, error) {
	if in.Name == "" {
		return nil, errors.New("plan name is required")
	}
	p := &domain.Plan{
		ID:            uuid.New(),
		Name:          in.Name,
		Description:   in.Description,
		DataLimit:     in.DataLimit,
		Duration:      in.DurationDays,
		DeviceLimit:   in.DeviceLimit,
		ResetStrategy: in.ResetStrategy,
		InboundIDs:    in.InboundIDs,
		PriceToman:    in.PriceToman,
		PriceUSD:      in.PriceUSD,
		MaxUsers:      in.MaxUsers,
		Enabled:       true,
		CreatedAt:     s.now(),
	}
	if err := s.repo.CreatePlan(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// ListPlans returns all plans.
func (s *PlanService) ListPlans(ctx context.Context) ([]*domain.Plan, error) {
	return s.repo.ListPlans(ctx)
}

// GetPlan returns one plan.
func (s *PlanService) GetPlan(ctx context.Context, id uuid.UUID) (*domain.Plan, error) {
	return s.repo.GetPlan(ctx, id)
}

// DeletePlan removes a plan.
func (s *PlanService) DeletePlan(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePlan(ctx, id)
}

// FulfillOrder applies a paid plan to a user: sets their limits, expiry, and
// binds inbounds. Called after payment verification succeeds.
func (s *PlanService) FulfillOrder(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != domain.OrderPaid {
		return errors.New("order is not paid")
	}
	plan, err := s.repo.GetPlan(ctx, order.PlanID)
	if err != nil {
		return err
	}

	// If order has a user, apply additive renewal: stack data and extend expiry
	// on top of the user's current subscription rather than replacing it. This
	// lets existing users renew or upgrade without losing remaining quota.
	if order.UserID != nil {
		u, err := s.users.users.GetByID(ctx, *order.UserID)
		if err != nil {
			return err
		}

		// Add purchased traffic to remaining limit (stack on existing).
		u.DataLimit += plan.DataLimit

		// Extend expiry: if user is still active (not expired), add duration on
		// top of current expire; if expired, start from now. Never shrink.
		now := s.now()
		base := now
		if u.ExpireAt != nil && u.ExpireAt.After(now) {
			base = *u.ExpireAt
		}
		expire := base.Add(time.Duration(plan.Duration) * 24 * time.Hour)
		u.ExpireAt = &expire

		// Do NOT reset UsedTraffic — user keeps their current usage accounting.

		// Update device limit and reset strategy to the new plan's values.
		if plan.DeviceLimit > 0 {
			u.DeviceLimit = plan.DeviceLimit
		}
		if plan.ResetStrategy != "" {
			u.ResetStrategy = plan.ResetStrategy
		}

		// Reactivate if expired/limited.
		u.Status = domain.UserStatusActive

		return s.users.users.Update(ctx, u)
	}

	// Create new user from order
	expire := s.now().Add(time.Duration(plan.Duration) * 24 * time.Hour)
	_, err = s.users.Create(ctx, CreateUserInput{
		Username:      order.Username,
		DataLimit:     plan.DataLimit,
		ExpireAt:      &expire,
		DeviceLimit:   plan.DeviceLimit,
		ResetStrategy: plan.ResetStrategy,
		InboundIDs:    plan.InboundIDs,
		AdminID:       order.AdminID,
	})
	return err
}

// CreateOrder persists a new order.
func (s *PlanService) CreateOrder(ctx context.Context, o *domain.Order) error {
	return s.repo.CreateOrder(ctx, o)
}

// GetOrder retrieves an order by ID.
func (s *PlanService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.repo.GetOrder(ctx, id)
}

// UpdateOrder persists order changes (status, gateway_id, paid_at).
func (s *PlanService) UpdateOrder(ctx context.Context, o *domain.Order) error {
	return s.repo.UpdateOrder(ctx, o)
}

// ListOrders returns orders, optionally filtered by user.
func (s *PlanService) ListOrders(ctx context.Context, userID *uuid.UUID) ([]*domain.Order, error) {
	return s.repo.ListOrders(ctx, userID, 100)
}
