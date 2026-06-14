package singbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/domain"
)

// Options configures a Driver. Zero values fall back to sensible defaults.
type Options struct {
	BinPath       string
	ConfigPath    string
	APIPort       int
	StatsInterval time.Duration
	StatsDialer   func(addr string) (statsClient, error) // injected for tests
	Logger        *slog.Logger
}

func (o *Options) withDefaults() {
	if o.APIPort == 0 {
		o.APIPort = 9090
	}
	if o.StatsInterval == 0 {
		o.StatsInterval = 10 * time.Second
	}
	if o.StatsDialer == nil {
		o.StatsDialer = dialStats
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
}

// applier writes a rendered config and (re)starts the core. It is an interface
// so the driver's state/rebuild logic is testable without spawning a process.
type applier interface {
	Apply(ctx context.Context, raw []byte) error
	Stop()
	Running() bool
	Logs(limit int) []string
}

// Driver implements core.CoreDriver for sing-box. Because sing-box cannot add a
// user to a live inbound, the driver holds the full user set and rebuilds +
// reloads the config on every membership change.
type Driver struct {
	opts    Options
	builder Builder
	run     applier
	log     *slog.Logger

	mu       sync.Mutex
	inbounds map[string]domain.Inbound          // tag -> inbound
	users    map[string]map[string]*domain.User // tag -> email -> user
	stats    statsClient
}

var _ core.CoreDriver = (*Driver)(nil)

// New constructs a Driver. The process is not started until Start is called.
func New(opts Options) *Driver {
	opts.withDefaults()
	return &Driver{
		opts:     opts,
		builder:  Builder{APIPort: opts.APIPort},
		run:      newFileRunner(opts.BinPath, opts.ConfigPath, opts.Logger),
		log:      opts.Logger,
		inbounds: map[string]domain.Inbound{},
		users:    map[string]map[string]*domain.User{},
	}
}

func (d *Driver) Type() domain.CoreType { return domain.CoreSingbox }

// Start seeds the in-memory state from the config and applies it.
func (d *Driver) Start(ctx context.Context, cfg *core.GeneratedConfig) error {
	d.mu.Lock()
	d.inbounds = make(map[string]domain.Inbound, len(cfg.Inbounds))
	d.users = make(map[string]map[string]*domain.User, len(cfg.Inbounds))
	for _, in := range cfg.Inbounds {
		d.inbounds[in.Tag] = in
		set := map[string]*domain.User{}
		for _, u := range cfg.UsersByInbound[in.Tag] {
			set[u.ID.String()] = u
		}
		d.users[in.Tag] = set
	}
	err := d.applyLocked(ctx)
	d.mu.Unlock()
	if err != nil {
		return err
	}
	return d.connectStats()
}

// Reload is identical to Start for sing-box (state is replaced wholesale).
func (d *Driver) Reload(ctx context.Context, cfg *core.GeneratedConfig) error {
	return d.Start(ctx, cfg)
}

// AddUser adds the user to its inbound's set and reloads. No live API needed.
func (d *Driver) AddUser(ctx context.Context, inboundTag string, u *domain.User) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.inbounds[inboundTag]; !ok {
		return fmt.Errorf("singbox: unknown inbound %q; run a config sync first", inboundTag)
	}
	if d.users[inboundTag] == nil {
		d.users[inboundTag] = map[string]*domain.User{}
	}
	d.users[inboundTag][u.ID.String()] = u
	return d.applyLocked(ctx)
}

// RemoveUser removes the user (by email/UUID) from its inbound and reloads.
func (d *Driver) RemoveUser(ctx context.Context, inboundTag, email string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if set := d.users[inboundTag]; set != nil {
		delete(set, email)
	}
	return d.applyLocked(ctx)
}

// applyLocked rebuilds the config from current state and applies it. Caller holds d.mu.
func (d *Driver) applyLocked(ctx context.Context) error {
	cfg := &core.GeneratedConfig{UsersByInbound: map[string][]*domain.User{}}
	for tag, in := range d.inbounds {
		cfg.Inbounds = append(cfg.Inbounds, in)
		for _, u := range d.users[tag] {
			cfg.UsersByInbound[tag] = append(cfg.UsersByInbound[tag], u)
		}
	}
	raw, err := d.builder.Build(cfg)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}
	return d.run.Apply(ctx, raw)
}

func (d *Driver) connectStats() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stats != nil {
		return nil
	}
	sc, err := d.opts.StatsDialer(fmt.Sprintf("127.0.0.1:%d", d.opts.APIPort))
	if err != nil {
		return fmt.Errorf("connect singbox stats: %w", err)
	}
	d.stats = sc
	return nil
}

// Stop tears down stats and the process.
func (d *Driver) Stop(_ context.Context) error {
	d.mu.Lock()
	if d.stats != nil {
		_ = d.stats.Close()
		d.stats = nil
	}
	d.mu.Unlock()
	d.run.Stop()
	return nil
}

// StreamTraffic polls the V2Ray API counters with reset=true and emits deltas,
// mirroring the Xray driver's accounting model.
func (d *Driver) StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error) {
	d.mu.Lock()
	sc := d.stats
	d.mu.Unlock()
	if sc == nil {
		return nil, fmt.Errorf("singbox stats not connected")
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
				samples, err := sc.QueryTraffic(ctx, true)
				if err != nil {
					d.log.Warn("singbox stats query failed", "err", err)
					continue
				}
				now := time.Now()
				for _, s := range samples {
					if s.Up == 0 && s.Down == 0 {
						continue
					}
					uid, perr := uuid.Parse(s.Email)
					if perr != nil {
						continue
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
	return domain.NodeHealth{CoreRunning: d.run.Running()}, nil
}

// OnlineStats is not supported by the sing-box V2Ray API (it has no
// GetAllOnlineUsers equivalent), so it reports an empty set rather than erroring.
func (d *Driver) OnlineStats(context.Context) (map[string]int, error) {
	return map[string]int{}, nil
}

// OnlineIPList is likewise unsupported by sing-box; reports an empty set.
func (d *Driver) OnlineIPList(context.Context, string) (map[string]int64, error) {
	return map[string]int64{}, nil
}

// Logs returns the most recent core log lines captured by the supervisor.
func (d *Driver) Logs(_ context.Context, limit int) ([]string, error) {
	return d.run.Logs(limit), nil
}

func (d *Driver) Version(_ context.Context) (string, error) { return "sing-box", nil }
