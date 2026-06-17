<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [EN](../en/15-troubleshooting-faq.md) · [AR](../ar/15-troubleshooting-faq.md) · [TR](../tr/15-troubleshooting-faq.md)

</div>

<div dir="rtl">

# ۱۵. عیب‌یابی و سوالات متداول

[← عملیات](./14-operations-maintenance.md) · [فهرست](./README.md)

> [!TIP]
> اول **`vortexui logs`** و **`/api/health`** — بیشتر مشکلات از JWT، DB یا firewall است.

---

## مشکلات رایج

### پنل بالا نمی‌آید

```bash
vortexui status
vortexui logs
curl http://127.0.0.1:8080/api/health
```

| علت | راه‌حل |
|-----|--------|
| JWT secret خالی | `deploy/.env` → `JWT_SECRET=$(openssl rand -hex 32)` |
| DB down | `docker compose ps` — restart `db` |
| پورت اشغال | `ss -tlnp \| grep 8080` |

---

### HTTPS / Let's Encrypt fail

| علت | راه‌حل |
|-----|--------|
| DNS اشتباه | A record به IP سرور |
| پورت 80 بسته | firewall: `ufw allow 80,443` |
| rate limit LE | صبر ۱h یا staging test |

---

### نود offline / قرمز

| علت | راه‌حل |
|-----|--------|
| Agent down | `systemctl status vortex-node` |
| mTLS mismatch | regenerate certs، SAN شامل IP |
| Firewall | پورت 50051 gRPC باز |
| Core crash | Nodes → Logs |

---

### کاربر وصل نمی‌شود

| بررسی | |
|-------|---|
| Inbound فعال؟ | Nodes → Inbounds |
| User status `active`? | Users |
| Expire / limited? | User detail |
| پورت inbound باز? | `ufw` / cloud security group |
| REALITY keys match? | regenerate + new sub |

---

### ترافیک ثبت نمی‌شود

| علت | راه‌حل |
|-----|--------|
| Core API port | `VORTEX_CORE_API_PORT=10085` |
| Stats disabled in core | panel config renders stats |
| Redis down | restart redis |

---

### Subscription خالی

| علت | راه‌حل |
|-----|--------|
| No inbound assigned | Edit user → select inbounds |
| Node down | fix node first |
| Wrong endpoint | set Custom Endpoint |

---

### SSE / live update کار نمی‌کند

| علت | راه‌حل |
|-----|--------|
| Caddy buffering | default OK — check proxy timeout |
| Token expired | re-login |
| Ad blocker | disable for panel domain |

---

## FAQ

### تفاوت VortexUI با 3x-ui چیست؟

مدل **کاربر‌محور**، push delta traffic، outbound/routing/balancer کامل، audit، API token، failover پیشرفته.

### آیا SQLite دارد؟

خیر — **PostgreSQL + TimescaleDB** (production-grade، سری زمانی ترافیک).

### چند نود پشتیبانی می‌شود؟

نامحدود — هر نود agent جدا یا یک local node.

### sing-box یا xray؟

هر نود جدا — Hysteria2/TUIC فقط sing-box؛ REALITY روی هر دو.

### import از Marzban؟

بله — Users → Import.

### اشتراک‌گذاری اکانت؟

Device limit + online IP guard + optional autolimit.

### فروش با زرین‌پال؟

Plans + ZarinPal gateway — [فصل ۹](./09-plans-payments.md).

### backup قبل از update؟

**همیشه** — `vortexui update` ایمن است ولی backup توصیه می‌شود.

### license?

GPL-3.0 — مشتقات باید source باز باشند.

---

## گزارش باگ

1. [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues)
2. نسخه: `vortexui settings` یا sidebar
3. لاگ: `vortexui logs` (بدون secret)
4. [SECURITY.md](../../../SECURITY.md) برای آسیب‌پذیری

---

## Community

- ⭐ Star on GitHub
- [Contributing](../../../CONTRIBUTING.md)
- [Changelog](../../../CHANGELOG.md)

</div>
