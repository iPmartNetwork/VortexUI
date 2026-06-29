# Notifications

!!! info "Multi-Channel"
    VortexUI supports webhooks, Telegram, and in-app notifications. All channels
    can be configured independently and receive different event subsets.

---

## Webhook (HMAC-SHA256)

**Settings → Notifications → Webhook**

Receive structured JSON payloads for panel events at your endpoint.

### Configuration

| Setting | Description |
|---------|-------------|
| URL | Your webhook endpoint |
| Secret | HMAC-SHA256 signing key |
| Events | Select which events to send |
| Enabled | Toggle on/off |

### Verification

Every request includes an `X-Signature-256` header:

```
X-Signature-256: sha256=<hex_hmac>
```

Verify by computing `HMAC-SHA256(request_body, secret)` and comparing.

### Event Payload Structure

```json
{
  "event": "user.limited",
  "timestamp": "2025-01-15T10:30:00Z",
  "data": {
    "user_id": "uuid",
    "username": "testuser",
    "data_limit": 53687091200,
    "used_traffic": 53687091200
  }
}
```

### Available Events

| Event | Trigger |
|-------|---------|
| `user.created` | New user created |
| `user.deleted` | User deleted |
| `user.limited` | User hit data limit |
| `user.expired` | User subscription expired |
| `user.expiry_warning` | 3 days before expiry |
| `user.enabled` | User re-enabled |
| `user.disabled` | User manually disabled |
| `node.offline` | Node lost connection |
| `node.online` | Node reconnected |
| `node.unhealthy` | Node health check failed |
| `order.created` | New purchase order |
| `order.paid` | Order payment confirmed |
| `backup.completed` | Backup finished |
| `admin.login` | Admin login event |

---

## Telegram Bot

**Settings → Notifications → Telegram**

### Admin Bot

Sends notifications to the admin's Telegram chat:

| Setting | Description |
|---------|-------------|
| Bot token | From [@BotFather](https://t.me/BotFather) |
| Admin chat ID | Your personal or group chat ID |
| Events | Select notification events |

### User-Facing Bot

Users can interact with the bot using their subscription token:

| Command | Action |
|---------|--------|
| `/start` | Link account with sub token |
| `/usage` | View current data usage |
| `/renew` | Get renewal link |
| `/status` | Check account status |
| `/help` | List available commands |

### Notification Templates

Customize message format with template variables:

```
🔔 User Limited
Username: {username}
Used: {used_traffic}
Limit: {data_limit}
```

---

## Quota Notifications

**Settings → Notifications → Quota Alerts**

Alert users when they approach their data limit:

| Setting | Description |
|---------|-------------|
| Enabled | Activate quota notifications |
| Thresholds | Percentages to trigger (e.g. 80%, 90%, 100%) |
| Telegram | Send via bot |
| Webhook | Send to webhook URL |
| Cooldown | Minutes between repeated alerts |

---

## Reseller Quota Alerts

**Sidebar → Reseller Quota Alerts** (sudo admin only)

Monitor when resellers approach their traffic/user pool limits:

| Setting | Description |
|---------|-------------|
| Enabled | Global toggle |
| Telegram | Send to panel bot |
| Webhook URL | Optional external endpoint |
| Thresholds | Percentages (e.g. 80, 90, 100) |
| Cooldown | Minutes between alerts |
| Recent alerts | Table of triggered alerts |

---

## Notification Center (Bell Dropdown)

The bell icon in the panel header shows recent notifications:

- Unread count badge
- Click to expand dropdown
- Each notification shows: event type, description, timestamp
- Click to navigate to the relevant resource
- Mark as read / mark all as read

Persisted per admin account.

---

## SSE Live Events

The panel uses Server-Sent Events for real-time UI updates:

| Stream | Content |
|--------|---------|
| `/api/sse/events` | System events (node status, user limits, etc.) |
| `/api/sse/stats` | Live statistics (connections, traffic counters) |
| `/api/sse/monitor` | Active connection updates |

Frontend components subscribe automatically. No configuration needed — SSE is always active.

!!! tip
    SSE events power the real-time dashboard gauges, monitor page, and notification bell.
    If your reverse proxy buffers responses, ensure SSE streaming is allowed (Caddy handles this by default).
