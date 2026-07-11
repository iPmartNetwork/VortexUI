# Troubleshooting & FAQ

Common issues, debugging tips, and frequently asked questions.

---

## Installation Issues

### Installer fails with "Go version too old"

VortexUI requires Go 1.26+.

```bash
# Remove old Go
sudo rm -rf /usr/local/go

# Install Go 1.26
wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
go version  # verify
```

### Database connection failed

```bash
# Check PostgreSQL is running
systemctl status postgresql

# Verify connection
psql -U vortex -d vortexui -h localhost

# Check connection string in .env
echo $VORTEX_DB_URL
```

### Port 8080 already in use

```bash
# Find what's using it
sudo lsof -i :8080

# Change panel port
echo "VORTEX_LISTEN=:8081" >> .env
```

### Docker containers won't start

```bash
# Check logs
docker compose logs

# Verify .env is complete
cat .env

# Rebuild
docker compose down
docker compose up -d --force-recreate
```

---

## Login Issues

### Can't login after installation

```bash
# Reset admin password
vortexui admin reset-password admin

# Or create new admin
vortexui admin create --username admin2 --password newpass --sudo
```

### "Invalid token" after restart

The JWT secret may have changed. Ensure `VORTEX_JWT_SECRET` is stable in `.env`.

### 2FA locked out

Use a recovery code, or reset via CLI:

```bash
vortexui admin disable-2fa admin
```

### Session expires too quickly

Increase session TTL in **Settings → Security** or:

```env
VORTEX_SESSION_TTL=72h
```

---

## Node Issues

### Node shows "Offline"

**Checklist:**

1. Network reachable?
   ```bash
   ping node-ip
   ```

2. Agent running?
   ```bash
   systemctl status vortex-node
   ```

3. Port open?
   ```bash
   telnet node-ip 62050
   ```

4. Check agent logs:
   ```bash
   journalctl -u vortex-node -f
   ```

5. Firewall allows the port?
   ```bash
   sudo ufw allow 62050
   ```

### Config push fails

```bash
# Check agent logs for the error
journalctl -u vortex-node -n 100

# Verify core binary exists
which xray  # or sing-box

# Restart agent
systemctl restart vortex-node
```

### mTLS certificate errors

Certificates may have expired or mismatched. Re-enroll the node:

1. Delete node from panel
2. Re-add via enrollment wizard
3. Run new install command on node

### High node latency

```bash
# Check network path
traceroute panel-ip

# Check node load
top

# Check for packet loss
mtr panel-ip
```

---

## Connection Issues

### Users can't connect

**Checklist:**

1. Inbound port open on node?
   ```bash
   telnet node-ip 443
   ```

2. User not expired/limited?
   - Check user status in panel

3. Protocol settings correct?
   - Verify in client app

4. Firewall on node?
   ```bash
   sudo ufw status
   ```

### Subscription URL returns 404

1. User has at least one inbound assigned?
2. User not expired or suspended?
3. Domain resolves to panel?
   ```bash
   dig panel.example.com
   ```

### Slow speeds

1. Check node bandwidth/load
2. Enable BBR on node
3. Try different transport (WS vs gRPC)
4. Check for ISP throttling → use TLS Tricks

### Reality connection fails

1. Verify dest domain is reachable from node
2. Check SNI matches server_names
3. Ensure short IDs are configured
4. Use Reality Scanner for better SNI

---

## Performance Issues

### Panel is slow

1. Check database performance:
   ```sql
   SELECT * FROM pg_stat_activity;
   ```

2. Increase DB pool:
   ```env
   VORTEX_DB_MAX_CONNS=50
   ```

3. Ensure Redis is running:
   ```bash
   redis-cli ping
   ```

4. Check disk I/O:
   ```bash
   iostat -x 1
   ```

### High memory usage

1. Set retention policy for old traffic data
2. Configure Redis maxmemory
3. Check for memory leaks in logs

### Database growing too fast

Add retention policy:

```sql
SELECT add_retention_policy('traffic_stats', INTERVAL '90 days');
```

---

## Certificate / HTTPS Issues

### ACME certificate not issuing

> **1.3.0+**

1. Verify Cloudflare token has DNS edit permission
2. Check domain is in Cloudflare
3. Review ACME events in Audit Log
4. Check logs:
   ```bash
   journalctl -u vortexui | grep acme
   ```

### Certificate expired

With auto-renewal, this shouldn't happen. If it does:

```bash
vortexui doctor  # diagnose
# Force renewal via Settings → ACME → Renew
```

---

## Frequently Asked Questions

### General

**Q: What's the difference between VortexUI and 3x-ui/Marzban?**

A: VortexUI is user-centric (one identity across all nodes), supports dual cores (Xray + sing-box), has push-based traffic (not polling), and includes a full reseller platform, federation, and anti-censorship suite. See [comparison](01-introduction.md#comparison-with-other-panels).

**Q: Can I migrate from 3x-ui or Marzban?**

A: Yes. Use **Users → Import** to migrate from 3x-ui database or Marzban export.

**Q: Does it support IPv6?**

A: Yes, both panel and nodes support IPv6.

**Q: How many nodes can one panel manage?**

A: Dozens comfortably. For hundreds, use federation.

### Cores & Protocols

**Q: Xray or sing-box — which should I use?**

A: Xray for VLESS+Reality (most common). sing-box for Hysteria2, TUIC, WireGuard. You can mix cores across nodes.

**Q: What's the best protocol for censorship resistance?**

A: VLESS + Reality is the gold standard. For lossy networks, Hysteria2.

**Q: Can one user use multiple protocols?**

A: Yes! That's the user-centric model. One user, all assigned inbounds.

### Reseller Platform

**Q: How do reseller wallets work?**

A: Resellers have credits (traffic + user). Credits are consumed as they create users or users consume traffic. See [Plans & Payments](09-plans-payments.md).

**Q: Can resellers have their own branding?**

A: Yes, full whitelabel — logo, colors, portal branding. See [Settings → Branding](11-settings-backup.md#branding-whitelabel).

**Q: Can resellers create sub-resellers?**

A: Yes, with inherited scope (can't exceed parent's limits).

### Security

**Q: Is 2FA required?**

A: Not required by default, but strongly recommended for sudo. Can be enforced in settings.

**Q: How does probing protection work?**

A: It detects active GFW/TSPU probes and can block, honeypot, or log them. See [Security](08-security-administration.md#probing-protection).

**Q: What is account-sharing guard?**

A: It monitors concurrent IPs per user and can warn, limit, or disconnect on sharing detection.

### Backup & Recovery

**Q: How often should I backup?**

A: Daily minimum. Auto-backup to Telegram + S3 recommended.

**Q: Can I restore to a different server?**

A: Yes. Install VortexUI on new host, then `vortexui restore backup.sql.gz`.

### Updates

**Q: Will updates break my setup?**

A: Migrations run automatically and are backward-compatible. Always backup before major updates.

**Q: How do I update from 1.2.x to 1.3.x?**

A: See the [Changelog](16-changelog.md) and migration notes. Generally `vortexui update` handles it.

---

## Getting Help

### Self-Diagnosis

```bash
vortexui doctor
```

### Logs

```bash
# Panel
journalctl -u vortexui -f

# Node
journalctl -u vortex-node -f

# Docker
docker compose logs -f
```

### Debug Mode

```env
VORTEX_LOG_LEVEL=debug
```

### Community Support

- 💬 [Telegram](https://t.me/vortex_ui)
- 🐛 [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues)
- 💡 [GitHub Discussions](https://github.com/iPmartNetwork/VortexUI/discussions)

### Reporting Bugs

Include:
1. VortexUI version (`vortexui version`)
2. OS and architecture
3. Relevant logs
4. Steps to reproduce
5. Output of `vortexui doctor`
