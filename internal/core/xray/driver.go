package xray

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/proc"
	"github.com/vortexui/vortexui/internal/domain"
)

// Options configures a Driver. Zero values fall back to sensible defaults.
type Options struct {
	BinPath        string        // path to the xray binary
	ConfigPath     string        // where the rendered config is written
	APIPort        int           // loopback port for Xray's gRPC API
	StatsInterval  time.Duration // how often to poll counters for deltas
	APIDialer      APIDialer     // injected for testing; nil → defaultDialer
	Logger         *slog.Logger
}

func (o *Options) withDefaults() {
	if o.APIPort == 0 {
		o.APIPort = 10085
	}
	if o.StatsInterval == 0 {
		o.StatsInterval = 10 * time.Second
	}
	if o.APIDialer == nil {
		o.APIDialer = defaultDialer
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
}

// Driver implements core.CoreDriver for Xray-core, orchestrating the config
// builder, the process supervisor, and the runtime gRPC API.
type Driver struct {
	opts    Options
	builder Builder
	proc    *proc.Supervisor
	log     *slog.Logger

	mu       sync.Mutex
	api      xrayAPI
	inbounds map[string]domain.Inbound // tag -> inbound, for runtime account building
}

var _ core.CoreDriver = (*Driver)(nil)

// New constructs a Driver. The process is not started until Start is called.
func New(opts Options) *Driver {
	opts.withDefaults()
	return &Driver{
		opts:    opts,
		builder: Builder{APIPort: opts.APIPort},
		proc:    proc.New(opts.BinPath, []string{"run", "-config", opts.ConfigPath}, opts.Logger),
		log:     opts.Logger,
	}
}

func (d *Driver) Type() domain.CoreType { return domain.CoreXray }

// Start renders the config, writes it, (re)starts the process, and connects the
// runtime API. Idempotent: calling it on a running core hot-applies the new
// config via a supervised restart.
func (d *Driver) Start(ctx context.Context, cfg *core.GeneratedConfig) error {
	raw, err := d.builder.Build(cfg)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	if err := os.WriteFile(d.opts.ConfigPath, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	// Cache the inbound shapes so runtime AddUser can build the right account.
	d.mu.Lock()
	d.inbounds = make(map[string]domain.Inbound, len(cfg.Inbounds))
	for _, in := range cfg.Inbounds {
		d.inbounds[in.Tag] = in
	}
	d.mu.Unlock()
	if d.proc.Running() {
		if err := d.proc.Restart(ctx); err != nil {
			return fmt.Errorf("restart: %w", err)
		}
	} else if err := d.proc.Start(ctx); err != nil {
		return fmt.Errorf("start: %w", err)
	}
	return d.connectAPI()
}

// Reload is an alias for Start since user changes go through the API and config
// changes require a supervised restart.
func (d *Driver) Reload(ctx context.Context, cfg *core.GeneratedConfig) error {
	return d.Start(ctx, cfg)
}

// Stop tears down the API connection and the process.
func (d *Driver) Stop(_ context.Context) error {
	d.mu.Lock()
	if d.api != nil {
		_ = d.api.Close()
		d.api = nil
	}
	d.mu.Unlock()
	d.proc.Stop()
	return nil
}

func (d *Driver) connectAPI() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.api != nil {
		return nil
	}
	api, err := d.opts.APIDialer(fmt.Sprintf("127.0.0.1:%d", d.opts.APIPort))
	if err != nil {
		return fmt.Errorf("connect xray api: %w", err)
	}
	d.api = api
	return nil
}

func (d *Driver) currentAPI() (xrayAPI, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.api == nil {
		return nil, fmt.Errorf("xray api not connected")
	}
	return d.api, nil
}

func (d *Driver) AddUser(ctx context.Context, inboundTag string, u *domain.User) error {
	api, err := d.currentAPI()
	if err != nil {
		return err
	}
	d.mu.Lock()
	in, ok := d.inbounds[inboundTag]
	d.mu.Unlock()
	if !ok {
		return fmt.Errorf("xray: unknown inbound %q; run a config sync first", inboundTag)
	}
	return api.AddUser(ctx, in, u)
}

func (d *Driver) RemoveUser(ctx context.Context, inboundTag, userID string) error {
	api, err := d.currentAPI()
	if err != nil {
		return err
	}
	// Stats email == user UUID, so userID is the removal key.
	return api.RemoveUser(ctx, inboundTag, userID)
}

// StreamTraffic polls Xray's per-user counters with reset=true on a ticker and
// emits the result as deltas. Because the query resets the counters, each tick
// yields exactly the bytes since the last tick — no absolute-counter bookkeeping
// and no double counting across panel/node restarts.
func (d *Driver) StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error) {
	api, err := d.currentAPI()
	if err != nil {
		return nil, err
	}
	out := make(chan domain.TrafficDelta, 256)
	go func() {
		defer close(out)
		ticker := time.NewTicker(d.opts.StatsInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				samples, err := api.QueryTraffic(ctx, true)
				if err != nil {
					d.log.Warn("xray stats query failed", "err", err)
					continue
				}
				now := time.Now()
				for _, s := range samples {
					if s.Up == 0 && s.Down == 0 {
						continue
					}
					uid, perr := uuid.Parse(s.Email)
					if perr != nil {
						continue // non-user counter (e.g. inbound-level); skip
					}
					select {
					case <-ctx.Done():
						return
					case out <- domain.TrafficDelta{UserID: uid, Up: s.Up, Down: s.Down, Timestamp: now}:
					}
				}
			}
		}
	}()
	return out, nil
}

func (d *Driver) Health(_ context.Context) (domain.NodeHealth, error) {
	return domain.NodeHealth{CoreRunning: d.proc.Running()}, nil
}

func (d *Driver) Version(_ context.Context) (string, error) {
	// TODO: shell out to `xray version` once and cache; placeholder for now.
	return "xray-core", nil
}
