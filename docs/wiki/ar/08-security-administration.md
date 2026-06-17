<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/08-security-administration.md) · [EN](../en/08-security-administration.md) · [TR](../tr/08-security-administration.md)

</div>

<div dir="rtl">

# ٨. الأمان والإدارة

[← سياسة الشبكة](./07-network-policy.md) · [الفهرس](./README.md) · [التالي: الخطط →](./09-plans-payments.md)

> [!IMPORTANT]
> فعّل **2FA** وJWT secret قوي (≥32 بايت) للمسؤول الرئيسي.

---

## المصادقة

| الطبقة | الآلية |
|-------|-----------|
| تسجيل دخول اللوحة | JWT (Bearer) — TTL افتراضي 1h |
| 2FA | TOTP (Google Authenticator، إلخ) |
| أتمتة API | Personal Access Token (PAT) |
| Panel ↔ Node | mTLS (شهادات متبادلة) |

### تفعيل 2FA

**Settings → Two-Factor Authentication**

1. بدء الإعداد → مسح QR
2. أدخل رمز 6 أرقام → Confirm
3. للتعطيل: الرمز الحالي + Disable

---

## RBAC (التحكم في الوصول القائم على الأدوار)

**Admins → Roles**

| المفهوم | الوصف |
|---------|-------------|
| **Role** | مجموعة صلاحيات |
| **Permission** | `users.read`, `nodes.write`, … |
| **Reseller quota** | حد المستخدم/الحركة للمسؤول الفرعي |
| **Sub-panel** | المسؤول يرى فقط المستخدمين في نطاقه |

### إنشاء موزّع

1. أنشئ دوراً بصلاحيات محدودة
2. مسؤول جديد + حصة
3. الموزّع يدير مستخدميه فقط

---

## رموز API (PAT)

**Settings → API Tokens**

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- كل رمز يمكن إبطاله بشكل فردي
- الصلاحيات تُورث من دور المسؤول المنشئ

---

## حارس مشاركة الحساب

حلقة خلفية تقارن IP المتصلة (`GetStatsOnlineIpList`) مع حد الأجهزة.

| الوضع | السلوك |
|------|----------|
| Detection (افتراضي) | حدث `user.ip_limit` + webhook/TG |
| `VORTEX_SHARE_AUTOLIMIT=true` | حد تلقائي للمستخدم (قابل للعكس) |

---

## IP Guard (Whitelist/Blacklist)

**Settings → IP Guard**

- تقييد وصول API/الاشتراك حسب IP
- مفيد لتقييد وصول اللوحة لـ IP المسؤولين

---

## حماية من Brute-Force

- حد على محاولات تسجيل الدخول الفاشلة
- قفل مؤقت

---

## سجل التدقيق

**Audit** — يسجل جميع تغييرات المسؤول:

| الحقل | مثال |
|-------|---------|
| Actor | اسم مستخدم المسؤول |
| Action | `user.create`, `inbound.update` |
| Target | معرف user/node |
| Timestamp | ISO8601 |
| Diff | قبل/بعد |

---

## حد النطاق الترددي لكل Inbound

حد سرعة على inbound — يمنع خدمة واحدة من تشبع الرابط.

---

## Geo-Blocking لكل Inbound

تقييد البلد/المنطقة للاتصال بـ inbound محدد.

---

## قائمة التحقق الأمنية

- [ ] JWT secret قوي (≥32 بايت عشوائي)
- [ ] HTTPS مفعّل (Let's Encrypt)
- [ ] 2FA لمسؤول sudo
- [ ] PAT بأقل امتياز
- [ ] نسخ احتياطي مشفر خارج الموقع
- [ ] webhook secret لـ HMAC
- [ ] منفذ اللوحة مغلق من الإنترنت العام (Caddy 443 فقط)

</div>
