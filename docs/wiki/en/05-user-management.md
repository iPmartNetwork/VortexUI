<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/05-user-management.md) · [AR](../ar/05-user-management.md) · [TR](../tr/05-user-management.md)

</div>

<div>

# 5. User Management

[← Dashboard](./04-dashboard.md) · [Index](./README.md) · [Next: Nodes →](./06-node-management.md)

> [!WARNING]
> **Revoke Sub** invalidates the old link — use only if the token was leaked.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![Users page — user management and subscriptions](../assets/panel/User_light.png) | ![Users page — user management and subscriptions](../assets/panel/User_dark.png) |

*Users page — user management and subscriptions*

</div>

---

## Philosophy: One User, Multiple Protocols

Each **User** is a single record. With one **subscription token** they access all permitted inbounds — no need to create separate entries for VLESS and VMess.

---

## Creating a User

**Users → New User**

| Field | Description |
|-------|-------------|
| **Username** | Unique identifier |
| **Data limit** | Traffic cap (bytes) — 0 = unlimited |
| **Expire at** | Expiration date |
| **Device limit** | Max concurrent devices (IP/HWID) |
| **Reset strategy** | `none` / `daily` / `weekly` / `monthly` |
| **Status** | `active` / `disabled` / `limited` |
| **Inbounds** | Permitted inbounds |
| **Note** | Admin note |

---

## Bulk Operations

| Action | Path |
|--------|------|
| **Bulk Create** | Users → Add Bulk — CSV/count |
| **Multi-select** | Select multiple users → action |
| **Import** | Users → Import — 3x-ui / Marzban |

---

## Subscription

### Links

| Endpoint | Output |
|----------|--------|
| `GET /sub/{token}` | base64 (default) |
| `GET /sub/{token}?format=clash` | Clash YAML |
| `GET /sub/{token}?format=singbox` | sing-box JSON |
| `GET /sub/info/{token}` | User HTML page |

### Revoke

**Users → Revoke Sub** — a new token is issued; the previous link is invalidated.

---

## Traffic Accounting

- **Delta push** method: the core pushes deltas (not polling)
- **Restart-safe**: counters stored in DB
- **Reset**: manual or scheduled (monthly, etc.)
- **Quota enforcement**: exceeding the cap → status `limited` + `user.limited` event

---

## Device Limit & HWID

| Mechanism | Description |
|-----------|-------------|
| **Device limit** | Concurrent IP/distinct device count |
| **HWID allowlist** | Registered devices only |
| **Online IP guard** | Compare online IP with limit — [Chapter 8](./08-security-administration.md) |

---

## User Detail Page

**Users → click username** → `/users/:id`

- Usage chart
- Online IPs
- Reset history
- Inline editing

---

## Auto-Select Best Node

In subscriptions you can enable url-test to automatically select the lowest-latency active node (in the config template).

---

## User Notifications

- **Telegram user bot**: user logs in with token — `/usage`, `/renew`
- **Expiry warning**: 3 days before — `user.expiry_warning` event

</div>
