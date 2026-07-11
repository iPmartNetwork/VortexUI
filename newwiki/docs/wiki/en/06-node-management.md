# Node Management

Nodes are servers running proxy cores (Xray-core or sing-box) that handle user traffic.

---

## Node Types

### Local Node

- Runs **in-process** with the panel
- No separate agent needed
- Perfect for single-server setups
- Limited to one per panel

### Remote Node

- Runs on a **separate server**
- Communicates via gRPC + mTLS
- Supports fleet of many nodes
- Health monitoring + auto-migration

---

## Node List

**Sidebar → Nodes**

| Column | Description |
|--------|-------------|
| Name | Node identifier |
| Address | IP or hostname |
| Core | Xray or sing-box |
| Status | Online 🟢, Offline 🔴, Migrating 🟡 |
| Users | Connected user count |
| Traffic | Current bandwidth |
| Health | CPU, RAM, latency |

### Status Indicators

- 🟢 **Online**: Node is healthy and serving traffic
- 🟡 **Migrating**: Users being moved to/from this node
- 🔴 **Offline**: Node unreachable
- ⚠️ **Degraded**: High latency or resource usage

---

## Adding a Remote Node

### Step 1: Create in Panel

**Nodes → Add Node → Remote**

| Field | Description |
|-------|-------------|
| Name | Friendly name (e.g., "Frankfurt-01") |
| Address | Server IP or hostname |
| API Port | Agent API port (default: 62050) |
| Core | Xray or sing-box |
| Region | Geographic region (for routing) |
| Tags | Custom tags for filtering |

### Step 2: Install Agent

After creating, copy the install command and run on the node server:

```bash
bash <(curl -Ls https://panel.example.com/node-install.sh) --token=xyz123
```

Or manually:

```bash
# Download agent
curl -Lo vortex-node https://github.com/iPmartNetwork/VortexUI/releases/latest/download/vortex-node-linux-amd64
chmod +x vortex-node

# Configure
cat > /etc/vortex-node/config.yaml <<EOF
panel:
  address: https://panel.example.com
  token: xyz123
core:
  type: xray
  path: /usr/local/bin/xray
listen:
  grpc: 0.0.0.0:62050
EOF

# Start
./vortex-node serve
```

### Step 3: Verify Connection

Back in the panel, the node should show 🟢 Online within 30 seconds.

---

## Node Enrollment Wizard

For easier setup, use the 4-step wizard:

1. **Basic Info**: Name, region, expected load
2. **Connection**: Address, ports, firewall check
3. **Certificate Exchange**: Automatic mTLS setup
4. **Verify**: Test connection and create

The wizard handles certificate generation and validates connectivity before finalizing.

---

## Inbound Management

Each node has its own set of inbounds (protocol listeners).

**Nodes → [Node Name] → Inbounds**

### Adding an Inbound

| Field | Options |
|-------|---------|
| Protocol | VLESS, VMess, Trojan, Shadowsocks, Hysteria2, TUIC, etc. |
| Port | Listen port (or range: 10000-10100) |
| Transport | TCP, WS, gRPC, HTTPUpgrade, QUIC |
| Security | None, TLS, Reality |

### Protocol-Specific Settings

**VLESS + Reality:**
- Dest (target server)
- Server Names (SNI list)
- Private Key (auto-generated)
- Short IDs (client auth)
- Flow (xtls-rprx-vision for TCP)

**VMess + WebSocket:**
- Path (/ws)
- Host header
- Early data size

**Hysteria2:**
- Up/Down bandwidth
- Obfs type + password

---

## Health Monitoring

Each node reports health metrics every 10 seconds:

| Metric | Description |
|--------|-------------|
| CPU | Current utilization % |
| RAM | Used / Total |
| Disk | Used / Total |
| Network | Bytes in/out per second |
| Connections | Active tunnel count |
| Latency | Panel ↔ Node round-trip time |

### Health Alerts

Configure thresholds at **Settings → Notifications**:

- CPU > 90% for 5 minutes
- RAM > 95%
- Node offline for 30 seconds
- Latency > 500ms

---

## Auto-Migration

When a node becomes unhealthy, users are automatically moved to healthy nodes.

### Enable Auto-Migration

**Settings → Nodes → Auto-Migration**

| Setting | Description |
|---------|-------------|
| Enable | Turn on auto-migration |
| Threshold | Offline time before migration (e.g., 60s) |
| Target Selection | Least loaded, same region, random |
| Migrate Back | Return users when node recovers |

### Migration Process

1. Node goes offline
2. Wait for threshold (e.g., 60 seconds)
3. Select target nodes based on strategy
4. Push new configs to target nodes
5. Update subscription endpoints
6. Notify affected users (optional)
7. When original node recovers, migrate back (if enabled)

---

## Cloudflare DNS Automation

Automatically manage DNS records for your nodes.

### Setup

1. **Settings → Integrations → Cloudflare**
2. Add API token with DNS edit permissions
3. Select zone (domain)

### Usage

1. When adding a node, enable "Auto DNS"
2. Enter subdomain (e.g., "node1")
3. Panel creates A record pointing to node IP
4. On node IP change, record updates automatically

---

## Node Configuration

### Xray-core Settings

| Setting | Description |
|---------|-------------|
| Log Level | Debug, info, warning, error |
| DNS | DNS servers for outbound resolution |
| Routing | Custom routing rules |
| Policy | Timeout, buffer size settings |

### sing-box Settings

| Setting | Description |
|---------|-------------|
| Log Level | trace, debug, info, warn, error |
| DNS | DNS server configuration |
| Route | Route rules (similar to Xray routing) |
| Experimental | Clash API, cache settings |

---

## Load Balancing

Distribute users across multiple nodes.

**Network → Balancers → Add Balancer**

### Strategies

| Strategy | Description |
|----------|-------------|
| Round Robin | Rotate through nodes sequentially |
| Weighted | Distribute based on assigned weights |
| Least Connections | Send to node with fewest active connections |
| Latency | Send to node with lowest latency |

### Health Probing

Balancer continuously probes nodes:
- Interval: 30 seconds
- Timeout: 5 seconds
- Unhealthy after: 3 failures
- Healthy after: 2 successes

---

## Node Fleet Operations

### Bulk Actions

Select multiple nodes, then:
- **Restart Core**: Restart proxy cores
- **Push Config**: Force config sync
- **Enable/Disable**: Toggle node availability
- **Delete**: Remove nodes from panel

### Config Sync

When you change inbound settings:
1. Panel generates new core config
2. Pushes to node via gRPC
3. Node hot-reloads (no restart needed for most changes)
4. Restart triggered only for port changes

---

## Troubleshooting

### Node Shows Offline

1. Check network connectivity: `ping node-ip`
2. Verify agent is running: `systemctl status vortex-node`
3. Check agent logs: `journalctl -u vortex-node -f`
4. Verify firewall allows port 62050
5. Check mTLS certificates haven't expired

### Config Push Fails

1. Check agent logs for errors
2. Verify core binary is installed
3. Check disk space for config files
4. Restart agent: `systemctl restart vortex-node`

### High Latency

1. Check network route: `traceroute panel-ip`
2. Verify no bandwidth throttling
3. Check for DDoS or high traffic
4. Consider node location relative to panel
