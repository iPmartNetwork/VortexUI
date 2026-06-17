# 12. API Referansı

!!! note
    Tam şema: [`docs/openapi.yaml`](../../openapi.yaml) — tüm endpoint'ler ve RBAC.

---

## OpenAPI

Tam spesifikasyon: [`docs/openapi.yaml`](../../openapi.yaml)

### Etkileşimli görüntüleme

```bash
npx @redocly/cli preview-docs docs/openapi.yaml
```

```bash
docker run -p 8081:8080 \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  -v "$PWD/docs:/spec" \
  swaggerapi/swagger-ui
```

---

## Kimlik Doğrulama

### Login

```bash
curl -X POST https://panel.example.com/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"secret","totp":"123456"}'
```

Response: `{ "token": "eyJ..." }`

### Usage

```http
Authorization: Bearer eyJ...
```

### Public endpoints (no JWT)

| Endpoint | Açıklama |
|----------|-------------|
| `GET /api/health` | Health check |
| `GET /sub/{token}` | Subscription |
| `GET /sub/info/{token}` | User HTML page |
| `POST /api/payment/ipn/*` | Payment webhooks |

---

## Main Endpoints

### Account

| Method | Path | Permission |
|--------|------|------------|
| POST | `/api/login` | public |
| POST | `/api/account/password` | authenticated |
| POST | `/api/account/2fa/setup` | authenticated |
| POST | `/api/account/2fa/confirm` | authenticated |
| POST | `/api/account/2fa/disable` | authenticated |

### Overview & Logs

| Method | Path |
|--------|------|
| GET | `/api/overview` |
| GET | `/api/logs` |
| GET | `/api/events/stream` |

### Users

| Method | Path |
|--------|------|
| GET/POST | `/api/users` |
| GET/PATCH/DELETE | `/api/users/{id}` |
| GET | `/api/users/{id}/sub` |
| GET | `/api/users/{id}/online` |
| POST | `/api/users/{id}/reset` |
| POST | `/api/users/{id}/revoke-sub` |

### Nodes & Inbounds

| Method | Path |
|--------|------|
| GET/POST | `/api/nodes` |
| GET/PATCH/DELETE | `/api/nodes/{id}` |
| GET | `/api/nodes/{id}/logs` |
| POST | `/api/nodes/{id}/geo-update` |
| GET/POST | `/api/inbounds` |
| GET/PATCH/DELETE | `/api/inbounds/{id}` |

### Network Policy

| Method | Path |
|--------|------|
| GET/POST | `/api/outbounds` |
| GET/PATCH/DELETE | `/api/outbounds/{id}` |
| GET/POST | `/api/routing` |
| GET/PATCH/DELETE | `/api/routing/{id}` |
| GET/POST | `/api/balancers` |
| GET/PATCH/DELETE | `/api/balancers/{id}` |
| POST | `/api/reality/keypair` |

### Admin & Backup

| Method | Path |
|--------|------|
| GET/POST | `/api/admins` |
| GET/PATCH/DELETE | `/api/admins/{id}` |
| GET/POST | `/api/roles` |
| GET | `/api/backup` |
| POST | `/api/backup/restore` |
| GET/POST/DELETE | `/api/tokens` |

### Plans

| Method | Path |
|--------|------|
| GET/POST | `/api/plans` |
| GET/PATCH/DELETE | `/api/plans/{id}` |
| GET | `/api/orders` |

---

## RBAC Permissions

Her mutation route bir izin gerektirir, örneğin:

- `users.read`, `users.write`, `users.delete`
- `nodes.read`, `nodes.write`
- `inbounds.write`, `routing.write`
- `admins.write`, `backup.restore`

Hem PAT hem JWT taşıyıcının rol izinlerini yansıtır.

---

## Example: Create a User

```bash
TOKEN="eyJ..."

curl -X POST https://panel.example.com/api/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "api-user",
    "data_limit": 53687091200,
    "expire_at": "2026-07-17T00:00:00Z",
    "device_limit": 3,
    "inbound_ids": ["inbound-uuid-here"]
  }'
```

---

## Client Generation

```bash
npx openapi-typescript docs/openapi.yaml -o web/src/api/types.ts
```
