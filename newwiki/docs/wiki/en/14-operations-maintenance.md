# Operations & Maintenance

Running VortexUI in production: HTTPS, monitoring, scaling, database maintenance, and performance tuning.

---

## HTTPS Setup

### Option 1: Built-in ACME (Recommended)

> **New in 1.3.0**

Automatic Let's Encrypt via Cloudflare DNS-01:

```env
VORTEX_ACME_EMAIL=admin@example.com
CLOUDFLARE_API_TOKEN=cf_token_with_dns_edit
VORTEX_DOMAIN=panel.example.com
```

Certificates issue and renew automatically. No port 80 needed.

### Option 2: Caddy (Auto HTTPS)

The default Docker setup includes Caddy, which handles HTTPS automatically:

```
panel.example.com {
    reverse_proxy api:8080
}
```

### Option 3: Manual Certificate

```bash
sudo certbot certonly --standalone -d panel.example.com
```

Then configure in Nginx (see [Manual Install](02-installation.md#native-build)).

---

## Monitoring with Prometheus

### Enable Metrics

```env
VORTEX_METRICS_ENABLED=true
VORTEX_METRICS_LISTEN=:9090
```

### Available Metrics

| Metric | Description |
|--------|-------------|
| vortex_users_total | Total users by status |
| vortex_nodes_online | Online node count |
| vortex_traffic_bytes | Traffic counters |
| vortex_connections_active | Active connections |
| vortex_api_requests_total | API request count |
| vortex_api_duration_seconds | Request latency |
| vortex_node_cpu | Per-node CPU |
| vortex_node_ram | Per-node RAM |

### Prometheus Config

```yaml
scrape_configs:
  - job_name: vortexui
    static_configs:
      - targets: ['panel.example.com:9090']
    metrics_path: /metrics
```

### Grafana Dashboard

Import the ready-made dashboard:

1. Grafana → Import
2. Use dashboard ID or paste JSON from `docs/grafana-dashboard.json`
3. Select Prometheus data source

Includes panels for:
- User growth
- Traffic trends
- Node health
- API performance

---

## Database Maintenance

### PostgreSQL + TimescaleDB

VortexUI uses TimescaleDB for time-series traffic data.

### Backup

```bash
# Manual
vortexui backup

# Or direct pg_dump
pg_dump -U vortex vortexui | gzip > backup.sql.gz
```

### Vacuum & Analyze

TimescaleDB auto-vacuums, but for manual optimization:

```sql
VACUUM ANALYZE;
```

### Data Retention

Configure how long to keep detailed traffic data:

```sql
SELECT add_retention_policy('traffic_stats', INTERVAL '90 days');
```

Older data is automatically dropped, keeping the database lean.

### Continuous Aggregates

TimescaleDB pre-computes hourly/daily rollups for fast analytics:

```sql
-- Hourly traffic aggregate (auto-maintained)
SELECT * FROM traffic_hourly WHERE time > now() - INTERVAL '24 hours';
```

---

## Scaling

### Vertical Scaling

For most deployments, add resources to the panel host:

| Users | RAM | CPU |
|-------|-----|-----|
| < 1,000 | 2 GB | 2 vCPU |
| 1,000-5,000 | 4 GB | 4 vCPU |
| 5,000-20,000 | 8 GB | 8 vCPU |
| 20,000+ | 16 GB+ | 8+ vCPU |

### Horizontal Scaling (Nodes)

Add more nodes to handle traffic:
- Each node handles its own proxy traffic
- Panel only manages coordination
- Use load balancers to distribute users

### Federation for Scale

> **New in 1.3.0**

Split load across multiple panels:
- One panel per region
- Peers sync counts
- Reduces single-panel load

### Database Scaling

For very large deployments:
- Move PostgreSQL to dedicated host
- Use connection pooling (PgBouncer)
- Consider read replicas for analytics

---

## Performance Tuning

### Panel Performance

```env
# Increase DB connection pool
VORTEX_DB_MAX_CONNS=50

# Redis for session caching
VORTEX_REDIS_URL=redis://localhost:6379/0

# Worker concurrency
VORTEX_WORKERS=4
```

### Node Performance

Enable BBR congestion control on nodes:

```bash
echo "net.core.default_qdisc=fq" >> /etc/sysctl.conf
echo "net.ipv4.tcp_congestion_control=bbr" >> /etc/sysctl.conf
sysctl -p
```

Increase file descriptor limits:

```bash
echo "* soft nofile 1000000" >> /etc/security/limits.conf
echo "* hard nofile 1000000" >> /etc/security/limits.conf
```

### Redis Optimization

```
maxmemory 512mb
maxmemory-policy allkeys-lru
```

---

## Log Management

### Panel Logs

```bash
# Systemd
journalctl -u vortexui -f

# Docker
docker compose logs -f api
```

### Log Levels

```env
VORTEX_LOG_LEVEL=info  # debug, info, warn, error
```

### Structured Logging

Logs are JSON-formatted for parsing:

```json
{
  "level": "info",
  "time": "2026-01-15T14:32:05Z",
  "msg": "user created",
  "user_id": "abc123",
  "actor": "admin"
}
```

### Log Rotation

For file logs, configure logrotate:

```
/var/log/vortexui/*.log {
    daily
    rotate 14
    compress
    missingok
}
```

---

## Health Checks

### Endpoint Monitoring

```bash
curl https://panel.example.com/api/health
```

Set up external monitoring:
- UptimeRobot
- Prometheus Blackbox Exporter
- Healthchecks.io

### Diagnostics

```bash
vortexui doctor
```

Checks:
- ✅ Database connection
- ✅ Redis connection
- ✅ Node reachability
- ✅ Port availability
- ✅ Certificate validity
- ✅ Disk space

---

## Upgrade Procedures

### Standard Upgrade

```bash
vortexui update
```

### Zero-Downtime (Advanced)

For critical deployments:
1. Deploy new version to staging
2. Run migrations on replica
3. Switch traffic via load balancer
4. Verify, then decommission old

### Rollback

```bash
# Restore from backup
vortexui restore backup-before-upgrade.sql.gz

# Or pin to previous version
git checkout v1.3.0
go build -o vortexui ./cmd/panel
```

---

## Security Maintenance

### Regular Tasks

| Task | Frequency |
|------|-----------|
| Review audit log | Weekly |
| Rotate API tokens | Quarterly |
| Update dependencies | Monthly |
| Test backups | Monthly |
| Check certificate expiry | Automated |
| Review admin access | Quarterly |

### Certificate Renewal

With built-in ACME, renewal is automatic at T-30 days. Verify:

```bash
vortexui doctor  # checks cert validity
```

---

## Disaster Recovery

### Backup Strategy (3-2-1)

- **3** copies of data
- **2** different media types
- **1** off-site (S3, Telegram)

### Recovery Steps

1. Provision new host
2. Install VortexUI
3. Restore database:
   ```bash
   vortexui restore latest-backup.sql.gz
   ```
4. Verify nodes reconnect
5. Test user connectivity

### RTO/RPO Targets

| Metric | Target |
|--------|--------|
| RPO (data loss) | ≤ 24h (with daily backup) |
| RTO (recovery time) | ≤ 1h |

---

## Capacity Planning

### Monitoring Growth

Track these metrics over time:
- User count growth rate
- Traffic per user
- Peak concurrent connections
- Database size growth

### When to Scale

| Signal | Action |
|--------|--------|
| CPU > 80% sustained | Add vCPU |
| RAM > 90% | Add RAM |
| DB > 50% disk | Add disk / retention policy |
| Node CPU > 80% | Add nodes |
| Panel slow | Consider federation |
