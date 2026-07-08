# Protocols & Configuration

> **Capability Matrix:** The panel exposes a live per-protocol capability matrix (`GET /api/capabilities`) as the single source of truth. The inbound editor only offers combinations the selected node's core actually supports.

---

## Protocol Overview

| Protocol | Core | Inbound | Outbound | Transport | Security |
|----------|------|---------|----------|-----------|----------|
| VLESS | Both | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, xHTTP, mKCP | None, TLS, REALITY |
| VMess | Both | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, mKCP | None, TLS |
| Trojan | Both | ✅ | ✅ | TCP, WS, gRPC, mKCP | TLS, REALITY |
| Shadowsocks | Both | ✅ | ✅ | TCP (+ SS-2022 multi-user) | None |
| Hysteria2 | sing-box | ✅ | ✅ | UDP (QUIC) | TLS |
| TUIC | sing-box | ✅ | ✅ | UDP (QUIC) | TLS |
| WireGuard | sing-box | ✅ | ✅ | UDP | Native |
| Hysteria (v1) | sing-box | ✅ | — | UDP | TLS |
| ShadowTLS | sing-box | ✅ | ✅ | TCP | TLS |
| AnyTLS | sing-box | ✅ | — | TCP | TLS |
| Naive | sing-box | ✅ | — | — | TLS (mandatory) |
| SOCKS | Both | ✅ | ✅ | — (raw TCP) | plaintext |
| HTTP | Both | ✅ | ✅ | — (raw TCP) | plaintext |
| Dokodemo | Xray | ✅ | — | — (raw TCP/UDP) | plaintext |

---

## Per-Protocol Configuration

### VLESS + REALITY (Recommended)

The gold standard for censorship resistance. REALITY eliminates the need for a TLS certificate.

**Inbound config:**

| Field | Example |
|-------|---------|
| Protocol | vless |
| Port | 443 |
| Transport | tcp |
| Security | reality |
| Dest (target) | www.google.com:443 |
| Server Names | www.google.com |
| Private Key | Auto-generated |
| Short IDs | Auto-generated (up to 8) |
| Flow | xtls-rprx-vision (for TCP) |

> **Tip:** Use the **Reality Scanner** to find the best SNI domains for your server location.

### VMess + WebSocket + TLS

Classic setup compatible with CDN fronting (Cloudflare):

| Field | Example |
|-------|---------|
| Protocol | vmess |
| Port | 443 |
| Transport | ws |
| Path | /vmws |
| Security | tls |
| SNI | cdn.example.com |

Works behind Cloudflare with WebSocket enabled on the domain.

### Trojan + gRPC + TLS

High-performance option with multiplexing:

| Field | Example |
|-------|---------|
| Protocol | trojan |
| Port | 443 |
| Transport | grpc |
| Service Name | trojangrpc |
| Security | tls |
| SNI | your-domain.com |

### Shadowsocks 2022 (Multi-User)

Modern Shadowsocks with per-user keys:

| Field | Example |
|-------|---------|
| Protocol | shadowsocks |
| Port | 8388 |
| Method | 2022-blake3-aes-128-gcm |
| Server Key | Auto-generated |
| Security | none (SS handles its own encryption) |

Each user gets a derived key — no shared password.

### Hysteria2

QUIC-based protocol with built-in congestion control. Excellent for lossy networks:

| Field | Example |
|-------|---------|
| Protocol | hysteria2 |
| Port | 4443 |
| Security | tls (mandatory) |
| Up/Down bandwidth | Client-reported for congestion control |
| Obfs type | salamander (optional) |
| Obfs password | Shared secret |

> **Note:** Hysteria2 requires sing-box core. Not available on Xray nodes.

### TUIC

QUIC-based UDP relay with zero-RTT:

| Field | Example |
|-------|---------|
| Protocol | tuic |
| Port | 4444 |
| Security | tls (mandatory) |
| Congestion | bbr or cubic |
| UUID | Per-user authentication |

### WireGuard

Native WireGuard tunnel via sing-box:

| Field | Example |
|-------|---------|
| Protocol | wireguard |
| Port | 51820 |
| Private Key | Server private key |
| Peer public keys | Per-user public keys |
| Allowed IPs | 0.0.0.0/0, ::/0 |
| MTU | 1280 |

### Naive (NaiveProxy)

HTTP/2 or HTTP/3 proxy disguised as normal HTTPS traffic:

| Field | Example |
|-------|---------|
| Protocol | naive |
| Port | 443 |
| Security | tls (mandatory) |
| Username/Password | Per-user credentials |

> **Warning:** Naive requires sing-box core and **mandates TLS**. Cannot run without a valid certificate.

### ShadowTLS

TLS camouflage — makes traffic look like a normal TLS connection to a popular website:

| Field | Example |
|-------|---------|
| Protocol | shadowtls |
| Port | 443 |
| Version | 3 (recommended) |
| Handshake server | www.microsoft.com:443 |
| Password | Shared secret |

---

## Capability Matrix (Xray vs sing-box)

### Xray-core

| Category | Supported |
|----------|-----------|
| Protocols | vless, vmess, trojan, shadowsocks, socks, http, dokodemo |
| Transports | tcp, ws, grpc, httpupgrade, http/h2, xhttp, mkcp |
| Security | none, tls, reality |
| Special | xtls-rprx-vision flow, xhttp mode selector, mKCP headers |

### sing-box

| Category | Supported |
|----------|-----------|
| Protocols | vless, vmess, trojan, shadowsocks, hysteria2, tuic, wireguard, hysteria, shadowtls, anytls, naive, socks, http |
| Transports | tcp, ws, grpc, httpupgrade, http/h2, quic |
| Security | none, tls, reality (limited) |
| Special | QUIC-based protocols, multiplex, brutal congestion |

---

## Transport Details

### TCP

Default transport. Supports optional HTTP camouflage header (Xray).

### WebSocket (WS)

HTTP upgrade to WebSocket. CDN-compatible (Cloudflare, etc.).

| Setting | Description |
|---------|-------------|
| Path | URL path (e.g. /ws) |
| Host | HTTP Host header |
| Max early data | Bytes in first WS frame (0-RTT) |

### gRPC

HTTP/2 based. High performance with multiplexing.

| Setting | Description |
|---------|-------------|
| Service name | gRPC service path |
| Multi-mode | Enable multi-stream mode |

### HTTPUpgrade

HTTP/1.1 upgrade (like WS but simpler). Supported by both cores.

| Setting | Description |
|---------|-------------|
| Path | URL path |
| Host | HTTP Host header |

### xHTTP (Xray only)

Advanced HTTP transport with multiple modes:

| Mode | Description |
|------|-------------|
| auto | Auto-detect best mode |
| packet-up | Packet framing for upload |
| stream-up | Streaming upload |

### mKCP (Xray only)

UDP-based transport with FEC (Forward Error Correction). Good for lossy networks.

| Setting | Description |
|---------|-------------|
| Header type | none, srtp, utp, wechat-video, dtls, wireguard |
| Seed | Obfuscation seed |
| MTU | Maximum transmission unit |

---

## Security Layers

### None

No encryption on the transport layer. Protocol handles its own encryption (e.g. VMess, Shadowsocks).

### TLS

Standard TLS 1.2/1.3. Requires a valid certificate (auto-provisioned via Caddy, or manually configured).

| Setting | Description |
|---------|-------------|
| SNI | Server name indication |
| ALPN | Application-layer protocol (h2, http/1.1) |
| Certificate | Auto (ACME) or manual (file path) |
| Min version | 1.2 or 1.3 |
| Fingerprint | uTLS impersonation |

### REALITY

TLS 1.3 imitation without needing a real certificate. The server impersonates a legitimate website.

| Setting | Description |
|---------|-------------|
| Dest | Target server to impersonate |
| Server Names | Allowed SNI values |
| Private Key | X25519 server key |
| Short IDs | Client authentication IDs |
| Spider X | Path for active probing evasion |

---

## Subscription Output Formats

| Format | Content-Type | Description |
|--------|--------------|-------------|
| base64 | text/plain | V2Ray-compatible base64 share links |
| clash | text/yaml | Clash Meta YAML config |
| singbox | application/json | sing-box client JSON |
| xray | application/json | Raw Xray/V2Ray JSON |
| outline | text/plain | ss:// links for Outline |
| links | text/plain | One share link per line |

Auto-detection from User-Agent:

| Client | Detected format |
|--------|-----------------|
| Clash / ClashX / Clash Meta | clash |
| sing-box | singbox |
| Outline | outline |
| v2rayNG / V2RayN | base64 |
| Other | base64 |
