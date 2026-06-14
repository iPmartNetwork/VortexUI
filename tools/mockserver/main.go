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
	Status      string    `json:"status"`
	UsageRatio  float64   `json:"usage_ratio"`
	Health      health    `json:"health"`
	CoreVersion string    `json:"core_version"`
	AgentVer    string    `json:"agent_version"`
	CreatedAt   time.Time `json:"created_at"`
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
	n1 := &node{ID: id(), Name: "de-1", Address: "5.5.5.5:50051", Core: "xray", Status: "connected", UsageRatio: 1, Health: health{CPU: 12, Mem: 41, Disk: 28, CoreRunning: true, Connections: 87}, CoreVersion: "Xray 1.8.24", AgentVer: "0.1.0", CreatedAt: time.Now()}
	n2 := &node{ID: id(), Name: "nl-2", Address: "6.6.6.6:50051", Core: "singbox", Status: "connected", UsageRatio: 2, Health: health{CPU: 73, Mem: 58, Disk: 44, CoreRunning: true, Connections: 41}, CoreVersion: "sing-box 1.9.3", AgentVer: "0.1.0", CreatedAt: time.Now()}
	nodes[n1.ID] = n1
	nodes[n2.ID] = n2
	inbounds[id()] = &inbound{ID: id(), NodeID: n1.ID, Tag: "vless-ws", Protocol: "vless", Port: 443, Network: "ws", Security: "tls", Enabled: true}
	inbounds[id()] = &inbound{ID: id(), NodeID: n1.ID, Tag: "trojan-tcp", Protocol: "trojan", Port: 8443, Network: "tcp", Security: "tls", Enabled: true}
}

func firstNodeID() string {
	mu.Lock()
	defer mu.Unlock()
	for id := range nodes {
		return id
	}
	return ""
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
	mux.HandleFunc("PUT /api/nodes/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		n := nodes[r.PathValue("id")]
		if n == nil {
			writeJSON(w, 404, map[string]string{"message": "not found"})
			return
		}
		var in struct {
			Name, Address string
			UsageRatio    float64 `json:"usage_ratio"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Name != "" {
			n.Name = in.Name
		}
		if in.Address != "" {
			n.Address = in.Address
		}
		if in.UsageRatio > 0 {
			n.UsageRatio = in.UsageRatio
		}
		writeJSON(w, 200, map[string]any{"node": n})
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
	mux.HandleFunc("PUT /api/inbounds/{id}", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		ib := inbounds[r.PathValue("id")]
		if ib == nil {
			writeJSON(w, 404, map[string]string{"message": "not found"})
			return
		}
		var in struct {
			Port            int
			Network, Security string
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Port > 0 {
			ib.Port = in.Port
		}
		if in.Network != "" {
			ib.Network = in.Network
		}
		if in.Security != "" {
			ib.Security = in.Security
		}
		writeJSON(w, 200, map[string]any{"inbound": ib})
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

	mux.HandleFunc("POST /api/account/password", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"ok": true})
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

	// --- overview / logs / reality ---
	mux.HandleFunc("GET /api/overview", func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		nl := []map[string]any{}
		for _, n := range nodes {
			nl = append(nl, map[string]any{"id": n.ID, "name": n.Name, "core": n.Core, "online": n.Health.CoreRunning, "health": n.Health})
		}
		mu.Unlock()
		writeJSON(w, 200, map[string]any{
			"users": map[string]any{"total": 128, "total_used": int64(214) << 30, "by_status": map[string]int{"active": 96, "limited": 18, "expired": 9, "disabled": 5}},
			"nodes": map[string]any{"total": len(nl), "online": len(nl), "items": nl},
		})
	})
	mux.HandleFunc("GET /api/logs", func(w http.ResponseWriter, _ *http.Request) {
		now := time.Now()
		es := []map[string]any{
			{"time": now.Add(-90 * time.Second).Format(time.RFC3339), "level": 0, "message": "node resync on connect", "attrs": map[string]any{"node": "de-1"}},
			{"time": now.Add(-40 * time.Second).Format(time.RFC3339), "level": 4, "message": "traffic stream ended, reconnecting", "attrs": map[string]any{"node": "nl-2"}},
			{"time": now.Add(-10 * time.Second).Format(time.RFC3339), "level": 0, "message": "enforced user limit", "attrs": map[string]any{"user": "bob", "status": "limited"}},
		}
		writeJSON(w, 200, map[string]any{"entries": es})
	})
	mux.HandleFunc("GET /api/reality/keypair", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]any{"private_key": "yJ" + id(), "public_key": "Xk" + id(), "short_id": id()[:8]})
	})

	// --- user extras ---
	mux.HandleFunc("GET /api/users/{id}/sub", func(w http.ResponseWriter, _ *http.Request) {
		base := "https://panel.example.com/sub/Xk7Qa9demo"
		writeJSON(w, 200, map[string]any{
			"token": "Xk7Qa9demo", "subscription_url": base,
			"formats": map[string]any{"auto": base, "clash": base + "?format=clash", "singbox": base + "?format=singbox", "base64": base + "?format=base64"},
			"links": []string{"vless://11111111-1111-1111-1111-111111111111@5.5.5.5:443?type=ws&security=tls&sni=ex.com#alice", "trojan://pw@5.5.5.5:8443?security=tls#alice"},
		})
	})
	mux.HandleFunc("POST /api/users/{id}/reset", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]any{"ok": true}) })
	mux.HandleFunc("POST /api/users/{id}/revoke-sub", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]any{"ok": true}) })

	// --- per-node policy: outbounds / routing / balancers ---
	policy := func(name string, seed []map[string]any) {
		store := map[string]map[string]any{}
		for _, s := range seed {
			store[s["id"].(string)] = s
		}
		mux.HandleFunc("GET /api/"+name, func(w http.ResponseWriter, r *http.Request) {
			nid := r.URL.Query().Get("node_id")
			list := []map[string]any{}
			mu.Lock()
			for _, v := range store {
				if v["node_id"] == nid || nid == "" {
					list = append(list, v)
				}
			}
			mu.Unlock()
			writeJSON(w, 200, map[string]any{name: list})
		})
		mux.HandleFunc("POST /api/"+name, func(w http.ResponseWriter, r *http.Request) {
			var v map[string]any
			_ = json.NewDecoder(r.Body).Decode(&v)
			v["id"] = id()
			mu.Lock()
			store[v["id"].(string)] = v
			mu.Unlock()
			writeJSON(w, 201, map[string]any{name[:len(name)-1]: v})
		})
		mux.HandleFunc("DELETE /api/"+name+"/{id}", func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			delete(store, r.PathValue("id"))
			mu.Unlock()
			w.WriteHeader(204)
		})
	}
	policy("outbounds", []map[string]any{{"id": id(), "node_id": firstNodeID(), "tag": "direct-out", "protocol": "freedom", "enabled": true}})
	policy("routing", []map[string]any{{"id": id(), "node_id": firstNodeID(), "name": "block-ads", "priority": 1, "outbound_tag": "block", "enabled": true}})
	policy("balancers", []map[string]any{{"id": id(), "node_id": firstNodeID(), "tag": "auto", "strategy": "leastPing", "selectors": []string{"proxy-"}, "enabled": true}})

	log.Println("mock panel API listening on :8080 (DEV ONLY, in-memory)")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
