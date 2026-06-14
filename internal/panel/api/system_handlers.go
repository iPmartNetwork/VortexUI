package api

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/labstack/echo/v4"
)

var panelStartTime = time.Now()

// SystemInfo is returned by GET /api/system and powers the dashboard system card.
type SystemInfo struct {
	Uptime     int64  `json:"uptime_seconds"`
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	GoVersion  string `json:"go_version"`
	Goroutines int    `json:"goroutines"`
	MemAlloc   uint64 `json:"mem_alloc_bytes"`
	MemSys     uint64 `json:"mem_sys_bytes"`
	Hostname   string `json:"hostname"`
}

// GetSystem reports live panel process information for the dashboard system card.
func (h *Handlers) GetSystem(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	host, _ := os.Hostname()
	return c.JSON(http.StatusOK, SystemInfo{
		Uptime:     int64(time.Since(panelStartTime).Seconds()),
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GoVersion:  runtime.Version(),
		Goroutines: runtime.NumGoroutine(),
		MemAlloc:   m.Alloc,
		MemSys:     m.Sys,
		Hostname:   host,
	})
}
