package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
	"github.com/vortexui/vortexui/internal/subscription"
)

// maxWebhookRetries is the maximum number of delivery attempts.
const maxWebhookRetries = 5

// NotificationService dispatches events to matching notification channels,
// delivers webhooks with HMAC signatures, and handles retry logic.
type NotificationService struct {
	channels    port.NotificationChannelRepository
	deliveries  port.WebhookDeliveryRepository
	varResolver *subscription.VarResolver
}

// NewNotificationService constructs a NotificationService with required dependencies.
func NewNotificationService(
	channels port.NotificationChannelRepository,
	deliveries port.WebhookDeliveryRepository,
	varResolver *subscription.VarResolver,
) *NotificationService {
	return &NotificationService{
		channels:    channels,
		deliveries:  deliveries,
		varResolver: varResolver,
	}
}

// Dispatch routes an event to all matching channels based on scope and event
// type filtering. For webhook channels it enqueues a delivery record.
func (s *NotificationService) Dispatch(ctx context.Context, eventType string, data any) error {
	channels, err := s.channels.List(ctx)
	if err != nil {
		return fmt.Errorf("notification dispatch: list channels: %w", err)
	}

	for _, ch := range channels {
		if !ch.Enabled {
			continue
		}
		if !eventMatches(ch.Events, eventType) {
			continue
		}

		switch ch.Type {
		case "webhook":
			if err := s.enqueueWebhook(ctx, ch, eventType, data); err != nil {
				return fmt.Errorf("notification dispatch: enqueue webhook for channel %s: %w", ch.ID, err)
			}
		case "telegram":
			// Telegram delivery is handled via the bot adapter; here we
			// could enqueue a message job. For now we skip inline delivery
			// as the bot adapter handles Telegram dispatch separately.
		}
	}
	return nil
}

// DeliverWebhook performs an HTTP POST to the channel's configured URL with
// the payload as JSON body and an HMAC-SHA256 signature in the
// X-VortexUI-Signature header.
func (s *NotificationService) DeliverWebhook(ctx context.Context, delivery *domain.WebhookDelivery) error {
	// Look up the channel to get config (url, hmac_secret).
	ch, err := s.channels.GetByID(ctx, delivery.ChannelID)
	if err != nil {
		return fmt.Errorf("deliver webhook: get channel: %w", err)
	}

	url, _ := ch.Config["url"].(string)
	if url == "" {
		return fmt.Errorf("deliver webhook: channel %s has no url configured", ch.ID)
	}
	secret, _ := ch.Config["hmac_secret"].(string)

	payloadBytes, err := json.Marshal(delivery.Payload)
	if err != nil {
		return fmt.Errorf("deliver webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("deliver webhook: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		sig := SignPayload(secret, payloadBytes)
		req.Header.Set("X-VortexUI-Signature", sig)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// Network failure — increment attempt for retry.
		nextRetry := s.computeNextRetry(delivery.Attempts)
		_ = s.deliveries.IncrementAttempt(ctx, delivery.ID, nextRetry, 0)
		return fmt.Errorf("deliver webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return s.deliveries.MarkDelivered(ctx, delivery.ID)
	}

	// Non-2xx response — schedule retry if attempts remain.
	nextRetry := s.computeNextRetry(delivery.Attempts)
	return s.deliveries.IncrementAttempt(ctx, delivery.ID, nextRetry, resp.StatusCode)
}

// SignPayload computes HMAC-SHA256 of the payload using the given secret and
// returns the hex-encoded signature.
func SignPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// RetryFailed picks up pending (undelivered) webhook deliveries whose
// next_retry time has passed and attempts redelivery with exponential backoff.
// Maximum 5 attempts per delivery.
func (s *NotificationService) RetryFailed(ctx context.Context) error {
	pending, err := s.deliveries.ListPending(ctx)
	if err != nil {
		return fmt.Errorf("retry failed: list pending: %w", err)
	}

	for _, d := range pending {
		if d.Attempts >= maxWebhookRetries {
			// Exhausted retries; mark as delivered (failed) to stop retrying.
			_ = s.deliveries.MarkDelivered(ctx, d.ID)
			continue
		}
		// Attempt delivery.
		_ = s.DeliverWebhook(ctx, d)
	}
	return nil
}

// RenderTemplate resolves format variables in a notification template string
// using the event data context. Unresolved variables are replaced with empty strings.
func (s *NotificationService) RenderTemplate(template string, eventData any) string {
	if template == "" {
		return ""
	}

	// Convert event data to a map for variable substitution.
	dataMap := toStringMap(eventData)

	result := template
	for key, val := range dataMap {
		placeholder := "{" + strings.ToUpper(key) + "}"
		result = strings.ReplaceAll(result, placeholder, val)
	}

	// Remove any remaining unresolved {VARIABLE} tokens.
	result = removeUnresolved(result)
	return result
}

// --- internal helpers ---

// enqueueWebhook creates a webhook delivery record for later processing.
func (s *NotificationService) enqueueWebhook(ctx context.Context, ch *domain.NotificationChannel, eventType string, data any) error {
	delivery := &domain.WebhookDelivery{
		ID:        uuid.New(),
		ChannelID: ch.ID,
		EventType: eventType,
		Payload:   data,
		Attempts:  0,
		Delivered: false,
		CreatedAt: time.Now(),
	}
	return s.deliveries.Create(ctx, delivery)
}

// computeNextRetry calculates exponential backoff: 1s, 2s, 4s, 8s, 16s.
func (s *NotificationService) computeNextRetry(currentAttempts int) *time.Time {
	if currentAttempts >= maxWebhookRetries-1 {
		return nil // no more retries
	}
	delay := time.Duration(math.Pow(2, float64(currentAttempts))) * time.Second
	next := time.Now().Add(delay)
	return &next
}

// eventMatches checks if the given event type is in the channel's enabled events list.
func eventMatches(enabledEvents []string, eventType string) bool {
	for _, e := range enabledEvents {
		if e == eventType {
			return true
		}
	}
	return false
}

// toStringMap converts event data into a flat map[string]string for template substitution.
func toStringMap(data any) map[string]string {
	result := make(map[string]string)
	if data == nil {
		return result
	}

	// Attempt JSON round-trip to get a flat map.
	b, err := json.Marshal(data)
	if err != nil {
		return result
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return result
	}

	for k, v := range m {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

// removeUnresolved strips remaining {VARIABLE} placeholders from the string.
func removeUnresolved(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '{' {
			// Find closing brace.
			j := i + 1
			for j < len(s) && s[j] != '}' && s[j] != '{' {
				j++
			}
			if j < len(s) && s[j] == '}' {
				// Check if content looks like a variable (all uppercase/underscores).
				content := s[i+1 : j]
				if isVariableToken(content) {
					i = j + 1
					continue
				}
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// isVariableToken returns true if the string matches [A-Z0-9_]+.
func isVariableToken(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// --- Channel CRUD Methods ---

// CreateChannel persists a new notification channel.
func (s *NotificationService) CreateChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	return s.channels.Create(ctx, ch)
}

// ListChannels returns all configured notification channels.
func (s *NotificationService) ListChannels(ctx context.Context) ([]*domain.NotificationChannel, error) {
	return s.channels.List(ctx)
}

// UpdateChannel persists changes to an existing notification channel.
func (s *NotificationService) UpdateChannel(ctx context.Context, ch *domain.NotificationChannel) error {
	return s.channels.Update(ctx, ch)
}

// DeleteChannel removes a notification channel by ID.
func (s *NotificationService) DeleteChannel(ctx context.Context, id uuid.UUID) error {
	return s.channels.Delete(ctx, id)
}

// SendTestNotification dispatches a test event to the specified channel to
// verify connectivity and configuration.
func (s *NotificationService) SendTestNotification(ctx context.Context, channelID uuid.UUID, message string) error {
	ch, err := s.channels.GetByID(ctx, channelID)
	if err != nil {
		return fmt.Errorf("get channel: %w", err)
	}

	testPayload := map[string]any{
		"event":   "test",
		"message": message,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}

	switch ch.Type {
	case "webhook":
		delivery := &domain.WebhookDelivery{
			ID:        uuid.New(),
			ChannelID: ch.ID,
			EventType: "test",
			Payload:   testPayload,
			Attempts:  0,
			Delivered: false,
			CreatedAt: time.Now(),
		}
		if err := s.deliveries.Create(ctx, delivery); err != nil {
			return fmt.Errorf("enqueue test delivery: %w", err)
		}
		// Attempt immediate delivery.
		return s.DeliverWebhook(ctx, delivery)
	case "telegram":
		// For Telegram test notifications, the delivery is handled
		// by the bot adapter. We just verify the channel config is valid.
		chatID, _ := ch.Config["chat_id"].(string)
		if chatID == "" {
			return fmt.Errorf("telegram channel has no chat_id configured")
		}
		return nil
	default:
		return fmt.Errorf("unsupported channel type: %s", ch.Type)
	}
}
