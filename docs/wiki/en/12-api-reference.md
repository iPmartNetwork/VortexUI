# API Reference

!!! info "OpenAPI Spec"
    Full OpenAPI 3.0 specification available at
    [`docs/openapi.yaml`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml).

---

## Authentication

### JWT Login

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password",
  "totp_code": "123456"  // optional, if 2FA enabled
}
```

Response:

```json
{
  "access_token": "eyJhbG...",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

Use the token in subsequent requests:

```
Authorization: Bearer <access_token>
```

### API Tokens (PAT)

For automation, create a Personal Access Token:

1. **Settings → API Tokens → Create**
2. Use like a JWT: `Authorization: Bearer <PAT>`

PATs don't expire unless configured to, and can be revoked individually.

---

## Base URL & Versioning

```
https://panel.example.com/api/
```

All endpoints are under `/api/`. No version prefix — the API is forward-compatible.

---

## Key Endpoints

### Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | Login, get JWT |
| POST | `/api/auth/refresh` | Refresh token |
| GET | `/api/auth/me` | Current admin info |

### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users` | List users (pagination, filters) |
| POST | `/api/users` | Create user |
| GET | `/api/users/:id` | Get user detail |
| PUT | `/api/users/:id` | Update user |
| DELETE | `/api/users/:id` | Delete user |
| POST | `/api/users/:id/reset-traffic` | Reset traffic counter |
| POST | `/api/users/:id/revoke-sub` | Revoke subscription token |
| GET | `/api/users/:id/usage` | Usage history |

### Nodes

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/nodes` | List nodes |
| POST | `/api/nodes` | Create node |
| GET | `/api/nodes/:id` | Get node detail |
| PUT | `/api/nodes/:id` | Update node |
| DELETE | `/api/nodes/:id` | Delete node |
| POST | `/api/nodes/:id/restart` | Restart node core |
| GET | `/api/nodes/:id/health` | Health status |
| GET | `/api/nodes/:id/stats` | Live statistics |

### Inbounds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/inbounds` | List inbounds |
| POST | `/api/inbounds` | Create inbound |
| PUT | `/api/inbounds/:id` | Update inbound |
| DELETE | `/api/inbounds/:id` | Delete inbound |
| GET | `/api/capabilities` | Per-protocol capability matrix |

### Plans

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/plans` | List plans (scoped to admin) |
| POST | `/api/plans` | Create plan |
| PUT | `/api/plans/:id` | Update plan |
| DELETE | `/api/plans/:id` | Delete plan |

### Orders

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/orders` | List orders |
| GET | `/api/orders/:id` | Get order detail |
| POST | `/api/orders/:id/approve` | Approve pending order |
| POST | `/api/orders/:id/reject` | Reject pending order |

### Payment Configuration

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/payment-config` | Get payment config |
| PUT | `/api/payment-config` | Update payment config |

### Admins & Reseller

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admins` | List admins |
| POST | `/api/admins` | Create admin |
| PUT | `/api/admins/:id` | Update admin |
| POST | `/api/admins/:id/quota-adjust` | Adjust reseller quota |
| POST | `/api/admins/:id/unsuspend` | Clear suspension |
| POST | `/api/admins/:id/impersonate` | Issue reseller token |

### Reseller Account

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/account/dashboard` | Reseller dashboard stats |
| GET | `/api/account/export/users` | CSV export of owned users |
| GET | `/api/account/wallet` | Wallet + ledger |
| GET/PUT | `/api/account/branding` | Whitelabel settings |
| GET/PUT | `/api/account/webhook` | Outbound webhook config |
| GET/POST | `/api/account/sub-admins` | Sub-reseller management |

### System

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | Health check |
| GET | `/api/stats` | System statistics |
| GET | `/api/audit` | Audit log |
| POST | `/api/backup` | Trigger backup |
| GET | `/api/settings` | Panel settings |
| PUT | `/api/settings` | Update settings |

### Subscriptions (Public)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/sub/{token}` | User subscription (auto-format) |
| GET | `/sub/{token}?format=clash` | Clash YAML |
| GET | `/sub/{token}?format=singbox` | sing-box JSON |
| GET | `/sub/{token}?format=xray` | Xray JSON |
| GET | `/sub/{token}?format=outline` | Outline links |
| GET | `/sub/{token}?format=links` | Plain share links |
| GET | `/sub/info/{token}` | User info HTML page |
| GET | `/sub/{token}/shop` | Self-service shop |

---

## Example Requests

### Login

```bash
curl -X POST https://panel.example.com/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret"}'
```

### Create User

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

### List Plans

```bash
curl https://panel.example.com/api/plans \
  -H "Authorization: Bearer <token>"
```

### Purchase (Portal)

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

## Pagination

List endpoints support pagination:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `page` | Page number (1-indexed) | 1 |
| `per_page` | Items per page | 20 |
| `sort` | Sort field | `created_at` |
| `order` | `asc` or `desc` | `desc` |

Response includes pagination metadata:

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

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "USER_NOT_FOUND",
    "message": "User with the specified ID does not exist",
    "status": 404
  }
}
```

Common HTTP status codes:

| Code | Meaning |
|------|---------|
| 400 | Bad request (validation error) |
| 401 | Unauthorized (invalid/expired token) |
| 403 | Forbidden (insufficient permissions) |
| 404 | Resource not found |
| 409 | Conflict (duplicate username, etc.) |
| 429 | Rate limited |
| 500 | Internal server error |
