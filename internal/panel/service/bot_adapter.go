package service

import (
	"context"
	"fmt"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/notify"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BotAdapter adapts panel services to the notify.BotPanel interface so the
// Telegram bot can query panel state without importing the full service layer.
type BotAdapter struct {
	users port.UserRepository
	nodes port.NodeRepository
	start time.Time
}

// NewBotAdapter builds the adapter.
func NewBotAdapter(users port.UserRepository, nodes port.NodeRepository) *BotAdapter {
	return &BotAdapter{users: users, nodes: nodes, start: time.Now()}
}

var _ notify.BotPanel = (*BotAdapter)(nil)

func (a *BotAdapter) StatusSummary(ctx context.Context) (notify.BotStatus, error) {
	users, total, err := a.users.List(ctx, port.UserFilter{Limit: 1, Offset: 0})
	_ = users
	if err != nil {
		return notify.BotStatus{}, err
	}
	nodes, err := a.nodes.List(ctx)
	if err != nil {
		return notify.BotStatus{}, err
	}
	online := 0
	for _, n := range nodes {
		if n.Status == domain.NodeConnected {
			online++
		}
	}
	uptime := time.Since(a.start)
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24

	return notify.BotStatus{
		TotalUsers:     total,
		ActiveUsers:    0, // would need a filtered count
		TotalNodes:     len(nodes),
		OnlineNodes:    online,
		TotalTrafficGB: 0,
		Uptime:         fmt.Sprintf("%dd %dh", days, hours),
	}, nil
}

func (a *BotAdapter) TopUsers(ctx context.Context, limit int) ([]notify.BotUser, error) {
	users, _, err := a.users.List(ctx, port.UserFilter{Limit: limit, Offset: 0})
	if err != nil {
		return nil, err
	}
	out := make([]notify.BotUser, 0, len(users))
	for _, u := range users {
		daysLeft := -1
		if u.ExpireAt != nil {
			daysLeft = int(time.Until(*u.ExpireAt).Hours() / 24)
		}
		out = append(out, notify.BotUser{
			Username: u.Username,
			Status:   string(u.Status),
			UsedGB:   float64(u.UsedTraffic) / (1024 * 1024 * 1024),
			LimitGB:  float64(u.DataLimit) / (1024 * 1024 * 1024),
			DaysLeft: daysLeft,
		})
	}
	return out, nil
}

func (a *BotAdapter) FindUser(ctx context.Context, username string) (*notify.BotUser, error) {
	users, _, err := a.users.List(ctx, port.UserFilter{Search: username, Limit: 1})
	if err != nil || len(users) == 0 {
		return nil, err
	}
	u := users[0]
	daysLeft := -1
	if u.ExpireAt != nil {
		daysLeft = int(time.Until(*u.ExpireAt).Hours() / 24)
	}
	return &notify.BotUser{
		Username: u.Username,
		Status:   string(u.Status),
		UsedGB:   float64(u.UsedTraffic) / (1024 * 1024 * 1024),
		LimitGB:  float64(u.DataLimit) / (1024 * 1024 * 1024),
		DaysLeft: daysLeft,
	}, nil
}

func (a *BotAdapter) SetUserStatus(ctx context.Context, username, status string) error {
	users, _, err := a.users.List(ctx, port.UserFilter{Search: username, Limit: 1})
	if err != nil || len(users) == 0 {
		return fmt.Errorf("user %q not found", username)
	}
	u := users[0]
	u.Status = domain.UserStatus(status)
	return a.users.Update(ctx, u)
}

func (a *BotAdapter) OnlineCount(ctx context.Context) (int, error) {
	nodes, err := a.nodes.List(ctx)
	if err != nil {
		return 0, err
	}
	total := 0
	for _, n := range nodes {
		total += n.Health.Connections
	}
	return total, nil
}

func (a *BotAdapter) NodesSummary(ctx context.Context) ([]notify.BotNode, error) {
	nodes, err := a.nodes.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]notify.BotNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, notify.BotNode{
			Name:   n.Name,
			Status: string(n.Status),
			Core:   string(n.Core),
			CPU:    n.Health.CPUPercent,
			RAM:    n.Health.MemPercent,
			Conns:  n.Health.Connections,
		})
	}
	return out, nil
}

func (a *BotAdapter) ListPlans(ctx context.Context) ([]notify.BotPlan, error) {
	// BotAdapter doesn't have plan repo — return empty for now.
	// Wire PlanService into BotAdapter when plan repo is available.
	return []notify.BotPlan{}, nil
}

func (a *BotAdapter) PurchasePlan(ctx context.Context, planName, username string) (string, error) {
	// Placeholder — will be wired to PlanService.FulfillOrder when plan repo is available.
	return fmt.Sprintf("Plan '%s' applied to user '%s' (manual)", planName, username), nil
}
