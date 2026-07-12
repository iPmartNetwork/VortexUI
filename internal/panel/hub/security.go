package hub

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/domain"
	"github.com/vortexui/vortexui/internal/node/security"
)

// SecurityIngest handles probe/fingerprint observations from node logs.
type SecurityIngest func(ctx context.Context, nodeID uuid.UUID, ev domain.SecurityEvent)

// runSecurity polls node core logs and forwards parsed security events.
func (m *managedNode) runSecurity(ctx context.Context) {
	seen := make(map[string]struct{})
	var mu sync.Mutex
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ingest := m.hub.opts.SecurityIngest
			if ingest == nil {
				continue
			}
			conn, err := m.ensureConn()
			if err != nil {
				continue
			}
			lines, err := conn.Logs(ctx, 200)
			if err != nil {
				continue
			}
			for _, line := range lines {
				ev, ok := security.ParseLogLine(line)
				if !ok {
					continue
				}
				key := ev.SourceIP + "|" + ev.Method + "|" + ev.Fingerprint
				mu.Lock()
				if _, dup := seen[key]; dup {
					mu.Unlock()
					continue
				}
				if len(seen) > 4096 {
					seen = make(map[string]struct{})
				}
				seen[key] = struct{}{}
				mu.Unlock()
				ingest(ctx, m.node.ID, ev)
			}
		}
	}
}
