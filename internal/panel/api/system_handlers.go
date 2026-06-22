package api

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/vortexui/vortexui/internal/core/hostmetrics"
)

var panelStartTime = time.Now()

// SystemInfo is returned by GET /api/system and powers the dashboard system card.
type SystemInfo struct {
	Uptime     int64   `json:"uptime_seconds"`
	OS         string  `json:"os"`
	Arch       string  `json:"arch"`
	GoVersion  string  `json:"go_version"`
	Goroutines int     `json:"goroutines"`
	MemAlloc   uint64  `json:"mem_alloc_bytes"`
	MemSys     uint64  `json:"mem_sys_bytes"`
	Hostname   string  `json:"hostname"`
	CPUPercent float64 `json:"cpu_percent"`  // host CPU utilisation
	MemPercent float64 `json:"mem_percent"`  // host memory utilisation
	DiskPercent float64 `json:"disk_percent"` // host disk utilisation
}

// VersionInfo is returned by GET /api/version so the UI can show the actually
// running build instead of a hardcoded constant.
type VersionInfo struct {
	Version string `json:"version"`
}

// GetVersion reports the panel build version (injected at build time via
// -ldflags "-X main.version=..."). The UI footer consumes it.
func (h *Handlers) GetVersion(c echo.Context) error {
	return c.JSON(http.StatusOK, VersionInfo{Version: h.Version})
}

// GetSystem reports live panel process information for the dashboard system card.
func (h *Handlers) GetSystem(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	host, _ := os.Hostname()
	hm := hostmetrics.Read()
	return c.JSON(http.StatusOK, SystemInfo{
		Uptime:      int64(time.Since(panelStartTime).Seconds()),
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemAlloc:    m.Alloc,
		MemSys:      m.Sys,
		Hostname:    host,
		CPUPercent:  hm.CPU,
		MemPercent:  hm.Mem,
		DiskPercent: hm.Disk,
	})
}
