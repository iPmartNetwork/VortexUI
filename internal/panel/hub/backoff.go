package hub

import (
	"context"
	"time"
)

// backoff is a simple capped exponential backoff used by the reconnect loops.
// It doubles the delay on each sleep up to max, and reset() returns to min after
// a successful connection.
type backoff struct {
	min, max time.Duration
	cur      time.Duration
}

func (b *backoff) reset() { b.cur = 0 }

// sleep waits for the current backoff interval (advancing it), returning false
// if ctx is cancelled during the wait so callers can exit promptly.
func (b *backoff) sleep(ctx context.Context) bool {
	if b.cur == 0 {
		b.cur = b.min
	} else {
		b.cur *= 2
		if b.cur > b.max {
			b.cur = b.max
		}
	}
	t := time.NewTimer(b.cur)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}
