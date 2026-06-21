package service

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

const (
	// cleanIPMaxCandidates caps the candidate list to bound scan work and
	// prevent the scanner being abused as an SSRF/DoS amplifier (Req 4.1.3).
	cleanIPMaxCandidates = 256
	// cleanIPWorkers bounds probe concurrency, mirroring the Reality scanner.
	cleanIPWorkers = 10
	// cleanIPProbeAttempts is the number of dials used to measure loss/latency.
	cleanIPProbeAttempts = 4
	// cleanIPDialTimeout is the per-dial connect timeout.
	cleanIPDialTimeout = 5 * time.Second
	// cleanIPUnreachableLatencyMS is the sentinel latency recorded when every
	// probe attempt fails (so unreachable IPs sort last but stay representable).
	cleanIPUnreachableLatencyMS = 9999
)

// cleanIPProbeResult is the raw measurement for one candidate IP.
type cleanIPProbeResult struct {
	latencyMS int
	lossPct   int // 0-100
	reachable bool
}

// cleanIPProbeFunc probes one IP:port and returns latency/loss. It is the
// network seam: tests inject a fake so no real connections are made.
type cleanIPProbeFunc func(ctx context.Context, ip string, port int) cleanIPProbeResult

// CleanIPScannerService probes candidate CDN IPs (e.g. Cloudflare) for latency
// and packet loss, scores them, and caches the best-first results. It mirrors
// RealityScannerService end to end.
type CleanIPScannerService struct {
	repo  port.CleanIPScanRepository
	probe cleanIPProbeFunc
	now   func() time.Time
}

// NewCleanIPScannerService wires the service with the default TCP-connect probe.
func NewCleanIPScannerService(repo port.CleanIPScanRepository) *CleanIPScannerService {
	return &CleanIPScannerService{repo: repo, probe: probeCleanIPTCP, now: time.Now}
}

// Scan validates the candidate IPs, probes them under a bounded worker pool,
// scores and sorts them best-first, then replaces the cached result set
// (DeleteAll + SaveBatch). The candidate list is capped and restricted to
// public IPs to prevent SSRF/DoS abuse.
func (s *CleanIPScannerService) Scan(ctx context.Context, ips []string, scanPort int) ([]*domain.CleanIPScan, error) {
	if scanPort <= 0 {
		scanPort = 443
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("ips list is required")
	}
	if len(ips) > cleanIPMaxCandidates {
		return nil, fmt.Errorf("too many candidates: %d (max %d)", len(ips), cleanIPMaxCandidates)
	}
	// SSRF guard: every candidate must be a syntactically valid public IP.
	for _, ip := range ips {
		if !isPublicIP(ip) {
			return nil, fmt.Errorf("invalid or non-public IP: %q", ip)
		}
	}

	results := make([]*domain.CleanIPScan, len(ips))
	now := s.now()
	var wg sync.WaitGroup
	sem := make(chan struct{}, cleanIPWorkers)

	for i, ip := range ips {
		wg.Add(1)
		go func(idx int, candidate string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r := s.probe(ctx, candidate, scanPort)
			results[idx] = &domain.CleanIPScan{
				ID:        uuid.New(),
				IP:        candidate,
				LatencyMS: r.latencyMS,
				LossPct:   r.lossPct,
				Score:     scoreCleanIP(r.latencyMS, r.lossPct, r.reachable),
				Reachable: r.reachable,
				ScannedAt: now,
			}
		}(i, ip)
	}
	wg.Wait()

	// Sort best-first: highest score, then lowest latency as a tie-breaker.
	sort.SliceStable(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].LatencyMS < results[j].LatencyMS
	})

	_ = s.repo.DeleteAll(ctx)
	_ = s.repo.SaveBatch(ctx, results)

	return results, nil
}

// GetCachedResults returns the last scan's scored IPs (score DESC).
func (s *CleanIPScannerService) GetCachedResults(ctx context.Context) ([]*domain.CleanIPScan, error) {
	return s.repo.List(ctx)
}

// scoreCleanIP maps latency + loss to a 0-100 score (higher is better):
// unreachable IPs score 0; otherwise score = 100 - latencyMS/10 - lossPct,
// clamped to [0,100]. A ~100ms IP with no loss scores 90; 1000ms or 100% loss
// drops to 0.
func scoreCleanIP(latencyMS, lossPct int, reachable bool) int {
	if !reachable {
		return 0
	}
	score := 100 - latencyMS/10 - lossPct
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score
}

// isPublicIP reports whether s parses as an IP that is safe to probe: it must be
// a valid IP and must NOT be loopback, private, link-local, multicast, or the
// unspecified address. This is the SSRF guard for candidate inputs.
func isPublicIP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	return true
}

// probeCleanIPTCP performs cleanIPProbeAttempts TCP connects to ip:port,
// measuring connect latency and packet loss. LatencyMS is the average of
// successful connects; LossPct is failed/total*100; Reachable is true if any
// connect succeeded.
func probeCleanIPTCP(ctx context.Context, ip string, scanPort int) cleanIPProbeResult {
	addr := net.JoinHostPort(ip, strconv.Itoa(scanPort))
	var successes int
	var totalLatencyMS int64
	for i := 0; i < cleanIPProbeAttempts; i++ {
		d := net.Dialer{Timeout: cleanIPDialTimeout}
		start := time.Now()
		conn, err := d.DialContext(ctx, "tcp", addr)
		elapsed := time.Since(start)
		if err != nil {
			continue
		}
		_ = conn.Close() //nolint:errcheck // best-effort probe cleanup
		successes++
		totalLatencyMS += elapsed.Milliseconds()
	}
	failed := cleanIPProbeAttempts - successes
	res := cleanIPProbeResult{
		lossPct:   failed * 100 / cleanIPProbeAttempts,
		reachable: successes > 0,
	}
	if successes > 0 {
		res.latencyMS = int(totalLatencyMS / int64(successes))
	} else {
		res.latencyMS = cleanIPUnreachableLatencyMS
	}
	return res
}
