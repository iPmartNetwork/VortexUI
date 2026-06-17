// Package metrics exposes panel metrics in Prometheus format at /metrics.
// Grafana can scrape this endpoint to build real-time dashboards.
package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// Collector holds panel-wide counters that the metrics endpoint exposes.
type Collector struct {
	ActiveUsers    atomic.Int64
	TotalUsers     atomic.Int64
	OnlineNodes    atomic.Int64
	TotalNodes     atomic.Int64
	TotalTraffic   atomic.Int64  // bytes
	ActiveConns    atomic.Int64
	UptimeStart    time.Time
}

// New creates a collector with uptime starting now.
func New() *Collector {
	return &Collector{UptimeStart: time.Now()}
}

// Handler returns an HTTP handler that serves Prometheus-format metrics.
func (c *Collector) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		uptime := time.Since(c.UptimeStart).Seconds()

		fmt.Fprintf(w, "# HELP vortexui_uptime_seconds Panel uptime in seconds.\n")
		fmt.Fprintf(w, "# TYPE vortexui_uptime_seconds gauge\n")
		fmt.Fprintf(w, "vortexui_uptime_seconds %.0f\n\n", uptime)

		fmt.Fprintf(w, "# HELP vortexui_users_total Total registered users.\n")
		fmt.Fprintf(w, "# TYPE vortexui_users_total gauge\n")
		fmt.Fprintf(w, "vortexui_users_total %d\n\n", c.TotalUsers.Load())

		fmt.Fprintf(w, "# HELP vortexui_users_active Currently active users.\n")
		fmt.Fprintf(w, "# TYPE vortexui_users_active gauge\n")
		fmt.Fprintf(w, "vortexui_users_active %d\n\n", c.ActiveUsers.Load())

		fmt.Fprintf(w, "# HELP vortexui_nodes_total Total registered nodes.\n")
		fmt.Fprintf(w, "# TYPE vortexui_nodes_total gauge\n")
		fmt.Fprintf(w, "vortexui_nodes_total %d\n\n", c.TotalNodes.Load())

		fmt.Fprintf(w, "# HELP vortexui_nodes_online Currently online nodes.\n")
		fmt.Fprintf(w, "# TYPE vortexui_nodes_online gauge\n")
		fmt.Fprintf(w, "vortexui_nodes_online %d\n\n", c.OnlineNodes.Load())

		fmt.Fprintf(w, "# HELP vortexui_traffic_bytes_total Total traffic processed (bytes).\n")
		fmt.Fprintf(w, "# TYPE vortexui_traffic_bytes_total counter\n")
		fmt.Fprintf(w, "vortexui_traffic_bytes_total %d\n\n", c.TotalTraffic.Load())

		fmt.Fprintf(w, "# HELP vortexui_connections_active Current live connections.\n")
		fmt.Fprintf(w, "# TYPE vortexui_connections_active gauge\n")
		fmt.Fprintf(w, "vortexui_connections_active %d\n\n", c.ActiveConns.Load())
	}
}
