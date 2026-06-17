<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/10-notifications.md) · [EN](../en/10-notifications.md) · [TR](../tr/10-notifications.md)

</div>

<div dir="rtl">

# ١٠. الإشعارات

[← الخطط](./09-plans-payments.md) · [الفهرس](./README.md) · [التالي: الإعدادات →](./11-settings-backup.md)

> [!TIP]
> Webhooks موقّعة بـ `X-Vortex-Signature: sha256=...` — اضبط secret في env.

---

## Event Bus

جميع أحداث النطاق تمر عبر الحافلة الداخلية:

| الحدث | متى |
|-------|------|
| `user.created` | إنشاء مستخدم |
| `user.deleted` | حذف مستخدم |
| `user.limited` | تجاوز حد الحركة |
| `user.expired` | انتهى |
| `user.reset` | إعادة تعيين الحركة |
| `user.ip_limit` | مشاركة الحساب |
| `user.expiry_warning` | 3 أيام قبل الانتهاء |
| `node.down` | العقدة غير قابلة للوصول |
| `node.up` | تعافت العقدة |

---

## Webhook

```env
VORTEX_WEBHOOK_URL=https://your-server.com/hook
VORTEX_WEBHOOK_SECRET=your-hmac-secret
```

### Payload

```json
{
  "type": "user.limited",
  "time": "2026-06-17T12:00:00Z",
  "user_id": "uuid",
  "username": "john",
  "message": "User john exceeded data limit"
}
```

### التوقيع

Header: `X-Vortex-Signature: sha256=<hex>`

```python
import hmac, hashlib
sig = hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()
```

---

## Telegram Notifier

```env
VORTEX_TELEGRAM_TOKEN=123456:ABC...
VORTEX_TELEGRAM_CHAT_ID=-1001234567890
```

تُرسل الأحداث إلى محادثة المسؤول.

---

## Telegram Bot (تفاعلي — Admin)

بوت مع long-polling:

| الأمر | الإجراء |
|---------|--------|
| `/status` | حالة العقد |
| `/users` | إحصائيات المستخدمين |
| `/node <name>` | تفاصيل العقدة |
| `/limit <user>` | تحديد المستخدم |

---

## Telegram User Bot

المستخدمون النهائيون يصادقون برمز الاشتراك:

| الأمر | الإجراء |
|---------|--------|
| `/start` | مساعدة |
| `/login <token>` | ربط الحساب |
| `/usage` | الاستخدام الحالي |
| `/sub` | رابط الاشتراك |

---

## Auto-Backup إلى Telegram/S3

**Settings → Auto Backup**

- جدولة (cron-like)
- الوجهة: مستند Telegram أو S3 bucket
- الملف: نسخ احتياطي JSON transactional

---

## SSE (UI)

بالإضافة إلى webhooks، الواجهة تشترك في نفس الحافلة عبر SSE — toasts وتحديث تلقائي.

</div>
