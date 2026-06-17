# ١١. الإعدادات والنسخ الاحتياطي

!!! tip
    خذ نسخة احتياطية حالية دائماً قبل **Restore**.

---

## المظهر

**Settings → Appearance**

| الخيار | القيم |
|--------|--------|
| Theme | Light / Dark / System |
| Language | EN, FA, TR, AR, RU, ZH, JA, ES |

---

## تغيير كلمة المرور

كلمة المرور الحالية + الجديدة — جلسة JWT الحالية محفوظة.

---

## رموز API

إنشاء، نسخ لمرة واحدة، قائمة، إبطال — [الفصل 8](./08-security-administration.md)

---

## النسخ الاحتياطي والاستعادة

### التصدير

**Settings → Backup → Download**

- لقطة transactional كاملة لـ DB (users، nodes، inbounds، routing، …)
- تنسيق JSON

### الاستعادة

**Upload JSON** — دمج أو استبدال (يعتمد على API)

```bash
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @backup.json \
  https://panel.example.com/api/backup/restore
```

> خذ نسخة احتياطية دائماً قبل الاستعادة.

---

## قالب تكوين الاشتراك

**Settings → Config Template**

- تجاوز قالب Clash/sing-box
- قواعد افتراضية، DNS، proxy-groups
- Placeholders: `{{USER}}`, `{{NODES}}`, …

---

## العلامة التجارية المخصصة

**Settings → Branding**

| الحقل | التأثير |
|-------|--------|
| Panel title | عنوان الواجهة |
| Logo URL | شعار مخصص |
| Sub page title | صفحة `/sub/info` |

---

## Auto Backup

- الفاصل الزمني
- Telegram / S3
- سياسة الاحتفاظ

---

## Update Checker

**Settings → Updates**

- التحقق من إصدار GitHub release
- تحديث تلقائي للوحة + ثنائيات النواة (اختياري)

---

## PWA

اللوحة هي **Progressive Web App** — من متصفح الجوال استخدم "Add to Home Screen" لتجربة شبيهة بالتطبيق.

`web/public/manifest.json` — الاسم، الأيقونات، لون المظهر.

---

## السجلات

**Logs** — سجلات مستوى اللوحة (وليس النواة):

- تصفية المستوى
- بحث
- tail مباشر

لسجلات النواة: **Nodes → Logs**
