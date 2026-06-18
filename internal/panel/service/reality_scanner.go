package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// RealityScannerService probes SNIs for REALITY compatibility.
type RealityScannerService struct {
	repo  port.RealityScanRepository
	nodes port.NodeRepository
}

// NewRealityScannerService wires dependencies.
func NewRealityScannerService(repo port.RealityScanRepository, nodes port.NodeRepository) *RealityScannerService {
	return &RealityScannerService{repo: repo, nodes: nodes}
}

// ScanResult holds the outcome of probing one SNI.
type ScanResult struct {
	SNI       string `json:"sni"`
	LatencyMS int    `json:"latency_ms"`
	Score     int    `json:"score"`
	Valid     bool   `json:"valid"`
	Error     string `json:"error,omitempty"`
}

// Scan probes a list of SNIs from the given node and returns scored results.
// Probing is done concurrently with a limit of 10 goroutines.
func (s *RealityScannerService) Scan(ctx context.Context, nodeID uuid.UUID, snis []string, port int) ([]ScanResult, error) {
	if port <= 0 {
		port = 443
	}
	if _, err := s.nodes.GetByID(ctx, nodeID); err != nil {
		return nil, fmt.Errorf("node not found: %w", err)
	}

	results := make([]ScanResult, len(snis))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)

	for i, sni := range snis {
		wg.Add(1)
		go func(idx int, serverName string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Probe the SNI domain directly (not through the node) to check
			// if it supports TLS 1.3 and measure latency. This tells us if
			// the domain is suitable as a REALITY camouflage target.
			r := probeSNI(serverName, port, serverName)
			results[idx] = r
		}(i, sni)
	}
	wg.Wait()

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Persist results.
	var toSave []domain.RealityScan
	now := time.Now()
	for _, r := range results {
		toSave = append(toSave, domain.RealityScan{
			ID:        uuid.New(),
			NodeID:    nodeID,
			SNI:       r.SNI,
			LatencyMS: r.LatencyMS,
			Score:     r.Score,
			Valid:     r.Valid,
			ScannedAt: now,
		})
	}
	_ = s.repo.DeleteByNode(ctx, nodeID)
	_ = s.repo.SaveBatch(ctx, toSave)

	return results, nil
}

// GetCachedResults returns previously scanned results for a node.
func (s *RealityScannerService) GetCachedResults(ctx context.Context, nodeID uuid.UUID) ([]domain.RealityScan, error) {
	return s.repo.ListByNode(ctx, nodeID)
}

// probeSNI performs a TLS handshake to the target host with the given SNI and
// measures latency. Scores are based on latency and TLS version.
func probeSNI(host string, port int, sni string) ScanResult {
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		ServerName:         sni,
		InsecureSkipVerify: true,
	})
	elapsed := time.Since(start)
	latencyMS := int(elapsed.Milliseconds())

	if err != nil {
		return ScanResult{SNI: sni, LatencyMS: latencyMS, Score: 0, Valid: false, Error: err.Error()}
	}
	defer conn.Close() //nolint:errcheck // best-effort TLS cleanup

	// Score calculation: lower latency = higher score, TLS 1.3 preferred.
	score := 100
	if latencyMS > 500 {
		score -= 40
	} else if latencyMS > 200 {
		score -= 20
	} else if latencyMS > 100 {
		score -= 10
	}

	state := conn.ConnectionState()
	if state.Version < tls.VersionTLS13 {
		score -= 20
	}
	if !state.HandshakeComplete {
		score -= 30
	}
	if score < 0 {
		score = 0
	}

	return ScanResult{SNI: sni, LatencyMS: latencyMS, Score: score, Valid: true}
}
