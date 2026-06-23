package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

// Webhook POSTs each event as JSON to a configured URL. When a secret is set it
// signs the body with HMAC-SHA256 in the X-Vortex-Signature header so the
// receiver can verify authenticity. Delivery retries a few times with backoff.
type Webhook struct {
	url      string
	secret   string
	client   *http.Client
	log      *slog.Logger
	attempts int
	backoff  time.Duration
}

// NewWebhook builds a dispatcher for url. secret may be empty (no signature).
func NewWebhook(url, secret string, log *slog.Logger) *Webhook {
	if log == nil {
		log = slog.Default()
	}
	return &Webhook{
		url:      url,
		secret:   secret,
		client:   &http.Client{Timeout: 10 * time.Second},
		log:      log,
		attempts: 3,
		backoff:  500 * time.Millisecond,
	}
}

// Run consumes events until ctx is cancelled or the channel closes, delivering
// each to the webhook.
func (w *Webhook) Run(ctx context.Context, ch <-chan events.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-ch:
			if !ok {
				return
			}
			if err := w.deliver(ctx, e); err != nil {
				w.log.Warn("webhook delivery failed", "type", string(e.Type), "err", err)
			}
		}
	}
}

// Deliver marshals and POSTs a single event (exported for reseller dispatchers).
func (w *Webhook) Deliver(ctx context.Context, e events.Event) error {
	return w.deliver(ctx, e)
}

// deliver marshals the event and POSTs it, retrying with exponential backoff.
func (w *Webhook) deliver(ctx context.Context, e events.Event) error {
	body, err := json.Marshal(e)
	if err != nil {
		return err
	}
	backoff := w.backoff
	var lastErr error
	for attempt := 1; attempt <= w.attempts; attempt++ {
		if err := w.post(ctx, body); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt == w.attempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
	}
	return lastErr
}

func (w *Webhook) post(ctx context.Context, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "VortexUI")
	if w.secret != "" {
		mac := hmac.New(sha256.New, []byte(w.secret))
		mac.Write(body)
		req.Header.Set("X-Vortex-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body) // drain so the connection can be reused
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook responded %d", resp.StatusCode)
	}
	return nil
}
