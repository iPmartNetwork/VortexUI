package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TelegramBot is a long-polling Telegram bot that handles admin commands
// (interactive management) in addition to the one-way event notifier. Commands:
//
//	/start   — greeting + help
//	/status  — panel status (users, nodes, uptime)
//	/users   — list top users by traffic
//	/online  — current online count
//	/node    — node fleet health summary
//	/find <username> — lookup a user
//	/limit <username> — disable a user
//	/unlimit <username> — re-enable a user
type TelegramBot struct {
	token   string
	chatID  int64 // authorized admin chat
	apiBase string
	client  *http.Client
	log     *slog.Logger
	panel   BotPanel // interface to panel services
}

// BotPanel is the narrow interface the bot needs to query the panel.
type BotPanel interface {
	StatusSummary(ctx context.Context) (BotStatus, error)
	TopUsers(ctx context.Context, limit int) ([]BotUser, error)
	FindUser(ctx context.Context, username string) (*BotUser, error)
	SetUserStatus(ctx context.Context, username, status string) error
	OnlineCount(ctx context.Context) (int, error)
	NodesSummary(ctx context.Context) ([]BotNode, error)
	ListPlans(ctx context.Context) ([]BotPlan, error)
	PurchasePlan(ctx context.Context, planName, username string) (string, error)
}

// BotStatus is a snapshot for the /status command.
type BotStatus struct {
	TotalUsers  int
	ActiveUsers int
	TotalNodes  int
	OnlineNodes int
	TotalTrafficGB float64
	Uptime      string
}

// BotUser is a user summary for bot commands.
type BotUser struct {
	Username    string
	Status      string
	UsedGB      float64
	LimitGB     float64
	DaysLeft    int
	DeviceCount int
}

// BotNode is a node summary for the /node command.
type BotNode struct {
	Name    string
	Status  string
	Core    string
	CPU     float64
	RAM     float64
	Conns   int
}

// BotPlan is a plan summary for the /plans command.
type BotPlan struct {
	Name       string
	DataGB     float64
	Days       int
	PriceToman int64
	PriceUSD   float64
}

// NewTelegramBot creates a new interactive bot.
func NewTelegramBot(token string, chatID int64, panel BotPanel, log *slog.Logger) *TelegramBot {
	if log == nil {
		log = slog.Default()
	}
	return &TelegramBot{
		token:   token,
		chatID:  chatID,
		apiBase: defaultTelegramAPI,
		client:  &http.Client{Timeout: 30 * time.Second},
		log:     log,
		panel:   panel,
	}
}

// Run starts the long-polling loop. Blocks until ctx is cancelled.
func (b *TelegramBot) Run(ctx context.Context) {
	b.log.Info("telegram bot started (long-polling)")
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		updates, err := b.getUpdates(ctx, offset)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			b.log.Warn("telegram getUpdates failed", "err", err)
			time.Sleep(3 * time.Second)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			b.handleUpdate(ctx, u)
		}
	}
}

type tgUpdate struct {
	UpdateID int       `json:"update_id"`
	Message  *tgMessage `json:"message"`
}

type tgMessage struct {
	Chat tgChat `json:"chat"`
	Text string `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

func (b *TelegramBot) getUpdates(ctx context.Context, offset int) ([]tgUpdate, error) {
	url := fmt.Sprintf("%s/bot%s/getUpdates?offset=%d&timeout=25", b.apiBase, b.token, offset)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		OK     bool       `json:"ok"`
		Result []tgUpdate `json:"result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func (b *TelegramBot) handleUpdate(ctx context.Context, u tgUpdate) {
	if u.Message == nil || u.Message.Text == "" {
		return
	}
	// Only respond to the authorized admin chat
	if u.Message.Chat.ID != b.chatID {
		return
	}
	text := strings.TrimSpace(u.Message.Text)
	parts := strings.Fields(text)
	cmd := strings.ToLower(parts[0])

	var reply string
	switch cmd {
	case "/start", "/help":
		reply = b.cmdHelp()
	case "/status":
		reply = b.cmdStatus(ctx)
	case "/users":
		reply = b.cmdUsers(ctx)
	case "/online":
		reply = b.cmdOnline(ctx)
	case "/node", "/nodes":
		reply = b.cmdNodes(ctx)
	case "/find":
		if len(parts) < 2 {
			reply = "⚠️ Usage: /find <username>"
		} else {
			reply = b.cmdFind(ctx, parts[1])
		}
	case "/plans":
		reply = b.cmdPlans(ctx)
	case "/buy":
		if len(parts) < 2 {
			reply = "⚠️ Usage: /buy <plan_name> <username>"
		} else {
			username := ""
			if len(parts) >= 3 {
				username = parts[2]
			}
			reply = b.cmdBuy(ctx, parts[1], username)
		}
	case "/limit":
		if len(parts) < 2 {
			reply = "⚠️ Usage: /limit <username>"
		} else {
			reply = b.cmdSetStatus(ctx, parts[1], "disabled")
		}
	case "/unlimit":
		if len(parts) < 2 {
			reply = "⚠️ Usage: /unlimit <username>"
		} else {
			reply = b.cmdSetStatus(ctx, parts[1], "active")
		}
	default:
		reply = "❓ Unknown command. Try /help"
	}

	if reply != "" {
		b.sendMessage(ctx, reply)
	}
}

func (b *TelegramBot) cmdHelp() string {
	return `🔹 *VortexUI Bot*

/status — Panel overview
/users — Top users by traffic
/online — Current online count
/nodes — Node fleet health
/find <user> — Lookup user info
/limit <user> — Disable a user
/unlimit <user> — Re-enable a user
/plans — List available plans
/buy <plan> <user> — Purchase a plan for a user`
}

func (b *TelegramBot) cmdStatus(ctx context.Context) string {
	s, err := b.panel.StatusSummary(ctx)
	if err != nil {
		return "❌ Failed to get status"
	}
	return fmt.Sprintf(`📊 *Panel Status*

👥 Users: %d total, %d active
🖥 Nodes: %d/%d online
📡 Traffic: %.1f GB total
⏱ Uptime: %s`, s.TotalUsers, s.ActiveUsers, s.OnlineNodes, s.TotalNodes, s.TotalTrafficGB, s.Uptime)
}

func (b *TelegramBot) cmdUsers(ctx context.Context) string {
	users, err := b.panel.TopUsers(ctx, 10)
	if err != nil {
		return "❌ Failed to get users"
	}
	if len(users) == 0 {
		return "No users found."
	}
	var sb strings.Builder
	sb.WriteString("👥 *Top Users (by traffic)*\n\n")
	for i, u := range users {
		status := statusEmoji(u.Status)
		fmt.Fprintf(&sb, "%d. %s `%s` — %.1f/%.1f GB %s\n", i+1, status, u.Username, u.UsedGB, u.LimitGB, u.Status)
	}
	return sb.String()
}

func (b *TelegramBot) cmdOnline(ctx context.Context) string {
	n, err := b.panel.OnlineCount(ctx)
	if err != nil {
		return "❌ Failed to get online count"
	}
	return fmt.Sprintf("🟢 Online connections: *%d*", n)
}

func (b *TelegramBot) cmdNodes(ctx context.Context) string {
	nodes, err := b.panel.NodesSummary(ctx)
	if err != nil {
		return "❌ Failed to get nodes"
	}
	if len(nodes) == 0 {
		return "No nodes registered."
	}
	var sb strings.Builder
	sb.WriteString("🖥 *Nodes*\n\n")
	for _, n := range nodes {
		dot := "🔴"
		if n.Status == "connected" {
			dot = "🟢"
		}
		fmt.Fprintf(&sb, "%s *%s* [%s]\n   CPU: %.0f%% | RAM: %.0f%% | Conns: %d\n\n", dot, n.Name, n.Core, n.CPU, n.RAM, n.Conns)
	}
	return sb.String()
}

func (b *TelegramBot) cmdFind(ctx context.Context, username string) string {
	u, err := b.panel.FindUser(ctx, username)
	if err != nil || u == nil {
		return fmt.Sprintf("❌ User `%s` not found", username)
	}
	return fmt.Sprintf(`👤 *%s*

Status: %s %s
Traffic: %.2f / %.2f GB
Days left: %d
Devices: %d`, u.Username, statusEmoji(u.Status), u.Status, u.UsedGB, u.LimitGB, u.DaysLeft, u.DeviceCount)
}

func (b *TelegramBot) cmdSetStatus(ctx context.Context, username, status string) string {
	if err := b.panel.SetUserStatus(ctx, username, status); err != nil {
		return fmt.Sprintf("❌ Failed: %v", err)
	}
	action := "disabled"
	if status == "active" {
		action = "re-enabled"
	}
	return fmt.Sprintf("✅ User `%s` %s", username, action)
}

func (b *TelegramBot) cmdPlans(ctx context.Context) string {
	plans, err := b.panel.ListPlans(ctx)
	if err != nil {
		return "❌ Failed to get plans"
	}
	if len(plans) == 0 {
		return "No plans configured."
	}
	var sb strings.Builder
	sb.WriteString("📦 *Available Plans*\n\n")
	for i, p := range plans {
		price := ""
		if p.PriceToman > 0 {
			price = fmt.Sprintf("%d تومان", p.PriceToman)
		} else if p.PriceUSD > 0 {
			price = fmt.Sprintf("$%.2f", p.PriceUSD)
		} else {
			price = "Free"
		}
		fmt.Fprintf(&sb, "%d. *%s* — %.0f GB / %d days — %s\n", i+1, p.Name, p.DataGB, p.Days, price)
	}
	sb.WriteString("\n💡 Use `/buy <plan_name> <username>` to purchase")
	return sb.String()
}

func (b *TelegramBot) cmdBuy(ctx context.Context, planName, username string) string {
	if username == "" {
		return "⚠️ Usage: /buy <plan_name> <username>"
	}
	msg, err := b.panel.PurchasePlan(ctx, planName, username)
	if err != nil {
		return fmt.Sprintf("❌ Purchase failed: %v", err)
	}
	return fmt.Sprintf("✅ %s", msg)
}

func (b *TelegramBot) sendMessage(ctx context.Context, text string) {
	payload, _ := json.Marshal(map[string]any{
		"chat_id":    b.chatID,
		"text":       text,
		"parse_mode": "Markdown",
	})
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.apiBase, b.token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		b.log.Warn("telegram bot sendMessage failed", "err", err)
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
}

func statusEmoji(s string) string {
	switch s {
	case "active":
		return "🟢"
	case "limited":
		return "🟡"
	case "expired":
		return "🔴"
	case "disabled":
		return "⚫"
	default:
		return "⚪"
	}
}

// ParseChatID converts a string chat ID (from env) to int64.
func ParseChatID(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}
