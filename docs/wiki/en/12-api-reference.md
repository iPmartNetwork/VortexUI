# API Reference

VortexUI exposes a complete REST API. Every UI action is backed by a documented endpoint.

---

## Base URL

```
https://your-domain.com/api
```

## Authentication

All API requests require authentication via JWT or Personal Access Token (PAT).

### Using JWT (Login)

```bash
curl -X POST https://panel.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"your-password"}'
```

Response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2026-01-16T14:32:05Z",
  "admin": {
    "id": "abc123",
    "username": "admin",
    "role": "sudo"
  }
}
```

### Using PAT (Recommended for Automation)

Create a token at **Settings → API Tokens**, then:

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

---

## Response Format

### Success

```json
{
  "data": { ... },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150
  }
}
```

### Error

```json
{
  "error": {
    "code": "user_not_found",
    "message": "User does not exist",
    "status": 404
  }
}
```

> **Translated errors (1.3.1):** The frontend maps error codes to localized messages via `getApiErrorMessage()`.

---

## Users API

### List Users

```http
GET /api/users
```

Query parameters:

| Param | Description |
|-------|-------------|
| page | Page number |
| per_page | Items per page |
| search | Filter by username |
| status | active, limited, expired |
| sort | Field to sort by |

### Get User

```http
GET /api/users/{id}
```

### Create User

```http
POST /api/users
Content-Type: application/json

{
  "username": "alice",
  "data_limit": 107374182400,
  "expire_at": "2026-02-15T00:00:00Z",
  "device_limit": 2,
  "inbound_ids": ["inb1", "inb2"]
}
```

### Update User

```http
PATCH /api/users/{id}
Content-Type: application/json

{
  "data_limit": 214748364800
}
```

### Delete User

```http
DELETE /api/users/{id}
```

### Reset User Traffic

```http
POST /api/users/{id}/reset-traffic
```

### Get User Subscription

```http
GET /api/users/{id}/subscription
```

---

## Nodes API

### List Nodes

```http
GET /api/nodes
```

### Create Node

```http
POST /api/nodes
Content-Type: application/json

{
  "name": "Frankfurt-01",
  "address": "1.2.3.4",
  "api_port": 62050,
  "core": "xray",
  "region": "EU"
}
```

### Get Node Health

```http
GET /api/nodes/{id}/health
```

Response:

```json
{
  "status": "online",
  "cpu": 42.5,
  "ram": {"used": 512, "total": 2048},
  "connections": 145,
  "latency_ms": 23
}
```

### Restart Node Core

```http
POST /api/nodes/{id}/restart
```

---

## Inbounds API

### List Inbounds

```http
GET /api/inbounds
```

### Create Inbound

```http
POST /api/inbounds
Content-Type: application/json

{
  "node_id": "n1",
  "protocol": "vless",
  "port": 443,
  "transport": "tcp",
  "security": "reality",
  "settings": {
    "dest": "www.google.com:443",
    "server_names": ["www.google.com"]
  }
}
```

---

## Capabilities API

Get the live per-protocol capability matrix:

```http
GET /api/capabilities
```

Response:

```json
{
  "xray": {
    "protocols": ["vless", "vmess", "trojan", ...],
    "transports": ["tcp", "ws", "grpc", ...],
    "security": ["none", "tls", "reality"]
  },
  "sing-box": {
    "protocols": ["vless", "hysteria2", "tuic", ...],
    ...
  }
}
```

---

## Plans & Orders API

### List Plans

```http
GET /api/plans
```

### Create Plan

```http
POST /api/plans
Content-Type: application/json

{
  "name": "Premium 100GB",
  "data_limit": 107374182400,
  "duration_days": 30,
  "price_toman": 150000,
  "price_usd": 5.0
}
```

### List Orders

```http
GET /api/orders
```

### Approve Order

```http
POST /api/orders/{id}/approve
```

---

## Wallet API

### Get Wallet Balance

```http
GET /api/wallet
```

### Get Wallet Ledger

```http
GET /api/wallet/ledger
```

### Request Top-Up (Reseller)

```http
POST /api/wallet/topup
Content-Type: application/json

{
  "amount": 100,
  "note": "Monthly top-up"
}
```

---

## Audit Log API

> **New in 1.3.0**

### List Audit Events

```http
GET /api/audit
```

Query parameters:

| Param | Description |
|-------|-------------|
| actor | Filter by admin |
| action | Filter by action type |
| from | Start date |
| to | End date |

Response:

```json
{
  "data": [
    {
      "actor": "admin",
      "action": "user.create",
      "target": "alice",
      "timestamp": "2026-01-15T14:32:05Z",
      "diff": {"before": null, "after": {...}}
    }
  ]
}
```

---

## Settings API

> **Persisted since 1.3.0**

### Get Settings

```http
GET /api/settings
```

### Update Settings

```http
PATCH /api/settings
Content-Type: application/json

{
  "panel_name": "My VPN Panel",
  "timezone": "Asia/Tehran"
}
```

---

## Federation API

> **New in 1.3.0**

### List Peers

```http
GET /api/federation/peers
```

### Add Peer

```http
POST /api/federation/peers
Content-Type: application/json

{
  "address": "https://panel-b.example.com",
  "secret": "shared-secret"
}
```

---

## Statistics API

### Dashboard Stats

```http
GET /api/stats/overview
```

### Traffic Analytics

```http
GET /api/stats/traffic?range=7d
```

### Top Users

```http
GET /api/stats/top-users?limit=10
```

---

## Health Endpoint

Public endpoint for monitoring (no auth):

```http
GET /api/health
```

Response:

```json
{
  "status": "ok",
  "version": "1.3.1",
  "components": {
    "database": "ok",
    "redis": "ok",
    "nodes": {"online": 18, "total": 20}
  }
}
```

---

## Rate Limiting

API requests are rate-limited:

| Header | Description |
|--------|-------------|
| X-RateLimit-Limit | Max requests per window |
| X-RateLimit-Remaining | Remaining requests |
| X-RateLimit-Reset | Window reset time |

When exceeded: `429 Too Many Requests`.

---

## Webhooks

See [Notifications → Webhooks](10-notifications.md#webhooks) for outbound webhook events.

---

## OpenAPI Spec

The complete machine-readable specification:

```
https://panel.example.com/api/openapi.yaml
```

Or view interactive docs at:

```
https://panel.example.com/api/docs
```

Import into Postman, Insomnia, or generate client SDKs.

---

## CLI Alternative

For scripting without HTTP, use `vortexctl`:

```bash
vortexctl user create --username alice --traffic 100GB --days 30
vortexctl node list
vortexctl backup create
```

See [Installation → CLI](02-installation.md#cli-management).

---

## SDK Examples

### Python

```python
import requests

class VortexClient:
    def __init__(self, base_url, token):
        self.base = base_url
        self.headers = {"Authorization": f"Bearer {token}"}

    def create_user(self, username, data_limit, days):
        return requests.post(
            f"{self.base}/api/users",
            headers=self.headers,
            json={
                "username": username,
                "data_limit": data_limit,
                "expire_at": days
            }
        ).json()

client = VortexClient("https://panel.example.com", "your-pat")
client.create_user("alice", 107374182400, "2026-02-15T00:00:00Z")
```

### JavaScript

```javascript
const createUser = async (username, dataLimit) => {
  const res = await fetch('https://panel.example.com/api/users', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${TOKEN}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ username, data_limit: dataLimit })
  });
  return res.json();
};
```
