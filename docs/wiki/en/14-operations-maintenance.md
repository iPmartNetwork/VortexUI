<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [FA](../fa/14-operations-maintenance.md) · [AR](../ar/14-operations-maintenance.md) · [TR](../tr/14-operations-maintenance.md)

</div>

<div>

# 14. Operations & Maintenance

[← Protocols](./13-protocols-config.md) · [Index](./README.md) · [Next: Troubleshooting →](./15-troubleshooting-faq.md)

> [!TIP]
> After install run `vortexui status` and `curl .../api/health` as a sanity check.

---

## `vortexui` Console

```bash
vortexui              # interactive menu
vortexui start        # start stack
vortexui stop         # stop
vortexui restart      # restart
vortexui status       # status
vortexui logs         # tail logs
vortexui update       # git pull + rebuild
vortexui admin        # create admin
vortexui settings     # URL and settings
vortexui uninstall    # remove (with confirmation)
```

Install path: `VORTEXUI_DIR` (default `/opt/vortexui`)

---

## HTTPS / SSL

### Docker (Caddy)

`deploy/.env`:

```env
SITE_ADDRESS=panel.example.com
ACME_EMAIL=admin@example.com
```

- Ports 80 + 443 open
- DNS A record to server
- Cert in `caddy-data` volume

Change domain:

```bash
vortexui   # → option 9) Domain / SSL
```

### HTTP only

```env
SITE_ADDRESS=:8080
```

---

## Cluster Mode (HA)

Multiple panel instances with shared DB — for high availability. Details in env and deploy docs.

---

## Prometheus / Grafana

Panel and node metrics for Prometheus scraping. Sample Grafana dashboard in releases.

---

## Auto-Update

- Panel binary from GitHub releases
- Core binaries (xray/sing-box) from upstream
- `vortexui update` or Settings → Updates

---

## Migration

```bash
# with goose
export VORTEX_DATABASE_URL=postgres://...
make migrate-up
```

Docker: migration runs on panel startup.

---

## Makefile (Development)

| Target | Action |
|--------|--------|
| `make build` | panel + node binaries |
| `make test` | tests with race detector |
| `make certs` | mTLS dev certs |
| `make stack-up` | full docker stack |
| `make proto` | regenerate gRPC |
| `make sqlc` | regenerate DB code |

---

## systemd (Native)

| Service | Role |
|---------|------|
| `vortexui-panel` | panel API |
| `vortex-node` | node agent (optional) |
| `caddy` | web + HTTPS |

```bash
systemctl status vortexui-panel caddy
journalctl -u vortexui-panel -f
```

---

## Recommended Backup Strategy

| Layer | Method | Frequency |
|-------|--------|-----------|
| DB | `GET /api/backup` | daily |
| Auto | Telegram/S3 | daily |
| Config | `deploy/.env` + certs | after changes |
| Off-site | copy to separate storage | weekly |

---

## Monitoring Checklist

- [ ] `/api/health` in uptime monitor
- [ ] Alert on `node.down`
- [ ] Disk >85% warning
- [ ] DB connection pool
- [ ] Certificate expiry (Caddy auto — verify port 80)

</div>
