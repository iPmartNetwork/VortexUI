# User Management

!!! abstract "User-Centric Model"
    One user = one subscription token = access to all assigned inbounds across all nodes.
    No need for separate accounts per protocol.

---

## User CRUD

### Creating a User

**Users → New User**

| Field | Description |
|-------|-------------|
| Username | Unique identifier |
| Data limit | Traffic cap (bytes) — `0` = unlimited |
| Expire at | Subscription expiration date |
| Device limit | Max concurrent devices |
| Reset strategy | `none` / `daily` / `weekly` / `monthly` |
| Status | `active` / `disabled` / `limited` |
| Inbounds | Permitted inbounds (select one or more) |
| Note | Admin-only note |

### Bulk Operations

| Action | Description |
|--------|-------------|
| **Bulk Create** | Users → Add Bulk — specify count or upload CSV |
| **Multi-select** | Checkbox multiple users → apply action |
| **Bulk actions** | Enable, disable, delete, extend, reset traffic |
| **Import** | Users → Import — from 3x-ui or Marzban database |

---

## Quotas & Limits

### Data Limit

Set a traffic cap per user. When exceeded, the user's status changes to `limited` and a `user.limited` event fires.

### Device Limit

Maximum concurrent connections (distinct IPs). Enforcement is handled by the [IP-limit system](08-security-administration.md).

### Expiration

Users expire at the configured date. The system fires `user.expiry_warning` 3 days before and `user.expired` at expiration.

### Reset Strategy

| Strategy | Behavior |
|----------|----------|
| `none` | Never resets — traffic accumulates forever |
| `daily` | Reset at midnight UTC |
| `weekly` | Reset every Monday at midnight UTC |
| `monthly` | Reset on the 1st at midnight UTC |

---

## Subscription Delivery

### Subscription Link

Every user gets a unique subscription URL:

```
https://panel.example.com/sub/{token}
```

The response format is auto-detected from the client's User-Agent, or forced with `?format=`:

| Format | Parameter | Output |
|--------|-----------|--------|
| Base64 | `?format=base64` | V2Ray-compatible base64 encoded links |
| Clash | `?format=clash` | Clash YAML configuration |
| sing-box | `?format=singbox` | sing-box JSON configuration |
| Xray JSON | `?format=xray` | Raw Xray/V2Ray client JSON |
| Outline | `?format=outline` | `ss://` Shadowsocks links for Outline |
| Plain links | `?format=links` | V2rayN-style share links, one per line |

### Subscription Hosts

Per-inbound overrides that change what the subscription advertises — without touching the live core config.

| Field | Description |
|-------|-------------|
| Remark | Display name (supports template variables) |
| Address | Advertised host/IP (e.g. a CDN domain) |
| Port | Override port (`0` = inherit) |
| SNI | TLS server name |
| Host | HTTP `Host` header |
| Path | WS/HTTPUpgrade/gRPC path |
| ALPN | Comma-separated (e.g. `h2,http/1.1`) |
| Fingerprint | uTLS fingerprint (e.g. `chrome`) |
| Security | `inbound_default`, `none`, `tls`, or `reality` |
| Allow insecure | Skip certificate verification |
| Mux | Enable multiplexing |
| Fragment | TLS fragment spec (e.g. `1,40-60,30-50`) |
| Priority | Order within the inbound (lower renders first) |

**Template Variables** — rendered per user:

| Variable | Expands to |
|----------|-----------|
| `{USERNAME}` | The user's username |
| `{SERVER_IP}` | The node's address |
| `{SERVER_PORT}` | The advertised port |
| `{PROTOCOL}` | The inbound protocol |
| `{NETWORK}` | The transport type |
| `{SECURITY}` | The transport security |
| `{REMARK}` | The configured remark |

!!! tip
    Use subscription hosts to front a single inbound behind multiple CDN domains, each with different SNI and path — all from one host definition using template variables.

---

## Self-Service Portal

**End-user URL:** `/portal/login`

Users log in with their subscription token and access:

| Feature | Description |
|---------|-------------|
| Dashboard | Usage stats, remaining data/time, active devices |
| Plans | Browse and purchase subscription plans (from their reseller's shop) |
| Tickets | Open support tickets, reply to admin messages |
| Referral | View/share referral code, see earned rewards |
| QR Code | Scan to import subscription into mobile apps |

### Self-Service Shop

**URL:** `/sub/{token}/shop`

Each reseller configures their own plans and payment methods. End-users see only their reseller's offerings:

1. User browses available plans
2. Selects a plan and payment method (ZarinPal, card-to-card, or crypto)
3. Completes payment (or uploads proof)
4. Order is fulfilled automatically (ZarinPal) or after admin approval (card/crypto)

See [Plans & Payments](09-plans-payments.md) for full details.

---

## Family/Group Subscriptions

**Users → Family Groups**

Let users share a data pool:

1. Create a **Family Group** with a shared data limit
2. Add existing users as members
3. Each member's traffic draws from the shared pool
4. Optional per-member cap within the shared pool

| Field | Description |
|-------|-------------|
| Name | Group name |
| Owner | Primary account |
| Data limit | Total shared pool |
| Max members | Limit (default: 5) |
| Member quota | Per-member cap within the pool |

---

## Smart Quota

Progressive speed reduction instead of hard-cutting users at their limit.

Configure tiers as JSON:

```json
[
  { "threshold_pct": 80, "action": "warn", "speed_limit": 0 },
  { "threshold_pct": 95, "action": "throttle", "speed_limit": 524288 },
  { "threshold_pct": 100, "action": "disable" }
]
```

- At **80%** → warn the user (notification event)
- At **95%** → throttle to 512 KB/s
- At **100%** → disable the account

!!! info
    Smart Quota is configured per plan or globally. Per-plan settings override the global default.

---

## Referral System

**Users → Referrals**

| Setting | Description | Default |
|---------|-------------|---------|
| Enabled | Turn referrals on/off | Off |
| Reward type | `data` (extra GB) or `days` (extra time) | `data` |
| Reward amount | How much per referral | 1 GB |
| Max referrals | Limit per user (`0` = unlimited) | `0` |
| Require paid | Only reward for paying referrals | Off |
| Both rewarded | Reward referrer + new user | Yes |

Users access their referral code via the portal. When a friend signs up using the code, both parties are rewarded.

---

## Config Templates

Customize the Clash/sing-box output that users receive in their subscription:

- Add custom routing rules (block ads, direct local traffic)
- Configure DNS settings
- Set proxy group strategies (url-test, fallback, load-balance)
- Per-user or global templates

---

## Deep Links & QR Codes

**System → Deep Links**

Generate subscription deep links for one-tap app setup:

| Setting | Description |
|---------|-------------|
| Base URL | Panel's public URL |
| App scheme | URL scheme for native apps |
| Include name | Add server name to the link |
| QR logo | Custom logo in QR center |

---

## Import from Other Panels

**Users → Import**

Migrate users from:

- **3x-ui** — provide database file path
- **Marzban** — provide database connection or export file

The importer maps users, preserving usernames, traffic limits, and expiration dates.
