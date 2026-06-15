# Supported Protocols & Configuration Examples

This document covers every protocol/transport combination VortexUI supports, with
copy-paste inbound configurations. All examples assume the panel is running and
you create inbounds via the API or UI — the panel renders the native Xray/sing-box
JSON automatically.

---

## Table of Contents

- [VLESS](#vless)
  - [VLESS + TCP + REALITY](#vless--tcp--reality)
  - [VLESS + WebSocket + TLS](#vless--websocket--tls)
  - [VLESS + gRPC + TLS](#vless--grpc--tls)
  - [VLESS + HTTPUpgrade + TLS](#vless--httpupgrade--tls)
  - [VLESS + TCP + XTLS-Vision](#vless--tcp--xtls-vision)
- [VMess](#vmess)
  - [VMess + WebSocket + TLS](#vmess--websocket--tls)
  - [VMess + TCP](#vmess--tcp)
  - [VMess + gRPC + TLS](#vmess--grpc--tls)
- [Trojan](#trojan)
  - [Trojan + TCP + TLS](#trojan--tcp--tls)
  - [Trojan + WebSocket + TLS](#trojan--websocket--tls)
  - [Trojan + gRPC + TLS](#trojan--grpc--tls)
- [Shadowsocks](#shadowsocks)
  - [Shadowsocks 2022](#shadowsocks-2022)
  - [Shadowsocks Classic](#shadowsocks-classic)
- [Hysteria2](#hysteria2)
- [TUIC](#tuic)
- [Outbound Examples](#outbound-examples)
- [Routing Examples](#routing-examples)

---

## VLESS

### VLESS + TCP + REALITY

The most censorship-resistant setup. No real certificate needed — the server
impersonates a legitimate website.

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `reality` |
| SNI | `www.microsoft.com` (or any high-traffic site) |
| Flow | `xtls-rprx-vision` |

**Panel UI steps:**
1. Go to **Nodes → Inbounds → Add**
2. Set protocol `vless`, port `443`, network `tcp`, security `reality`
3. Click **Generate** in the REALITY section to get keys
4. Set SNI to a target domain (e.g. `www.microsoft.com`)
5. Set flow to `xtls-rprx-vision`

**Raw JSON (for advanced use):**
```json
{
  "reality": {
    "private_key": "<generated>",
    "short_ids": ["<generated>"],
    "server_names": ["www.microsoft.com"],
    "dest": "www.microsoft.com:443"
  }
}
```

---

### VLESS + WebSocket + TLS

Best for CDN (Cloudflare) fronting.

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `ws` |
| Security | `tls` |
| SNI | `your-domain.com` |
| Path | `/vless-ws` |

---

### VLESS + gRPC + TLS

Low-latency multiplexed transport, CDN-compatible.

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `grpc` |
| Security | `tls` |
| SNI | `your-domain.com` |
| Path | `vless-grpc` (service name) |

---

### VLESS + HTTPUpgrade + TLS

Modern alternative to WebSocket, lower overhead.

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `httpupgrade` |
| Security | `tls` |
| Path | `/vless-hu` |

---

### VLESS + TCP + XTLS-Vision

Direct TLS with XTLS splice for maximum performance.

| Field | Value |
|-------|-------|
| Protocol | `vless` |
| Port | `443` |
| Network | `tcp` |
| Security | `tls` |
| Flow | `xtls-rprx-vision` |

---

## VMess

### VMess + WebSocket + TLS

Classic CDN-frontable setup.

| Field | Value |
|-------|-------|
| Protocol | `vmess` |
| Port | `443` |
| Network | `ws` |
| Security | `tls` |
| Path | `/vmess-ws` |
| Host | `your-domain.com` |

---

### VMess + TCP

Simple direct connection (no TLS needed for trusted networks).

| Field | Value |
|-------|-------|
| Protocol | `vmess` |
| Port | `10086` |
| Network | `tcp` |
| Security | `none` |

---

### VMess + gRPC + TLS

| Field | Value |
|-------|-------|
| Protocol | `vmess` |
| Port | `443` |
| Network | `grpc` |
| Security | `tls` |
| Path | `vmess-grpc` |

---

## Trojan

### Trojan + TCP + TLS

| Field | Value |
|-------|-------|
| Protocol | `trojan` |
| Port | `443` |
| Network | `tcp` |
| Security | `tls` |
| SNI | `your-domain.com` |

---

### Trojan + WebSocket + TLS

CDN-compatible Trojan.

| Field | Value |
|-------|-------|
| Protocol | `trojan` |
| Port | `443` |
| Network | `ws` |
| Security | `tls` |
| Path | `/trojan-ws` |

---

### Trojan + gRPC + TLS

| Field | Value |
|-------|-------|
| Protocol | `trojan` |
| Port | `443` |
| Network | `grpc` |
| Security | `tls` |
| Path | `trojan-grpc` |

---

## Shadowsocks

### Shadowsocks 2022

Modern AEAD cipher (recommended).

| Field | Value |
|-------|-------|
| Protocol | `shadowsocks` |
| Port | `8388` |
| Network | `tcp` |
| Security | `none` |
| Method | `2022-blake3-aes-128-gcm` |

---

### Shadowsocks Classic

| Field | Value |
|-------|-------|
| Protocol | `shadowsocks` |
| Port | `8388` |
| Network | `tcp` |
| Security | `none` |
| Method | `aes-128-gcm` or `chacha20-ietf-poly1305` |

---

## Hysteria2

UDP-based, high-speed protocol (sing-box only).

| Field | Value |
|-------|-------|
| Protocol | `hysteria2` |
| Port | `443` |
| Network | `udp` |
| Security | `tls` |
| SNI | `your-domain.com` |

**Note:** Requires sing-box core on the node. Set `VORTEX_CORE=singbox`.

---

## TUIC

QUIC-based protocol with multiplexing (sing-box only).

| Field | Value |
|-------|-------|
| Protocol | `tuic` |
| Port | `443` |
| Network | `udp` |
| Security | `tls` |
| SNI | `your-domain.com` |

**Note:** Requires sing-box core on the node.

---

## Outbound Examples

### Direct (freedom)
```json
{ "tag": "direct", "protocol": "freedom" }
```

### Block (blackhole)
```json
{ "tag": "blocked", "protocol": "blackhole" }
```

### Proxy chain (VLESS upstream)
```json
{
  "tag": "proxy-de",
  "protocol": "vless",
  "address": "de-server.example.com",
  "port": 443,
  "uuid": "your-uuid",
  "network": "ws",
  "security": "tls",
  "sni": "de-server.example.com",
  "path": "/chain"
}
```

### SOCKS5 upstream
```json
{
  "tag": "socks-exit",
  "protocol": "socks",
  "address": "127.0.0.1",
  "port": 1080,
  "username": "user",
  "password": "pass"
}
```

---

## Routing Examples

### Block ads + Iran direct
```json
[
  {
    "priority": 1,
    "name": "block-ads",
    "domains": ["geosite:category-ads-all"],
    "outbound_tag": "blocked"
  },
  {
    "priority": 2,
    "name": "iran-direct",
    "domains": ["geosite:ir"],
    "ip": ["geoip:ir"],
    "outbound_tag": "direct"
  },
  {
    "priority": 10,
    "name": "everything-else",
    "port": "1-65535",
    "outbound_tag": "proxy-de"
  }
]
```

### Load balance across proxies
```json
{
  "balancer": {
    "tag": "auto-best",
    "selectors": ["proxy-"],
    "strategy": "leastPing",
    "probe_url": "https://www.gstatic.com/generate_204",
    "probe_interval": "10s"
  },
  "rule": {
    "priority": 5,
    "name": "balance-all",
    "inbound_tags": ["vless-ws"],
    "balancer_tag": "auto-best"
  }
}
```

---

## Best Practices

1. **REALITY** is the gold standard for censorship bypass — use it as default
2. **WebSocket + CDN** for Cloudflare protection (add Cloudflare DNS)
3. **gRPC** for lowest latency when CDN supports it
4. **Hysteria2** for maximum speed on unrestricted networks
5. **Separate ports** for each protocol to avoid conflicts
6. **Use GeoIP/Geosite** rules to route Iranian traffic directly
7. **Enable the balancer** for multi-proxy failover (leastPing recommended)
8. **XTLS-Vision flow** only works with TCP transport (not WS/gRPC)

---

## Client Compatibility

| Client | VLESS+Reality | VLESS+WS | VMess+WS | Trojan | SS | Hysteria2 | TUIC |
|--------|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| v2rayNG | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ |
| Hiddify | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Clash Meta | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| sing-box | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Shadowrocket | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| Streisand | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| NekoBox | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |

---

*For more examples, see the [XTLS/Xray-examples](https://github.com/XTLS/Xray-examples) repository.*
