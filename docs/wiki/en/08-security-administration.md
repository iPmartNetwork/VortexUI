# Security & Administration

!!! important "First Priority"
    Enable **2FA** and set a strong JWT secret (≥32 bytes) for the primary admin before exposing the panel to the internet.

---

## Authentication

| Layer | Mechanism |
|-------|-----------|
| Panel login | JWT (Bearer) — configurable TTL |
| 2FA | TOTP (Google Authenticator, Authy, etc.) |
| API automation | Personal Access Token (PAT) |
| Panel ↔ Node | mTLS (mutual certificates) |
| Portal login | Subscription token |

### Enabling 2FA (TOTP)

1. **Settings → Two-Factor Authentication**
2. Click **Enable** → scan QR code with your TOTP app
3. Enter the 6-digit code to confirm
4. Store recovery codes securely
5. To disable: enter current code → click **Disable**

---

## RBAC & Roles

**Settings → Admins** (sudo) — sub-tabs: **Admins**, **Roles**, **Reseller access**.

Click a reseller username to open their **profile** (`/settings/admins/:id`) with wallet,
quota usage, ledger, and policy limits.

| Concept | Description |
|---------|-------------|
| Role | Named set of permissions |
| Permission | Granular access: `users.read`, `nodes.write`, `plans.manage`, etc. |
| Reseller quota | User count + traffic cap for non-sudo admins |
| Scoped view | Admin sees only their own users/data |

### Built-in Roles

| Role | Access |
|------|--------|
| **Sudo** | Full access to everything |
| **Admin** | Manage users, nodes, inbounds (no system settings) |
| **Reseller** | Manage own users, own plans, wallet, branding |

---

## Reseller Platform

VortexUI's reseller system turns each non-sudo admin into a complete sub-panel operator.

### Reseller Features

| Feature | Description |
|---------|-------------|
| **Owned users** | Reseller manages only users they created |
| **Owned plans** | Reseller creates their own plans with their own pricing |
| **Payment config** | Reseller configures their own payment methods |
| **Wallet** | Traffic + user credits, ledger, top-up requests |
| **Sub-resellers** | Create child resellers with inherited scope |
| **Whitelabel** | Custom branding (logo, title, accent color, footer) |
| **Webhooks** | Outbound events (user.created, user.deleted) for automation |
| **Dashboard** | Reseller-specific stats (account count, traffic, top users) |
| **CSV export** | Export owned users list |

### Scoped Allowlists

**Admins → Edit admin** — restrict what a reseller can access:

| Allowlist | Effect |
|-----------|--------|
| Plans | Only listed plans visible in their shop |
| Nodes | Reseller can only bind users to these nodes |
| Inbounds | Reseller can only attach users to these inbounds |

Empty allowlist = no restriction (all items visible). Sudo admins are never scoped.

### Traffic Quota Modes

| Mode | Counts against pool when... |
|------|---------------------------|
| **Allocated** | Reseller assigns data limits to users (sum of limits) |
| **Consumed** | Users actually consume traffic |

### Per-Reseller Payment Configuration

Each reseller configures their own payment methods:

| Method | Configuration |
|--------|--------------|
| ZarinPal | Merchant ID |
| Card-to-card | Card number + holder name |
| Crypto | Wallet addresses (BTC, USDT, etc.) |

Users in that reseller's shop see only their reseller's payment options.

### Per-Reseller Owned Plans

Each reseller creates plans visible only to their users:

- Reseller sets price, data limit, duration, device limit
- Plans appear in `/sub/{token}/shop` for that reseller's users only
- Admin plans are separate from reseller plans

### Reseller Wallet

| Credit Type | Purpose |
|-------------|---------|
| Traffic credits | Consumed when users use data |
| User credits | Consumed when creating new users |

Wallet operations:

- **Top-up request** — reseller requests credits, sudo approves
- **Auto-deduct** — system deducts as users consume/are created
- **Ledger** — full history of all credit changes with reasons

### Sub-Resellers

A reseller can create child resellers:

- Child inherits parent's scope (can't exceed parent's allowlists)
- Parent manages child's wallet
- Parent sees child's users and stats

### Policy Limits (per reseller)

Set on **Admins → Edit admin → Policy**:

| Setting | Effect |
|---------|--------|
| Max data limit (GB) | Cap per-user data limit the reseller may set |
| Max expire (days) | Cap subscription length |
| Allow bulk create/import | Gate bulk user creation |
| Allow bulk delete | Gate multi-delete |

### Auto-Suspend

| Setting | Effect |
|---------|--------|
| Enable auto-suspend | Master toggle |
| IP violations (7d) | Suspend after N sharing-detection events |
| Quota grace (minutes) | Grace period after quota exceeded |

Suspended resellers cannot log in until a sudo admin clicks **Unsuspend**.

### Sudo Admin Tools

| Action | Description |
|--------|-------------|
| Quota usage table | Per-reseller accounts, traffic, pool remaining |
| Quick quota adjust | +50 accounts / +10 GB / +50 GB buttons |
| Login as (impersonate) | Issue reseller JWT session |
| Unsuspend | Clear auto-suspension flag |
| Quota alerts | Configure threshold notifications for all resellers |

---

## TLS Tricks Manager

**Security → TLS Tricks**

ISP-specific profiles that combine multiple DPI bypass techniques:

| Technique | Description |
|-----------|-------------|
| Fragment | Split TLS ClientHello into small packets |
| Mux | Multiplex connections into one stream |
| Padding | Add random padding to packets |

### Creating a Profile

1. Click **New Profile**
2. Name it (e.g. "Iran — Fragment + Chrome")
3. Configure:
    - Fingerprint: Chrome, Firefox, Safari, Random, Randomized
    - Fragment: Enable + length range (e.g. `10-30`) + interval
    - Mux: Enable + protocol (smux, yamux, h2mux)
4. Save → assign to inbounds

!!! example "Country Presets"
    - **Iran**: Fragment `10-30` + Chrome fingerprint
    - **China**: Mux h2mux + Randomized fingerprint
    - **Russia**: Fragment `1-3` + Firefox fingerprint

---

## Probing Protection

**Security → Probing Protection**

Detect and block active probing attempts from censors (GFW, TSPU).

| Action | Behavior |
|--------|----------|
| **Block** | Drop connection + ban IP for configured duration |
| **Honeypot** | Return a fake website to fool the prober |
| **Log only** | Record without action (monitoring mode) |

Configuration:

- Block duration (default: 3600s)
- Max probes/min threshold (default: 5)
- Trusted IP whitelist
- Telegram alert on detection

---

## Client Fingerprint Validation (JA3)

**Security → Fingerprint**

Block connections based on TLS ClientHello fingerprints. Known scanner tools (curl, Go HTTP, Python requests) produce distinctive fingerprints.

| Setting | Description |
|---------|-------------|
| Enabled | Activate fingerprint checking |
| Default action | Unknown fingerprints: Allow / Block / Log |
| Rules | Explicit allow/block per fingerprint or JA3 hash |

!!! example
    Block scanner tools:
    - Rule 1: `fingerprint=curl`, action=block
    - Rule 2: `fingerprint=python`, action=block

---

## Decoy Website

**Security → Decoy Website**

Show a fake website when someone visits your server's IP directly:

| Mode | Behavior |
|------|----------|
| **Proxy** | Reverse-proxy an existing website (mirrors it) |
| **Static** | Serve custom HTML |

Makes your server look like a normal website to censors and casual visitors.

---

## Evasion Profiles

**Security → Evasion Profiles**

Pre-configured anti-DPI technique bundles assigned to inbounds for one-click censorship evasion. Combines fragment, mux, and fingerprint settings into a named profile.

---

## WARP+ Integration

**Network → Outbounds → WARP+**

Route traffic through Cloudflare WARP to get a clean IP:

- Free tier or with WARP+ license key
- Assign to routing rules for specific domains
- Useful when node IP is flagged by services

---

## DNS-over-HTTPS (DoH)

**Security → DNS-over-HTTPS**

Built-in DoH server preventing DNS leaks:

| Setting | Description | Default |
|---------|-------------|---------|
| Enabled | Turn DoH on/off | Off |
| Listen address | Bind address | `:8053` |
| Upstream DNS | Resolvers to forward to | `1.1.1.1`, `8.8.8.8` |
| Block ads | Filter ad domains | Off |
| Block malware | Filter malware domains | On |
| Custom blocklist | Your own blocked domains | Empty |
| Cache TTL | Cache duration (seconds) | 300 |

---

## Clean-IP Scanner

**Network → Clean IP**

Find well-performing CDN edge IPs by scanning and scoring on latency and packet loss.

1. Paste candidate IPs (one per line)
2. Set port (default: 443)
3. Click **Scan** — results scored and sorted best-first
4. Use top IPs in subscription hosts or CDN chains

!!! warning "SSRF Protection"
    Private, loopback, and link-local ranges are rejected to prevent internal network scanning.

---

## IP-Limit Enforcement

**Security → IP Limit**

Per-user concurrent IP/device caps to prevent account sharing.

| Setting | Description |
|---------|-------------|
| Enabled | Toggle enforcement |
| Action | `warn`, `disable_temporarily`, or `kill_connections` |
| Alert cooldown | Seconds between repeated alerts |
| Restore after | Seconds before a temp-disabled user is restored |

!!! note
    `kill_connections` is Xray-only. On sing-box nodes it degrades to `disable_temporarily`.

---

## IP Whitelist/Blacklist

**Settings → IP Guard**

Restrict panel API and subscription access by IP:

- **Whitelist mode** — only listed IPs can access
- **Blacklist mode** — listed IPs are blocked
- Supports CIDR ranges

---

## Geo-Blocking per Inbound

Restrict which countries can connect to a specific inbound:

- Configure on the inbound edit page
- Comma-separated ISO 3166-1 alpha-2 codes
- Empty = all countries allowed

---

## Account-Sharing Guard

Background loop comparing online IPs with device limits:

| Mode | Behavior |
|------|----------|
| Detection (default) | Fire `user.ip_limit` event + webhook/Telegram |
| Auto-limit (`VORTEX_SHARE_AUTOLIMIT=true`) | Automatically limit the user |

---

## Audit Log

**Audit** — records all admin mutations:

| Field | Content |
|-------|---------|
| Actor | Admin username |
| Action | `user.create`, `inbound.update`, `admin.login`, etc. |
| Target | Resource ID |
| Timestamp | ISO 8601 |
| Diff | Before/after JSON |

Resellers see only their own actions. Sudo admins see all.

---

## API Tokens (PAT)

**Settings → API Tokens**

Create personal access tokens for automation:

```bash
curl -H "Authorization: Bearer <PAT>" \
  https://panel.example.com/api/users
```

- Each token can be revoked individually
- Permissions inherit from the creating admin's role
- Set optional expiration date

---

## Security Checklist

- [ ] Strong JWT secret (≥32 byte random)
- [ ] HTTPS enabled (Let's Encrypt via Caddy)
- [ ] TOTP 2FA for sudo admin
- [ ] API tokens with least privilege
- [ ] Encrypted off-site backup
- [ ] Webhook secret for HMAC verification
- [ ] Panel port closed from public (Caddy 443 only)
- [ ] Probing protection enabled
- [ ] IP-limit enforcement configured
- [ ] Decoy website active
