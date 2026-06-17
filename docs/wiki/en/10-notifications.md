<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/10-notifications.md) · [AR](../ar/10-notifications.md) · [TR](../tr/10-notifications.md)

</div>

<div>

# 10. Notifications

[← Plans](./09-plans-payments.md) · [Index](./README.md) · [Next: Settings →](./11-settings-backup.md)

> [!TIP]
> Webhooks are signed with `X-Vortex-Signature: sha256=...` — set the secret in env.

---

## Event Bus

All domain events flow through the internal bus:

| Event | When |
|-------|------|
| `user.created` | User created |
| `user.deleted` | User deleted |
| `user.limited` | Traffic cap exceeded |
| `user.expired` | Expired |
| `user.reset` | Traffic reset |
| `user.ip_limit` | Account sharing |
| `user.expiry_warning` | 3 days before expiry |
| `node.down` | Node unreachable |
| `node.up` | Node recovered |

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

### Signature

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

Events are sent to the admin chat.

---

## Telegram Bot (Interactive — Admin)

Bot with long-polling:

| Command | Action |
|---------|--------|
| `/status` | Node status |
| `/users` | User stats |
| `/node <name>` | Node details |
| `/limit <user>` | Limit user |

---

## Telegram User Bot

End users authenticate with subscription token:

| Command | Action |
|---------|--------|
| `/start` | Help |
| `/login <token>` | Link account |
| `/usage` | Current usage |
| `/sub` | Subscription link |

---

## Auto-Backup to Telegram/S3

**Settings → Auto Backup**

- Schedule (cron-like)
- Destination: Telegram document or S3 bucket
- File: JSON transactional backup

---

## SSE (UI)

In addition to webhooks, the UI subscribes to the same bus via SSE — toasts and automatic refresh.

</div>
