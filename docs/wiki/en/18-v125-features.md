# 18. VortexUI v1.2.5 — Reseller Platform Guide

!!! info "Version 1.2.5"
    This page documents the reseller platform introduced in VortexUI v1.2.5 — scoped
    allowlists, quota enforcement, wallet hierarchy, whitelabel branding, automation
    webhooks, policy limits, auto-suspend, and full UI translations.

---

## Overview

VortexUI v1.2.5 turns the basic reseller role into a **complete sub-panel**:

| Audience | What they get |
|----------|---------------|
| **Sudo admin** | Quota usage table, bulk adjust, impersonate, wallet top-up, policy tuning |
| **Reseller** | Dashboard, owned users, wallet, sub-resellers, branding, outbound webhook |
| **End user** | Unchanged — still uses subscription links and optional portal |

**Sidebar → Reseller:** Reseller dashboard · Reseller account · Reseller quota alerts · My quota (redirects to dashboard).

---

## Allowlists (plans, nodes, inbounds)

**Location:** Admins → Edit admin (non-sudo)

When editing a reseller admin, assign:

| Picker | Effect |
|--------|--------|
| **Plans** | Only listed plans appear in the portal and can be sold |
| **Nodes** | Reseller may only bind users to these nodes |
| **Inbounds** | Reseller may only attach users to these inbounds |

Empty allowlist = no restriction (all items visible). Sudo admins are never scoped.

---

## Traffic quota modes

**Location:** Admins → Edit admin → Traffic quota mode

| Mode | Counts against pool when… |
|------|---------------------------|
| **Allocated** | You assign data limits to users (sum of limits) |
| **Consumed** | Users actually use traffic |

Set **User quota** and **Traffic quota (GB)** to `0` for unlimited. The reseller dashboard and Overview cards show remaining pool.

---

## Reseller dashboard

**Location:** Sidebar → Reseller dashboard (`/reseller-dashboard`)

| Widget | Content |
|--------|---------|
| Summary cards | Account count, new users (7d/30d), expiring (7d), quota mode |
| Traffic | Assigned vs consumed bytes |
| Users by status | Active / limited / expired breakdown |
| Top consumers | Highest-traffic owned users |
| **Export CSV** | Download owned-users report (`GET /api/account/export/users`) |

Sudo admins see a hint to use the main Overview for fleet-wide stats.

---

## Quota alerts (sudo)

**Location:** Sidebar → Reseller quota alerts (`/reseller-quota-alerts`)

Configure threshold notifications when **any** reseller approaches limits:

- Enable/disable globally
- Telegram (panel bot)
- Optional webhook URL
- Threshold percentages (e.g. `80, 90, 100`)
- Cooldown minutes
- Recent alerts table

---

## Reseller account

**Location:** Sidebar → Reseller account (`/reseller-account`)

### Wallet

- **Traffic credits** and **user credits** balance
- Ledger table: time, traffic delta, user delta, reason
- Top-up: sudo uses **Admins** (wallet credit API)

### Sub-resellers

- List child resellers with user/traffic quotas
- **Create sub-reseller:** username, password, role, quotas
- Child inherits parent's scope; parent can manage wallet

### Portal branding (whitelabel)

| Field | Purpose |
|-------|---------|
| Panel title | Shown in browser tab / header |
| Logo URL | Brand image |
| Accent color | `#hex` theme accent |
| Portal slug | Optional URL slug |
| Footer text | Portal footer line |

### Outbound webhook

- URL + signing secret (HMAC-SHA256)
- Events: `user.created`, `user.deleted` for owned accounts
- Enable/disable toggle; rotate secret without clearing URL

---

## Sudo admin tools

**Location:** Admins (`/admins`)

| Action | Description |
|--------|-------------|
| **Quota usage table** | Per-reseller accounts, assigned/consumed traffic, pool left |
| **+50 accounts / +10 GB / +50 GB** | Quick quota adjust (`POST /api/admins/:id/quota-adjust`) |
| **Login as** | Impersonate reseller (new JWT, refresh session) |
| **Unsuspend** | Clear auto-suspension flag |
| **Edit** | Role, quotas, allowlists, policy, auto-suspend |

### Scoped audit log

**Location:** Sidebar → Audit

- Sudo: all entries
- Reseller: only actions performed by that admin

### Policy limits (per reseller)

Set on **Edit admin**:

| Setting | Effect |
|---------|--------|
| Max data limit (GB) | Cap per-user data limit resellers may set (`0` = ∞) |
| Max expire (days) | Cap subscription length resellers may set |
| Allow bulk create/import | Gate bulk user creation and CSV import |
| Allow bulk delete | Gate multi-delete |

### Auto-suspend

| Setting | Effect |
|---------|--------|
| Enable auto-suspend | Master toggle |
| IP violations (7d) | Suspend after N sharing-detection events (`0` = off) |
| Quota grace (minutes) | Grace after quota exceeded before suspend |

A background worker evaluates resellers periodically. Suspended resellers cannot log in until sudo **Unsuspend**.

---

## Overview integration

**Location:** Dashboard → Overview

Non-sudo resellers see a **Your pool** card with account count, assigned traffic, consumed traffic, and a link to the reseller dashboard — without hunting through the sidebar.

---

## Internationalization

All reseller pages support **8 languages**: English, Persian (FA), Turkish, Arabic (RTL), Russian, Chinese, Japanese, Spanish.

Switch language from the panel header. Keys live under `reseller.*` in `web/src/i18n/dict.ts`.

---

## API summary

| Endpoint | Method | Who | Purpose |
|----------|--------|-----|---------|
| `/api/account/dashboard` | GET | Reseller | Dashboard stats |
| `/api/account/export/users` | GET | Reseller | CSV export |
| `/api/account/wallet` | GET | Reseller | Wallet + ledger |
| `/api/account/branding` | GET/PUT | Reseller | Whitelabel |
| `/api/account/webhook` | GET/PUT | Reseller | Outbound webhook |
| `/api/account/sub-admins` | GET/POST | Reseller | Sub-resellers |
| `/api/admins/:id/quota-adjust` | POST | Sudo | Bulk quota delta |
| `/api/admins/:id/unsuspend` | POST | Sudo | Clear suspension |
| `/api/admins/:id/impersonate` | POST | Sudo | Issue reseller token |
| `/api/admin-quota-notify/config` | GET/PUT | Sudo | Quota alert config |
| `/api/admin-quota-notify/events` | GET | Sudo | Recent alerts |

See [API reference](12-api-reference.md) and [`docs/openapi.yaml`](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) for request bodies.

---

## Upgrade notes

1. Pull v1.2.5 and rebuild/restart the panel.
2. Migrations run automatically on startup:
   - `0021_reseller_enhancements.sql`
   - `0022_reseller_advanced.sql`
   - `0023_reseller_policy_suspend.sql`
3. Create or edit **Roles** with reseller permissions.
4. Assign allowlists and quotas on **Admins → Edit**.
5. Optional: configure **Reseller quota alerts** and test **Login as** before handing credentials to partners.

!!! tip
    Start with **allocated** quota mode if resellers pre-sell fixed data packages; use
    **consumed** when you bill by actual usage.
