package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

// UserNotifier sends personal notifications to users via their registered
// Telegram chat ID. It listens to the event bus and routes user-specific
// events (expiry warning, limit reached) to the user's Telegram.
type UserNotifier struct {
	token  string // bot token (same bot used for user-facing commands)
	client *http.Client
	log    *slog.Logger
	// userTelegramResolver looks up a user's Telegram chat ID by their user ID.
	resolver UserTelegramResolver
}

// UserTelegramResolver returns the Telegram chat ID for a given user ID.
type UserTelegramResolver interface {
	GetTelegramChatID(ctx context.Context, userID string) (string, error)
}

// NewUserNotifier builds the notifier.
func NewUserNotifier(token string, resolver UserTelegramResolver, log *slog.Logger) *UserNotifier {
	if log == nil {
		log = slog.Default()
	}
	return &UserNotifier{
		token:    token,
		client:   &http.Client{Timeout: 10 * time.Second},
		log:      log,
		resolver: resolver,
	}
}

// Run consumes events and sends personal notifications to affected users.
func (n *UserNotifier) Run(ctx context.Context, ch <-chan events.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			n.handle(ctx, e)
		}
	}
}

func (n *UserNotifier) handle(ctx context.Context, e events.Event) {
	// Only send user-specific events
	if e.UserID == "" {
		return
	}
	switch e.Type {
	case events.UserLimited, events.UserExpired, events.UserExpiryWarning, events.UserReset:
	default:
		return
	}

	chatID, err := n.resolver.GetTelegramChatID(ctx, e.UserID)
	if err != nil || chatID == "" {
		return // user hasn't registered telegram
	}

	msg := n.formatUserMessage(e)
	if msg == "" {
		return
	}
	n.send(ctx, chatID, msg)
}

func (n *UserNotifier) formatUserMessage(e events.Event) string {
	switch e.Type {
	case events.UserExpiryWarning:
		days := e.Data["days_left"]
		return fmt.Sprintf("⚠️ Your subscription expires in %v days. Renew to stay connected!", days)
	case events.UserLimited:
		return "🚫 You've reached your data limit. Renew or upgrade your plan to continue."
	case events.UserExpired:
		return "⏰ Your subscription has expired. Renew to restore access."
	case events.UserReset:
		return "🔄 Your traffic counter has been reset. Enjoy!"
	default:
		return ""
	}
}

func (n *UserNotifier) send(ctx context.Context, chatID, text string) {
	payload, _ := json.Marshal(map[string]any{"chat_id": chatID, "text": text, "parse_mode": "Markdown"})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		n.log.Warn("user notification send failed", "chatID", chatID, "err", err)
		return
	}
	defer resp.Body.Close()
}
