# Operations & Maintenance

---

## HTTPS with Caddy

VortexUI uses Caddy as its web server for automatic HTTPS:

- **Automatic Let's Encrypt** — certificates issued and renewed with zero config
- **HTTP → HTTPS redirect** — all traffic forced to HTTPS
- **SPA routing** — React frontend served correctly
- **Reverse proxy** — API requests forwarded to the Go backend
- **DoH endpoint** — `/dns-query` served if DoH is enabled

### Caddyfile Structure

```
{$VORTEX_DOMAIN} {
    encode gzip
    handle /api/* {
        reverse_proxy localhost:8080
    }
    handle /sub/* {
        reverse_proxy localhost:8080
    }
    handle /dns-query {
        reverse_proxy localhost:8053
    }
    handle {
        root * /opt/vortexui/web/dist
        try_files {path} /index.html
        file_server
    }
}
```

!!! tip
    If you need custom TLS settings (e.g. client certificate, specific cipher suites), edit `deploy/Caddyfile`.

---

## Prometheus Metrics

**Enable:** Set `VORTEX_METRICS_ENABLED=true` and `VORTEX_METRICS_LISTEN=:9090`.

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `vortex_users_total` | Gauge | Total user count by status |
| `vortex_traffic_bytes` | Counter | Traffic bytes (upload/download) |
| `vortex_connections_active` | Gauge | Current active connections |
| `vortex_nodes_status` | Gauge | Node status (1=online, 0=offline) |
| `vortex_node_cpu_percent` | Gauge | Node CPU utilization |
| `vortex_node_memory_percent` | Gauge | Node memory utilization |
| `vortex_orders_total` | Counter | Orders by status |
| `vortex_api_requests_total` | Counter | API requests by endpoint/status |
| `vortex_api_latency_seconds` | Histogram | API response latency |

### Grafana Dashboard

Import the ready-made dashboard from `deploy/grafana/vortexui-dashboard.json`:

1. Open Grafana → Dashboards → Import
2. Upload the JSON file or paste its contents
3. Select your Prometheus data source
4. Dashboard shows: traffic trends, node health, user activity, API performance

---

## Scaling & High Availability

### Horizontal Scaling

For large deployments:

| Component | Scaling Strategy |
|-----------|-----------------|
| Panel API | Run multiple instances behind a load balancer |
| Database | PostgreSQL with read replicas |
| Redis | Redis Cluster or Sentinel |
| Nodes | Add as many as needed (auto-managed) |
| Federation | Connect multiple panels for regional distribution |

### Cluster Mode

Run multiple panel instances:

1. All instances share the same PostgreSQL and Redis
2. Use a load balancer (Caddy, nginx, HAProxy) in front
3. SSE connections are sticky (session affinity recommended)
4. Background jobs use Redis-based distributed locks

### Database Considerations

| Concern | Solution |
|---------|----------|
| Connection pooling | pgBouncer in front of PostgreSQL |
| Read scaling | Read replicas for analytics queries |
| Write throughput | TimescaleDB hypertables for traffic data |
| Backup | pg_dump + WAL archiving for point-in-time recovery |

---

## Database Maintenance

### TimescaleDB Tips

Traffic data is stored in TimescaleDB hypertables for efficient time-series queries:

```sql
-- Check chunk sizes
SELECT show_chunks('traffic_data');

-- Set retention policy (keep 90 days)
SELECT add_retention_policy('traffic_data', INTERVAL '90 days');

-- Compress old chunks
SELECT add_compression_policy('traffic_data', INTERVAL '7 days');
```

### Routine Maintenance

```bash
# Vacuum and analyze
psql -U vortex -d vortex -c "VACUUM ANALYZE;"

# Check database size
psql -U vortex -d vortex -c "SELECT pg_size_pretty(pg_database_size('vortex'));"

# List largest tables
psql -U vortex -d vortex -c "
  SELECT relname, pg_size_pretty(pg_relation_size(oid))
  FROM pg_class
  ORDER BY pg_relation_size(oid) DESC
  LIMIT 10;
"
```

### Migrations

Migrations run automatically on panel startup. To run manually:

```bash
vortexui migrate
```

Check migration status:

```bash
vortexui migrate status
```

---

## Log Management

### Panel Logs

```bash
# Live logs
journalctl -u vortexui -f

# Last 100 lines
journalctl -u vortexui -n 100

# Filter by level
journalctl -u vortexui | grep ERROR
```

### Node Agent Logs

```bash
journalctl -u vortex-node -f
```

### Log Levels

Configure via environment:

```
VORTEX_LOG_LEVEL=info   # debug, info, warn, error
VORTEX_LOG_FORMAT=json  # json or text
```

### Log Rotation

Systemd journal handles rotation by default. Configure limits:

```ini
# /etc/systemd/journald.conf
SystemMaxUse=500M
MaxRetentionSec=30d
```

---

## Performance Tuning

### System Limits

```bash
# Increase file descriptor limit
echo "* soft nofile 65535" >> /etc/security/limits.conf
echo "* hard nofile 65535" >> /etc/security/limits.conf

# Increase for systemd service
# In /etc/systemd/system/vortexui.service
[Service]
LimitNOFILE=65535
```

### Network Tuning

```bash
# /etc/sysctl.conf
net.core.somaxconn = 65535
net.ipv4.tcp_max_syn_backlog = 65535
net.ipv4.ip_local_port_range = 1024 65535
net.ipv4.tcp_tw_reuse = 1
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
```

Apply: `sysctl -p`

### Redis Tuning

```
# /etc/redis/redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
```

### PostgreSQL Tuning

Key settings for VortexUI workloads:

```
# postgresql.conf
shared_buffers = 512MB          # 25% of RAM
effective_cache_size = 1536MB   # 75% of RAM
work_mem = 16MB
maintenance_work_mem = 128MB
random_page_cost = 1.1          # SSD
```

---

## Systemd Services

### Panel Service

```ini
# /etc/systemd/system/vortexui.service
[Unit]
Description=VortexUI Panel
After=postgresql.service redis.service

[Service]
Type=simple
User=vortex
WorkingDirectory=/opt/vortexui
ExecStart=/opt/vortexui/vortexui serve
Restart=always
RestartSec=5
LimitNOFILE=65535
EnvironmentFile=/opt/vortexui/.env

[Install]
WantedBy=multi-user.target
```

### Node Agent Service

```ini
# /etc/systemd/system/vortex-node.service
[Unit]
Description=VortexUI Node Agent
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/vortex-node
ExecStart=/opt/vortex-node/vortex-node
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

### Common Commands

```bash
sudo systemctl start vortexui
sudo systemctl stop vortexui
sudo systemctl restart vortexui
sudo systemctl status vortexui
sudo systemctl enable vortexui   # start on boot
```
