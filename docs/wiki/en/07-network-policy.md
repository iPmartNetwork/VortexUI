# 7. Network Policy

!!! tip "Tip"
    Common pattern: `geosite:ir` and `geoip:ir` → direct, everything else → proxy.

---

## Outbounds

**Outbounds** define the egress path for traffic after the inbound.

### Types

| Tag | Role |
|-----|------|
| `freedom` | Direct — no proxy |
| `blackhole` | Drop |
| `dns` | Internal resolver |
| `proxy` | Chain to another upstream |
| `warp` | Cloudflare WARP+ |

### CRUD

- **Outbounds → Add** — JSON editor + share link import
- Link to inbound/routing rule

### Example: Proxy chain

```json
{
  "tag": "chain-de",
  "protocol": "vless",
  "settings": {
    "vnext": [{
      "address": "upstream.example.com",
      "port": 443,
      "users": [{"id": "uuid", "encryption": "none", "flow": "xtls-rprx-vision"}]
    }]
  }
}
```

---

## Routing

**Routing → Add Rule**

| Matcher | Example |
|---------|---------|
| Domain | `geosite:ir`, `domain:google.com` |
| IP | `geoip:ir`, `192.168.0.0/16` |
| Port | `80,443` |
| Protocol | `tcp,udp` |
| Inbound tag | `vless-in` |

### Common Iran pattern

```
geosite:ir  → outbound: direct (freedom)
geoip:ir    → outbound: direct
default     → outbound: proxy
```

---

## Balancers

**Balancers → Add**

| Strategy | Behavior |
|----------|----------|
| `random` | Random selection |
| `roundRobin` | Round-robin |
| `leastPing` | Lowest ping |
| `leastLoad` | Lowest load |

### Observatory

- Health probes on outbounds
- Automatically removes unhealthy from pool

---

## Evasion Profiles

**Evasion** — anti-DPI presets:

| Preset | Content |
|--------|---------|
| Iran (Fragment + Chrome) | TLS fragment + Chrome fingerprint |
| China (Mux + Random) | h2mux + randomized FP |
| Russia (Fragment + Firefox) | Short fragment |

Linked to inbound — fragment, mux, fingerprint in one place.

---

## WARP+ Integration

WARP-type outbound for Cloudflare tunneling — useful for bypass or privacy layer.

---

## Config Templates

**Settings → Subscription Template**

- Customize Clash/sing-box output
- Default rules, proxy-groups, DNS
