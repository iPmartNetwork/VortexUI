<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [EN](../en/04-dashboard.md) · [AR](../ar/04-dashboard.md) · [TR](../tr/04-dashboard.md)

</div>

<div dir="rtl">

# ۴. داشبورد (Overview)

[← اولین قدم‌ها](./03-first-steps.md) · [فهرست](./README.md) · [بعدی: کاربران →](./05-user-management.md)

> [!NOTE]
> داشبورد با **SSE** بدون refresh به‌روز می‌شود — polling لازم نیست.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![داشبورد Overview — آمار زنده و نمودار ترافیک](../assets/panel/overview_light.png) | ![داشبورد Overview — آمار زنده و نمودار ترافیک](../assets/panel/overview_dark.png) |

*داشبورد Overview — آمار زنده و نمودار ترافیک*

</div>

---

## نمای کلی

صفحه **Overview** نمای مرکزی عملیات است: وضعیت ناوگان، ترافیک، کاربران فعال و رویدادهای اخیر — همه با **به‌روزرسانی زنده (SSE)**.

---

## کارت‌های آماری

| کارت | محتوا |
|------|-------|
| **Users** | کل کاربران، فعال، محدودشده، منقضی |
| **Traffic** | آپلود/دانلود کل، روند زمانی |
| **Nodes** | تعداد آنلاین/آفلاین |
| **Connections** | اتصالات فعال proxy |

---

## نمودار ترافیک

- سری زمانی بر پایه **TimescaleDB**
- بازه‌های قابل انتخاب (۲۴h، ۷d، ۳۰d)
- تفکیک upload/download

---

## Live Updates (SSE)

پنل از **Server-Sent Events** استفاده می‌کند:

```
GET /api/events/stream?access_token=<JWT>
```

وقتی رویدادی رخ می‌دهد (نود down، user limited، …) UI بدون refresh به‌روز می‌شود.

| رویداد | اثر در UI |
|--------|-----------|
| `node.down` | badge قرمز نود + toast |
| `user.limited` | به‌روز status کاربر |
| `user.ip_limit` | هشدار اشتراک‌گذاری |
| `user.expiry_warning` | اعلان ۳ روز قبل انقضا |

> Caddy این stream را transparent proxy می‌کند. توکن از query string می‌آید چون `EventSource` header سفارشی نمی‌فرستد.

---

## Prometheus / Grafana

متریک‌ها در endpoint Prometheus در دسترس هستند (برای مانیتورینگ خارجی). جزئیات: [فصل ۱۴ — عملیات](./14-operations-maintenance.md).

</div>
