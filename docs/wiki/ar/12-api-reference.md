# مرجع الـ API

!!! info "مواصفات OpenAPI"
    مواصفات OpenAPI 3.0 الكاملة متاحة في
    [`docs/openapi.yaml`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml).

---

## المصادقة

### تسجيل الدخول بـ JWT

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password",
  "totp_code": "123456"  // اختياري، إذا كانت المصادقة الثنائية مفعّلة
}
```

الاستجابة:

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

استخدم التوكن في الطلبات اللاحقة:

```
Authorization: Bearer <access_token>
```

### توكنات API (PAT)

للأتمتة، أنشئ توكن وصول شخصي:

1. **الإعدادات → توكنات API → إنشاء**
2. استخدمه كـ JWT: `Authorization: Bearer <PAT>`

توكنات PAT لا تنتهي صلاحيتها إلا إذا أُعدّت كذلك، ويمكن إلغاؤها فردياً.

---

## الرابط الأساسي والإصدار

```
https://panel.example.com/api/
```

جميع نقاط النهاية تحت `/api/`. لا بادئة إصدار — الـ API متوافق مع المستقبل.

---

## نقاط النهاية الرئيسية

### المصادقة (Auth)

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| POST | `/api/auth/login` | تسجيل الدخول، الحصول على JWT |
| POST | `/api/auth/refresh` | تجديد التوكن |
| GET | `/api/auth/me` | معلومات المسؤول الحالي |

### المستخدمون

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/users` | قائمة المستخدمين (ترقيم، فلاتر) |
| POST | `/api/users` | إنشاء مستخدم |
| GET | `/api/users/:id` | تفاصيل المستخدم |
| PUT | `/api/users/:id` | تحديث المستخدم |
| DELETE | `/api/users/:id` | حذف المستخدم |
| POST | `/api/users/:id/reset-traffic` | إعادة تعيين عدّاد الحركة |
| POST | `/api/users/:id/revoke-sub` | إلغاء توكن الاشتراك |
| GET | `/api/users/:id/usage` | سجل الاستخدام |

### العقد

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/nodes` | قائمة العقد |
| POST | `/api/nodes` | إنشاء عقدة |
| GET | `/api/nodes/:id` | تفاصيل العقدة |
| PUT | `/api/nodes/:id` | تحديث العقدة |
| DELETE | `/api/nodes/:id` | حذف العقدة |
| POST | `/api/nodes/:id/restart` | إعادة تشغيل نواة العقدة |
| GET | `/api/nodes/:id/health` | حالة السلامة |
| GET | `/api/nodes/:id/stats` | إحصائيات حيّة |

### الاتصالات الواردة

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/inbounds` | قائمة الاتصالات الواردة |
| POST | `/api/inbounds` | إنشاء اتصال وارد |
| PUT | `/api/inbounds/:id` | تحديث اتصال وارد |
| DELETE | `/api/inbounds/:id` | حذف اتصال وارد |
| GET | `/api/capabilities` | مصفوفة القدرات لكل بروتوكول |

### الخطط

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/plans` | قائمة الخطط (محدودة بنطاق المسؤول) |
| POST | `/api/plans` | إنشاء خطة |
| PUT | `/api/plans/:id` | تحديث خطة |
| DELETE | `/api/plans/:id` | حذف خطة |

### الطلبات

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/orders` | قائمة الطلبات |
| GET | `/api/orders/:id` | تفاصيل الطلب |
| POST | `/api/orders/:id/approve` | الموافقة على طلب معلّق |
| POST | `/api/orders/:id/reject` | رفض طلب معلّق |

### إعداد الدفع

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/payment-config` | الحصول على إعداد الدفع |
| PUT | `/api/payment-config` | تحديث إعداد الدفع |

### المسؤولون والموزّع

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/admins` | قائمة المسؤولين |
| POST | `/api/admins` | إنشاء مسؤول |
| PUT | `/api/admins/:id` | تحديث مسؤول |
| POST | `/api/admins/:id/quota-adjust` | تعديل حصة الموزّع |
| POST | `/api/admins/:id/unsuspend` | مسح التعليق |
| POST | `/api/admins/:id/impersonate` | إصدار توكن الموزّع |

### حساب الموزّع

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/account/dashboard` | إحصائيات لوحة معلومات الموزّع |
| GET | `/api/account/export/users` | تصدير CSV للمستخدمين المملوكين |
| GET | `/api/account/wallet` | المحفظة + دفتر الحسابات |
| GET/PUT | `/api/account/branding` | إعدادات العلامة التجارية |
| GET/PUT | `/api/account/webhook` | إعداد الويب هوك الصادر |
| GET/POST | `/api/account/sub-admins` | إدارة الموزّعين الفرعيين |

### النظام

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/api/health` | فحص السلامة |
| GET | `/api/stats` | إحصائيات النظام |
| GET | `/api/audit` | سجل التدقيق |
| POST | `/api/backup` | تشغيل النسخ الاحتياطي |
| GET | `/api/settings` | إعدادات لوحة التحكم |
| PUT | `/api/settings` | تحديث الإعدادات |

### الاشتراكات (عامة)

| الطريقة | نقطة النهاية | الوصف |
|---------|-------------|-------|
| GET | `/sub/{token}` | اشتراك المستخدم (صيغة تلقائية) |
| GET | `/sub/{token}?format=clash` | Clash YAML |
| GET | `/sub/{token}?format=singbox` | sing-box JSON |
| GET | `/sub/{token}?format=xray` | Xray JSON |
| GET | `/sub/{token}?format=outline` | روابط Outline |
| GET | `/sub/{token}?format=links` | روابط مشاركة نصية |
| GET | `/sub/info/{token}` | صفحة HTML لمعلومات المستخدم |
| GET | `/sub/{token}/shop` | متجر الخدمة الذاتية |

---

## أمثلة على الطلبات

### تسجيل الدخول

```bash
curl -X POST https://panel.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'
```

### إنشاء مستخدم

```bash
curl -X POST https://panel.example.com/api/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "newuser",
    "data_limit": 53687091200,
    "expire_at": "2025-03-01T00:00:00Z",
    "device_limit": 3,
    "inbound_ids": ["uuid-1", "uuid-2"]
  }'
```

### قائمة الخطط

```bash
curl https://panel.example.com/api/plans \
  -H "Authorization: Bearer <token>"
```

### الشراء (البوابة)

```bash
curl -X POST https://panel.example.com/api/portal/purchase \
  -H "Authorization: Bearer <sub_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "uuid",
    "payment_method": "card",
    "proof_image": "base64_encoded_image",
    "reference_number": "123456789"
  }'
```

---

## ترقيم الصفحات

نقاط النهاية التي تعرض قوائم تدعم ترقيم الصفحات:

| المعامل | الوصف | الافتراضي |
|---------|-------|-----------|
| `page` | رقم الصفحة (يبدأ من 1) | 1 |
| `per_page` | عناصر لكل صفحة | 20 |
| `sort` | حقل الترتيب | `created_at` |
| `order` | `asc` أو `desc` | `desc` |

الاستجابة تتضمن بيانات الترقيم الوصفية:

```json
{
  "data": [...],
  "total": 150,
  "page": 1,
  "per_page": 20,
  "total_pages": 8
}
```

---

## استجابات الأخطاء

جميع الأخطاء تتبع صيغة موحّدة:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with the specified ID does not exist",
    "status": 404
  }
}
```

رموز حالة HTTP الشائعة:

| الرمز | المعنى |
|-------|--------|
| 400 | طلب غير صالح (خطأ تحقق) |
| 401 | غير مصرّح (توكن غير صالح/منتهي) |
| 403 | محظور (صلاحيات غير كافية) |
| 404 | المورد غير موجود |
| 409 | تعارض (اسم مستخدم مكرر، إلخ) |
| 429 | تجاوز حدّ المعدل |
| 500 | خطأ داخلي في الخادم |
