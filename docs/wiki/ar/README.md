# 🌀 توثيق VortexUI

<div align="center">

**لوحة إدارة البروكسي من الجيل التالي**

*محورها المستخدم · مستقلة عن النواة · جاهزة للمؤسسات*

[![الإصدار](https://img.shields.io/badge/الإصدار-1.3.1-7c3aed?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/releases)
[![الرخصة](https://img.shields.io/badge/الرخصة-MIT-green?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/blob/master/LICENSE)
[![Docker](https://img.shields.io/badge/docker-جاهز-blue?style=for-the-badge)](https://hub.docker.com/r/ipmartnetwork/vortexui)

</div>

---

## 🚀 التثبيت السريع

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

أمر واحد. إعداد تفاعلي. يتضمن HTTPS.

---

## 📖 خريطة التوثيق

| القسم | الوصف |
|-------|-------|
| [المقدمة](01-introduction.md) | البنية، نظرة عامة على الميزات، المقارنة |
| [التثبيت](02-installation.md) | التثبيت بسطر واحد، Docker، البناء الأصلي |
| [الخطوات الأولى](03-first-steps.md) | تسجيل الدخول، إضافة عقدة، إنشاء inbound، إضافة مستخدم |
| [لوحة التحكم](04-dashboard.md) | الأدوات، التحليلات، المراقبة، لوحة الأوامر |
| [المستخدمون](05-user-management.md) | الإدارة، الحصص، الاشتراكات، البوابة، المتجر |
| [العقد](06-node-management.md) | التسجيل، الصحة، الترحيل التلقائي، المراقبة |
| [الشبكة](07-network-policy.md) | المخرجات، حزم التوجيه، سلاسل CDN، موازن التحميل |
| [الأمان](08-security-administration.md) | RBAC، حيل TLS، حماية الفحص، حد IP |
| [الخطط والمدفوعات](09-plans-payments.md) | خطط الموزعين، إعداد الدفع، المحفظة |
| [الإشعارات](10-notifications.md) | Webhooks، Telegram، تنبيهات الحصة، SSE |
| [الإعدادات](11-settings-backup.md) | العلامة التجارية، التسمية البيضاء، النسخ الاحتياطي |
| [مرجع API](12-api-reference.md) | المصادقة، نقاط النهاية، مواصفات OpenAPI |
| [البروتوكولات](13-protocols-config.md) | 14 بروتوكول، النقل، طبقات الأمان |
| [العمليات](14-operations-maintenance.md) | HTTPS، Prometheus، التوسع، الأداء |
| [استكشاف الأخطاء](15-troubleshooting-faq.md) | المشاكل الشائعة، نصائح التصحيح، الأسئلة الشائعة |

---

## ✨ الميزات الرئيسية

### 🔧 المحرك والبنية التحتية
- **دعم النواة المزدوجة** — Xray-core و sing-box، اختر لكل عقدة
- **حركة دلتا** — مقاومة لإعادة التشغيل، بدون فقدان البيانات
- **أسطول عقد mTLS** — اتصالات مشفرة، تجاوز الفشل التلقائي
- **الترحيل التلقائي** — نقل المستخدمين من العقد غير الصحية
- **الاتحاد** — مزامنة المستخدمين/العقد عبر لوحات متعددة

### 🛡 الأمان ومكافحة الرقابة
- **ماسح Reality** — اكتشاف SNI المثالية مع تسجيل التأخير
- **مدير حيل TLS** — ملفات تعريف ISP (التجزئة، mux، الحشو)
- **حماية الفحص** — اكتشاف وحظر فحوصات GFW
- **موقع وهمي** — عرض موقع مزيف للمراقبين
- **DNS-over-HTTPS** — DoH مدمج مع حظر الإعلانات

### 👥 إدارة المستخدمين والتجارة
- **بوابة الخدمة الذاتية** — تسجيل الدخول بالرمز، عرض الاستخدام، التذاكر
- **متجر الخدمة الذاتية** — خطط الموزعين مع طرق دفع متعددة
- **الحصة الذكية** — تقليل السرعة التدريجي
- **مجموعات عائلية** — مجموعات بيانات مشتركة
- **نظام الإحالة** — رموز الدعوة مع مكافآت

### 💼 منصة الموزعين
- **محفظة** — نظام ائتمان مع طابور الشحن
- **موزعين فرعيين** — إنشاء موزعين أبناء
- **التسمية البيضاء** — علامة تجارية مخصصة
- **Webhooks** — أحداث صادرة للأتمتة
- **حدود السياسة** — الحد الأقصى للبيانات، الحد الأقصى لانتهاء الصلاحية

### 🎨 الواجهة الأمامية و UX
- **واجهة Veltrix الزجاجية** — نظام تصميم Glass حديث
- **لوحة الأوامر** — بحث Ctrl+K في كل مكان
- **أدوات لوحة التحكم** — سحب وإفلات وتغيير الحجم
- **8 لغات** — مع دعم RTL كامل للعربية والفارسية
- **السمة الداكنة والفاتحة** — انتقال متحرك سلس

---

## 🔗 روابط سريعة

| المورد | الرابط |
|--------|--------|
| مستودع GitHub | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
| قناة Telegram | [@vortex_ui](https://t.me/vortex_ui) |
| مواصفات OpenAPI | [openapi.yaml](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| سجل التغييرات | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
| الإبلاغ عن الأخطاء | [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues) |

---

## 🌍 اللغات

هذا التوثيق متاح بـ:

- 🇸🇦 **العربية** (الحالية)
- 🇬🇧 [English](../en/README.md)
- 🇮🇷 [فارسی](../fa/README.md)
- 🇹🇷 [Türkçe](../tr/README.md)
