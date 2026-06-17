<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/05-user-management.md) · [EN](../en/05-user-management.md) · [TR](../tr/05-user-management.md)

</div>

<div dir="rtl">

# ٥. إدارة المستخدمين

[← لوحة المعلومات](./04-dashboard.md) · [الفهرس](./README.md) · [التالي: العقد →](./06-node-management.md)

> [!WARNING]
> **Revoke Sub** يُبطل الرابط السابق — استخدمه فقط عند تسريب الرمز.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![صفحة Users — إدارة المستخدمين والاشتراك](../assets/panel/User_light.png) | ![صفحة Users — إدارة المستخدمين والاشتراك](../assets/panel/User_dark.png) |

*صفحة Users — إدارة المستخدمين والاشتراك*

</div>

---

## الفلسفة: مستخدم واحد، بروتوكولات متعددة

كل **User** هو سجل واحد. برمز **subscription token** واحد يصل إلى جميع inbounds المسموح بها — دون الحاجة لإنشاء إدخالات منفصلة لـ VLESS و VMess.

---

## إنشاء مستخدم

**Users → New User**

| الحقل | الوصف |
|-------|-------------|
| **Username** | معرف فريد |
| **Data limit** | حد الحركة (بايت) — 0 = غير محدود |
| **Expire at** | تاريخ الانتهاء |
| **Device limit** | الحد الأقصى للأجهزة المتزامنة (IP/HWID) |
| **Reset strategy** | `none` / `daily` / `weekly` / `monthly` |
| **Status** | `active` / `disabled` / `limited` |
| **Inbounds** | inbounds المسموح بها |
| **Note** | ملاحظة المسؤول |

---

## العمليات المجمّعة

| الإجراء | المسار |
|--------|------|
| **Bulk Create** | Users → Add Bulk — CSV/count |
| **Multi-select** | تحديد عدة مستخدمين → إجراء |
| **Import** | Users → Import — 3x-ui / Marzban |

---

## الاشتراك

### الروابط

| Endpoint | المخرج |
|----------|--------|
| `GET /sub/{token}` | base64 (افتراضي) |
| `GET /sub/{token}?format=clash` | Clash YAML |
| `GET /sub/{token}?format=singbox` | sing-box JSON |
| `GET /sub/info/{token}` | صفحة HTML للمستخدم |

### الإبطال

**Users → Revoke Sub** — يُصدر رمز جديد؛ الرابط السابق يُبطل.

---

## محاسبة الحركة

- طريقة **Delta push**: النواة تدفع الدeltas (وليس polling)
- **آمن عند إعادة التشغيل**: العدادات مخزنة في DB
- **Reset**: يدوي أو مجدول (شهري، إلخ)
- **تطبيق الحصة**: تجاوز الحد → حالة `limited` + حدث `user.limited`

---

## حد الأجهزة و HWID

| الآلية | الوصف |
|-----------|-------------|
| **Device limit** | عدد IP/جهاز متزامن |
| **HWID allowlist** | الأجهزة المسجلة فقط |
| **Online IP guard** | مقارنة IP المتصل بالحد — [الفصل 8](./08-security-administration.md) |

---

## صفحة تفاصيل المستخدم

**Users → انقر اسم المستخدم** → `/users/:id`

- رسم استخدام
- IP المتصلة
- سجل إعادة التعيين
- تحرير مضمّن

---

## اختيار أفضل عقدة تلقائياً

في الاشتراكات يمكن تفعيل url-test لاختيار العقدة النشطة ذات أقل latency تلقائياً (في قالب التكوين).

---

## إشعارات المستخدم

- **بوت Telegram للمستخدم**: المستخدم يسجل الدخول بالرمز — `/usage`، `/renew`
- **تحذير الانتهاء**: 3 أيام قبل — حدث `user.expiry_warning`

</div>
