# Menu & Usage Guide

This page explains every main VortexUI menu, what it is used for, and the most common workflows after installation.

---

## Main Navigation Overview

| Menu | Purpose | Who Uses It |
|------|---------|-------------|
| Overview / Dashboard | Real-time system overview, fleet status, traffic, widgets | Sudo, Admin, Reseller |
| Monitor | Live active connections table | Sudo, Admin, Support |
| Analytics | Traffic analytics by country, user, protocol, time | Sudo, Admin, Reseller |
| Users | Create, edit, suspend, renew, and export users | Sudo, Admin, Reseller |
| Families | Shared traffic pools for multiple users | Sudo, Admin, Reseller |
| Nodes | Register and monitor local or remote proxy servers | Sudo, Admin |
| Inbounds | Configure client-facing protocols and ports | Sudo, Admin |
| Routing | Routing packs, balancers, outbounds, WARP, CDN chains | Sudo, Admin |
| Evasion / Security Suite | Reality, Clean-IP, TLS Tricks, Decoy, probing protection | Sudo, Admin |
| Plans & Payments | Plans, orders, payment methods, invoices | Sudo, Reseller |
| Wallet Billing | Reseller wallet, top-ups, ledger, quotas | Sudo, Reseller |
| Tickets | End-user support tickets from portal | Sudo, Admin, Reseller |
| Audit | Admin action history and diffs | Sudo, Auditor |
| Settings | Global settings, admins, branding, backup, ACME, API tokens | Sudo, scoped Admins |

---

## Dashboard / Overview

Use the dashboard as the command center for daily operations.

### What You See

| Area | Meaning |
|------|---------|
| User summary | Total, active, expired, limited users |
| Traffic summary | Today, month, upload, download |
| Node fleet | Online/offline nodes and health indicators |
| Real-time gauges | CPU, RAM, disk, bandwidth, connections |
| Recent events | New users, node status changes, payments, alerts |
| World map | GeoIP traffic and connection heatmap |

### Typical Actions

1. Check if all nodes are online.
2. Watch traffic spikes and active connections.
3. Open the command palette with `Ctrl+K`.
4. Jump quickly to users, nodes, or actions.
5. Customize widgets by dragging and resizing.

---

## Monitor

The Monitor page shows live connections.

| Column | Description |
|--------|-------------|
| User | Connected username |
| Node | Node handling the connection |
| IP | Client source IP |
| Protocol | VLESS, VMess, Trojan, Hysteria2, etc. |
| Duration | How long the session has been active |
| Traffic | Upload/download for the current session |

### Use Cases

- Find account sharing by checking multiple IPs.
- Debug whether a user is actually connected.
- Identify overloaded nodes.
- Kill suspicious sessions if supported by the core.

---

## Analytics

Analytics helps you understand consumption and plan capacity.

### Available Views

| View | Use |
|-----|-----|
| Traffic over time | Identify peak hours |
| Top users | Find heavy consumers |
| Traffic by country | Understand geographic demand |
| Protocol breakdown | See which protocols are most used |
| CSV export | Reporting and billing |

### Recommended Routine

1. Review the last 7 days weekly.
2. Export CSV for reseller or billing reports.
3. Add nodes when peak usage approaches capacity.

---

## Users

The Users menu is where you manage subscriber accounts.

### Main Actions

| Action | Description |
|--------|-------------|
| Add User | Create a new user with quota and expiry |
| Edit User | Change limit, expiry, device cap, inbounds |
| Suspend | Temporarily block access |
| Reset Traffic | Set used traffic to zero |
| Copy Subscription | Copy `/sub/{token}` URL |
| QR Code | Show QR for mobile import |
| Export | Download users as CSV |

### Create a Standard User

1. Go to **Users → Add User**.
2. Enter username.
3. Set data limit, expiry date, and device limit.
4. Select inbounds the user can access.
5. Save.
6. Copy the subscription link and send it to the user.

### Renew a User

1. Open the user profile.
2. Increase expiry date or add duration.
3. Increase traffic limit if needed.
4. Save.

---

## Families

Families group multiple users under one shared data pool.

### When to Use

- Household plans
- Team subscriptions
- Shared company packages

### Workflow

1. Create a family group.
2. Set the total traffic pool.
3. Add member users.
4. Monitor family usage from the family detail page.

---

## Nodes

Nodes are servers that run Xray-core or sing-box.

### Node Types

| Type | Use Case |
|------|----------|
| Local Node | Single-server deployments |
| Remote Node | Multi-server production deployments |

### Add a Remote Node

1. Go to **Nodes → Add Node**.
2. Enter name, address, core type, region.
3. Copy the generated install command.
4. Run it on the remote server.
5. Wait until the node becomes online.

### Daily Checks

- Online/offline status
- CPU and RAM usage
- Ping latency
- Active connections
- Core version

---

## Inbounds

Inbounds are the client-facing listeners: protocol, port, transport, and security.

### Recommended Inbounds

| Scenario | Recommended Setup |
|----------|-------------------|
| General censorship resistance | VLESS + TCP + REALITY |
| CDN fronting | VMess/VLESS + WebSocket + TLS |
| High performance | Trojan + gRPC + TLS |
| Lossy networks | Hysteria2 or TUIC |
| Native tunnel | WireGuard via sing-box |

### Create VLESS + REALITY

1. Go to **Inbounds → Add Inbound**.
2. Select node.
3. Protocol: `VLESS`.
4. Transport: `TCP`.
5. Security: `REALITY`.
6. Set port `443`.
7. Use Reality Scanner to choose SNI.
8. Save and assign users.

---

## Routing

Routing controls where traffic goes after entering a node.

### Tabs

| Tab | Purpose |
|-----|---------|
| Packs | Reusable routing rule bundles |
| Balancers | Distribute traffic across nodes |
| Outbounds | Direct, block, WARP, proxy chain |
| CDN Chains | Multi-hop CDN/relay paths |

### Common Tasks

- Enable ad-blocking pack.
- Route local country domains directly.
- Send selected domains through WARP+.
- Create a latency-based balancer.
- Build CDN → Relay → Exit chains.

---

## Evasion / Security Suite

The Evasion menu contains anti-censorship features.

### Tabs

| Tab | Purpose |
|-----|---------|
| Reality | Manage REALITY targets, keys, short IDs |
| Clean IP | Scan CDN edge IPs and latency |
| TLS Tricks | Fragment, mux, padding, fingerprint presets |
| Decoy | Show fake website to direct visitors/probers |
| Probing Protection | Detect and block active probing attempts |

### Recommended Anti-Censorship Setup

1. Run Reality Scanner.
2. Create VLESS + REALITY inbound.
3. Enable probing protection.
4. Enable Decoy Website.
5. Apply a TLS Tricks profile for the target ISP/country.
6. Test from the real user network.

---

## Plans & Payments

This menu powers the self-service shop.

### Tabs

| Tab | Purpose |
|-----|---------|
| Plans | Create packages with traffic, duration, price |
| Orders | Review purchases and payment proofs |
| Payment Methods | Configure ZarinPal, card-to-card, crypto |
| Invoices | Export PDF or CSV records |

### Create a Shop Plan

1. Go to **Plans & Payments → Plans**.
2. Add name, traffic, duration, device limit, price.
3. Select allowed inbounds.
4. Enable the plan.

### Payment Flow

1. User opens `/sub/{token}/shop`.
2. Selects a plan.
3. Pays through configured method.
4. Order becomes pending/review/approved.
5. Approved order applies the plan automatically.

---

## Wallet Billing

Wallet Billing is mainly for resellers.

### What Resellers See

| Area | Description |
|------|-------------|
| Balance | Current traffic/user credits |
| Top-up Request | Ask sudo admin for more credit |
| Ledger | Full credit history |
| Quota Bars | Users, traffic, sessions, policy usage |
| Owned Plans | Plans created by this reseller |
| Owned Orders | Orders from this reseller's users |

### Sudo Workflow

1. Create reseller admin.
2. Set user quota, traffic quota, policy limits.
3. Give initial wallet credit.
4. Review top-up requests.
5. Audit reseller actions.

---

## Tickets

Tickets provide user support inside the portal.

### User Flow

1. User logs into portal.
2. Opens a ticket.
3. Selects category and priority.
4. Admin/reseller replies from panel.
5. Ticket is closed when resolved.

### Admin Actions

- Assign ticket.
- Change priority.
- Add internal note.
- Close/reopen.
- Filter by reseller or status.

---

## Audit

> Added in 1.3.0

Audit records every admin mutation.

### What Is Logged

- Admin login/logout
- User create/update/delete
- Node changes
- Inbound changes
- Settings changes
- Order approval/rejection
- Wallet credit changes
- ACME and federation events

### How to Use

1. Go to **Audit**.
2. Filter by actor, action, target, or date.
3. Open an event to view before/after diff.
4. Export when needed for compliance.

---

## Settings

Settings control global behavior.

### Tabs

| Tab | Use |
|-----|-----|
| General | Panel name, timezone, language, accent |
| Security | IP Guard, 2FA, session TTL |
| Appearance | Theme, logo, favicon |
| Branding | Reseller portal whitelabel |
| Notifications | Telegram, email, webhooks |
| Backup | Scheduled backups to Telegram/S3 |
| ACME | Let's Encrypt certificates via DNS-01 |
| API Tokens | Personal access tokens |
| Admins | Roles, resellers, policy limits |
| System Info | Version, diagnostics, health |

### Recommended First Settings

1. Set panel name and timezone.
2. Enable 2FA.
3. Configure IP whitelist.
4. Configure notifications.
5. Enable auto-backup.
6. Configure ACME for HTTPS.

---

## Complete First-Run Workflow

Use this sequence for a fresh production deployment.

### 1. Install

Choose one:

- Docker: see [Installation → Docker](02-installation.md#docker-compose)
- Native: see [Installation → Native Build](02-installation.md#native-build)

### 2. Secure the Panel

1. Login as sudo.
2. Change password.
3. Enable 2FA.
4. Set IP Guard.
5. Configure backups.

### 3. Add Infrastructure

1. Add local or remote node.
2. Create VLESS + REALITY inbound.
3. Run Reality Scanner.
4. Enable Decoy and Probing Protection.

### 4. Create Service Offering

1. Create plans.
2. Configure payment methods.
3. Enable portal branding.

### 5. Add Users

1. Create user manually or import CSV.
2. Assign inbounds.
3. Copy subscription link or QR.
4. Verify traffic in Monitor.

### 6. Operate Daily

1. Check dashboard.
2. Review node health.
3. Review tickets and orders.
4. Check audit log.
5. Watch backup status.

---

## Role-Based Usage

### Sudo Admin

- Owns the whole platform.
- Manages system settings, nodes, security, federation, backups.
- Creates resellers and approves top-ups.

### Admin

- Manages users, nodes, inbounds depending on permissions.
- Does not necessarily control global settings.

### Reseller

- Manages only owned users.
- Creates own plans and payment methods.
- Uses wallet credits.
- Can customize portal branding.

### End User

- Uses self-service portal.
- Views usage.
- Downloads subscription.
- Purchases/renews plans.
- Opens support tickets.

---

## Related Pages

- [Installation](02-installation.md)
- [First Steps](03-first-steps.md)
- [Users](05-user-management.md)
- [Nodes](06-node-management.md)
- [Security](08-security-administration.md)
- [Plans & Payments](09-plans-payments.md)
- [Settings & Backup](11-settings-backup.md)