package xray

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/hostmetrics"
	"github.com/vortexui/vortexui/internal/core/proc"
	"github.com/vortexui/vortexui/internal/domain"
)

// Options configures a Driver. Zero values fall back to sensible defaults.
type Options struct {
	BinPath       string        // path to the xray binary
	ConfigPath    string        // where the rendered config is written
	APIPort       int           // loopback port for Xray's gRPC API
	StatsInterval time.Duration // how often to poll counters for deltas
	AssetDir      string        // geoip.dat / geosite.dat location (XRAY_LOCATION_ASSET)
	APIDialer     APIDialer     // injected for testing; nil → defaultDialer
	Logger        *slog.Logger
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
	if o.AssetDir == "" {
		if env := os.Getenv("XRAY_LOCATION_ASSET"); env != "" {
			o.AssetDir = env
		} else {
			o.AssetDir = "/usr/local/share/xray"
		}
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

	mu         sync.Mutex
	api        xrayAPI
	inbounds   map[string]domain.Inbound // tag -> inbound, for runtime account building
	lastConfig []byte                    // last config bytes actually written/applied; guards redundant restarts
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
// runtime API. Idempotent: when the rendered config is byte-identical to what is
// already applied and the core is running, the disruptive restart is skipped so a
// redundant resync (e.g. from a health-reconnect) does not bounce live
// connections. A real config change still triggers a supervised restart.
func (d *Driver) Start(ctx context.Context, cfg *core.GeneratedConfig) error {
	raw, err := d.builder.Build(cfg)
	if err != nil {
		return fmt.Errorf("build config: %w", err)
	}

	d.mu.Lock()
	// Always refresh the inbound-shape cache so runtime AddUser stays correct.
	d.inbounds = make(map[string]domain.Inbound, len(cfg.Inbounds))
	for _, in := range cfg.Inbounds {
		d.inbounds[in.Tag] = in
	}
	unchanged := d.proc.Running() && bytes.Equal(raw, d.lastConfig)
	d.mu.Unlock()

	if unchanged {
		// Config matches the running core exactly — do not restart; just make
		// sure the runtime API link is up.
		return d.connectAPI()
	}

	if err := os.WriteFile(d.opts.ConfigPath, raw, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	d.mu.Lock()
	d.lastConfig = raw
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
	m := hostmetrics.Read()
	return domain.NodeHealth{
		CoreRunning: d.proc.Running(),
		CPUPercent:  m.CPU,
		MemPercent:  m.Mem,
		DiskPercent: m.Disk,
	}, nil
}

// OnlineStats reports live connection counts per user via the runtime API.
func (d *Driver) OnlineStats(ctx context.Context) (map[string]int, error) {
	api, err := d.currentAPI()
	if err != nil {
		return nil, err
	}
	return api.OnlineUsers(ctx)
}

// OnlineIPList reports the distinct source IPs currently online for one user.
func (d *Driver) OnlineIPList(ctx context.Context, email string) (map[string]int64, error) {
	api, err := d.currentAPI()
	if err != nil {
		return nil, err
	}
	return api.OnlineIPs(ctx, email)
}

// UpdateGeoAssets downloads geoip.dat and geosite.dat into the asset directory
// and restarts the core so the new routing data takes effect. Empty URLs skip
// that file. Returns the byte size written for each. Writes are atomic (temp +
// rename) so a failed download never corrupts the running asset.
func (d *Driver) UpdateGeoAssets(ctx context.Context, geoipURL, geositeURL string) (geoip, geosite int64, err error) {
	if err := os.MkdirAll(d.opts.AssetDir, 0o755); err != nil {
		return 0, 0, fmt.Errorf("asset dir: %w", err)
	}
	if geoipURL != "" {
		if geoip, err = downloadFile(ctx, geoipURL, filepath.Join(d.opts.AssetDir, "geoip.dat")); err != nil {
			return 0, 0, fmt.Errorf("geoip: %w", err)
		}
	}
	if geositeURL != "" {
		if geosite, err = downloadFile(ctx, geositeURL, filepath.Join(d.opts.AssetDir, "geosite.dat")); err != nil {
			return geoip, 0, fmt.Errorf("geosite: %w", err)
		}
	}
	// Restart so xray reloads the dat files (they are read at startup).
	if d.proc.Running() {
		if err := d.proc.Restart(ctx); err != nil {
			return geoip, geosite, fmt.Errorf("restart: %w", err)
		}
	}
	return geoip, geosite, nil
}

// downloadFile fetches url into dst atomically, returning the byte count.
func downloadFile(ctx context.Context, url, dst string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".geo-*")
	if err != nil {
		return 0, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	n, err := io.Copy(tmp, resp.Body)
	_ = tmp.Close()
	if err != nil {
		return 0, err
	}
	if err := os.Rename(tmpName, dst); err != nil {
		return 0, err
	}
	return n, nil
}

// Logs returns the most recent core log lines captured by the supervisor.
func (d *Driver) Logs(_ context.Context, limit int) ([]string, error) {
	return d.proc.Logs(limit), nil
}

func (d *Driver) Version(_ context.Context) (string, error) {
	out, err := exec.Command(d.opts.BinPath, "version").Output()
	if err != nil {
		return "xray-core", nil
	}
	// First line: "Xray 1.8.24 (Xray, Penetrates Everything.) ..."
	line := strings.SplitN(string(out), "\n", 2)[0]
	return strings.TrimSpace(line), nil
}
