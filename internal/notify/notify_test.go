package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vortexui/vortexui/internal/events"
)

func quietLog() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestWebhookPostsSignedJSON(t *testing.T) {
	var gotBody []byte
	var gotSig string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		gotSig = r.Header.Get("X-Vortex-Signature")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := NewWebhook(srv.URL, "s3cret", quietLog())
	if err := wh.deliver(context.Background(), events.Event{Type: events.UserLimited, Username: "alice"}); err != nil {
		t.Fatalf("deliver: %v", err)
	}

	var e events.Event
	if err := json.Unmarshal(gotBody, &e); err != nil {
		t.Fatalf("payload not JSON: %v", err)
	}
	if e.Type != events.UserLimited || e.Username != "alice" {
		t.Errorf("payload wrong: %+v", e)
	}
	// Verify the HMAC signature matches the body.
	mac := hmac.New(sha256.New, []byte("s3cret"))
	mac.Write(gotBody)
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if gotSig != want {
		t.Errorf("signature = %q, want %q", gotSig, want)
	}
}

func TestWebhookRetriesThenSucceeds(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&hits, 1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := NewWebhook(srv.URL, "", quietLog())
	wh.backoff = time.Millisecond // keep the test fast
	if err := wh.deliver(context.Background(), events.Event{Type: events.NodeDown}); err != nil {
		t.Fatalf("deliver should eventually succeed: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 3 {
		t.Errorf("attempts = %d, want 3 (2 failures + 1 success)", got)
	}
}

func TestWebhookFailsAfterAllAttempts(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	wh := NewWebhook(srv.URL, "", quietLog())
	wh.backoff = time.Millisecond
	if err := wh.deliver(context.Background(), events.Event{Type: events.NodeDown}); err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
}

func TestTelegramSendsMessage(t *testing.T) {
	var path string
	var body map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	tg := NewTelegram("BOTTOKEN", "12345", quietLog())
	tg.apiBase = srv.URL
	if err := tg.send(context.Background(), "hello"); err != nil {
		t.Fatalf("send: %v", err)
	}
	if path != "/botBOTTOKEN/sendMessage" {
		t.Errorf("path = %q, want /botBOTTOKEN/sendMessage", path)
	}
	if body["chat_id"] != "12345" || body["text"] != "hello" {
		t.Errorf("payload wrong: %+v", body)
	}
}

func TestTelegramRunFormatsEvents(t *testing.T) {
	// The handler runs on the httptest server's goroutine; pass the received
	// text back over a channel so the test reads it without a data race.
	got := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		text, _ := body["text"].(string)
		select {
		case got <- text:
		default:
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	tg := NewTelegram("T", "C", quietLog())
	tg.apiBase = srv.URL

	ch := make(chan events.Event, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go tg.Run(ctx, ch)

	ch <- events.Event{Type: events.UserExpired, Username: "bob", Time: time.Unix(0, 0)}

	select {
	case text := <-got:
		if !strings.Contains(text, "bob") {
			t.Errorf("text should name the user: %q", text)
		}
		if !strings.Contains(text, "expired") {
			t.Errorf("text should describe expiry: %q", text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("message not sent in time")
	}
}
