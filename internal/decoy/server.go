// Package decoy serves static/honeypot fallback pages for invalid inbound probes.
package decoy

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/vortexui/vortexui/internal/domain"
)

// Server is a loopback HTTP server for decoy and honeypot HTML.
type Server struct {
	mu      sync.Mutex
	srv     *http.Server
	body    string
	listen  string
	enabled bool
}

// NewServer wires a decoy HTTP server.
func NewServer() *Server {
	return &Server{listen: domain.DefaultDecoyListen}
}

// ReloadFromInbounds picks the first enabled decoy/honeypot page from inbounds.
func (s *Server) ReloadFromInbounds(inbounds []domain.Inbound) error {
	body, enabled := extractBody(inbounds)
	return s.Reload(body, enabled)
}

// Reload starts or stops the loopback listener.
func (s *Server) Reload(body string, enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.srv != nil {
		_ = s.srv.Shutdown(context.Background())
		s.srv = nil
	}
	s.body = body
	s.enabled = enabled
	if !enabled || strings.TrimSpace(body) == "" {
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(s.body))
	})

	s.srv = &http.Server{
		Addr:              s.listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() { _ = s.srv.ListenAndServe() }()
	return waitListen(s.listen)
}

// Stop shuts down the decoy listener.
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

func extractBody(inbounds []domain.Inbound) (string, bool) {
	for _, in := range inbounds {
		raw, ok := in.Raw["decoy"].(map[string]any)
		if !ok {
			continue
		}
		mode, _ := raw["mode"].(string)
		switch domain.DecoyMode(mode) {
		case domain.DecoyStatic:
			if html, _ := raw["static_html"].(string); strings.TrimSpace(html) != "" {
				return html, true
			}
		case domain.DecoyProxy:
			continue
		default:
			if html, _ := raw["static_html"].(string); strings.TrimSpace(html) != "" {
				return html, true
			}
		}
	}
	return "", false
}

func waitListen(addr string) error {
	deadline := 20
	for i := 0; i < deadline; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			return nil
		}
	}
	return nil
}
