package api

import (
	"sync"
	"time"
)

// LoginThrottle is an in-memory brute-force guard for the login endpoint. It
// counts consecutive failures per key (client IP + username) and, once the
// threshold is crossed, locks that key out for a cooldown window. A successful
// login resets the counter. It is process-local (no shared store) which is fine
// for a single panel instance; multi-instance deployments should front it with a
// shared rate limiter.
type LoginThrottle struct {
	mu      sync.Mutex
	m       map[string]*loginAttempt
	max     int
	lockout time.Duration
	now     func() time.Time
}

type loginAttempt struct {
	fails     int
	lockUntil time.Time
	seen      time.Time
}

// NewLoginThrottle builds a throttle: max consecutive failures before a lockout
// of the given duration. Zero values fall back to 5 failures / 15 minutes.
func NewLoginThrottle(max int, lockout time.Duration) *LoginThrottle {
	if max <= 0 {
		max = 5
	}
	if lockout <= 0 {
		lockout = 15 * time.Minute
	}
	return &LoginThrottle{m: map[string]*loginAttempt{}, max: max, lockout: lockout, now: time.Now}
}

// RetryAfter returns the remaining lockout for key, or 0 if it may attempt now.
func (t *LoginThrottle) RetryAfter(key string) time.Duration {
	if t == nil {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.prune()
	a := t.m[key]
	if a == nil {
		return 0
	}
	if d := a.lockUntil.Sub(t.now()); d > 0 {
		return d
	}
	return 0
}

// Fail records a failed attempt, arming a lockout once the threshold is reached.
func (t *LoginThrottle) Fail(key string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	a := t.m[key]
	if a == nil {
		a = &loginAttempt{}
		t.m[key] = a
	}
	a.fails++
	a.seen = t.now()
	if a.fails >= t.max {
		a.lockUntil = t.now().Add(t.lockout)
		a.fails = 0 // restart counting after the lockout expires
	}
}

// Reset clears a key's failure state after a successful login.
func (t *LoginThrottle) Reset(key string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.m, key)
}

// prune drops stale entries; caller holds the lock.
func (t *LoginThrottle) prune() {
	cutoff := t.now().Add(-2 * t.lockout)
	for k, a := range t.m {
		if a.lockUntil.Before(cutoff) && a.seen.Before(cutoff) {
			delete(t.m, k)
		}
	}
}
