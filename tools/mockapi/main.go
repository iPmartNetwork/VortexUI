//go:build ignore

// Mock API server — serves fake data for all panel endpoints so the frontend
// can be developed and demoed without a real database/redis/node.
// Run: go run tools/mockapi/main.go
package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var start = time.Now()

func j(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func main() {
	mux := http.NewServeMux()

	// --- auth ---
	mux.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]string{"token": "mock-jwt-token-for-dev"})
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]string{"status": "ok"})
	})

	// --- system ---
	mux.HandleFunc("/api/system", func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		host, _ := os.Hostname()
		j(w, map[string]any{
			"uptime_seconds":  int(time.Since(start).Seconds()),
			"os":             runtime.GOOS,
			"arch":           runtime.GOARCH,
			"go_version":     runtime.Version(),
			"goroutines":     runtime.NumGoroutine(),
			"mem_alloc_bytes": m.Alloc,
			"mem_sys_bytes":   m.Sys,
			"hostname":       host,
		})
	})

	// --- overview ---
	mux.HandleFunc("/api/overview", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{
			"users": map[string]any{
				"total":      127,
				"total_used": 483921061888 + rand.Int63n(10000000),
				"by_status":  map[string]int{"active": 89, "limited": 14, "expired": 8, "disabled": 12, "on_hold": 4},
			},
			"nodes": map[string]any{
				"total":  3,
				"online": 2,
			},
		})
	})

	// --- nodes ---
	nodes := []map[string]any{
		{
			"id": "n1-uuid", "name": "Germany-1", "address": "185.220.101.34:50051",
			"core": "xray", "status": "connected", "usage_ratio": 1.0,
			"last_seen": time.Now().Add(-5 * time.Second).Format(time.RFC3339),
			"core_version": "1.8.24", "agent_version": "0.1.0",
			"health": map[string]any{"cpu_percent": 23.5 + rand.Float64()*5, "mem_percent": 61.2 + rand.Float64()*3, "disk_percent": 44.0, "core_running": true, "connections": 34 + rand.Intn(10)},
		},
		{
			"id": "n2-uuid", "name": "Netherlands-1", "address": "91.132.145.78:50051",
			"core": "singbox", "status": "connected", "usage_ratio": 1.5,
			"last_seen": time.Now().Add(-3 * time.Second).Format(time.RFC3339),
			"core_version": "1.9.0", "agent_version": "0.1.0",
			"health": map[string]any{"cpu_percent": 45.1 + rand.Float64()*8, "mem_percent": 72.4 + rand.Float64()*4, "disk_percent": 55.0, "core_running": true, "connections": 52 + rand.Intn(15)},
		},
		{
			"id": "n3-uuid", "name": "Finland-1", "address": "95.216.1.100:50051",
			"core": "xray", "status": "disconnected", "usage_ratio": 1.0,
			"last_seen": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"core_version": "1.8.23", "agent_version": "0.1.0",
			"health": map[string]any{"cpu_percent": 0.0, "mem_percent": 0.0, "disk_percent": 38.0, "core_running": false, "connections": 0},
		},
	}
	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		// Randomize live values on each request
		for i := range nodes {
			if h, ok := nodes[i]["health"].(map[string]any); ok && h["core_running"] == true {
				h["cpu_percent"] = 20 + rand.Float64()*40
				h["mem_percent"] = 50 + rand.Float64()*30
				h["connections"] = 20 + rand.Intn(50)
			}
			if nodes[i]["status"] == "connected" {
				nodes[i]["last_seen"] = time.Now().Add(-time.Duration(rand.Intn(10)) * time.Second).Format(time.RFC3339)
			}
		}
		j(w, map[string]any{"nodes": nodes})
	})

	// --- users ---
	users := make([]map[string]any, 0)
	statuses := []string{"active", "active", "active", "limited", "expired", "disabled", "on_hold"}
	for i := 1; i <= 25; i++ {
		users = append(users, map[string]any{
			"id": fmt.Sprintf("u%d-uuid", i), "username": fmt.Sprintf("user_%d", i),
			"status": statuses[rand.Intn(len(statuses))], "note": "",
			"data_limit": int64(10+rand.Intn(90)) * 1024 * 1024 * 1024,
			"used_traffic": int64(rand.Intn(80)) * 1024 * 1024 * 1024,
			"expire_at": time.Now().Add(time.Duration(rand.Intn(60)-30) * 24 * time.Hour).Format(time.RFC3339),
			"reset_strategy": "monthly", "device_limit": 3,
			"proxies": map[string]string{"vmess_uuid": "vm-uuid", "vless_uuid": "vl-uuid", "trojan_password": "trpw", "ss_password": "sspw", "ss_method": "aes-128-gcm"},
			"sub_token":  fmt.Sprintf("tok-%d", i),
			"created_at": time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Format(time.RFC3339),
		})
	}
	mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"users": users, "total": len(users)})
	})

	// --- user sub ---
	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Route: /api/users/{id}/usage
		if len(path) > 16 && strings.HasSuffix(path, "/usage") {
			now := time.Now()
			points := make([]map[string]any, 7)
			for i := range points {
				points[i] = map[string]any{
					"time": now.AddDate(0, 0, -(6 - i)).Format(time.RFC3339),
					"up":   int64(rand.Intn(500)) * 1024 * 1024,
					"down": int64(rand.Intn(2000)) * 1024 * 1024,
				}
			}
			j(w, map[string]any{"points": points})
			return
		}
		// Route: /api/users/{id}/online
		if len(path) > 16 && strings.HasSuffix(path, "/online") {
			j(w, map[string]any{
				"live_connections": 2 + rand.Intn(5),
				"live_tracking":    true,
				"active_devices":   1 + rand.Intn(3),
				"device_tracking":  true,
			})
			return
		}
		// Route: /api/users/{id}/sub
		if len(path) > 16 && strings.HasSuffix(path, "/sub") {
			j(w, map[string]any{
				"token":             "mock-sub-token",
				"subscription_url":  "http://localhost:8080/sub/mock-sub-token",
				"subscription_path": "/sub/mock-sub-token",
				"formats": map[string]string{
					"auto":    "http://localhost:8080/sub/mock-sub-token",
					"clash":   "http://localhost:8080/sub/mock-sub-token?format=clash",
					"singbox": "http://localhost:8080/sub/mock-sub-token?format=singbox",
					"base64":  "http://localhost:8080/sub/mock-sub-token?format=base64",
				},
				"links": []string{"vless://uuid@1.2.3.4:443?type=ws&security=tls#Germany", "trojan://pass@5.6.7.8:443#Netherlands"},
			})
			return
		}
		// Route: /api/users/{id}/reset or /api/users/{id}/revoke-sub (POST)
		if r.Method == http.MethodPost {
			j(w, map[string]any{"user": users[0]})
			return
		}
		// Route: /api/users/{id} (GET single user)
		j(w, users[0])
	})

	// --- inbounds ---
	mux.HandleFunc("/api/inbounds", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"inbounds": []map[string]any{
			{"id": "ib1", "node_id": "n1-uuid", "tag": "vless-ws", "protocol": "vless", "port": 443, "network": "ws", "security": "tls", "enabled": true},
			{"id": "ib2", "node_id": "n1-uuid", "tag": "trojan-tcp", "protocol": "trojan", "port": 8443, "network": "tcp", "security": "tls", "enabled": true},
			{"id": "ib3", "node_id": "n2-uuid", "tag": "vless-reality", "protocol": "vless", "port": 443, "network": "tcp", "security": "reality", "enabled": true},
		}})
	})

	// --- outbounds / routing / balancers ---
	mux.HandleFunc("/api/outbounds", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"outbounds": []map[string]any{
			{"id": "ob1", "node_id": "n1-uuid", "tag": "direct", "protocol": "freedom", "enabled": true},
			{"id": "ob2", "node_id": "n1-uuid", "tag": "blocked", "protocol": "blackhole", "enabled": true},
			{"id": "ob3", "node_id": "n1-uuid", "tag": "proxy-nl", "protocol": "vless", "address": "91.132.145.78", "port": 443, "enabled": true},
		}})
	})
	mux.HandleFunc("/api/routing", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"routing": []map[string]any{
			{"id": "r1", "node_id": "n1-uuid", "priority": 1, "name": "block-ads", "domains": []string{"geosite:category-ads"}, "outbound_tag": "blocked", "enabled": true},
			{"id": "r2", "node_id": "n1-uuid", "priority": 10, "name": "proxy-all", "inbound_tags": []string{"vless-ws"}, "outbound_tag": "proxy-nl", "enabled": true},
		}})
	})
	mux.HandleFunc("/api/balancers", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"balancers": []map[string]any{
			{"id": "b1", "node_id": "n1-uuid", "tag": "auto-lb", "selectors": []string{"proxy-"}, "strategy": "leastPing", "observe": true, "probe_url": "https://www.gstatic.com/generate_204", "probe_interval": "10s", "enabled": true},
		}})
	})

	// --- logs ---
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		entries := []map[string]any{
			{"time": time.Now().Add(-2 * time.Minute).Format(time.RFC3339), "level": 0, "message": "panel started", "attrs": map[string]any{"addr": ":8080"}},
			{"time": time.Now().Add(-90 * time.Second).Format(time.RFC3339), "level": 0, "message": "node Germany-1 connected"},
			{"time": time.Now().Add(-60 * time.Second).Format(time.RFC3339), "level": 0, "message": "node Netherlands-1 connected"},
			{"time": time.Now().Add(-45 * time.Second).Format(time.RFC3339), "level": 4, "message": "node Finland-1 health check failed", "attrs": map[string]any{"err": "connection refused"}},
			{"time": time.Now().Add(-30 * time.Second).Format(time.RFC3339), "level": 0, "message": "enforced user limit", "attrs": map[string]any{"user": "user_7", "status": "limited"}},
			{"time": time.Now().Add(-10 * time.Second).Format(time.RFC3339), "level": 0, "message": "traffic flush", "attrs": map[string]any{"users": 12, "bytes": 48291840}},
		}
		j(w, map[string]any{"entries": entries})
	})

	// --- admins ---
	mux.HandleFunc("/api/admins", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"admins": []map[string]any{
			{"id": "a1", "username": "root", "sudo": true, "totp_enabled": true, "created_at": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339)},
		}})
	})

	// --- node logs ---
	mux.HandleFunc("/api/nodes/", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"lines": []string{
			"2026/06/14 12:00:01 [Info] Xray 1.8.24 started",
			"2026/06/14 12:00:01 [Info] Reading config: /etc/vortex/core.json",
			"2026/06/14 12:00:02 [Info] Listening on 0.0.0.0:443 (vless-ws)",
			"2026/06/14 12:00:02 [Info] Listening on 0.0.0.0:8443 (trojan-tcp)",
			"2026/06/14 12:01:15 [Info] user_3 connected from 78.46.12.34",
			"2026/06/14 12:02:30 [Warning] user_7 exceeded data limit",
		}})
	})

	// --- reality keygen ---
	mux.HandleFunc("/api/reality/keypair", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"private_key": "qBxz7xKl3eMkPbSfGQd5Rw4kN1uX9v2Z", "public_key": "Aj8xKl3eMkPbSfGQd5Rw4kN1uX9v2ZqB", "short_id": "a1b2c3d4"})
	})

	// --- backup ---
	mux.HandleFunc("/api/backup", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"version": 1, "exported_at": time.Now().Format(time.RFC3339), "nodes": nodes, "users": users[:3], "inbounds": []any{}, "outbounds": []any{}, "routing": []any{}, "balancers": []any{}, "bindings": []any{}})
	})

	// --- account ---
	mux.HandleFunc("/api/account/", func(w http.ResponseWriter, r *http.Request) {
		j(w, map[string]any{"ok": true})
	})

	// --- catch-all for POST/PUT/DELETE ---
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			j(w, map[string]any{"ok": true})
			return
		}
		http.NotFound(w, r)
	})

	fmt.Println("🚀 Mock API running on http://localhost:8080")
	fmt.Println("   Frontend should proxy /api → here")
	http.ListenAndServe(":8080", mux)
}
