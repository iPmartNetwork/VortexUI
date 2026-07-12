// Package doh implements a built-in DNS-over-HTTPS (RFC 8484) resolver for the panel.
package doh

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/miekg/dns"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/panel/port"
)

// Server serves DNS queries over HTTPS using upstream resolvers from config.
type Server struct {
	repo port.DoHRepository
	log  *slog.Logger

	mu   sync.Mutex
	srv  *http.Server
	cfg  atomic.Pointer[domain.DoHConfig]
}

// NewServer wires the DoH server.
func NewServer(repo port.DoHRepository, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	return &Server{repo: repo, log: log}
}

// Reload reads config from the repository and starts or stops the listener.
func (s *Server) Reload(ctx context.Context) error {
	cfg, err := s.repo.GetConfig(ctx)
	if err != nil || cfg == nil {
		def := domain.DefaultDoHConfig()
		cfg = &def
	}
	s.cfg.Store(cfg)

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.srv != nil {
		_ = s.srv.Shutdown(context.Background())
		s.srv = nil
	}
	if !cfg.Enabled {
		return nil
	}

	addr := strings.TrimSpace(cfg.ListenAddr)
	if addr == "" {
		addr = ":8053"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/dns-query", s.handleDNS)
	mux.HandleFunc("/", s.handleDNS)

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		s.log.Info("doh server listening", "addr", addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Error("doh server failed", "err", err)
		}
	}()
	return nil
}

// Stop shuts down the DoH listener.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv == nil {
		return nil
	}
	err := s.srv.Shutdown(ctx)
	s.srv = nil
	return err
}

func (s *Server) handleDNS(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg.Load()
	if cfg == nil || !cfg.Enabled {
		http.Error(w, "doh disabled", http.StatusServiceUnavailable)
		return
	}

	var wire []byte
	var err error
	switch r.Method {
	case http.MethodGet:
		wire, err = dnsMsgUnpack(r.URL.Query().Get("dns"))
	case http.MethodPost:
		if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/dns-message") {
			http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
			return
		}
		wire, err = readAllLimited(r.Body, 4096)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err != nil || len(wire) == 0 {
		http.Error(w, "invalid dns query", http.StatusBadRequest)
		return
	}

	req := new(dns.Msg)
	if err := req.Unpack(wire); err != nil {
		http.Error(w, "invalid dns query", http.StatusBadRequest)
		return
	}

	start := time.Now()
	clientIP := clientIPFromRequest(r)
	qname, qtype := queryMeta(req)

	if blocked := s.shouldBlock(cfg, qname); blocked {
		s.logQuery(cfg, qname, qtype, clientIP, true, start)
		writeEmptyDNS(w, req)
		return
	}

	resp, err := s.exchange(cfg, req)
	if err != nil {
		http.Error(w, "upstream failed", http.StatusBadGateway)
		return
	}
	out, err := resp.Pack()
	if err != nil {
		http.Error(w, "encode failed", http.StatusInternalServerError)
		return
	}

	s.logQuery(cfg, qname, qtype, clientIP, false, start)
	w.Header().Set("Content-Type", "application/dns-message")
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge(cfg)))
	_, _ = w.Write(out)
}

func (s *Server) exchange(cfg *domain.DoHConfig, req *dns.Msg) (*dns.Msg, error) {
	upstreams := cfg.UpstreamDNS
	if len(upstreams) == 0 {
		upstreams = domain.DefaultDoHConfig().UpstreamDNS
	}
	client := &dns.Client{Timeout: 4 * time.Second}
	var lastErr error
	for _, upstream := range upstreams {
		addr := normalizeUpstream(upstream)
		if addr == "" {
			continue
		}
		resp, _, err := client.Exchange(req, addr)
		if err == nil && resp != nil {
			return resp, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no upstream dns configured")
	}
	return nil, lastErr
}

func (s *Server) shouldBlock(cfg *domain.DoHConfig, qname string) bool {
	name := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(qname)), ".")
	if name == "" {
		return false
	}
	for _, blocked := range cfg.CustomBlocklist {
		b := strings.ToLower(strings.TrimSpace(blocked))
		if b == "" {
			continue
		}
		if name == b || strings.HasSuffix(name, "."+b) {
			return true
		}
	}
	return false
}

func (s *Server) logQuery(cfg *domain.DoHConfig, qname, qtype, clientIP string, blocked bool, start time.Time) {
	if cfg == nil || !cfg.LogQueries || s.repo == nil {
		return
	}
	_ = s.repo.SaveQueryLog(context.Background(), &domain.DoHQueryLog{
		Domain:    qname,
		Type:      qtype,
		ClientIP:  clientIP,
		Blocked:   blocked,
		LatencyMS: int(time.Since(start).Milliseconds()),
		Timestamp: time.Now(),
	})
}

func writeEmptyDNS(w http.ResponseWriter, req *dns.Msg) {
	resp := new(dns.Msg)
	resp.SetReply(req)
	out, _ := resp.Pack()
	w.Header().Set("Content-Type", "application/dns-message")
	_, _ = w.Write(out)
}

func queryMeta(req *dns.Msg) (name, qtype string) {
	if req == nil || len(req.Question) == 0 {
		return "", "A"
	}
	q := req.Question[0]
	return q.Name, dns.TypeToString[q.Qtype]
}

func normalizeUpstream(upstream string) string {
	upstream = strings.TrimSpace(upstream)
	if upstream == "" {
		return ""
	}
	if strings.Contains(upstream, ":") {
		return upstream
	}
	return net.JoinHostPort(upstream, "53")
}

func maxAge(cfg *domain.DoHConfig) int {
	if cfg == nil || cfg.CacheTTL <= 0 {
		return 300
	}
	return cfg.CacheTTL
}

func clientIPFromRequest(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func dnsMsgUnpack(raw string) ([]byte, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty dns param")
	}
	return base64.RawURLEncoding.DecodeString(raw)
}

func readAllLimited(r interface{ Read([]byte) (int, error) }, limit int) ([]byte, error) {
	buf := make([]byte, 0, 512)
	tmp := make([]byte, 512)
	for len(buf) < limit {
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty body")
	}
	if len(buf) > limit {
		return nil, fmt.Errorf("body too large")
	}
	return buf, nil
}
