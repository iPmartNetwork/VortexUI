# 10. Bildirimler

!!! tip "İpucu"
    Webhook'lar `X-Vortex-Signature: sha256=...` ile imzalanır — secret'ı env'de ayarlayın.

---

## Event Bus

Tüm domain olayları dahili bus üzerinden akar:

| Olay | Ne zaman |
|-------|------|
| `user.created` | Kullanıcı oluşturuldu |
| `user.deleted` | Kullanıcı silindi |
| `user.limited` | Trafik limiti aşıldı |
| `user.expired` | Süresi doldu |
| `user.reset` | Trafik sıfırlandı |
| `user.ip_limit` | Hesap paylaşımı |
| `user.expiry_warning` | Süre dolumundan 3 gün önce |
| `node.down` | Node erişilemez |
| `node.up` | Node kurtarıldı |

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

### İmza

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

Olaylar admin sohbetine gönderilir.

---

## Telegram Bot (Etkileşimli — Admin)

Long-polling ile bot:

| Komut | Eylem |
|---------|--------|
| `/status` | Node durumu |
| `/users` | Kullanıcı istatistikleri |
| `/node <name>` | Node ayrıntıları |
| `/limit <user>` | Kullanıcıyı sınırla |

---

## Telegram User Bot

Son kullanıcılar abonelik token'ı ile kimlik doğrular:

| Komut | Eylem |
|---------|--------|
| `/start` | Yardım |
| `/login <token>` | Hesap bağla |
| `/usage` | Mevcut kullanım |
| `/sub` | Abonelik bağlantısı |

---

## Telegram/S3'e Otomatik Yedekleme

**Settings → Auto Backup**

- Zamanlama (cron benzeri)
- Hedef: Telegram belgesi veya S3 bucket
- Dosya: JSON transactional yedekleme

---

## SSE (UI)

Webhook'lara ek olarak UI aynı bus'a SSE ile abone olur — toast'lar ve otomatik yenileme.
