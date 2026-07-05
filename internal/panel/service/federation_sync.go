package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
)

// RunSyncWorker periodically health-checks federation peers and syncs metadata.
func (s *FederationService) RunSyncWorker(ctx context.Context, log *slog.Logger) {
	if log == nil {
		log = slog.Default()
	}
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.syncTick(ctx, log)
		}
	}
}

func (s *FederationService) syncTick(ctx context.Context, log *slog.Logger) {
	cfg, err := s.GetConfig(ctx)
	if err != nil || cfg == nil || !cfg.Enabled {
		return
	}

	peers, err := s.ListPeers(ctx)
	if err != nil {
		return
	}
	for _, peer := range peers {
		if peer == nil {
			continue
		}
		s.syncPeer(ctx, peer, log)
	}
}

func (s *FederationService) syncPeer(ctx context.Context, peer *domain.FederationPeer, log *slog.Logger) {
	base := strings.TrimRight(peer.Endpoint, "/")
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/api/health", nil)
	if err != nil {
		return
	}
	if peer.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+peer.APIKey)
	}
	resp, err := client.Do(req)
	now := s.now()
	peer.LastSync = &now

	if err != nil || resp.StatusCode >= 300 {
		peer.Status = domain.PeerDisconnected
		_ = s.repo.UpdatePeer(ctx, peer)
		s.recordEvent(ctx, peer, "health", 0, "failed", syncErrMsg(err, resp))
		if resp != nil && resp.Body != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}
		return
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	peer.Status = domain.PeerConnected
	_ = s.repo.UpdatePeer(ctx, peer)

	if peer.SyncUsers {
		count := s.fetchRemoteCount(ctx, client, base+"/api/users?limit=1", peer.APIKey)
		s.recordEvent(ctx, peer, "users", count, "success", "")
	}
	if peer.SyncNodes {
		count := s.fetchRemoteCount(ctx, client, base+"/api/nodes", peer.APIKey)
		s.recordEvent(ctx, peer, "nodes", count, "success", "")
	}
	log.Debug("federation peer synced", "peer", peer.Name, "status", peer.Status)
}

func (s *FederationService) fetchRemoteCount(ctx context.Context, client *http.Client, url, apiKey string) int {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return 0
	}
	var payload map[string]json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0
	}
	for _, key := range []string{"users", "nodes", "entries"} {
		if raw, ok := payload[key]; ok {
			var list []json.RawMessage
			if json.Unmarshal(raw, &list) == nil {
				return len(list)
			}
		}
	}
	return 0
}

func (s *FederationService) recordEvent(ctx context.Context, peer *domain.FederationPeer, entity string, count int, status, errMsg string) {
	ev := &domain.FederationSyncEvent{
		ID:         uuid.New(),
		PeerID:     peer.ID,
		PeerName:   peer.Name,
		Direction:  "pull",
		EntityType: entity,
		Count:      count,
		Status:     status,
		Error:      errMsg,
		CreatedAt:  s.now(),
	}
	_ = s.repo.SaveSyncEvent(ctx, ev)
}

func syncErrMsg(err error, resp *http.Response) string {
	if err != nil {
		return err.Error()
	}
	if resp != nil {
		return fmt.Sprintf("status %d", resp.StatusCode)
	}
	return "unknown error"
}
