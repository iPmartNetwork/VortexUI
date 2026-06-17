<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [EN](../en/14-operations-maintenance.md) · [AR](../ar/14-operations-maintenance.md) · [TR](../tr/14-operations-maintenance.md)

</div>

<div dir="rtl">

# ۱۴. عملیات و نگهداری

[← پروتکل‌ها](./13-protocols-config.md) · [فهرست](./README.md) · [بعدی: عیب‌یابی →](./15-troubleshooting-faq.md)

> [!TIP]
> بعد از نصب `vortexui status` و `curl .../api/health` را برای sanity check اجرا کنید.

---

## کنسول `vortexui`

```bash
vortexui              # منوی تعاملی
vortexui start        # start stack
vortexui stop         # stop
vortexui restart      # restart
vortexui status       # وضعیت
vortexui logs         # tail logs
vortexui update       # git pull + rebuild
vortexui admin        # ساخت admin
vortexui settings     # URL و تنظیمات
vortexui uninstall    # حذف (با تأیید)
```

مسیر نصب: `VORTEXUI_DIR` (پیش‌فرض `/opt/vortexui`)

---

## HTTPS / SSL

### Docker (Caddy)

`deploy/.env`:

```env
SITE_ADDRESS=panel.example.com
ACME_EMAIL=admin@example.com
```

- پورت 80 + 443 باز
- DNS A record به سرور
- cert در volume `caddy-data`

تغییر domain:

```bash
vortexui   # → گزینه 9) Domain / SSL
```

### HTTP only

```env
SITE_ADDRESS=:8080
```

---

## Cluster Mode (HA)

چند instance panel با DB مشترک — برای availability بالا. جزئیات در env و deploy docs.

---

## Prometheus / Grafana

متریک‌های panel و node برای scraping Prometheus. dashboard Grafana نمونه در releaseها.

---

## Auto-update

- Panel binary از GitHub releases
- Core binaries (xray/sing-box) از upstream
- `vortexui update` یا Settings → Updates

---

## Migration

```bash
# با goose
export VORTEX_DATABASE_URL=postgres://...
make migrate-up
```

Docker: migration در startup panel اجرا می‌شود.

---

## Makefile (توسعه)

| Target | عمل |
|--------|-----|
| `make build` | panel + node binaries |
| `make test` | tests با race detector |
| `make certs` | mTLS dev certs |
| `make stack-up` | full docker stack |
| `make proto` | regenerate gRPC |
| `make sqlc` | regenerate DB code |

---

## systemd (Native)

| سرویس | نقش |
|-------|-----|
| `vortexui-panel` | panel API |
| `vortex-node` | node agent (optional) |
| `caddy` | web + HTTPS |

```bash
systemctl status vortexui-panel caddy
journalctl -u vortexui-panel -f
```

---

## Backup Strategy پیشنهادی

| لایه | روش | فرکانس |
|------|-----|---------|
| DB | `GET /api/backup` | روزانه |
| Auto | Telegram/S3 | روزانه |
| Config | `deploy/.env` + certs | پس از تغییر |
| Off-site | copy به storage جدا | هفتگی |

---

## Monitoring Checklist

- [ ] `/api/health` در uptime monitor
- [ ] Alert روی `node.down`
- [ ] Disk >85% warning
- [ ] DB connection pool
- [ ] Certificate expiry (Caddy auto — verify port 80)

</div>
