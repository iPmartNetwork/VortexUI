<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [FA](../fa/04-dashboard.md) · [EN](../en/04-dashboard.md) · [TR](../tr/04-dashboard.md)

</div>

<div dir="rtl">

# ٤. لوحة المعلومات (Overview)

[← الخطوات الأولى](./03-first-steps.md) · [الفهرس](./README.md) · [التالي: المستخدمون →](./05-user-management.md)

> [!NOTE]
> لوحة المعلومات تتحدّث عبر **SSE** دون refresh — لا حاجة لـ polling.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![لوحة Overview — إحصائيات مباشرة ومخطط الحركة](../assets/panel/overview_light.png) | ![لوحة Overview — إحصائيات مباشرة ومخطط الحركة](../assets/panel/overview_dark.png) |

*لوحة Overview — إحصائيات مباشرة ومخطط الحركة*

</div>

---

## نظرة عامة

صفحة **Overview** هي عرض العمليات المركزي: حالة الأسطول، الحركة، المستخدمون النشطون، والأحداث الأخيرة — كلها مع **تحديثات مباشرة (SSE)**.

---

## بطاقات الإحصائيات

| البطاقة | المحتوى |
|------|---------|
| **Users** | الإجمالي، النشط، المحدود، المنتهي |
| **Traffic** | إجمالي الرفع/التنزيل، سلسلة زمنية |
| **Nodes** | عدد المتصل/غير المتصل |
| **Connections** | اتصالات البروكسي النشطة |

---

## رسم الحركة البياني

- سلسلة زمنية مدعومة بـ **TimescaleDB**
- نطاقات قابلة للاختيار (24h، 7d، 30d)
- تفصيل الرفع/التنزيل

---

## التحديثات المباشرة (SSE)

تستخدم اللوحة **Server-Sent Events**:

```
GET /api/events/stream?access_token=<JWT>
```

عند حدوث حدث (عقدة متوقفة، مستخدم محدود، إلخ) تتحدث الواجهة دون إعادة تحميل الصفحة.

| الحدث | تأثير الواجهة |
|-------|-----------|
| `node.down` | شارة عقدة حمراء + toast |
| `user.limited` | تحديث حالة المستخدم |
| `user.ip_limit` | تحذير مشاركة الحساب |
| `user.expiry_warning` | إشعار انتهاء خلال 3 أيام |

> Caddy يمرّر هذا التدفق بشفافية. يأتي الرمز من query string لأن `EventSource` لا يمكنه إرسال رؤوس مخصصة.

---

## Prometheus / Grafana

المقاييس متاحة على endpoint Prometheus (للمراقبة الخارجية). التفاصيل: [الفصل 14 — العمليات](./14-operations-maintenance.md).

</div>
