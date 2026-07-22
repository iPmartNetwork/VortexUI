# Protocol Groups

## Overview

**Protocol Groups** let you bundle multiple inbounds into a logical unit and serve them as a single subscription block. Instead of exposing every inbound individually, users receive an organized, prioritized set of connections tailored to their ISP, region, or use case.

> **Note:** Protocol Groups require VortexUI v1.4.0+ with sing-box nodes. Xray nodes receive a flat list fallback.

---

## Why Protocol Groups?

| Problem | Solution |
|---------|----------|
| Users connect to suboptimal protocols | Groups auto-order by quality score |
| ISPs block a protocol → all users break | Groups rotate to next inbound seamlessly |
| Admins can't A/B test protocol stacks | Groups allow per-ISP profiles |
| Subscription bloat (20+ configs) | Groups collapse into `urltest` blocks |

---

## Creating a Protocol Group

### Panel UI

1. Navigate to **Inbounds → Protocol Groups** tab
2. Click **+ New Group**
3. Name the group (e.g. "Iran MCI Optimized")
4. Drag inbounds from the left panel into the group
5. Set priority order (top = highest priority)
6. Save

### API

```bash
curl -X POST https://panel.example.com/api/protocol-groups \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Iran MCI Optimized",
    "inbound_ids": [12, 7, 3, 19],
    "isp_profile": "MCI",
    "enabled": true
  }'
```

**Response:**

```json
{
  "id": 5,
  "name": "Iran MCI Optimized",
  "inbound_ids": [12, 7, 3, 19],
  "isp_profile": "MCI",
  "quality_score": 87,
  "created_at": "2025-01-15T10:30:00Z"
}
```

---

## ISP Profiles

Each group can target a specific ISP. When a user's subscription request arrives, VortexUI detects the ISP via MaxMind GeoIP and serves the matching group.

| ISP Code | Target | Behavior |
|----------|--------|----------|
| `MCI` | Mobile Communication Iran | Aggressive anti-DPI, Reality preferred |
| `IRANCELL` | MTN Irancell | Moderate filtering, WS+TLS works |
| `TCI` | Telecommunication Company of Iran | Heavy filtering, gRPC+Reality required |
| `SHATEL` | Shatel | Moderate, standard TLS sufficient |
| `DEFAULT` | All other / unknown | Balanced mix of protocols |

Configure ISP profiles via:

```bash
curl -X PATCH https://panel.example.com/api/protocol-groups/5 \
  -H "Authorization: Bearer $TOKEN" \
  -d '{ "isp_profile": "MCI" }'
```

---

## Reordering Inbounds

Priority order determines which inbound clients try first within a `urltest` block.

### Drag & Drop (UI)

On the **ProtocolGroupsPanel**, drag inbound chips to reorder. Changes save automatically.

### API

```bash
curl -X POST https://panel.example.com/api/protocol-groups/5/reorder \
  -H "Authorization: Bearer $TOKEN" \
  -d '{ "inbound_ids": [19, 12, 7, 3] }'
```

---

## Subscription Rendering

When a client fetches `/sub/{token}`, VortexUI renders protocol groups as `urltest` blocks in the generated config:

```json
{
  "outbounds": [
    {
      "type": "urltest",
      "tag": "Iran-MCI-Optimized",
      "outbounds": ["vless-reality-mci", "trojan-ws-cdn", "vmess-grpc-relay"],
      "url": "https://cp.cloudflare.com/generate_204",
      "interval": "3m",
      "tolerance": 100
    }
  ]
}
```

Clash Meta users receive equivalent `proxy-group` with `url-test` strategy.

---

## Auto-Protocol Switching

Protocol Groups support **automatic switching** — when the active protocol's quality drops, clients seamlessly failover to the next priority inbound.

### Switch Events

Monitor switch events via the API:

```bash
curl -X POST https://panel.example.com/sub/{token}/switch \
  -H "Content-Type: application/json" \
  -d '{
    "from_inbound": 12,
    "to_inbound": 7,
    "reason": "timeout",
    "latency_ms": 3200
  }'
```

Switch events feed the adaptive ordering engine.

### Adaptive Ordering

VortexUI collects switch-event telemetry and **self-heals** group ordering:

1. Aggregate switch events per group over 24h window
2. Calculate success rate per inbound (connections / timeouts)
3. Re-rank inbounds: highest success rate → position 1
4. Apply new order to future subscription renders

> **Note:** Adaptive ordering runs every 6 hours. Manual reorder overrides it until next cycle.

---

## Frontend UI

The **ProtocolGroupsPanel** is accessible from the Inbounds page sidebar:

| Element | Description |
|---------|-------------|
| Group list | All groups with status badge and quality score |
| Inbound pool | Available inbounds (drag to add) |
| Reorder area | Drag chips to set priority |
| ISP selector | Dropdown for ISP profile assignment |
| Stats card | 24h switch events, avg latency, success rate |
| Sync button | Force re-render all subscriptions with new order |

---

## Best Practices

1. **One group per ISP** — avoid overlapping profiles
2. **3-5 inbounds per group** — too many increases subscription size and urltest overhead
3. **Mix transport types** — combine Reality + WS + gRPC for resilience
4. **Monitor switch events** — frequent switches indicate an inbound needs attention
5. **Let adaptive ordering stabilize** — avoid manual reorder during first 48h of a new group

---

## Related Pages

- [Inbounds Configuration](07-inbounds-config.md)
- [Anti-Censorship](09-anti-censorship.md)
- [Subscription Pipeline](12-subscription.md)
- [API Reference](17-api-reference.md)
