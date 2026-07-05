package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
)

// fakeCleanIPRepo is an in-memory CleanIPScanRepository for service unit tests.
type fakeCleanIPRepo struct {
	saved      []*domain.CleanIPScan
	deleteCalls int
}

func (r *fakeCleanIPRepo) UpdateThroughput(_ context.Context, id uuid.UUID, mbps float64) error {
	for _, s := range r.saved {
		if s.ID == id {
			s.ThroughputMbps = mbps
			return nil
		}
	}
	return nil
}

func (r *fakeCleanIPRepo) SaveBatch(_ context.Context, results []*domain.CleanIPScan) error {
	r.saved = results
	return nil
}

func (r *fakeCleanIPRepo) List(_ context.Context) ([]*domain.CleanIPScan, error) {
	return r.saved, nil
}

func (r *fakeCleanIPRepo) DeleteAll(_ context.Context) error {
	r.deleteCalls++
	r.saved = nil
	return nil
}

// newTestScanner wires a scanner with a deterministic, network-free probe whose
// result is looked up per IP from the supplied map.
func newTestScanner(repo *fakeCleanIPRepo, byIP map[string]cleanIPProbeResult) *CleanIPScannerService {
	s := NewCleanIPScannerService(repo)
	s.probe = func(_ context.Context, ip string, _ int) cleanIPProbeResult {
		return byIP[ip]
	}
	return s
}

func TestScoreCleanIP(t *testing.T) {
	tests := []struct {
		name      string
		latencyMS int
		lossPct   int
		reachable bool
		want      int
	}{
		{"unreachable scores zero", 10, 0, false, 0},
		{"fast no loss", 100, 0, true, 90},
		{"zero latency no loss", 0, 0, true, 100},
		{"latency and loss penalised", 200, 10, true, 70},
		{"high latency clamps to zero", 2000, 0, true, 0},
		{"full loss with low latency", 50, 100, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scoreCleanIP(tt.latencyMS, tt.lossPct, tt.reachable)
			if got != tt.want {
				t.Errorf("scoreCleanIP(%d,%d,%v) = %d, want %d", tt.latencyMS, tt.lossPct, tt.reachable, got, tt.want)
			}
		})
	}
}

func TestIsPublicIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"1.1.1.1", true},
		{"104.16.0.1", true},
		{"2606:4700:4700::1111", true},
		{"192.168.1.1", false},     // private
		{"10.0.0.5", false},        // private
		{"172.16.5.4", false},      // private
		{"127.0.0.1", false},       // loopback
		{"169.254.1.1", false},     // link-local
		{"224.0.0.1", false},       // multicast
		{"0.0.0.0", false},         // unspecified
		{"::1", false},             // loopback v6
		{"not-an-ip", false},       // garbage
		{"", false},                // empty
		{"999.999.999.999", false}, // invalid octets
	}
	for _, tt := range tests {
		if got := isPublicIP(tt.ip); got != tt.want {
			t.Errorf("isPublicIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestScanSortsBestFirst(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	byIP := map[string]cleanIPProbeResult{
		"1.1.1.1":     {latencyMS: 300, lossPct: 0, reachable: true},  // score 70
		"104.16.0.1":  {latencyMS: 50, lossPct: 0, reachable: true},   // score 95
		"104.16.0.2":  {latencyMS: 0, lossPct: 0, reachable: false},   // score 0
		"8.8.8.8":     {latencyMS: 100, lossPct: 25, reachable: true}, // score 65
	}
	ips := []string{"1.1.1.1", "104.16.0.1", "104.16.0.2", "8.8.8.8"}
	s := newTestScanner(repo, byIP)

	out, err := s.Scan(context.Background(), ips, 443)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("got %d results, want 4", len(out))
	}
	wantOrder := []string{"104.16.0.1", "1.1.1.1", "8.8.8.8", "104.16.0.2"}
	for i, ip := range wantOrder {
		if out[i].IP != ip {
			t.Errorf("position %d = %q, want %q (scores: %v)", i, out[i].IP, ip, scores(out))
		}
	}
	// Scan replaces the cached set: DeleteAll then SaveBatch.
	if repo.deleteCalls != 1 {
		t.Errorf("DeleteAll called %d times, want 1", repo.deleteCalls)
	}
	if len(repo.saved) != 4 {
		t.Errorf("SaveBatch saved %d, want 4", len(repo.saved))
	}
}

func scores(results []*domain.CleanIPScan) string {
	var b strings.Builder
	for _, r := range results {
		fmt.Fprintf(&b, "%s=%d ", r.IP, r.Score)
	}
	return b.String()
}

func TestScanRejectsOversizeList(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	s := newTestScanner(repo, nil)
	ips := make([]string, cleanIPMaxCandidates+1)
	for i := range ips {
		ips[i] = "1.1.1.1"
	}
	if _, err := s.Scan(context.Background(), ips, 443); err == nil {
		t.Fatal("expected error for oversize candidate list")
	}
	if repo.deleteCalls != 0 {
		t.Error("repo mutated despite rejected request")
	}
}

func TestScanRejectsEmptyList(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	s := newTestScanner(repo, nil)
	if _, err := s.Scan(context.Background(), nil, 443); err == nil {
		t.Fatal("expected error for empty candidate list")
	}
}

func TestScanRejectsNonPublicIPs(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	s := newTestScanner(repo, nil)
	bad := []string{"192.168.0.1", "127.0.0.1", "10.1.2.3", "garbage", "169.254.0.1"}
	for _, ip := range bad {
		if _, err := s.Scan(context.Background(), []string{"1.1.1.1", ip}, 443); err == nil {
			t.Errorf("expected rejection for candidate list containing %q", ip)
		}
	}
	if repo.deleteCalls != 0 {
		t.Error("repo mutated despite rejected SSRF-guard request")
	}
}

func TestScanAcceptsPublicIPsAndDefaultsPort(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	captured := -1
	s := NewCleanIPScannerService(repo)
	s.probe = func(_ context.Context, _ string, p int) cleanIPProbeResult {
		captured = p
		return cleanIPProbeResult{latencyMS: 20, lossPct: 0, reachable: true}
	}
	if _, err := s.Scan(context.Background(), []string{"1.1.1.1"}, 0); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if captured != 443 {
		t.Errorf("default port = %d, want 443", captured)
	}
}

func TestMeasureThroughputPersistsAndDefaultsPort(t *testing.T) {
	id := uuid.New()
	repo := &fakeCleanIPRepo{saved: []*domain.CleanIPScan{{ID: id, IP: "1.1.1.1"}}}
	s := NewCleanIPScannerService(repo)
	capturedPort := -1
	s.throughput = func(_ context.Context, ip string, port int) (float64, error) {
		capturedPort = port
		return 123.45, nil
	}

	got, err := s.MeasureThroughput(context.Background(), id, "1.1.1.1", 0)
	if err != nil {
		t.Fatalf("MeasureThroughput: %v", err)
	}
	if got != 123.45 {
		t.Errorf("got %v, want 123.45", got)
	}
	if capturedPort != 443 {
		t.Errorf("default port = %d, want 443", capturedPort)
	}
	if repo.saved[0].ThroughputMbps != 123.45 {
		t.Errorf("repo not updated: %+v", repo.saved[0])
	}
}

func TestMeasureThroughputRejectsNonPublicIP(t *testing.T) {
	repo := &fakeCleanIPRepo{}
	s := NewCleanIPScannerService(repo)
	s.throughput = func(context.Context, string, int) (float64, error) {
		t.Fatal("throughput func should not be called for a rejected IP")
		return 0, nil
	}
	if _, err := s.MeasureThroughput(context.Background(), uuid.New(), "192.168.1.1", 443); err == nil {
		t.Fatal("expected error for non-public IP")
	}
}

func TestMeasureThroughputPropagatesProbeError(t *testing.T) {
	repo := &fakeCleanIPRepo{saved: []*domain.CleanIPScan{{ID: uuid.New(), IP: "1.1.1.1"}}}
	s := NewCleanIPScannerService(repo)
	s.throughput = func(context.Context, string, int) (float64, error) {
		return 0, fmt.Errorf("boom")
	}
	if _, err := s.MeasureThroughput(context.Background(), uuid.New(), "1.1.1.1", 443); err == nil {
		t.Fatal("expected error to propagate")
	}
}

func TestGetCachedResultsPassthrough(t *testing.T) {
	repo := &fakeCleanIPRepo{saved: []*domain.CleanIPScan{{IP: "1.1.1.1", Score: 90}}}
	s := NewCleanIPScannerService(repo)
	out, err := s.GetCachedResults(context.Background())
	if err != nil {
		t.Fatalf("GetCachedResults: %v", err)
	}
	if len(out) != 1 || out[0].IP != "1.1.1.1" {
		t.Errorf("passthrough mismatch: %+v", out)
	}
}
