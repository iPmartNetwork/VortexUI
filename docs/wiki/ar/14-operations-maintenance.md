# ١٤. العمليات والصيانة

!!! tip
    بعد التثبيت نفّذ `vortexui status` و`curl .../api/health` للتحقق السريع.

---

## وحدة `vortexui`

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

مسار التثبيت: `VORTEXUI_DIR` (افتراضي `/opt/vortexui`)

---

## HTTPS / SSL

### Docker (Caddy)

`deploy/.env`:

```env
SITE_ADDRESS=panel.example.com
ACME_EMAIL=admin@example.com
```

- المنافذ 80 + 443 مفتوحة
- سجل DNS A إلى الخادم
- الشهادة في volume `caddy-data`

تغيير النطاق:

```bash
vortexui   # → option 9) Domain / SSL
```

### HTTP only

```env
SITE_ADDRESS=:8080
```

---

## Cluster Mode (HA)

عدة instances للوحة مع DB مشترك — للتوفر العالي. التفاصيل في env و deploy docs.

---

## Prometheus / Grafana

مقاييس اللوحة والعقد لـ Prometheus scraping. لوحة Grafana نموذجية في الإصدارات.

---

## Auto-Update

- ثنائي اللوحة من GitHub releases
- ثنائيات النواة (xray/sing-box) من upstream
- `vortexui update` أو Settings → Updates

---

## Migration

```bash
# with goose
export VORTEX_DATABASE_URL=postgres://...
make migrate-up
```

Docker: migration يعمل عند بدء panel.

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

- [ ] `/api/health` في uptime monitor
- [ ] تنبيه على `node.down`
- [ ] تحذير Disk >85%
- [ ] DB connection pool
- [ ] انتهاء الشهادة (Caddy auto — تحقق من المنفذ 80)
