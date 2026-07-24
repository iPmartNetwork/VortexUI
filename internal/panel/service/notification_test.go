package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// Property 15: Notification routing by scope — notifications are delivered
// only to channels whose scope matches the event source.
func TestProperty_NotificationRoutingByScope(t *testing.T) {
	channels := []testChannel{
		{ID: "ch1", Scope: "global", Events: []string{"user_created", "user_expired"}},
		{ID: "ch2", Scope: "admin:admin1", Events: []string{"user_created"}},
		{ID: "ch3", Scope: "global", Events: []string{"node_offline"}},
	}

	matched := routeNotification(channels, "user_created", "global")
	if len(matched) != 1 || matched[0].ID != "ch1" {
		t.Fatalf("expected only ch1 for global user_created, got %v", matched)
	}

	matched = routeNotification(channels, "user_created", "admin:admin1")
	if len(matched) != 1 || matched[0].ID != "ch2" {
		t.Fatalf("expected ch2 for admin-scoped user_created, got %v", matched)
	}
}

// Property 16: Event type filtering — channels only fire for subscribed events.
func TestProperty_EventTypeFiltering(t *testing.T) {
	ch := testChannel{ID: "ch1", Scope: "global", Events: []string{"user_expired", "node_offline"}}

	if !channelMatchesEvent(ch, "user_expired") {
		t.Fatal("channel should match subscribed event")
	}
	if channelMatchesEvent(ch, "user_created") {
		t.Fatal("channel should not match unsubscribed event")
	}
}

// Property 17: Webhook HMAC signature correctness — the signature header
// validates against the payload with the shared secret.
func TestProperty_WebhookHMACSignatureCorrectness(t *testing.T) {
	secret := "webhook-secret-key"
	payload := `{"event":"user_created","user_id":"abc123"}`

	signature := computeHMAC(secret, payload)

	// Verify signature is valid hex.
	if len(signature) != 64 {
		t.Fatalf("HMAC-SHA256 hex should be 64 chars, got %d", len(signature))
	}

	// Verify it validates.
	if !validateHMAC(secret, payload, signature) {
		t.Fatal("HMAC signature should validate")
	}

	// Wrong secret fails.
	if validateHMAC("wrong-secret", payload, signature) {
		t.Fatal("wrong secret should fail validation")
	}

	// Tampered payload fails.
	if validateHMAC(secret, payload+"tampered", signature) {
		t.Fatal("tampered payload should fail validation")
	}
}

// Property 18: Notification template variable resolution — template variables
// in notification messages resolve correctly.
func TestProperty_NotificationTemplateVariableResolution(t *testing.T) {
	template := "User {USERNAME} on {NODE_NAME} has {DAYS_LEFT} days left"
	vars := map[string]string{
		"{USERNAME}":  "john",
		"{NODE_NAME}": "DE-1",
		"{DAYS_LEFT}": "5",
	}

	result := resolveTemplate(template, vars)
	expected := "User john on DE-1 has 5 days left"
	if result != expected {
		t.Fatalf("expected %q, got %q", expected, result)
	}
}

// --- helpers ---

type testChannel struct {
	ID     string
	Scope  string
	Events []string
}

func routeNotification(channels []testChannel, event, scope string) []testChannel {
	var matched []testChannel
	for _, ch := range channels {
		if ch.Scope == scope && channelMatchesEvent(ch, event) {
			matched = append(matched, ch)
		}
	}
	return matched
}

func channelMatchesEvent(ch testChannel, event string) bool {
	for _, e := range ch.Events {
		if e == event {
			return true
		}
	}
	return false
}

func computeHMAC(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func validateHMAC(secret, payload, signature string) bool {
	expected := computeHMAC(secret, payload)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func resolveTemplate(template string, vars map[string]string) string {
	result := template
	for k, v := range vars {
		for i := 0; i < len(result); i++ {
			idx := indexOf(result, k)
			if idx == -1 {
				break
			}
			result = result[:idx] + v + result[idx+len(k):]
		}
	}
	return result
}

func indexOf(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
