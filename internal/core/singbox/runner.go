package singbox

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/vortexui/vortexui/internal/core/proc"
)

// fileRunner is the production applier: it writes the rendered config to disk and
// (re)starts sing-box via the shared supervisor. sing-box reads "-c <path>".
type fileRunner struct {
	sup  *proc.Supervisor
	path string

	mu         sync.Mutex
	lastConfig []byte // last config bytes actually written/applied; guards redundant restarts
}

func newFileRunner(binPath, configPath string, log *slog.Logger) *fileRunner {
	return &fileRunner{
		sup:  proc.New(binPath, []string{"run", "-c", configPath}, log),
		path: configPath,
	}
}

// Apply writes the config then starts the core, or restarts it if already up so
// the new user set takes effect (sing-box has no live membership API).
// Idempotent: when the rendered config is byte-identical to what is already
// applied and the core is running, the disruptive restart is skipped so a
// redundant resync (e.g. from a health-reconnect) does not bounce live
// connections. A real config change still triggers a supervised restart.
func (r *fileRunner) Apply(ctx context.Context, raw []byte) error {
	r.mu.Lock()
	unchanged := r.sup.Running() && bytes.Equal(raw, r.lastConfig)
	r.mu.Unlock()
	if unchanged {
		return nil
	}
	if err := os.WriteFile(r.path, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	r.mu.Lock()
	r.lastConfig = raw
	r.mu.Unlock()
	if r.sup.Running() {
		return r.sup.Restart(ctx)
	}
	return r.sup.Start(ctx)
}

func (r *fileRunner) Stop()         { r.sup.Stop() }
func (r *fileRunner) Running() bool { return r.sup.Running() }

// Logs returns recent core log lines captured by the supervisor.
func (r *fileRunner) Logs(limit int) []string { return r.sup.Logs(limit) }
