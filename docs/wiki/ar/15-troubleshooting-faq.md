# ١٥. استكشاف الأخطاء والأسئلة الشائعة

!!! tip
    ابدأ بـ **`vortexui logs`** و**`/api/health`** — أغلب المشاكل JWT أو DB أو firewall.

---

## المشاكل الشائعة

### اللوحة لا تبدأ

```bash
vortexui status
vortexui logs
curl http://127.0.0.1:8080/api/health
```

| السبب | الحل |
|-------|----------|
| Empty JWT secret | `deploy/.env` → `JWT_SECRET=$(openssl rand -hex 32)` |
| DB down | `docker compose ps` — restart `db` |
| Port in use | `ss -tlnp \| grep 8080` |

---

### فشل HTTPS / Let's Encrypt

| السبب | الحل |
|-------|----------|
| Wrong DNS | A record to server IP |
| Port 80 closed | firewall: `ufw allow 80,443` |
| LE rate limit | wait 1h or staging test |

---

### العقدة offline / حمراء

| السبب | الحل |
|-------|----------|
| Agent down | `systemctl status vortex-node` |
| mTLS mismatch | regenerate certs, SAN includes IP |
| Firewall | port 50051 gRPC open |
| Core crash | Nodes → Logs |

---

### المستخدم لا يستطيع الاتصال

| التحقق | |
|-------|---|
| Inbound active? | Nodes → Inbounds |
| User status `active`? | Users |
| Expired / limited? | User detail |
| Inbound port open? | `ufw` / cloud security group |
| REALITY keys match? | regenerate + new sub |

---

### الحركة غير مسجّلة

| السبب | الحل |
|-------|----------|
| Core API port | `VORTEX_CORE_API_PORT=10085` |
| Stats disabled in core | panel config renders stats |
| Redis down | restart redis |

---

### اشتراك فارغ

| السبب | الحل |
|-------|----------|
| No inbound assigned | Edit user → select inbounds |
| Node down | fix node first |
| Wrong endpoint | set Custom Endpoint |

---

### SSE / التحديث المباشر لا يعمل

| السبب | الحل |
|-------|----------|
| Caddy buffering | default OK — check proxy timeout |
| Token expired | re-login |
| Ad blocker | disable for panel domain |

---

## FAQ

### كيف يختلف VortexUI عن 3x-ui?

**نموذج يركز على المستخدم**، حركة push delta، outbound/routing/balancer كامل، audit، API token، failover متقدم.

### هل يدعم SQLite?

لا — **PostgreSQL + TimescaleDB** (إنتاجي، سلاسل زمنية للحركة).

### كم عقدة مدعومة?

غير محدود — كل عقدة لها agent منفصل أو local node واحد.

### sing-box أم xray?

لكل عقدة — Hysteria2/TUIC فقط على sing-box؛ REALITY على كليهما.

### استيراد من Marzban?

نعم — Users → Import.

### مشاركة الحساب?

حد الأجهزة + online IP guard + autolimit اختياري.

### البيع مع ZarinPal?

Plans + ZarinPal gateway — [الفصل 9](./09-plans-payments.md).

### نسخ احتياطي قبل التحديث?

**دائماً** — `vortexui update` آمن لكن النسخ الاحتياطي موصى به.

### الترخيص?

GPL-3.0 — المشتقات يجب أن تكون مفتوحة المصدر.

---

## الإبلاغ عن الأخطاء

1. [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues)
2. الإصدار: `vortexui settings` أو الشريط الجانبي
3. السجلات: `vortexui logs` (بدون أسرار)
4. [SECURITY.md](../../../SECURITY.md) للثغرات

---

## المجتمع

- ⭐ Star on GitHub
- [Contributing](../../../CONTRIBUTING.md)
- [Changelog](../../../CHANGELOG.md)
