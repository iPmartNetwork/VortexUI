package events

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func quietBus() *Bus {
	return New(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func TestPublishFansOutToAllSubscribers(t *testing.T) {
	b := quietBus()
	a := b.Subscribe(8)
	c := b.Subscribe(8)

	b.Publish(Event{Type: UserLimited, Username: "alice"})

	for _, ch := range []<-chan Event{a, c} {
		select {
		case e := <-ch:
			if e.Type != UserLimited || e.Username != "alice" {
				t.Errorf("unexpected event: %+v", e)
			}
			if e.Time.IsZero() {
				t.Error("publish should stamp the event time")
			}
		case <-time.After(time.Second):
			t.Fatal("subscriber did not receive the event")
		}
	}
}

func TestPublishDropsWhenSubscriberFull(t *testing.T) {
	b := quietBus()
	ch := b.Subscribe(1)

	// Fill the buffer, then publish more than it can hold; must not block.
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			b.Publish(Event{Type: NodeDown})
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a full subscriber")
	}
	// At least the first event is buffered.
	if len(ch) == 0 {
		t.Error("expected at least one buffered event")
	}
}

func TestNopPublisherIsSafe(t *testing.T) {
	var p Publisher = Nop{}
	p.Publish(Event{Type: UserExpired}) // must not panic
}
