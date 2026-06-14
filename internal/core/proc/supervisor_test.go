package proc

import "testing"

func TestLineRingTailOrderingAndEviction(t *testing.T) {
	r := newLineRing(3)
	for _, l := range []string{"a", "b"} {
		r.add(l)
	}
	// Not yet full: oldest→newest.
	if got := r.tail(0); len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("partial ring = %v, want [a b]", got)
	}

	for _, l := range []string{"c", "d", "e"} {
		r.add(l)
	}
	// Capacity 3: only the last three remain, oldest→newest.
	got := r.tail(0)
	if len(got) != 3 || got[0] != "c" || got[2] != "e" {
		t.Errorf("full ring = %v, want [c d e]", got)
	}

	// Limit returns the most recent N.
	if last := r.tail(1); len(last) != 1 || last[0] != "e" {
		t.Errorf("tail(1) = %v, want [e]", last)
	}
}

func TestSupervisorLogsEmptyBeforeStart(t *testing.T) {
	s := New("noop", nil, nil)
	if got := s.Logs(0); len(got) != 0 {
		t.Errorf("expected no logs before start, got %v", got)
	}
}
