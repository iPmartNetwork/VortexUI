<div align="center" dir="rtl">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](../README.md) · [EN](../en/10-notifications.md) · [AR](../ar/10-notifications.md) · [TR](../tr/10-notifications.md)

</div>

<div dir="rtl">

# ۱۰. اعلان‌ها

[← پلن‌ها](./09-plans-payments.md) · [فهرست](./README.md) · [بعدی: تنظیمات →](./11-settings-backup.md)

> [!TIP]
> Webhook با `X-Vortex-Signature: sha256=...` امضا می‌شود — secret را در env تنظیم کنید.

---

## Event Bus

همه رویدادهای domain از bus داخلی عبور می‌کنند:

| رویداد | زمان |
|--------|------|
| `user.created` | ساخت کاربر |
| `user.deleted` | حذف |
| `user.limited` | عبور از سقف ترافیک |
| `user.expired` | انقضا |
| `user.reset` | reset ترافیک |
| `user.ip_limit` | اشتراک‌گذاری اکانت |
| `user.expiry_warning` | ۳ روز قبل انقضا |
| `node.down` | نود unreachable |
| `node.up` | بازیابی نود |

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

### امضا

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

رویدادها به chat ادمین ارسال می‌شوند.

---

## Telegram Bot (Interactive — Admin)

ربات با long-polling:

| Command | عمل |
|---------|-----|
| `/status` | وضعیت نودها |
| `/users` | آمار کاربران |
| `/node <name>` | جزئیات نود |
| `/limit <user>` | محدود کردن user |

---

## Telegram User Bot

کاربران end-user با subscription token authenticate می‌شوند:

| Command | عمل |
|---------|-----|
| `/start` | راهنما |
| `/login <token>` | اتصال اکانت |
| `/usage` | مصرف فعلی |
| `/sub` | لینک subscription |

---

## Auto-backup to Telegram/S3

**Settings → Auto Backup**

- زمان‌بندی (cron-like)
- مقصد: Telegram document یا S3 bucket
- فایل: JSON transactional backup

---

## SSE (UI)

علاوه بر webhook، UI از همان bus با SSE subscribe می‌کند — toast و refresh خودکار.

</div>
