# ۵. مدیریت کاربران

!!! warning "هشدار"
    **Revoke Sub** لینک قبلی را باطل می‌کند — فقط در صورت نشت token استفاده کنید.

---

## فلسفه: یک کاربر، چند پروتکل

هر **User** یک رکورد واحد است. با یک **subscription token** به همه inboundهای مجاز دسترسی دارد — نیازی به ساخت جداگانه برای VLESS و VMess نیست.

---

## ساخت کاربر

**Users → New User**

| فیلد | توضیح |
|------|-------|
| **Username** | شناسه یکتا |
| **Data limit** | سقف ترافیک (bytes) — 0 = نامحدود |
| **Expire at** | تاریخ انقضا |
| **Device limit** | حداکثر دستگاه همزمان (IP/HWID) |
| **Reset strategy** | `none` / `daily` / `weekly` / `monthly` |
| **Status** | `active` / `disabled` / `limited` |
| **Inbounds** | inboundهای مجاز |
| **Note** | یادداشت ادمین |

---

## عملیات گروهی (Bulk)

| عمل | مسیر |
|-----|------|
| **Bulk Create** | Users → Add Bulk — CSV/تعداد |
| **Multi-select** | انتخاب چند کاربر → عملیات |
| **Import** | Users → Import — 3x-ui / Marzban |

---

## Subscription

### لینک‌ها

| Endpoint | خروجی |
|----------|--------|
| `GET /sub/{token}` | base64 (پیش‌فرض) |
| `GET /sub/{token}?format=clash` | YAML Clash |
| `GET /sub/{token}?format=singbox` | JSON sing-box |
| `GET /sub/info/{token}` | صفحه HTML کاربر |

### Revoke

**Users → Revoke Sub** — token جدید صادر می‌شود؛ لینک قبلی باطل.

---

## حساب‌داری ترافیک

- روش **delta push**: هسته delta را push می‌کند (نه polling)
- **Restart-safe**: counter در DB ذخیره می‌شود
- **Reset**: دستی یا زمان‌بندی‌شده (ماهانه و …)
- **Quota enforcement**: عبور از سقف → status `limited` + رویداد `user.limited`

---

## محدودیت دستگاه و HWID

| مکانیزم | توضیح |
|---------|-------|
| **Device limit** | تعداد IP/distinct device همزمان |
| **HWID allowlist** | فقط دستگاه‌های ثبت‌شده |
| **Online IP guard** | مقایسه IP آنلاین با limit — [فصل ۸](./08-security-administration.md) |

---

## صفحه جزئیات کاربر

**Users → کلیک روی username** → `/users/:id`

- نمودار مصرف
- IPهای آنلاین
- تاریخچه reset
- ویرایش inline

---

## Auto-select بهترین نود

در subscription می‌توان url-test برای انتخاب خودکار کم‌تأخیرترین نود فعال کرد (در template کانفیگ).

---

## اعلان به کاربر

- **Telegram user bot**: کاربر با token لاگین می‌کند — `/usage`, `/renew`
- **Expiry warning**: ۳ روز قبل — رویداد `user.expiry_warning`
