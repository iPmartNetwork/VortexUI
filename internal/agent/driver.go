// Package agent builds core drivers for node agents and the panel local node.
package agent

import (
	"fmt"
	"log/slog"

	"github.com/vortexui/vortexui/internal/config"
	"github.com/vortexui/vortexui/internal/core"
	"github.com/vortexui/vortexui/internal/core/singbox"
	"github.com/vortexui/vortexui/internal/core/xray"
	"github.com/vortexui/vortexui/internal/domain"
)

// NewDriver builds a single or composite driver from node-agent config.
func NewDriver(cfg *config.Node, log *slog.Logger) (core.CoreDriver, error) {
	if cfg == nil {
		return nil, fmt.Errorf("agent driver: nil config")
	}
	enabled := cfg.EnabledCores
	if len(enabled) == 0 {
		enabled = []string{cfg.Core}
	}
	if len(enabled) <= 1 {
		return singleDriver(enabled[0], cfg, log)
	}
	drivers := make(map[domain.CoreType]core.CoreDriver, len(enabled))
	for _, ct := range enabled {
		d, err := singleDriver(ct, cfg, log)
		if err != nil {
			return nil, err
		}
		drivers[domain.CoreType(ct)] = d
	}
	return core.NewCompositeDriver(drivers)
}

// NewDriverFromPanel builds a driver for the in-process local node.
func NewDriverFromPanel(p *config.Panel, log *slog.Logger) (core.CoreDriver, error) {
	if p == nil {
		return nil, fmt.Errorf("agent driver: nil panel config")
	}
	enabled := p.EnabledCores
	if len(enabled) == 0 {
		enabled = []string{p.Core}
	}
	n := &config.Node{
		Core:            p.Core,
		EnabledCores:    enabled,
		CoreBin:         p.CoreBin,
		CoreConfig:      p.CoreConfig,
		APIPort:         p.CoreAPIPort,
		SingboxV2RayAPI: p.SingboxV2RayAPI,
	}
	return NewDriver(n, log)
}

func singleDriver(ct string, cfg *config.Node, log *slog.Logger) (core.CoreDriver, error) {
	switch ct {
	case "singbox":
		return singbox.New(singbox.Options{
			BinPath:      cfg.CoreBinFor(ct),
			ConfigPath:   cfg.CoreConfigFor(ct),
			APIPort:      cfg.APIPortFor(ct),
			OmitV2RayAPI: !cfg.SingboxV2RayAPIFor(),
			Logger:       log,
		}), nil
	case "xray", "":
		return xray.New(xray.Options{
			BinPath:    cfg.CoreBinFor(ct),
			ConfigPath: cfg.CoreConfigFor(ct),
			APIPort:    cfg.APIPortFor(ct),
			Logger:     log,
		}), nil
	default:
		return nil, fmt.Errorf("unknown core %q (want xray|singbox)", ct)
	}
}
