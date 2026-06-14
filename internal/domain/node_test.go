package domain

import (
	"testing"
	"time"
)

func TestNodeLive(t *testing.T) {
	now := time.Date(2026, 6, 14, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-10 * time.Second)
	stale := now.Add(-10 * time.Minute)

	cases := []struct {
		name string
		node Node
		want bool
	}{
		{"never polled is given benefit of the doubt", Node{LastSeen: nil}, true},
		{"fresh heartbeat + core running is live", Node{LastSeen: &fresh, Health: NodeHealth{CoreRunning: true}}, true},
		{"stale heartbeat is pruned", Node{LastSeen: &stale, Health: NodeHealth{CoreRunning: true}}, false},
		{"core reported down is pruned", Node{LastSeen: &fresh, Health: NodeHealth{CoreRunning: false}}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.node.Live(now, 90*time.Second); got != c.want {
				t.Errorf("Live = %v, want %v", got, c.want)
			}
		})
	}
}
