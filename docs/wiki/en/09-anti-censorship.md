# Anti-Censorship

## Smart Config Engine

VortexUI v1.4.0 introduces the **Smart Config Engine** — an adaptive layer that dynamically adjusts proxy configurations based on ISP severity, time of day, and historical block data. Instead of static configs, users receive optimized connections that evolve with censorship patterns.

> **Warning:** Anti-censorship features are tuned for Iran's filtering infrastructure. Other regions may require custom profiles.

---

## ISP Severity Classification

The engine classifies ISPs by filtering aggressiveness and applies protocol strategies accordingly:

| ISP | Severity | Behavior | Recommended Stack |
|-----|----------|----------|-------------------|
| MCI (Hamrah-e Aval) | Aggressive | Deep packet inspection, active probing, TLS fingerprint analysis | VLESS+Reality, gRPC, h2mux |
| Irancell (MTN) | Moderate | Basic DPI, occasional port blocking, less active probing | Trojan+WS+TLS, yamux |
| TCI (Mokhaberat) | Always Aggressive | Heavy filtering, SNI blocking, frequent protocol bans | VLESS+Reality+gRPC, full obfuscation |
| Shatel | Moderate | Standard HTTPS inspection, rarely blocks new protocols | VMess+WS+TLS, standard mux |
| Rightel | Low-Moderate | Mobile-focused, limited DPI capacity | Most protocols work |

---

## Time-of-Day Adaptation

Filtering intensity increases during peak hours. The engine adjusts severity dynamically:

| Time Window (IRST) | Modifier | Action |
|--------------------|----------|--------|
| 00:00–08:00 | -1 severity | Lighter obfuscation acceptable |
| 08:00–18:00 | Normal | Standard profile applies |
| 18:00–23:00 (Peak) | +1 severity | Maximum obfuscation engaged |
| 23:00–00:00 | Normal | Standard profile applies |

> **Note:** During national events or internet shutdowns, severity is forced to maximum regardless of time.

---

## Dynamic SNI Rotation

Static SNIs get blocked within days. VortexUI rotates SNIs daily per proxy from ISP-specific pools:

```json
{
  "sni_pools": {
    "MCI": ["dl.google.com", "update.microsoft.com", "cdn.cloudflare.com"],
    "IRANCELL": ["fonts.googleapis.com", "ajax.googleapis.com"],
    "TCI": ["www.google.com", "accounts.google.com", "play.google.com"]
  },
  "rotation_interval": "24h",
  "validation": "tls_handshake_check"
}
```

The panel validates each SNI candidate via TLS handshake before deployment. Failed SNIs are removed from the pool automatically.

### Managing SNI Pools

```bash
# Add SNIs to a pool
curl -X POST https://panel.example.com/api/anti-censorship/sni-pools \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "isp": "MCI",
    "snis": ["new.domain.com", "another.domain.com"],
    "validate": true
  }'

# Get current active SNIs
curl https://panel.example.com/api/anti-censorship/active-snis \
  -H "Authorization: Bearer $TOKEN"
```

---

## Transport Obfuscation

Each transport is disguised as legitimate traffic:

| Transport | Disguise | Header/Path |
|-----------|----------|-------------|
| gRPC | Google Pub/Sub API | `google.pubsub.v1.Publisher/Publish` |
| WebSocket | REST API polling | `/api/v2/ws`, `/api/v2/events` |
| HTTPUpgrade | Software update check | `/update/check`, `/api/version` |
| xHTTP | CDN asset fetch | `/assets/chunk-{random}.js` |

Configure per inbound:

```bash
curl -X PATCH https://panel.example.com/api/inbounds/12/transport \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "grpc",
    "service_name": "google.pubsub.v1.Publisher",
    "authority": "pubsub.googleapis.com"
  }'
```

---

## Smart Mux Configuration

Multiplexing strategy varies by ISP to avoid detection patterns:

| ISP | Mux Protocol | Padding | Max Streams | XUDP |
|-----|-------------|---------|-------------|------|
| MCI | h2mux | Enabled (100-300 bytes) | 8 | Enabled |
| Irancell | yamux | Enabled (50-150 bytes) | 12 | Enabled |
| TCI | h2mux | Enabled (200-500 bytes) | 4 | Enabled |
| Shatel | smux | Disabled | 16 | Optional |
| Default | yamux | Disabled | 8 | Enabled |

Example sing-box multiplex config generated:

```json
{
  "multiplex": {
    "enabled": true,
    "protocol": "h2mux",
    "max_streams": 8,
    "padding": true,
    "brutal": {
      "enabled": false
    }
  }
}
```

> **Note:** XUDP is required for UDP-over-TCP forwarding (gaming, VoIP). Always enabled for aggressive ISPs.

---

## CDN Routing

For ISPs that block direct connections, CDN routing provides a fallback path:

### Clean-IP Fallback

```json
{
  "cdn_fallback": {
    "enabled": true,
    "provider": "cloudflare",
    "clean_ips": ["172.67.x.x", "104.21.x.x"],
    "alpn": ["h2", "http/1.1"],
    "sni": "your-worker.domain.workers.dev"
  }
}
```

The panel's **Clean-IP Scanner** periodically discovers optimal CDN edge IPs:

```bash
curl -X POST https://panel.example.com/api/scanner/clean-ip \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "provider": "cloudflare",
    "count": 20,
    "timeout_ms": 3000,
    "target_latency_ms": 200
  }'
```

### ALPN Configuration

Proper ALPN prevents CDN connection failures:

| Scenario | ALPN | Notes |
|----------|------|-------|
| WS over CDN | `http/1.1` | CDN requires HTTP/1.1 for WebSocket upgrade |
| gRPC over CDN | `h2` | gRPC mandates HTTP/2 |
| HTTPUpgrade | `http/1.1` | Uses HTTP/1.1 101 Switching Protocols |
| Direct (no CDN) | `h2, http/1.1` | Allow server preference |

---

## DNS Leak Prevention

Prevent DNS queries from bypassing the tunnel:

| Method | Configuration | Purpose |
|--------|--------------|---------|
| DoH (DNS-over-HTTPS) | `https://1.1.1.1/dns-query` | Encrypted DNS for all queries |
| IR Direct | Route `.ir` domains directly | Iranian sites don't need proxy |
| Block port 53 | Firewall rule on client | Prevent plain DNS leaks |
| Fake DNS | sing-box `fakeip` mode | Eliminate DNS round-trip |

Generated client config includes:

```json
{
  "dns": {
    "servers": [
      { "tag": "doh", "address": "https://1.1.1.1/dns-query", "detour": "proxy" },
      { "tag": "direct-dns", "address": "https://dns.403.online/dns-query", "detour": "direct" }
    ],
    "rules": [
      { "domain_suffix": [".ir"], "server": "direct-dns" },
      { "clash_mode": "direct", "server": "direct-dns" }
    ]
  }
}
```

---

## Port Hopping

Rotate listening ports to evade port-based blocking:

```json
{
  "port_hopping": {
    "enabled": true,
    "port_range": "20000-40000",
    "hop_interval": "30s",
    "protocol": "hysteria2"
  }
}
```

| Parameter | Description | Default |
|-----------|-------------|---------|
| `port_range` | UDP port range for hopping | `20000-40000` |
| `hop_interval` | Time between port switches | `30s` |
| `protocol` | Must be UDP-based (Hysteria2, TUIC) | — |

> **Warning:** Port hopping requires the full port range opened in your firewall (`ufw allow 20000:40000/udp`).

---

## Emergency Fallback

The `🆘 outbound` activates when all primary connections fail:

1. Client detects all `urltest` outbounds unreachable
2. Switches to emergency outbound (pre-configured WARP or backup server)
3. Sends alert to panel via fallback endpoint
4. Panel logs event and notifies admin

Configure emergency outbound:

```bash
curl -X PUT https://panel.example.com/api/anti-censorship/emergency \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "wireguard",
    "server": "engage.cloudflareclient.com",
    "port": 2408,
    "private_key": "...",
    "peer_public_key": "...",
    "reserved": [0, 0, 0]
  }'
```

---

## Certificate Rotation

Automated TLS certificate rotation reduces fingerprinting:

| Setting | Value | Purpose |
|---------|-------|---------|
| Rotation interval | 15 days | Switch before patterns emerge |
| Providers | Let's Encrypt ↔ ZeroSSL | Alternate to avoid CA fingerprint |
| Method | DNS-01 (Cloudflare) | No port 80 required |
| Overlap | 2 days | Old cert valid during transition |

The panel handles rotation automatically. Monitor via:

```bash
curl https://panel.example.com/api/anti-censorship/certs \
  -H "Authorization: Bearer $TOKEN"
```

---

## Configuration Summary

All anti-censorship settings live under **Settings → Anti-Censorship** in the panel UI, or via the `/api/anti-censorship/*` endpoint family.

| Feature | UI Location | API Endpoint |
|---------|-------------|--------------|
| SNI Pools | Anti-Censorship → SNI | `/api/anti-censorship/sni-pools` |
| Mux Profiles | Anti-Censorship → Mux | `/api/anti-censorship/mux-profiles` |
| CDN Fallback | Anti-Censorship → CDN | `/api/anti-censorship/cdn` |
| Port Hopping | Inbounds → Advanced | `/api/inbounds/{id}/port-hopping` |
| Emergency | Anti-Censorship → Emergency | `/api/anti-censorship/emergency` |
| Cert Rotation | Settings → TLS | `/api/anti-censorship/certs` |

---

## Related Pages

- [Protocol Groups](08-protocol-groups.md)
- [Subscription Pipeline](12-subscription.md)
- [Network & Routing](10-network-routing.md)
- [Security](11-security.md)
