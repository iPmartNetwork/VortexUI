# 15. Troubleshooting & FAQ

!!! tip "Tip"
    Start with **`vortexui logs`** and **`/api/health`** — most issues are JWT, DB, or firewall.

---

## Common Issues

### Panel won't start

```bash
vortexui status
vortexui logs
curl http://127.0.0.1:8080/api/health
```

| Cause | Solution |
|-------|----------|
| Empty JWT secret | `deploy/.env` → `JWT_SECRET=$(openssl rand -hex 32)` |
| DB down | `docker compose ps` — restart `db` |
| Port in use | `ss -tlnp \| grep 8080` |

---

### HTTPS / Let's Encrypt failure

| Cause | Solution |
|-------|----------|
| Wrong DNS | A record to server IP |
| Port 80 closed | firewall: `ufw allow 80,443` |
| LE rate limit | wait 1h or staging test |

---

### Node offline / red

| Cause | Solution |
|-------|----------|
| Agent down | `systemctl status vortex-node` |
| mTLS mismatch | regenerate certs, SAN includes IP |
| Firewall | port 50051 gRPC open |
| Core crash | Nodes → Logs |

---

### User can't connect

| Check | |
|-------|---|
| Inbound active? | Nodes → Inbounds |
| User status `active`? | Users |
| Expired / limited? | User detail |
| Inbound port open? | `ufw` / cloud security group |
| REALITY keys match? | regenerate + new sub |

---

### Traffic not recorded

| Cause | Solution |
|-------|----------|
| Core API port | `VORTEX_CORE_API_PORT=10085` |
| Stats disabled in core | panel config renders stats |
| Redis down | restart redis |

---

### Empty subscription

| Cause | Solution |
|-------|----------|
| No inbound assigned | Edit user → select inbounds |
| Node down | fix node first |
| Wrong endpoint | set Custom Endpoint |

---

### SSE / live update not working

| Cause | Solution |
|-------|----------|
| Caddy buffering | default OK — check proxy timeout |
| Token expired | re-login |
| Ad blocker | disable for panel domain |

---

## FAQ

### How is VortexUI different from 3x-ui?

**User-centric** model, push delta traffic, full outbound/routing/balancer, audit, API token, advanced failover.

### Does it support SQLite?

No — **PostgreSQL + TimescaleDB** (production-grade, traffic time series).

### How many nodes are supported?

Unlimited — each node has a separate agent or one local node.

### sing-box or xray?

Per node — Hysteria2/TUIC only on sing-box; REALITY on both.

### Import from Marzban?

Yes — Users → Import.

### Account sharing?

Device limit + online IP guard + optional autolimit.

### Sell with ZarinPal?

Plans + ZarinPal gateway — [Chapter 9](./09-plans-payments.md).

### Backup before update?

**Always** — `vortexui update` is safe but backup is recommended.

### License?

GPL-3.0 — derivatives must be open source.

---

## Reporting Bugs

1. [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues)
2. Version: `vortexui settings` or sidebar
3. Logs: `vortexui logs` (no secrets)
4. [SECURITY.md](https://github.com/iPmartNetwork/VortexUI/blob/master/SECURITY.md) for vulnerabilities

---

## Community

- ⭐ Star on GitHub
- [Contributing](https://github.com/iPmartNetwork/VortexUI/blob/master/CONTRIBUTING.md)
- [Changelog](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md)
