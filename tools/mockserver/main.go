// Command mockserver is a DEV-ONLY in-memory stand-in for the VortexUI panel
// API, so the real frontend can be demoed without Postgres. It serves the subset
// of endpoints the web UI calls, with state held in maps. Not for production.
package main

import (
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"
)

type creds struct {
	VmessUUID   string `json:"vmess_uuid"`
	VlessUUID   string `json:"vless_uuid"`
	TrojanPass  string `json:"trojan_password"`
	SSPass      string `json:"ss_password"`
	SSMethod    string `json:"ss_method"`
}

type user struct {
	ID            string    `json:"id"`
	Username      string    `json:"username"`
	Status        string    `json:"status"`
	Note          string    `json:"note"`
	DataLimit     int64     `json:"data_limit"`
	UsedTraffic   int64     `json:"used_traffic"`
	ExpireAt      *string   `json:"expire_at"`
	ResetStrategy string    `json:"reset_strategy"`
	DeviceLimit   int       `json:"device_limit"`
	Proxies       creds     `json:"proxies"`
	SubToken      string    `json:"sub_token"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type health struct {
	CPU         float64 `json:"cpu_percent"`
	Mem         float64 `json:"mem_percent"`
	Disk        float64 `json:"disk_percent"`
	CoreRunning bool    `json:"core_running"`
	Connections int     `json:"connections"`
}

type node struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Core       string    `json:"core"`
	Status     string    `json:"status"`
	UsageRatio float64   `json:"usage_ratio"`
	Health     health    `json:"health"`
	CreatedAt  time.Time `json:"created_at"`
}

type inbound struct {
	ID       string `json:"id"`
	NodeID   string `json:"node_id"`
	Tag      string `json:"tag"`
	Protocol string `json:"protocol"`
	Listen   string `json:"listen"`
	Port     int    `json:"port"`
	Network  string `json:"network"`
	Security string `json:"security"`
	Enabled  bool   `json:"enabled"`
}

var (
	mu       sync.Mutex
	users    = map[string]*user{}
	nodes    = map[string]*node{}
	inbounds = map[string]*inbound{}
)

func id() string {
	b := make([]byte, 8)
	_, _ = crand.Read(b)
	return hex.EncodeToString(b)
}

func seed() {
	exp := time.Now().AddDate(0, 1, 0).Format(time.RFC3339)
	users[id()] = &user{ID: id(), Username: "alice", Status: "active", DataLimit: 50 << 30, UsedTraffic: 33 << 30, ExpireAt: &exp, ResetStrategy: "monthly", SubToken: "Xk7Qa9demo", CreatedAt: time.Now()}
	users[id()] = &user{ID: id(), Username: "bob", Status: "limited", DataLimit: 50 << 30, UsedTraffic: 50 << 30, ResetStrategy: "no_reset", SubToken: "Bb22demo", CreatedAt: time.Now()}
	users[id()] = &user{ID: id(), Username: "carol", Status: "active", DataLimit: 0, UsedTraffic: 1 << 30, ResetStrategy: "no_reset", SubToken: "Cc33demo", CreatedAt: time.Now()}
	n1 := &node{ID: id(), Name: "de-1", Address: "5.5.5.5:50051", Core: "xray", Status: "connected", UsageRatio: 1, Health: health{CPU: 12, Mem: 41, CoreRunning: true, Connections: 87}, CreatedAt: time.Now()}
	n2 := &node{ID: id(), Name: "nl-2", Address: "6.6.6.6:50051", Core: "singbox", Status: "connected", UsageRatio: 2, Health: health{CPU: 5, Mem: 33, CoreRunning: true, Connections: 41}, CreatedAt: time.Now()}
	nodes[n1.ID] = n1
	nodes[n2.ID] = n2
	inbounds[id()] = &inbound{ID: id(), NodeID: n1.ID, Tag: "vless-ws", Protocol: "vless", Port: 443, Network: "ws", Security: "tls", Enabled: true}
	inbounds[id()] = &inbound{ID: id(), NodeID: n1.ID, Tag: "trojan-tcp", Protocol: "trojan", Port: 8443, Network: "tcp", Security: "tls", Enabled: true}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func main() {
	seed()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/login", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]string{"token": "demo-token"})
	})

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		list := make([]*user, 0, len(users))
		for _, u := range users {
			list = append(list, u)
		}
		writeJSON(w, 200, map[string]any{"users": list, "total": len(list)})
	})
	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Username    string  `json:"username"`
			DataLimit   int64   `json:"data_limit"`
			ExpireAt    *string `json:"expire_at"`
			DeviceLimit int     `json:"device_limit"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		u := &user{ID: id(), Username: in.Username, Status: "active", DataLimit: in.DataLimit, ExpireAt: in.ExpireAt, DeviceLimit: in.DeviceLimit, ResetStrategy: "no_reset", SubToken: id() + "demo", CreatedAt: time.Now(), Proxies: creds{VlessUUID: id() + "-uuid", SSMethod: "aes-128-gcm"}}
		mu.Lock()
		users[u.ID] = u
		mu.Unlock()
		writeJSON(w, 201, map[string]any{"user": u})
	})
	mux.HandleFunc("PUT /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		u := users[r.PathValue("id")]
		if u == nil {
			writeJSON(w, 404, map[string]string{"message": "not found"})
			return
		}
		var in struct {
			Note        string  `json:"note"`
			Status      string  `json:"status"`
			DataLimit   int64   `json:"data_limit"`
			ExpireAt    *string `json:"expire_at"`
			DeviceLimit int     `json:"device_limit"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		u.Note, u.Status, u.DataLimit, u.ExpireAt, u.DeviceLimit = in.Note, in.Status, in.DataLimit, in.ExpireAt, in.DeviceLimit
		writeJSON(w, 200, map[string]any{"user": u})
	})
	mux.HandleFunc("DELETE /api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		delete(users, r.PathValue("id"))
		mu.Unlock()
		w.WriteHeader(204)
	})
	mux.HandleFunc("GET /api/users/{id}/usage", func(w http.ResponseWriter, _ *http.Request) {
		pts := make([]map[string]any, 7)
		for i := range pts {
			pts[i] = map[string]any{
				"time": time.Now().AddDate(0, 0, i-6).Format(time.RFC3339),
				"up":   int64(rand.IntN(2<<30) + (1 << 29)),
				"down": int64(rand.IntN(5<<30) + (1 << 30)),
			}
		}
		writeJSON(w, 200, map[string]any{"points": pts})
	})

	mux.HandleFunc("GET /api/nodes", func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		list := make([]*node, 0, len(nodes))
		for _, n := range nodes {
			list = append(list, n)
		}
		writeJSON(w, 200, map[string]any{"nodes": list})
	})
	mux.HandleFunc("POST /api/nodes", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Name, Address, Core string }
		_ = json.NewDecoder(r.Body).Decode(&in)
		n := &node{ID: id(), Name: in.Name, Address: in.Address, Core: in.Core, Status: "connected", UsageRatio: 1, Health: health{CoreRunning: true}, CreatedAt: time.Now()}
		mu.Lock()
		nodes[n.ID] = n
		mu.Unlock()
		writeJSON(w, 201, map[string]any{"node": n})
	})
	mux.HandleFunc("DELETE /api/nodes/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		delete(nodes, r.PathValue("id"))
		mu.Unlock()
		w.WriteHeader(204)
	})

	mux.HandleFunc("GET /api/inbounds", func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.URL.Query().Get("node_id")
		mu.Lock()
		defer mu.Unlock()
		list := []*inbound{}
		for _, ib := range inbounds {
			if ib.NodeID == nodeID {
				list = append(list, ib)
			}
		}
		writeJSON(w, 200, map[string]any{"inbounds": list})
	})
	mux.HandleFunc("POST /api/inbounds", func(w http.ResponseWriter, r *http.Request) {
		var ib inbound
		_ = json.NewDecoder(r.Body).Decode(&ib)
		ib.ID = id()
		mu.Lock()
		inbounds[ib.ID] = &ib
		mu.Unlock()
		writeJSON(w, 201, map[string]any{"inbound": ib})
	})
	mux.HandleFunc("DELETE /api/inbounds/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		delete(inbounds, r.PathValue("id"))
		mu.Unlock()
		w.WriteHeader(204)
	})

	mux.HandleFunc("GET /api/admins", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"admins": []map[string]any{
			{"id": "1", "username": "root", "sudo": true, "totp_enabled": false},
			{"id": "2", "username": "reseller", "sudo": false, "role_id": "r1", "totp_enabled": true},
		}})
	})
	mux.HandleFunc("POST /api/admins", func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		in["id"] = id()
		writeJSON(w, 201, map[string]any{"admin": in})
	})
	mux.HandleFunc("DELETE /api/admins/{id}", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) })

	mux.HandleFunc("GET /api/roles", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"roles": []map[string]any{
			{"id": "r1", "name": "reseller", "permissions": []string{"user:read", "user:write"}},
		}})
	})
	mux.HandleFunc("POST /api/roles", func(w http.ResponseWriter, r *http.Request) {
		var in map[string]any
		_ = json.NewDecoder(r.Body).Decode(&in)
		in["id"] = id()
		writeJSON(w, 201, map[string]any{"role": in})
	})

	mux.HandleFunc("POST /api/account/2fa/setup", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"secret": "JBSWY3DPEHPK3PXP", "url": "otpauth://totp/VortexUI:root?secret=JBSWY3DPEHPK3PXP&issuer=VortexUI"})
	})
	mux.HandleFunc("POST /api/account/2fa/confirm", func(w http.ResponseWriter, r *http.Request) {
		var in struct{ Code string }
		_ = json.NewDecoder(r.Body).Decode(&in)
		if len(in.Code) == 6 {
			writeJSON(w, 200, map[string]any{"enabled": true})
			return
		}
		writeJSON(w, 400, map[string]string{"message": "invalid code"})
	})
	mux.HandleFunc("POST /api/account/2fa/disable", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"enabled": false})
	})

	log.Println("mock panel API listening on :8080 (DEV ONLY, in-memory)")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
