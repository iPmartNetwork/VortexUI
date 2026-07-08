# Notifications

Stay informed about important events via Telegram, webhooks, email, and real-time SSE.

---

## Notification Channels

| Channel | Use Case |
|---------|----------|
| Telegram | Instant admin alerts |
| Webhooks | Automation & integrations |
| Email | Formal notifications |
| SSE | Real-time UI updates |
| In-app | Notification bell |

---

## Telegram Notifications

**Settings → Notifications → Telegram**

### Setup

1. Create a bot via [@BotFather](https://t.me/BotFather)
2. Copy the bot token
3. Paste in `VORTEX_TELEGRAM_TOKEN` or panel settings
4. Get your chat ID (message [@userinfobot](https://t.me/userinfobot))
5. Set `VORTEX_TELEGRAM_ADMIN` to your chat ID
6. Click **Test** to verify

### Configurable Events

| Event | Description |
|-------|-------------|
| User created | New user added |
| User expired | User reached expiry |
| Quota 80% | User approaching limit |
| Quota 100% | User reached limit |
| Node offline | A node went down |
| Node online | A node recovered |
| Probe detected | Active probing attempt |
| IP limit exceeded | Account sharing detected |
| Backup complete | Backup finished |
| Order received | New purchase |
| Top-up request | Reseller wants credits |

### Message Format

```
🔴 Node Offline
Node: Frankfurt-01
Address: 1.2.3.4
Offline since: 14:32:05
Affected users: 45
```

---

## Webhooks

**Settings → Notifications → Webhooks**

Send HTTP POST requests to your endpoint on events.

### Setup

1. Click **Add Webhook**
2. Configure:

| Field | Description |
|-------|-------------|
| URL | Your endpoint |
| Secret | HMAC-SHA256 signing key |
| Events | Which events to send |
| Active | Enable/disable |

### Payload Format

```json
{
  "event": "user.created",
  "timestamp": "2026-01-15T14:32:05Z",
  "data": {
    "user_id": "abc123",
    "username": "alice",
    "data_limit": 107374182400,
    "expire_at": "2026-02-15T00:00:00Z"
  }
}
```

### HMAC Verification

Verify the signature to ensure authenticity:

```python
import hmac, hashlib

def verify(payload_bytes, signature, secret):
    expected = hmac.new(
        secret.encode(),
        payload_bytes,
        hashlib.sha256
    ).hexdigest()
    return hmac.compare_digest(expected, signature)
```

The signature is sent in the `X-Vortex-Signature` header.

### Available Events

| Event | Trigger |
|-------|---------|
| user.created | User added |
| user.updated | User modified |
| user.deleted | User removed |
| user.expired | User expired |
| user.limited | User hit quota |
| user.ip_limit | Sharing detected |
| node.online | Node came up |
| node.offline | Node went down |
| order.created | New order |
| order.approved | Order confirmed |
| backup.complete | Backup done |
| acme.renewed | Certificate renewed |

---

## Email Notifications

**Settings → Notifications → Email**

### SMTP Configuration

| Field | Example |
|-------|---------|
| SMTP Host | smtp.gmail.com |
| SMTP Port | 587 |
| Username | panel@example.com |
| Password | App password |
| From Address | noreply@example.com |
| TLS | Enabled |

### Email Events

- User welcome (on creation)
- Expiry warning (3 days before)
- Quota warning (80%, 100%)
- Password reset

---

## Quota Alerts

Configure thresholds for user quota notifications.

**Settings → Notifications → Quota Thresholds**

| Threshold | Default | Action |
|-----------|---------|--------|
| Warning | 80% | Notify user + admin |
| Critical | 95% | Notify user + admin |
| Exceeded | 100% | Limit user + notify |

### Reseller Quota Alerts

Notify resellers when their pool runs low:

| Threshold | Notify |
|-----------|--------|
| 80% pool used | Reseller |
| 90% pool used | Reseller + sudo |
| 100% pool used | Reseller + sudo + auto-suspend |

---

## SSE (Server-Sent Events)

Real-time push updates to the panel UI. No configuration needed — works automatically.

### What Uses SSE

- Dashboard gauges (CPU, RAM, bandwidth)
- Live connection count
- Node status changes
- Traffic updates
- Recent events feed
- Notification bell

### Connection

The panel maintains an SSE connection to `/api/events`:

```
event: node.status
data: {"node_id":"n1","status":"offline"}

event: traffic.update
data: {"user_id":"u1","used":1073741824}
```

---

## In-App Notifications

The notification bell (top-right) shows recent events.

### Features

- Real-time updates via SSE
- Mark as read
- Clear all
- Click to navigate to related item
- Filter by type

### Keyboard Shortcut

Press `Ctrl+.` to open notifications.

---

## Notification Best Practices

### For Small Deployments
- Enable Telegram for instant alerts
- Set quota warnings at 80%
- Monitor node offline events

### For Large Deployments
- Use webhooks for automation
- Integrate with incident management (PagerDuty, Opsgenie)
- Set up email for formal user communications
- Use SSE for real-time dashboards

### For Resellers
- Enable quota alerts to avoid service interruption
- Get notified on new orders
- Monitor top-up approvals

---

## Troubleshooting

### Telegram not working
1. Verify bot token is correct
2. Ensure you've started a chat with the bot
3. Check chat ID is correct
4. Click Test button

### Webhooks not received
1. Verify endpoint URL is reachable
2. Check firewall allows outbound HTTPS
3. Verify HMAC secret matches
4. Check webhook logs in panel

### Emails not sending
1. Verify SMTP credentials
2. Check port (587 for TLS, 465 for SSL)
3. For Gmail, use app password (not account password)
4. Check spam folder
