# Network & Policy

Configure how traffic flows through your infrastructure: outbounds, routing rules, CDN chains, and load balancers.

---

## Outbounds

Outbounds define where traffic goes after reaching a node.

**Network → Outbounds**

| Type | Description |
|------|-------------|
| Freedom | Direct connection to the internet (default) |
| Blackhole | Drop traffic (for blocking) |
| Proxy Chain | Forward to another proxy server |
| WARP+ | Route through Cloudflare WARP for clean IP |
| SOCKS/HTTP | Forward to upstream SOCKS or HTTP proxy |

### Creating an Outbound

1. Click **Add Outbound**
2. Choose type
3. Configure:

**Freedom:**
| Field | Description |
|-------|-------------|
| Domain strategy | AsIs, UseIP, UseIPv4, UseIPv6 |
| Redirect | Optional local redirect |

**Proxy Chain:**
| Field | Description |
|-------|-------------|
| Address | Upstream proxy address |
| Port | Upstream port |
| Protocol | VLESS, VMess, Trojan, SS |
| Credentials | Auth for upstream |

---

## Routing Packs

Pre-built collections of routing rules you can enable with one click.

**Network → Routing → Packs**

### Built-in Packs

| Pack | Purpose |
|------|---------|
| Block Ads | Block known ad domains |
| Block Malware | Block malware/phishing domains |
| Iran Direct | Route Iranian domains directly (bypass proxy) |
| China Direct | Route Chinese domains directly |
| Cloudflare WARP | Route specific domains through WARP |
| Block Torrents | Block BitTorrent traffic |
| Block Porn | Block adult content |

### Custom Routing Pack

1. Click **New Pack**
2. Name it
3. Add rules:

```json
{
  "rules": [
    {
      "type": "field",
      "domain": ["geosite:category-ads-all"],
      "outboundTag": "blackhole"
    },
    {
      "type": "field",
      "domain": ["geosite:ir"],
      "outboundTag": "direct"
    }
  ]
}
```

4. Save and assign to nodes

### Rule Matching

Rules are evaluated top-to-bottom. First match wins.

| Match Type | Example |
|------------|---------|
| domain | `example.com`, `geosite:google` |
| ip | `8.8.8.8`, `geoip:cn` |
| port | `443`, `1000-2000` |
| protocol | `bittorrent`, `http`, `tls` |
| source | Client IP ranges |

---

## CDN / Relay Chains

Multi-hop paths for extra resilience and censorship resistance.

**Network → Chains**

```
Client → CDN (Cloudflare) → Relay Node → Exit Node → Internet
```

### Chain Types

| Hop | Purpose |
|-----|---------|
| CDN | Cloudflare/other CDN fronting |
| Relay | Intermediate hop (hides exit node) |
| Worker | Cloudflare Worker as proxy |
| Exit | Final node to internet |

### Creating a Chain

1. Click **New Chain**
2. Add hops in order
3. For each hop, configure address, port, protocol
4. Assign chain to inbounds or users

> **Tip:** CDN chains hide your exit node's real IP from clients, useful when the exit IP is precious.

---

## Load Balancers

Distribute users across multiple nodes for reliability and performance.

**Network → Balancers**

### Strategies

| Strategy | Description | Best For |
|----------|-------------|----------|
| Round Robin | Rotate sequentially | Even distribution |
| Weighted | Distribute by weight | Nodes with different capacity |
| Least Connections | Fewest active connections | Long-lived connections |
| Latency | Lowest latency node | Performance-critical |

### Creating a Balancer

1. Click **New Balancer**
2. Choose strategy
3. Add member nodes
4. Set weights (for weighted strategy)
5. Configure health probing:

| Setting | Default |
|---------|---------|
| Probe interval | 30s |
| Probe timeout | 5s |
| Unhealthy threshold | 3 failures |
| Healthy threshold | 2 successes |

6. Assign to inbounds

### Health Probing

The balancer continuously checks node health:

```
Node A: 🟢 Healthy   (latency: 45ms)
Node B: 🟢 Healthy   (latency: 78ms)
Node C: 🔴 Unhealthy (3 consecutive failures)
```

Unhealthy nodes are removed from rotation until they recover.

---

## Subscription Hosts

Override addresses per-inbound for CDN fronting or custom domains.

**Network → Subscription Hosts**

| Field | Purpose |
|-------|---------|
| Address | Override server address (e.g., CDN domain) |
| Port | Override port |
| SNI | Override TLS server name |
| Host | Override HTTP host header |
| Path | Override transport path |

### Template Variables

Use variables that resolve per-user/per-node:

| Variable | Resolves To |
|----------|-------------|
| `{node.address}` | Node IP address |
| `{node.name}` | Node friendly name |
| `{node.region}` | Node region code |
| `{user.username}` | Username |
| `{user.token}` | Subscription token |

Example: `{user.username}.cdn.example.com`

---

## DNS Configuration

**Network → DNS**

Configure how nodes resolve domain names.

| Setting | Description |
|---------|-------------|
| Servers | DNS servers (1.1.1.1, 8.8.8.8) |
| Strategy | prefer_ipv4, prefer_ipv6, ipv4_only |
| Cache | Enable DNS caching |
| Hosts | Static host mappings |

### DNS over HTTPS/TLS

For encrypted DNS on nodes:

```
https://cloudflare-dns.com/dns-query
tls://1.1.1.1
```

---

## Federation

> **New in 1.3.0**

Connect multiple VortexUI panels for cross-panel coordination.

**Network → Federation**

### How It Works

- Panels register each other as peers
- Periodic health checks between peers
- User/node count synchronization
- Event broadcasting across panels

### Setup

1. On Panel A: **Federation → Add Peer**
2. Enter Panel B's address and shared secret
3. On Panel B: accept the peer request
4. Peers now sync automatically

### Use Cases

- **Geographic distribution**: Separate panels per region
- **High availability**: Backup panel takes over if primary fails
- **Load distribution**: Spread users across panel instances

---

## Best Practices

### For Anti-Censorship
- Use CDN chains to hide exit node IPs
- Enable routing packs for direct local traffic
- Combine with TLS Tricks (see [Security](08-security-administration.md))

### For Performance
- Use latency-based load balancing
- Place relay nodes close to users
- Enable DNS caching on nodes

### For Reliability
- Set up load balancers with health probing
- Configure auto-migration for node failures
- Use federation for multi-panel redundancy
