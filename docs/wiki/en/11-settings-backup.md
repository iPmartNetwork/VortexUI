# Settings & Backup

Configure panel identity, branding, backups, and system behavior.

> **New in 1.3.0:** All panel settings are now **persisted in PostgreSQL** (previously browser localStorage). Settings survive browser changes and are shared across admin sessions.

---

## Settings Overview

**Sidebar → Settings**

| Tab | Contents |
|-----|----------|
| General | Name, timezone, language, accent color |
| Security | IP guard, 2FA, session TTL |
| Notifications | Telegram, email, webhooks |
| Appearance | Theme, logo, favicon |
| Branding | Whitelabel (reseller) |
| API | API keys, rate limits |
| Backup | Auto-backup schedule & destinations |
| ACME | TLS certificate management |
| Admins | List, roles, access matrix |
| System Info | Version, DB status, diagnostics |

---

## General Settings

| Field | Description |
|-------|-------------|
| Panel Name | Displayed in title and header |
| Timezone | For timestamps and scheduling |
| Default Language | For new admin accounts |
| Accent Color | UI theme accent |
| Date Format | How dates are displayed |

> **Validation (1.3.1):** Settings are validated on save. Invalid values show translated error messages via `getApiErrorMessage()`.

---

## Appearance

### Theme

- **Dark** — default glass design
- **Light** — bright glass variant
- **Auto** — follow system preference

Theme transitions are smoothly animated.

### Logo & Favicon

| Asset | Recommended |
|-------|-------------|
| Logo | SVG or PNG, 200x50px |
| Favicon | ICO or PNG, 32x32px |
| Login background | JPG, 1920x1080px |

---

## Branding (Whitelabel)

> **Portal whitelabel new in 1.3.0**

Resellers can brand their users' portal experience.

**Settings → Branding** (reseller view)

| Field | Description |
|-------|-------------|
| Brand Name | Shown in portal title |
| Logo | Portal logo |
| Accent Color | Portal theme color |
| Footer Text | Custom footer |
| Support Link | Contact URL |
| Custom CSS | Advanced styling |

### Per-Tenant Portals

Each reseller's users see a fully branded portal:
- Reseller's logo instead of VortexUI
- Reseller's colors
- Reseller's support contact
- No mention of the underlying panel

---

## Security Settings

### IP Guard

Restrict panel/API access by IP.

| Mode | Behavior |
|------|----------|
| Off | No restriction |
| Whitelist | Only listed IPs allowed |
| Blacklist | Listed IPs blocked |

- Supports CIDR ranges (e.g., `192.168.1.0/24`)
- **Runtime hot-reload (1.3.0):** Rules apply immediately on save, no restart

### Session Settings

| Setting | Default |
|---------|---------|
| Session TTL | 24h |
| Remember me | 30d |
| Force 2FA | Off |
| Max sessions | Unlimited |

### Two-Factor Authentication

See [Security](08-security-administration.md#enabling-2fa-totp) for TOTP setup.

---

## API Settings

### API Keys (PAT)

Create personal access tokens for automation.

| Field | Description |
|-------|-------------|
| Name | Token identifier |
| Scopes | Permissions (inherit from role) |
| Expiry | Optional expiration date |

### Rate Limiting

| Setting | Default |
|---------|---------|
| Requests per minute | 60 |
| Burst | 10 |
| Per-IP limit | Yes |

---

## Backup

> **Auto-backup new in 1.3.0**

### Manual Backup

```bash
vortexui backup
```

Or **Settings → Backup → Create Now**

Creates a full database export.

### Auto-Backup

**Settings → Backup → Schedule**

| Setting | Options |
|---------|---------|
| Schedule | Cron expression or preset (daily, weekly) |
| Time | e.g., 03:00 |
| Retention | Keep last N backups |

### Backup Destinations

| Destination | Configuration |
|-------------|---------------|
| Local | File path on server |
| Telegram | Bot sends backup file |
| S3 | Bucket, keys, endpoint |

#### S3 Configuration

```env
VORTEX_S3_ENDPOINT=https://s3.amazonaws.com
VORTEX_S3_BUCKET=vortex-backups
VORTEX_S3_ACCESS_KEY=AKIA...
VORTEX_S3_SECRET_KEY=secret...
```

#### Telegram Backup

```env
VORTEX_BACKUP_TELEGRAM=true
VORTEX_TELEGRAM_TOKEN=123456:ABC...
```

### What's Included

- All users and their configs
- All nodes and inbounds
- Settings and branding
- Plans and orders
- Wallet ledgers
- Audit log

### Restore

```bash
vortexui restore backup-2026-01-15.sql.gz
```

> **Warning:** Restore overwrites current data. Test on a staging instance first.

---

## ACME / TLS Certificates

> **Real ACME new in 1.3.0**

Automatic Let's Encrypt certificates via Cloudflare DNS-01 challenge.

**Settings → ACME**

### Setup

1. Create a Cloudflare API token with DNS edit permissions
2. Configure:

```env
VORTEX_ACME_EMAIL=admin@example.com
CLOUDFLARE_API_TOKEN=cf_token_here
```

3. Add domains to manage
4. Certificates issue automatically

### Features

| Feature | Description |
|---------|-------------|
| Auto-issue | Certificate on domain add |
| Auto-renew | Renews at T-30 days |
| Multi-domain | SAN certificates |
| Wildcard | `*.example.com` support |
| DNS-01 | No port 80 needed |

### Certificate Status

| Status | Meaning |
|--------|---------|
| 🟢 Valid | Certificate active |
| 🟡 Renewing | Renewal in progress |
| 🔴 Expired | Needs attention |
| ⚪ Pending | Being issued |

ACME events are logged in the [Audit Log](08-security-administration.md#audit-log).

---

## Admin Management

**Settings → Admins**

### Admin List

| Column | Description |
|--------|-------------|
| Username | Admin identifier |
| Role | Sudo, Admin, Reseller |
| Status | Active, Suspended |
| Users | Owned user count |
| Last Login | Recent activity |

### Adding an Admin

1. Click **Add Admin**
2. Set username, password, role
3. For resellers: set quotas, policy limits
4. Configure access matrix

### Access Matrix

Fine-grained permissions per admin:

| Permission | Options |
|------------|---------|
| users.* | read, write, delete |
| nodes.* | read, write |
| plans.* | read, manage |
| settings.* | read, write |

---

## System Info

**Settings → System Info**

| Item | Shows |
|------|-------|
| Version | Current VortexUI version |
| Uptime | Panel uptime |
| Database | PostgreSQL connection & size |
| Redis | Cache connection |
| Nodes | Total/online count |
| Users | Total/active count |
| Disk | Available space |

### Diagnostics

Run `vortexui doctor` or click **Run Diagnostics** to check:

- Database connectivity
- Redis connectivity
- Node reachability
- Port availability
- Certificate validity
- Disk space

---

## Updates

### Check for Updates

**Settings → System Info → Check Updates**

Shows current version vs latest release.

### Update

```bash
vortexui update
```

Or see [Installation → Updating](02-installation.md#updating).

---

## Settings Best Practices

### Initial Setup
1. Set panel name and timezone
2. Enable 2FA for sudo
3. Configure IP whitelist
4. Set up auto-backup
5. Configure ACME for HTTPS

### Ongoing
- Review audit log regularly
- Test backups periodically
- Rotate API tokens
- Monitor certificate expiry
