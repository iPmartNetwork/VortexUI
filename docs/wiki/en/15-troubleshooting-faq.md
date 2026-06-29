# Troubleshooting & FAQ

---

## Common Issues

### Connection Refused

**Symptom:** Clients cannot connect to the proxy.

**Check:**

1. Is the node online? Check **Nodes** page for status
2. Is the inbound port open? `ss -tlnp | grep <port>`
3. Is the firewall allowing traffic? `ufw status` or `iptables -L`
4. Is the core running? Check node logs for errors
5. Is the protocol/transport correct in the client config?

!!! tip
    Run `vortexui doctor` to check all components at once.

### TLS Errors

**Symptom:** `tls: handshake failure` or `certificate verify failed`

**Check:**

1. **REALITY:** Is the dest/SNI domain reachable from the node? Try `curl -I https://dest-domain`
2. **TLS:** Is the certificate valid? Check expiry with `openssl s_client -connect host:443`
3. **CDN:** Is the Cloudflare SSL mode set to "Full (Strict)"?
4. **Client:** Is the SNI field matching the server config?
5. **Fragment:** If using TLS fragment, try disabling it temporarily

### Node Disconnected

**Symptom:** Node shows "Offline" in the panel.

**Check:**

1. Is the node server running? SSH and check: `systemctl status vortex-node`
2. Network connectivity: `ping <panel-ip>` from the node
3. mTLS certificates: check for expiry or mismatch
4. Firewall: is the gRPC port (default 9090) open between panel and node?
5. Check node agent logs: `journalctl -u vortex-node -n 50`

### Subscription Empty

**Symptom:** Client receives empty subscription (no configs).

**Check:**

1. Does the user have inbounds assigned? Check user detail → Inbounds
2. Are the assigned inbounds on online nodes?
3. Are subscription hosts configured correctly (if using)?
4. Is the subscription token valid (not revoked)?
5. Check the raw response: `curl https://panel.example.com/sub/<token>`

### High CPU on Node

**Symptom:** Node CPU stays above 90%.

**Check:**

1. Too many users? Check active connections count
2. Is auto-migration configured? It should move users away
3. Core process: `top -p $(pgrep xray)` or `pgrep sing-box`
4. Consider adding more nodes and enabling load balancing

### Database Connection Issues

**Symptom:** Panel returns 500 errors, logs show PostgreSQL connection errors.

**Check:**

1. Is PostgreSQL running? `systemctl status postgresql`
2. Connection string correct in `.env`?
3. Max connections exhausted? `SELECT count(*) FROM pg_stat_activity;`
4. Consider adding pgBouncer for connection pooling

---

## Debug Tips

### `vortexui doctor`

Run comprehensive diagnostics:

```bash
vortexui doctor
```

Checks:

- ✅ PostgreSQL connection + schema version
- ✅ Redis connection + latency
- ✅ Node gRPC connectivity (per node)
- ✅ Certificate validity
- ✅ Port availability
- ✅ DNS resolution
- ✅ Disk space
- ✅ Core binary presence and version

### Health Endpoint

```bash
curl https://panel.example.com/api/health
```

Returns component status:

```json
{
  "status": "healthy",
  "components": {
    "database": "ok",
    "redis": "ok",
    "nodes": { "online": 3, "offline": 0 }
  },
  "version": "1.2.7"
}
```

### Enable Debug Logging

```bash
VORTEX_LOG_LEVEL=debug systemctl restart vortexui
```

Then watch logs:

```bash
journalctl -u vortexui -f
```

!!! warning
    Debug logging is verbose. Disable after troubleshooting to avoid disk fill.

### Test Subscription Manually

```bash
# Base64 format
curl -s https://panel.example.com/sub/<token>

# Clash format
curl -s "https://panel.example.com/sub/<token>?format=clash"

# With User-Agent detection
curl -s -A "clash-meta" https://panel.example.com/sub/<token>
```

### Check Node Health via API

```bash
curl -H "Authorization: Bearer <token>" \
  https://panel.example.com/api/nodes/<id>/health
```

---

## FAQ

### How do I reset the admin password?

```bash
vortexui admin reset-password --username admin
```

Or via the interactive menu:

```bash
vortexui
# Select option 7 → Reset password
```

### How do I migrate from 3x-ui?

1. Export your 3x-ui database (`x-ui.db`)
2. In VortexUI: **Users → Import → 3x-ui**
3. Upload the database file
4. Map inbounds (VortexUI assigns users to matching inbounds)
5. Review and confirm

### How do I migrate from Marzban?

1. In VortexUI: **Users → Import → Marzban**
2. Provide database connection string or export file
3. Users, traffic data, and expiry dates are preserved
4. Inbound mapping is done automatically where possible

### Can I run panel and node on the same server?

Yes — use the **Local Node** feature. The proxy core runs in-process alongside the panel. No agent needed.

### How does subscription auto-detection work?

The panel inspects the client's `User-Agent` header:

- Contains "clash" → Clash YAML
- Contains "sing-box" → sing-box JSON
- Contains "outline" → Outline `ss://` links
- Otherwise → base64 encoded share links

Override with `?format=` query parameter.

### How do I add a domain for HTTPS?

1. Point your domain's DNS A record to your server IP
2. Set `VORTEX_DOMAIN=your-domain.com` in `.env`
3. Restart the panel — Caddy auto-issues a certificate

### How do I backup and restore?

**Backup:**
```bash
vortexui backup
# or automatic: set VORTEX_BACKUP_CRON="0 3 * * *"
```

**Restore:**
```bash
vortexui restore /path/to/backup.tar.gz
```

### What's the difference between allocated and consumed quota mode?

- **Allocated:** Pool decreases when you assign data limits to users (sum of all user limits counts)
- **Consumed:** Pool decreases only when users actually use traffic

Use allocated for pre-sold packages. Use consumed for pay-per-use.

### How do I configure per-reseller payments?

1. Create a reseller admin with appropriate role
2. Reseller logs in → **Reseller Account → Payment Configuration**
3. Sets their own card number, crypto addresses, or ZarinPal merchant
4. Their users see these payment options in the shop

### Why is my node showing "Unhealthy"?

The node fails health checks. Common causes:

- High CPU (>90%) or RAM (>90%)
- Packet loss >10%
- Core process crashed (auto-restart should handle this)
- Certificate issue

Check: **Nodes → node → Health** for specific failure reasons.

### How do I use Cloudflare with VortexUI?

1. Point domain to your server via Cloudflare (orange cloud)
2. Set Cloudflare SSL mode to **Full (Strict)**
3. Use WebSocket transport (required for Cloudflare proxying)
4. Configure subscription hosts to advertise the CDN domain
5. Users connect to Cloudflare → Cloudflare forwards to your node

### How do I enable the self-service shop?

1. Create plans (**Plans → New Plan**)
2. Configure payment methods (**Settings → Payment Configuration**)
3. Share the portal link with users: `/portal/login`
4. Users can also access the shop via `/sub/{token}/shop`

### What happens when a reseller runs out of quota?

- **User credits exhausted:** Cannot create new users
- **Traffic credits exhausted:** Existing users continue until individually limited (consumed mode) or cannot assign more data (allocated mode)
- Auto-suspend can be configured to disable the reseller entirely
