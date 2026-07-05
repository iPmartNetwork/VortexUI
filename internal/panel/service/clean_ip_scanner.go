package service

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
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

	// cleanIPThroughputHost is the well-known, always-on endpoint used as the
	// TLS SNI/Host for the real download-speed test. The TCP connection is
	// made directly to the candidate IP (that's the whole point of a clean-IP
	// scan) while TLS is validated against this hostname, exactly mirroring
	// how a CDN-fronted proxy inbound would be reached in production.
	cleanIPThroughputHost = "speed.cloudflare.com"
	// cleanIPThroughputPath requests a bounded-size payload; the actual
	// transfer is additionally capped by duration/bytes below so a slow or
	// misbehaving candidate can't stall a scan.
	cleanIPThroughputPath = "/__down?bytes=25000000"
	// cleanIPThroughputMaxBytes hard-caps how much we ever read.
	cleanIPThroughputMaxBytes = 25 * 1024 * 1024
	// cleanIPThroughputMaxDuration bounds the download window; a good clean
	// IP saturates the sample well before this elapses.
	cleanIPThroughputMaxDuration = 4 * time.Second
	// cleanIPThroughputMinDuration guards against unreliable measurements
	// from samples that finished too quickly (e.g. a tiny cached response).
	cleanIPThroughputMinDuration = 300 * time.Millisecond

	// cleanIPSchedulePollInterval is how often the scheduler checks whether
	// a due recurring scan should run. It is independent of, and much
	// shorter than, the configurable interval_minutes itself.
	cleanIPSchedulePollInterval = time.Minute
	// cleanIPScheduleMinIntervalMinutes floors the configurable cadence so a
	// misconfigured schedule can't hammer the network.
	cleanIPScheduleMinIntervalMinutes = 15
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

// cleanIPThroughputFunc runs a real download-speed test against one IP:port
// and returns the measured throughput in Mbps. It is the network seam for
// MeasureThroughput; tests inject a fake so no real connections are made.
type cleanIPThroughputFunc func(ctx context.Context, ip string, port int) (float64, error)

// CleanIPScannerService probes candidate CDN IPs (e.g. Cloudflare) for latency
// and packet loss, scores them, and caches the best-first results. It mirrors
// RealityScannerService end to end.
type CleanIPScannerService struct {
	repo         port.CleanIPScanRepository
	scheduleRepo port.CleanIPScheduleRepository
	probe        cleanIPProbeFunc
	throughput   cleanIPThroughputFunc
	now          func() time.Time
}

// NewCleanIPScannerService wires the service with the default TCP-connect
// probe and real HTTPS download-speed test.
func NewCleanIPScannerService(repo port.CleanIPScanRepository) *CleanIPScannerService {
	return &CleanIPScannerService{repo: repo, probe: probeCleanIPTCP, throughput: measureCleanIPThroughput, now: time.Now}
}

// SetScheduleRepo wires the recurring-scan config store. Until this is
// called, GetSchedule reports defaults, UpdateSchedule is rejected, and
// RunScheduler is a no-op — so scheduling is entirely optional.
func (s *CleanIPScannerService) SetScheduleRepo(repo port.CleanIPScheduleRepository) {
	s.scheduleRepo = repo
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

// MeasureThroughput runs a real download-speed test against one previously
// scanned candidate (identified by its cached result ID) and persists the
// measured Mbps. It is deliberately separate from the bulk Scan so operators
// pay the (slower) throughput cost only for the handful of promising
// candidates they actually care about.
//
// The target IP is resolved from the cached scan row rather than trusted
// from the caller, so a request can't be used to make the server probe an
// arbitrary IP under someone else's result ID.
func (s *CleanIPScannerService) MeasureThroughput(ctx context.Context, id uuid.UUID, scanPort int) (float64, error) {
	target, err := s.findCached(ctx, id)
	if err != nil {
		return 0, err
	}
	if scanPort <= 0 {
		scanPort = 443
	}
	mbps, err := s.throughput(ctx, target.IP, scanPort)
	if err != nil {
		return 0, err
	}
	if err := s.repo.UpdateThroughput(ctx, id, mbps); err != nil {
		return 0, err
	}
	return mbps, nil
}

// findCached looks up one cached scan row by ID and re-validates its IP,
// guarding against a stale/tampered cache entry ever reaching the network
// probe.
func (s *CleanIPScannerService) findCached(ctx context.Context, id uuid.UUID) (*domain.CleanIPScan, error) {
	cached, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cached {
		if c.ID == id {
			if !isPublicIP(c.IP) {
				return nil, fmt.Errorf("invalid or non-public IP: %q", c.IP)
			}
			return c, nil
		}
	}
	return nil, fmt.Errorf("clean-ip result %s not found", id)
}

// MeasureAllThroughput runs the download-speed test against every reachable
// cached candidate, bounded by the same worker limit as Scan, and persists
// each measurement as it completes. It returns the refreshed, best-first
// result set.
func (s *CleanIPScannerService) MeasureAllThroughput(ctx context.Context, scanPort int) ([]*domain.CleanIPScan, error) {
	cached, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	if scanPort <= 0 {
		scanPort = 443
	}

	sem := make(chan struct{}, cleanIPWorkers)
	var wg sync.WaitGroup
	for _, c := range cached {
		if !c.Reachable || !isPublicIP(c.IP) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{}
		go func(target *domain.CleanIPScan) {
			defer wg.Done()
			defer func() { <-sem }()
			if mbps, err := s.throughput(ctx, target.IP, scanPort); err == nil {
				_ = s.repo.UpdateThroughput(ctx, target.ID, mbps)
			}
		}(c)
	}
	wg.Wait()

	return s.repo.List(ctx)
}

// GetSchedule returns the current recurring-scan configuration, falling
// back to defaults when unset or unreadable.
func (s *CleanIPScannerService) GetSchedule(ctx context.Context) (*domain.CleanIPSchedule, error) {
	if s.scheduleRepo == nil {
		def := domain.DefaultCleanIPSchedule()
		return &def, nil
	}
	sched, err := s.scheduleRepo.GetSchedule(ctx)
	if err != nil {
		def := domain.DefaultCleanIPSchedule()
		return &def, nil
	}
	return sched, nil
}

// UpdateSchedule validates and persists the recurring-scan configuration.
// The same SSRF guard and candidate cap used by Scan apply here, since the
// scheduler ultimately calls Scan with these IPs unattended.
func (s *CleanIPScannerService) UpdateSchedule(ctx context.Context, sched *domain.CleanIPSchedule) error {
	if s.scheduleRepo == nil {
		return fmt.Errorf("scheduling is not available")
	}
	if sched.IntervalMinutes < cleanIPScheduleMinIntervalMinutes {
		sched.IntervalMinutes = cleanIPScheduleMinIntervalMinutes
	}
	if sched.Port <= 0 {
		sched.Port = 443
	}
	if len(sched.IPs) > cleanIPMaxCandidates {
		return fmt.Errorf("too many candidates: %d (max %d)", len(sched.IPs), cleanIPMaxCandidates)
	}
	if sched.Enabled {
		if len(sched.IPs) == 0 {
			return fmt.Errorf("ips list is required to enable scheduling")
		}
		for _, ip := range sched.IPs {
			if !isPublicIP(ip) {
				return fmt.Errorf("invalid or non-public IP: %q", ip)
			}
		}
	}
	return s.scheduleRepo.SaveSchedule(ctx, sched)
}

// RunScheduler polls the schedule every cleanIPSchedulePollInterval and,
// when enabled and due, runs a full Scan over the configured candidates. It
// blocks until ctx is cancelled, so callers should invoke it in a goroutine.
// It is safe to call even when no schedule repo is wired (becomes a no-op)
// so callers don't need to special-case that.
func (s *CleanIPScannerService) RunScheduler(ctx context.Context) {
	if s.scheduleRepo == nil {
		return
	}
	ticker := time.NewTicker(cleanIPSchedulePollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tickScheduler(ctx)
		}
	}
}

func (s *CleanIPScannerService) tickScheduler(ctx context.Context) {
	sched, err := s.scheduleRepo.GetSchedule(ctx)
	if err != nil || !sched.Enabled || len(sched.IPs) == 0 {
		return
	}
	interval := time.Duration(sched.IntervalMinutes) * time.Minute
	if sched.LastRunAt != nil && s.now().Sub(*sched.LastRunAt) < interval {
		return
	}
	if _, err := s.Scan(ctx, sched.IPs, sched.Port); err != nil {
		slog.Default().Warn("scheduled clean-ip scan failed", "err", err)
		return
	}
	if err := s.scheduleRepo.MarkScheduleRun(ctx, s.now()); err != nil {
		slog.Default().Warn("failed to record scheduled clean-ip scan run", "err", err)
	}
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

// measureCleanIPThroughput downloads a bounded sample through a direct TCP
// connection to ip:scanPort, TLS-fronted as cleanIPThroughputHost (the
// connection goes to the candidate IP; the TLS handshake — SNI and
// certificate validation — is done against the well-known host, exactly as a
// CDN-fronted inbound would see it in production). It returns the observed
// download throughput in Mbps, measured over the response body transfer only
// (connect + TLS handshake + response headers are excluded).
func measureCleanIPThroughput(parentCtx context.Context, ip string, scanPort int) (float64, error) {
	ctx, cancel := context.WithTimeout(parentCtx, cleanIPThroughputMaxDuration+cleanIPDialTimeout)
	defer cancel()

	dialer := &net.Dialer{Timeout: cleanIPDialTimeout}
	addr := net.JoinHostPort(ip, strconv.Itoa(scanPort))
	transport := &http.Transport{
		DialContext: func(dialCtx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(dialCtx, network, addr)
		},
		TLSHandshakeTimeout:   cleanIPDialTimeout,
		ResponseHeaderTimeout: cleanIPDialTimeout,
		DisableCompression:    true,
	}
	client := &http.Client{Transport: transport}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+cleanIPThroughputHost+cleanIPThroughputPath, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("throughput probe failed: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort cleanup

	start := time.Now()
	n, _ := io.Copy(io.Discard, io.LimitReader(resp.Body, cleanIPThroughputMaxBytes))
	elapsed := time.Since(start)
	if elapsed < cleanIPThroughputMinDuration || n <= 0 {
		return 0, fmt.Errorf("insufficient sample: %d bytes in %s", n, elapsed)
	}
	mbps := float64(n*8) / elapsed.Seconds() / 1_000_000
	return mbps, nil
}
