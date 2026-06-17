package notify

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// UserBotPanel is the interface the user-facing bot needs.
type UserBotPanel interface {
	// AuthUser verifies a subscription token and returns user info.
	AuthUser(ctx context.Context, token string) (*BotUser, error)
	// GetUserConfigs returns subscription links for a user.
	GetUserConfigs(ctx context.Context, token string) (subURL string, links []string, err error)
	// GetUserPlans returns available plans for purchase.
	GetUserPlans(ctx context.Context) ([]BotPlan, error)
}

// TelegramUserBot is a user-facing Telegram bot where subscribers can:
// - /start <sub_token> — authenticate with their subscription token
// - /me — view usage, expiry, device count
// - /configs — get subscription links + individual configs
// - /plans — view available plans
// - /renew — get renewal/purchase link
type TelegramUserBot struct {
	token   string
	apiBase string
	client  *http.Client
	log     *slog.Logger
	panel   UserBotPanel
	// sessions maps telegram chat_id → subscription token (authenticated users)
	sessions map[int64]string
}

// NewTelegramUserBot creates the user-facing bot.
func NewTelegramUserBot(token string, panel UserBotPanel, log *slog.Logger) *TelegramUserBot {
	if log == nil {
		log = slog.Default()
	}
	return &TelegramUserBot{
		token:    token,
		apiBase:  defaultTelegramAPI,
		client:   &http.Client{Timeout: 30 * time.Second},
		log:      log,
		panel:    panel,
		sessions: make(map[int64]string),
	}
}

// Run starts the long-polling loop for user bot.
func (b *TelegramUserBot) Run(ctx context.Context) {
	b.log.Info("telegram user bot started")
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
			time.Sleep(3 * time.Second)
			continue
		}
		for _, u := range updates {
			offset = u.UpdateID + 1
			b.handle(ctx, u)
		}
	}
}

func (b *TelegramUserBot) handle(ctx context.Context, u tgUpdate) {
	if u.Message == nil || u.Message.Text == "" {
		return
	}
	chatID := u.Message.Chat.ID
	text := strings.TrimSpace(u.Message.Text)
	parts := strings.Fields(text)
	cmd := strings.ToLower(parts[0])

	var reply string
	switch cmd {
	case "/start":
		if len(parts) >= 2 {
			reply = b.cmdAuth(ctx, chatID, parts[1])
		} else {
			reply = "👋 Welcome to VortexUI!\n\nSend /start <your_subscription_token> to authenticate.\n\nYou can find your token in your subscription link."
		}
	case "/me", "/status":
		reply = b.cmdMe(ctx, chatID)
	case "/configs", "/config", "/sub":
		reply = b.cmdConfigs(ctx, chatID)
	case "/plans":
		reply = b.cmdPlans(ctx)
	case "/renew":
		reply = "💳 To renew, visit your panel or contact your provider."
	case "/help":
		reply = b.cmdHelp()
	default:
		reply = "❓ Unknown command. Try /help"
	}

	if reply != "" {
		b.send(ctx, chatID, reply)
	}
}

func (b *TelegramUserBot) cmdAuth(ctx context.Context, chatID int64, token string) string {
	user, err := b.panel.AuthUser(ctx, token)
	if err != nil || user == nil {
		return "❌ Invalid token. Check your subscription link and try again."
	}
	b.sessions[chatID] = token
	return fmt.Sprintf("✅ Authenticated as *%s*\n\nUse /me for status, /configs for your links.", user.Username)
}

func (b *TelegramUserBot) cmdMe(ctx context.Context, chatID int64) string {
	token := b.sessions[chatID]
	if token == "" {
		return "🔒 Not authenticated. Send /start <token> first."
	}
	user, err := b.panel.AuthUser(ctx, token)
	if err != nil || user == nil {
		return "❌ Session expired. Send /start <token> again."
	}
	return fmt.Sprintf(`👤 *%s*

📊 Traffic: %.2f / %.2f GB
📅 Days left: %d
📱 Devices: %d
🔵 Status: %s`, user.Username, user.UsedGB, user.LimitGB, user.DaysLeft, user.DeviceCount, user.Status)
}

func (b *TelegramUserBot) cmdConfigs(ctx context.Context, chatID int64) string {
	token := b.sessions[chatID]
	if token == "" {
		return "🔒 Not authenticated. Send /start <token> first."
	}
	subURL, links, err := b.panel.GetUserConfigs(ctx, token)
	if err != nil {
		return "❌ Failed to get configs."
	}
	var sb strings.Builder
	sb.WriteString("🔗 *Subscription Links*\n\n")
	fmt.Fprintf(&sb, "📋 Auto: `%s`\n", subURL)
	fmt.Fprintf(&sb, "📋 Clash: `%s?format=clash`\n", subURL)
	fmt.Fprintf(&sb, "📋 Sing-box: `%s?format=singbox`\n\n", subURL)
	if len(links) > 0 {
		sb.WriteString("⚡ *Configs*\n")
		for i, l := range links {
			if i >= 5 {
				fmt.Fprintf(&sb, "\n... and %d more", len(links)-5)
				break
			}
			fmt.Fprintf(&sb, "\n`%s`\n", l)
		}
	}
	return sb.String()
}

func (b *TelegramUserBot) cmdPlans(ctx context.Context) string {
	plans, err := b.panel.GetUserPlans(ctx)
	if err != nil || len(plans) == 0 {
		return "No plans available right now."
	}
	var sb strings.Builder
	sb.WriteString("📦 *Available Plans*\n\n")
	for i, p := range plans {
		price := ""
		if p.PriceToman > 0 {
			price = fmt.Sprintf("%d تومان", p.PriceToman)
		} else if p.PriceUSD > 0 {
			price = fmt.Sprintf("$%.2f", p.PriceUSD)
		}
		fmt.Fprintf(&sb, "%d. *%s* — %.0f GB / %d days — %s\n", i+1, p.Name, p.DataGB, p.Days, price)
	}
	return sb.String()
}

func (b *TelegramUserBot) cmdHelp() string {
	return `🔹 *VortexUI User Bot*

/start <token> — Authenticate
/me — Your usage & status
/configs — Subscription links
/plans — Available plans
/renew — Renewal info
/help — This message`
}

func (b *TelegramUserBot) send(ctx context.Context, chatID int64, text string) {
	payload, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text, "parse_mode": "Markdown"})
	url := fmt.Sprintf("%s/bot%s/sendMessage", b.apiBase, b.token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
}

func (b *TelegramUserBot) getUpdates(ctx context.Context, offset int) ([]tgUpdate, error) {
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
	_ = json.Unmarshal(body, &result)
	return result.Result, nil
}
