package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

// defaultTelegramAPI is the Bot API base; overridable in tests.
const defaultTelegramAPI = "https://api.telegram.org"

// Telegram sends a chat message for each event via the Bot API sendMessage
// method. It is a one-way notifier (no command handling).
type Telegram struct {
	token   string
	chatID  string
	apiBase string
	client  *http.Client
	log     *slog.Logger
}

// NewTelegram builds a notifier for the given bot token and chat id.
func NewTelegram(token, chatID string, log *slog.Logger) *Telegram {
	if log == nil {
		log = slog.Default()
	}
	return &Telegram{
		token:   token,
		chatID:  chatID,
		apiBase: defaultTelegramAPI,
		client:  &http.Client{Timeout: 10 * time.Second},
		log:     log,
	}
}

// Run consumes events until ctx is cancelled or the channel closes, sending each
// as a chat message.
func (t *Telegram) Run(ctx context.Context, ch <-chan events.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			text := describe(e) + "\n" + stamp(e.Time)
			if err := t.send(ctx, text); err != nil {
				t.log.Warn("telegram send failed", "type", string(e.Type), "err", err)
			}
		}
	}
}

func (t *Telegram) send(ctx context.Context, text string) error {
	payload, err := json.Marshal(map[string]any{"chat_id": t.chatID, "text": text})
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/bot%s/sendMessage", t.apiBase, t.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram responded %d", resp.StatusCode)
	}
	return nil
}
