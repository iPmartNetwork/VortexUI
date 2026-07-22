# Subscription Pipeline

## Overview

VortexUI v1.4.0 processes every subscription request through a **9-stage pipeline** that transforms a raw `/sub/{token}` request into an optimized, ISP-aware, format-specific configuration. Each stage adds intelligence to the final output.

---

## Pipeline Stages

```
┌─────────────────────────────────────────────────────────┐
│  Stage 1: ISP Auto-Detection                            │
│  Stage 2: Format Detection (UA-based)                   │
│  Stage 3: Protocol Group Resolution                     │
│  Stage 4: Smart Profiles Application                    │
│  Stage 5: Quality Score Calculation                     │
│  Stage 6: Multi-Path Outbound Selection                 │
│  Stage 7: Enhanced Probing Configuration                │
│  Stage 8: Share Link Parameter Injection                │
│  Stage 9: CDN Fallback Proxy Generation                 │
└─────────────────────────────────────────────────────────┘
```

---

## Stage 1: ISP Auto-Detection

The pipeline identifies the user's ISP from the subscription request IP:

| Method | Priority | Source |
|--------|----------|--------|
| MaxMind GeoIP2 ASN | Primary | ASN database lookup |
| Static IP ranges | Fallback | Hardcoded Iranian ISP CIDR blocks |
| X-Forwarded-For | CDN mode | When behind Cloudflare/CDN |
| Manual override | Highest | `?isp=MCI` query parameter |

Detection result determines which protocol group, mux profile, and SNI pool apply.

```bash
# Check detected ISP for a subscription
curl https://panel.example.com/sub/{token}?format=json&debug=true \
  -H "X-Debug-Key: $DEBUG_KEY"

# Response includes:
# { "detected_isp": "MCI", "asn": 197207, "severity": "aggressive" }
```

---

## Stage 2: Format Detection

User-Agent determines the output format:

| Client | UA Pattern | Format |
|--------|-----------|--------|
| sing-box | `sing-box` | sing-box JSON |
| Clash Meta | `clash-meta`, `mihomo` | Clash YAML |
| Clash Verge | `clash-verge` | Clash YAML |
| V2rayN | `v2rayn` | Base64 share links |
| V2rayNG | `v2rayng` | Base64 share links |
| Nekoray | `nekoray` | Base64 share links |
| Hiddify | `hiddify` | sing-box JSON |
| Streisand | `streisand` | sing-box JSON |
| Unknown | — | Base64 share links (default) |

Override with query parameter:

```
/sub/{token}?format=sing-box
/sub/{token}?format=clash
/sub/{token}?format=v2ray
```

---

## Stage 3: Protocol Group Resolution

Based on ISP detection, the pipeline selects the appropriate [Protocol Group](08-protocol-groups.md):

1. Find groups matching detected ISP profile
2. If multiple matches, prefer group with highest quality score
3. If no ISP-specific group, use `DEFAULT` group
4. Resolve inbound list with priority ordering

---

## Stage 4: Smart Profiles Application

Smart profiles inject anti-censorship settings based on the resolved ISP:

| Setting | Source | Applied To |
|---------|--------|-----------|
| Mux protocol | ISP mux profile | All outbounds |
| SNI | Daily rotation pool | TLS-based outbounds |
| Transport headers | Obfuscation rules | gRPC, WS, HTTPUpgrade |
| Fragment settings | ISP severity | TLS handshake |
| Padding | ISP mux profile | Mux-enabled outbounds |

See [Anti-Censorship](09-anti-censorship.md) for profile details.

---

## Stage 5: Quality Score

Every inbound receives a quality score (0–100) calculated from real-time telemetry:

### Scoring Formula

```
quality_score = (
    success_rate × 40 +
    latency_score × 25 +
    uptime_7d × 20 +
    switch_stability × 15
)
```

| Component | Weight | Measurement |
|-----------|--------|-------------|
| Success Rate | 40% | Successful connections / total attempts (24h) |
| Latency Score | 25% | `100 - min(latency_ms / 20, 100)` |
| Uptime (7-day) | 20% | Node uptime percentage over 7 days |
| Switch Stability | 15% | Inverse of switch-away events (24h) |

### Score Thresholds

| Score | Rating | Action |
|-------|--------|--------|
| 80–100 | Excellent | Prioritized in group |
| 60–79 | Good | Included normally |
| 40–59 | Fair | Deprioritized, monitoring increased |
| 0–39 | Poor | Excluded from subscription, alert sent |

```bash
# Get quality scores
curl https://panel.example.com/api/subscription/quality-scores \
  -H "Authorization: Bearer $TOKEN"
```

---

## Stage 6: Multi-Path Outbound

The pipeline selects the **top 4 inbounds** by quality score and generates multi-path outbound configuration:

```json
{
  "outbounds": [
    {
      "type": "urltest",
      "tag": "auto-select",
      "outbounds": ["proxy-1", "proxy-2", "proxy-3", "proxy-4"],
      "url": "https://cp.cloudflare.com/generate_204",
      "interval": "3m",
      "tolerance": 100
    },
    { "type": "vless", "tag": "proxy-1", "...": "..." },
    { "type": "trojan", "tag": "proxy-2", "...": "..." },
    { "type": "vmess", "tag": "proxy-3", "...": "..." },
    { "type": "hysteria2", "tag": "proxy-4", "...": "..." }
  ]
}
```

> **Note:** The number of outbounds is configurable per group (default: 4, max: 8).

---

## Stage 7: Enhanced Probing

Client-side health probing is tuned per ISP:

| Parameter | MCI | Irancell | TCI | Default |
|-----------|-----|----------|-----|---------|
| Probe URLs | 3 | 2 | 3 | 2 |
| Interval | 90s | 180s | 60s | 180s |
| Tolerance | 150ms | 200ms | 100ms | 200ms |
| Timeout | 5s | 5s | 3s | 5s |

Probe URLs (rotated to avoid fingerprinting):

```json
{
  "probe_urls": [
    "https://cp.cloudflare.com/generate_204",
    "https://www.gstatic.com/generate_204",
    "https://detectportal.firefox.com/success.txt"
  ]
}
```

---

## Stage 8: Share Link Parameters

For Base64 / share link formats, additional parameters are injected:

| Parameter | Value | Purpose |
|-----------|-------|---------|
| `port_range` | `20000-40000` | Port hopping range (Hysteria2/TUIC) |
| `hop_interval` | `30` | Seconds between port hops |
| `fragment` | `1-3,100-200,tlshello` | TLS fragment settings |
| `mux` | `h2mux,8,padding` | Mux protocol config |

Example generated share link:

```
vless://uuid@server:443?type=grpc&security=reality&sni=dl.google.com
  &serviceName=google.pubsub.v1.Publisher
  &pbk=...&sid=...&fp=chrome
  &fragment=1-3,100-200,tlshello
  &mux=h2mux,8,padding
```

---

## Stage 9: CDN Fallback Proxy

The final stage generates a CDN-routed fallback proxy for each primary connection:

```json
{
  "outbounds": [
    { "tag": "primary-vless", "server": "direct-ip", "...": "..." },
    {
      "tag": "cdn-fallback-vless",
      "type": "vless",
      "server": "172.67.x.x",
      "server_port": 443,
      "tls": {
        "enabled": true,
        "server_name": "your-worker.domain.workers.dev",
        "alpn": ["h2", "http/1.1"]
      },
      "transport": {
        "type": "ws",
        "path": "/sub-ws",
        "headers": { "Host": "your-worker.domain.workers.dev" }
      }
    }
  ]
}
```

CDN fallback activates when direct connection fails, providing a secondary path through Cloudflare's network.

---

## Connection Resilience

The subscription includes resilience settings for robust connectivity:

| Setting | Value | Description |
|---------|-------|-------------|
| `idle_timeout` | `15m` | Close idle connections after 15 minutes |
| `connect_timeout` | `5s` | Connection establishment timeout |
| `tcp_fast_open` | `true` | Reduce handshake latency |
| `tcp_multi_path` | `true` | MPTCP when available |
| `udp_timeout` | `5m` | UDP session timeout |
| `domain_strategy` | `prefer_ipv4` | IPv4 first for stability |

---

## Monitoring & Debug

### Subscription Analytics

```bash
# Get pipeline execution stats
curl https://panel.example.com/api/subscription/stats \
  -H "Authorization: Bearer $TOKEN"
```

Response:

```json
{
  "total_requests_24h": 12450,
  "by_format": { "sing-box": 6200, "clash": 3100, "v2ray": 3150 },
  "by_isp": { "MCI": 5400, "IRANCELL": 3800, "TCI": 2100, "other": 1150 },
  "avg_quality_score": 78,
  "cache_hit_rate": 0.92
}
```

### Debug Mode

Add `?debug=true` with a valid debug key to see pipeline decisions:

```bash
curl "https://panel.example.com/sub/{token}?debug=true" \
  -H "X-Debug-Key: $DEBUG_KEY"
```

---

## Configuration

Pipeline settings are managed at **Settings → Subscription** or via API:

```bash
curl -X PATCH https://panel.example.com/api/settings/subscription \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "max_outbounds": 4,
    "probe_interval": "3m",
    "quality_threshold": 40,
    "cdn_fallback_enabled": true,
    "cache_ttl": "5m"
  }'
```

---

## Related Pages

- [Protocol Groups](08-protocol-groups.md)
- [Anti-Censorship](09-anti-censorship.md)
- [Node Management](06-node-management.md)
- [API Reference](17-api-reference.md)
