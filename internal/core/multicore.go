package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/vortexui/vortexui/internal/domain"
)

// CompositeDriver runs several CoreDriver instances on one node agent, routing
// each inbound (by tag) to the engine declared on Inbound.Core.
type CompositeDriver struct {
	drivers map[domain.CoreType]CoreDriver

	mu      sync.RWMutex
	tagCore map[string]domain.CoreType
	lastCfg *GeneratedConfig
}

var _ CoreDriver = (*CompositeDriver)(nil)

// NewCompositeDriver wraps one driver per enabled core type. The map must hold
// at least one entry; callers typically pass xray and/or singbox.
func NewCompositeDriver(drivers map[domain.CoreType]CoreDriver) (*CompositeDriver, error) {
	if len(drivers) == 0 {
		return nil, fmt.Errorf("composite driver: no engines configured")
	}
	cp := make(map[domain.CoreType]CoreDriver, len(drivers))
	for k, v := range drivers {
		if v == nil {
			return nil, fmt.Errorf("composite driver: nil driver for %s", k)
		}
		cp[k] = v
	}
	return &CompositeDriver{drivers: cp, tagCore: make(map[string]domain.CoreType)}, nil
}

func (c *CompositeDriver) Type() domain.CoreType { return domain.CoreMulti }

func (c *CompositeDriver) Start(ctx context.Context, cfg *GeneratedConfig) error {
	if cfg == nil {
		return fmt.Errorf("composite driver: nil config")
	}
	split, tags := splitConfigByCore(cfg)
	c.mu.Lock()
	c.tagCore = tags
	c.lastCfg = cfg
	c.mu.Unlock()
	for ct, sub := range split {
		if err := c.drivers[ct].Start(ctx, sub); err != nil {
			return fmt.Errorf("start %s: %w", ct, err)
		}
	}
	return nil
}

func (c *CompositeDriver) Stop(ctx context.Context) error {
	var errs []error
	for ct, d := range c.drivers {
		if err := d.Stop(ctx); err != nil {
			errs = append(errs, fmt.Errorf("stop %s: %w", ct, err))
		}
	}
	return errorsJoin(errs)
}

func (c *CompositeDriver) Reload(ctx context.Context, cfg *GeneratedConfig) error {
	if cfg == nil {
		c.mu.RLock()
		cfg = c.lastCfg
		c.mu.RUnlock()
		if cfg == nil {
			return fmt.Errorf("composite driver: no config to reload")
		}
	}
	return c.Start(ctx, cfg)
}

func (c *CompositeDriver) AddUser(ctx context.Context, inboundTag string, u *domain.User) error {
	d, err := c.driverForTag(inboundTag)
	if err != nil {
		return err
	}
	return d.AddUser(ctx, inboundTag, u)
}

func (c *CompositeDriver) RemoveUser(ctx context.Context, inboundTag, userID string) error {
	d, err := c.driverForTag(inboundTag)
	if err != nil {
		return err
	}
	return d.RemoveUser(ctx, inboundTag, userID)
}

func (c *CompositeDriver) StreamTraffic(ctx context.Context) (<-chan domain.TrafficDelta, error) {
	out := make(chan domain.TrafficDelta, 32)
	var wg sync.WaitGroup
	for ct, d := range c.drivers {
		ch, err := d.StreamTraffic(ctx)
		if err != nil {
			close(out)
			return nil, fmt.Errorf("stream traffic %s: %w", ct, err)
		}
		wg.Add(1)
		go func(ch <-chan domain.TrafficDelta) {
			defer wg.Done()
			for d := range ch {
				select {
				case <-ctx.Done():
					return
				case out <- d:
				}
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out, nil
}

func (c *CompositeDriver) OnlineStats(ctx context.Context) (map[string]int, error) {
	merged := make(map[string]int)
	for ct, d := range c.drivers {
		stats, err := d.OnlineStats(ctx)
		if err != nil {
			return nil, fmt.Errorf("online stats %s: %w", ct, err)
		}
		for email, n := range stats {
			merged[email] += n
		}
	}
	return merged, nil
}

func (c *CompositeDriver) OnlineIPList(ctx context.Context, email string) (map[string]int64, error) {
	merged := make(map[string]int64)
	for ct, d := range c.drivers {
		ips, err := d.OnlineIPList(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("online ips %s: %w", ct, err)
		}
		for ip, ts := range ips {
			if prev, ok := merged[ip]; !ok || ts > prev {
				merged[ip] = ts
			}
		}
	}
	return merged, nil
}

func (c *CompositeDriver) UpdateGeoAssets(ctx context.Context, geoipURL, geositeURL string) (geoip, geosite int64, err error) {
	for ct, d := range c.drivers {
		g, s, e := d.UpdateGeoAssets(ctx, geoipURL, geositeURL)
		if e != nil {
			return 0, 0, fmt.Errorf("update geo %s: %w", ct, e)
		}
		geoip += g
		geosite += s
	}
	return geoip, geosite, nil
}

func (c *CompositeDriver) Logs(ctx context.Context, limit int) ([]string, error) {
	var lines []string
	for ct, d := range c.drivers {
		part, err := d.Logs(ctx, limit)
		if err != nil {
			return nil, fmt.Errorf("logs %s: %w", ct, err)
		}
		prefix := string(ct) + ": "
		for _, ln := range part {
			lines = append(lines, prefix+ln)
		}
	}
	if limit > 0 && len(lines) > limit {
		lines = lines[len(lines)-limit:]
	}
	return lines, nil
}

func (c *CompositeDriver) Health(ctx context.Context) (domain.NodeHealth, error) {
	var agg domain.NodeHealth
	first := true
	allRunning := true
	for _, d := range c.drivers {
		h, err := d.Health(ctx)
		if err != nil {
			return h, err
		}
		if !h.CoreRunning {
			allRunning = false
		}
		agg.Connections += h.Connections
		if first {
			agg = h
			first = false
		} else {
			agg.CPUPercent = maxFloat(agg.CPUPercent, h.CPUPercent)
			agg.MemPercent = maxFloat(agg.MemPercent, h.MemPercent)
			agg.DiskPercent = maxFloat(agg.DiskPercent, h.DiskPercent)
		}
	}
	agg.CoreRunning = allRunning
	return agg, nil
}

func (c *CompositeDriver) Version(ctx context.Context) (string, error) {
	parts := make([]string, 0, len(c.drivers))
	for ct, d := range c.drivers {
		ver, err := d.Version(ctx)
		if err != nil {
			return "", fmt.Errorf("version %s: %w", ct, err)
		}
		if ver != "" {
			parts = append(parts, string(ct)+"/"+ver)
		}
	}
	return strings.Join(parts, " + "), nil
}

func (c *CompositeDriver) driverForTag(tag string) (CoreDriver, error) {
	c.mu.RLock()
	ct, ok := c.tagCore[tag]
	c.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("composite driver: unknown inbound tag %q", tag)
	}
	d := c.drivers[ct]
	if d == nil {
		return nil, fmt.Errorf("composite driver: no driver for core %s (tag %q)", ct, tag)
	}
	return d, nil
}

func splitConfigByCore(cfg *GeneratedConfig) (map[domain.CoreType]*GeneratedConfig, map[string]domain.CoreType) {
	tags := make(map[string]domain.CoreType, len(cfg.Inbounds))
	out := make(map[domain.CoreType]*GeneratedConfig)

	newSub := func(ct domain.CoreType) *GeneratedConfig {
		if sub, ok := out[ct]; ok {
			return sub
		}
		sub := &GeneratedConfig{
			LogLevel:       cfg.LogLevel,
			UsersByInbound: make(map[string][]*domain.User),
			WireGuardPeers: make(map[string][]domain.WireGuardPeer),
			Outbounds:      cfg.Outbounds,
			Routing:        cfg.Routing,
			Balancers:      cfg.Balancers,
		}
		out[ct] = sub
		return sub
	}

	for _, in := range cfg.Inbounds {
		ct := in.Core
		if ct == "" {
			ct = domain.CoreXray
		}
		tags[in.Tag] = ct
		sub := newSub(ct)
		sub.Inbounds = append(sub.Inbounds, in)
		if users, ok := cfg.UsersByInbound[in.Tag]; ok {
			sub.UsersByInbound[in.Tag] = users
		}
		if peers, ok := cfg.WireGuardPeers[in.Tag]; ok {
			sub.WireGuardPeers[in.Tag] = peers
		}
	}
	return out, tags
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func errorsJoin(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	msgs := make([]string, len(errs))
	for i, e := range errs {
		msgs[i] = e.Error()
	}
	return fmt.Errorf("%s", strings.Join(msgs, "; "))
}
