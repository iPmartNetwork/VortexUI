# مرجع API

!!! info "مشخصات OpenAPI"
    مشخصات کامل OpenAPI 3.0 در
    [`docs/openapi.yaml`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) موجود است.

---

## احراز هویت

### ورود JWT

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password",
  "totp_code": "123456"  // اختیاری، اگر 2FA فعال باشد
}
```

پاسخ:

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

توکن را در درخواست‌های بعدی استفاده کنید:

```
Authorization: Bearer <access_token>
```

### توکن‌های API (PAT)

برای اتوماسیون، یک توکن دسترسی شخصی ایجاد کنید:

1. **تنظیمات → توکن‌های API → ایجاد**
2. مانند JWT استفاده کنید: `Authorization: Bearer <PAT>`

توکن‌های PAT منقضی نمی‌شوند مگر تنظیم شود و به‌صورت مستقل قابل لغو هستند.

---

## آدرس پایه و نسخه‌بندی

```
https://panel.example.com/api/
```

تمام اندپوینت‌ها زیر `/api/` هستند. بدون پیشوند نسخه — API سازگاری رو به جلو دارد.

---

## اندپوینت‌های اصلی

### احراز هویت

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| POST | `/api/auth/login` | ورود، دریافت JWT |
| POST | `/api/auth/refresh` | تمدید توکن |
| GET | `/api/auth/me` | اطلاعات ادمین فعلی |

### کاربران

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/users` | لیست کاربران (صفحه‌بندی، فیلتر) |
| POST | `/api/users` | ایجاد کاربر |
| GET | `/api/users/:id` | جزئیات کاربر |
| PUT | `/api/users/:id` | بروزرسانی کاربر |
| DELETE | `/api/users/:id` | حذف کاربر |
| POST | `/api/users/:id/reset-traffic` | ریست شمارنده ترافیک |
| POST | `/api/users/:id/revoke-sub` | لغو توکن سابسکریپشن |
| GET | `/api/users/:id/usage` | تاریخچه مصرف |

### نودها

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/nodes` | لیست نودها |
| POST | `/api/nodes` | ایجاد نود |
| GET | `/api/nodes/:id` | جزئیات نود |
| PUT | `/api/nodes/:id` | بروزرسانی نود |
| DELETE | `/api/nodes/:id` | حذف نود |
| POST | `/api/nodes/:id/restart` | ریستارت هسته نود |
| GET | `/api/nodes/:id/health` | وضعیت سلامت |
| GET | `/api/nodes/:id/stats` | آمار زنده |

### اینباندها

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/inbounds` | لیست اینباندها |
| POST | `/api/inbounds` | ایجاد اینباند |
| PUT | `/api/inbounds/:id` | بروزرسانی اینباند |
| DELETE | `/api/inbounds/:id` | حذف اینباند |
| GET | `/api/capabilities` | ماتریس قابلیت هر پروتکل |

### پلن‌ها

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/plans` | لیست پلن‌ها (محدوده ادمین) |
| POST | `/api/plans` | ایجاد پلن |
| PUT | `/api/plans/:id` | بروزرسانی پلن |
| DELETE | `/api/plans/:id` | حذف پلن |

### سفارشات

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/orders` | لیست سفارشات |
| GET | `/api/orders/:id` | جزئیات سفارش |
| POST | `/api/orders/:id/approve` | تأیید سفارش معلق |
| POST | `/api/orders/:id/reject` | رد سفارش معلق |

### تنظیم پرداخت

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/payment-config` | دریافت تنظیم پرداخت |
| PUT | `/api/payment-config` | بروزرسانی تنظیم پرداخت |

### ادمین‌ها و نمایندگی

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/admins` | لیست ادمین‌ها |
| POST | `/api/admins` | ایجاد ادمین |
| PUT | `/api/admins/:id` | بروزرسانی ادمین |
| POST | `/api/admins/:id/quota-adjust` | تنظیم سهمیه ریسلر |
| POST | `/api/admins/:id/unsuspend` | رفع تعلیق |
| POST | `/api/admins/:id/impersonate` | صدور توکن ریسلر |

### اکانت ریسلر

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/account/dashboard` | آمار داشبورد ریسلر |
| GET | `/api/account/export/users` | خروجی CSV کاربران اختصاصی |
| GET | `/api/account/wallet` | کیف پول + دفترکل |
| GET/PUT | `/api/account/branding` | تنظیمات وایت‌لیبل |
| GET/PUT | `/api/account/webhook` | تنظیم وب‌هوک خروجی |
| GET/POST | `/api/account/sub-admins` | مدیریت زیرنمایندگی |

### سیستم

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/api/health` | بررسی سلامت |
| GET | `/api/stats` | آمار سیستم |
| GET | `/api/audit` | لاگ حسابرسی |
| POST | `/api/backup` | فعال‌سازی بکاپ |
| GET | `/api/settings` | تنظیمات پنل |
| PUT | `/api/settings` | بروزرسانی تنظیمات |

### سابسکریپشن‌ها (عمومی)

| متد | اندپوینت | توضیحات |
|-----|----------|---------|
| GET | `/sub/{token}` | سابسکریپشن کاربر (فرمت خودکار) |
| GET | `/sub/{token}?format=clash` | YAML کلش |
| GET | `/sub/{token}?format=singbox` | JSON سینگ‌باکس |
| GET | `/sub/{token}?format=xray` | JSON ایکسری |
| GET | `/sub/{token}?format=outline` | لینک‌های Outline |
| GET | `/sub/{token}?format=links` | لینک‌های اشتراک ساده |
| GET | `/sub/info/{token}` | صفحه HTML اطلاعات کاربر |
| GET | `/sub/{token}/shop` | فروشگاه سلف‌سرویس |

---

## نمونه درخواست‌ها

### ورود

```bash
curl -X POST https://panel.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'
```

### ایجاد کاربر

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

### لیست پلن‌ها

```bash
curl https://panel.example.com/api/plans \
  -H "Authorization: Bearer <token>"
```

### خرید (پورتال)

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

## صفحه‌بندی

اندپوینت‌های لیست از صفحه‌بندی پشتیبانی می‌کنند:

| پارامتر | توضیحات | پیش‌فرض |
|---------|---------|---------|
| `page` | شماره صفحه (از ۱) | ۱ |
| `per_page` | آیتم در هر صفحه | ۲۰ |
| `sort` | فیلد مرتب‌سازی | `created_at` |
| `order` | `asc` یا `desc` | `desc` |

پاسخ شامل متادیتای صفحه‌بندی:

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

## پاسخ‌های خطا

همه خطاها از فرمت ثابت پیروی می‌کنند:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with the specified ID does not exist",
    "status": 404
  }
}
```

کدهای رایج HTTP:

| کد | معنی |
|----|------|
| 400 | درخواست نامعتبر (خطای اعتبارسنجی) |
| 401 | غیرمجاز (توکن نامعتبر/منقضی) |
| 403 | ممنوع (دسترسی ناکافی) |
| 404 | منبع یافت نشد |
| 409 | تعارض (نام کاربری تکراری و غیره) |
| 429 | محدودیت نرخ |
| 500 | خطای داخلی سرور |
