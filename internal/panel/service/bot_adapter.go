package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/notify"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// BotAdapter adapts panel services to the notify.BotPanel interface so the
// Telegram bot can query panel state without importing the full service layer.
// It also provides extended admin commands (/create, /status, /stats, /user)
// and user-facing notification support (personal expiry/quota alerts).
type BotAdapter struct {
	users     port.UserRepository
	nodes     port.NodeRepository
	templates port.UserTemplateRepository
	start     time.Time
	token     string // Telegram bot token for sending user notifications
	log       *slog.Logger
}

// NewBotAdapter builds the adapter.
func NewBotAdapter(users port.UserRepository, nodes port.NodeRepository) *BotAdapter {
	return &BotAdapter{users: users, nodes: nodes, start: time.Now(), log: slog.Default()}
}

// NewBotAdapterExtended builds the adapter with extended capabilities including
// template-based user creation and personal user notifications.
func NewBotAdapterExtended(
	users port.UserRepository,
	nodes port.NodeRepository,
	templates port.UserTemplateRepository,
	botToken string,
	log *slog.Logger,
) *BotAdapter {
	if log == nil {
		log = slog.Default()
	}
	return &BotAdapter{
		users:     users,
		nodes:     nodes,
		templates: templates,
		start:     time.Now(),
		token:     botToken,
		log:       log,
	}
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

// --- Extended Admin Commands ---

// CreateFromTemplate creates users from a template by name.
// Implements the /create <template_name> [count] command.
func (a *BotAdapter) CreateFromTemplate(ctx context.Context, templateName string, count int) ([]string, error) {
	if a.templates == nil {
		return nil, fmt.Errorf("template repository not configured")
	}
	if count < 1 {
		count = 1
	}
	if count > 100 {
		count = 100
	}

	tmpl, err := a.templates.GetByName(ctx, templateName)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", templateName, err)
	}

	createdNames := make([]string, 0, count)
	for i := 0; i < count; i++ {
		username := fmt.Sprintf("%s_%s", tmpl.Name, generateShortID())
		user := &domain.User{
			ID:        uuid.New(),
			Username:  username,
			Status:    domain.UserStatusOnHold,
			DataLimit: tmpl.DataLimit,
		}
		if tmpl.ExpireDuration != nil {
			expireAt := time.Now().Add(time.Duration(*tmpl.ExpireDuration) * time.Second)
			user.ExpireAt = &expireAt
		}
		if err := a.users.Create(ctx, user); err != nil {
			return createdNames, fmt.Errorf("failed to create user %d: %w", i+1, err)
		}
		createdNames = append(createdNames, username)
	}
	return createdNames, nil
}

// NodeStatus returns the health status of a named node.
// Implements the /status <node_name> command.
func (a *BotAdapter) NodeStatus(ctx context.Context, nodeName string) (*notify.BotNode, error) {
	nodes, err := a.nodes.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, n := range nodes {
		if n.Name == nodeName {
			return &notify.BotNode{
				Name:   n.Name,
				Status: string(n.Status),
				Core:   string(n.Core),
				CPU:    n.Health.CPUPercent,
				RAM:    n.Health.MemPercent,
				Conns:  n.Health.Connections,
			}, nil
		}
	}
	return nil, fmt.Errorf("node %q not found", nodeName)
}

// QuickStats returns a quick panel summary: total users, online, nodes.
// Implements the /stats command.
func (a *BotAdapter) QuickStats(ctx context.Context) (notify.BotStatus, error) {
	return a.StatusSummary(ctx)
}

// UserInfo returns details for a specific username.
// Implements the /user <username> command.
func (a *BotAdapter) UserInfo(ctx context.Context, username string) (*notify.BotUser, error) {
	return a.FindUser(ctx, username)
}

// --- User-Facing Personal Notification Support ---

// SendPersonalNotification sends a Telegram message to a user's registered chat ID.
// Used for expiry warnings, quota alerts, and other user-specific notifications.
func (a *BotAdapter) SendPersonalNotification(ctx context.Context, chatID string, message string) error {
	if a.token == "" {
		return fmt.Errorf("bot token not configured for user notifications")
	}
	if chatID == "" {
		return fmt.Errorf("user has no registered Telegram chat ID")
	}

	payload, _ := json.Marshal(map[string]any{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", a.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		a.log.Warn("personal notification send failed", "chatID", chatID, "err", err)
		return fmt.Errorf("send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("telegram API returned %d", resp.StatusCode)
	}
	return nil
}

// NotifyExpiryWarning sends an expiry warning to a user's Telegram.
func (a *BotAdapter) NotifyExpiryWarning(ctx context.Context, chatID string, username string, daysLeft int) error {
	msg := fmt.Sprintf("⚠️ *Expiry Warning*\n\nHi %s, your subscription expires in *%d days*. Please renew to stay connected!", username, daysLeft)
	return a.SendPersonalNotification(ctx, chatID, msg)
}

// NotifyQuotaAlert sends a quota/data limit warning to a user's Telegram.
func (a *BotAdapter) NotifyQuotaAlert(ctx context.Context, chatID string, username string, usedGB, limitGB float64) error {
	pct := 0.0
	if limitGB > 0 {
		pct = (usedGB / limitGB) * 100
	}
	msg := fmt.Sprintf("📊 *Quota Alert*\n\nHi %s, you've used *%.1f%%* of your data (%.1f / %.1f GB). Consider upgrading your plan!", username, pct, usedGB, limitGB)
	return a.SendPersonalNotification(ctx, chatID, msg)
}

// generateShortID creates a short random hex string for username uniqueness.
func generateShortID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}
