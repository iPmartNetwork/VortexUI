# Settings & Backup

---

## Panel Settings

**Settings** page contains general panel configuration:

| Setting | Description |
|---------|-------------|
| Panel title | Displayed in browser tab and header |
| Panel URL | Public URL (used for subscription links) |
| Language | Default UI language |
| Timezone | Panel timezone for display |
| JWT TTL | Token expiration time |
| Subscription update interval | How often clients should refresh |

---

## Custom Branding

**Settings → Branding**

Customize the panel's appearance:

| Field | Description |
|-------|-------------|
| Title | Panel title (header + browser tab) |
| Logo | Upload custom logo (SVG/PNG) |
| Favicon | Custom browser favicon |
| Accent color | Primary theme color (`#hex`) |
| Footer text | Custom footer line |
| Login background | Custom login page background |

---

## Reseller Whitelabel

Each reseller can brand their own portal independently:

**Reseller Account → Branding**

| Field | Description |
|-------|-------------|
| Panel title | Reseller's panel/portal title |
| Logo URL | Reseller's brand image |
| Accent color | `#hex` theme accent for their users |
| Portal slug | Optional custom URL slug |
| Footer text | Portal footer line |

Users accessing the portal see their reseller's branding, not the main panel's.

---

## Backup & Restore

### Manual Backup

**Settings → Backup → Create Backup**

Creates a full backup including:

- Database dump (users, nodes, inbounds, plans, orders, settings)
- Configuration files
- TLS certificates
- Uploaded assets (logos, payment proofs)

### Automatic Backup

| Setting | Description |
|---------|-------------|
| Schedule | Cron expression (e.g. `0 3 * * *` = daily at 3 AM) |
| Retain | Number of backups to keep |
| Telegram | Send backup file to admin chat |
| S3 | Upload to S3-compatible storage |

### Backup Destinations

=== "Telegram"

    Backups are sent as a file to the configured admin chat ID.
    Max file size: 50 MB (Telegram limit). Larger backups are split.

=== "S3"

    Configure S3-compatible storage:
    ```
    VORTEX_BACKUP_S3_BUCKET=my-backups
    VORTEX_BACKUP_S3_ENDPOINT=s3.amazonaws.com
    VORTEX_BACKUP_S3_ACCESS_KEY=xxxx
    VORTEX_BACKUP_S3_SECRET_KEY=xxxx
    VORTEX_BACKUP_S3_REGION=us-east-1
    ```

=== "Local"

    Backups stored in `/var/lib/vortexui/backups/` (default).
    Configure retention to prevent disk fill.

### Restore

```bash
vortexui restore /path/to/backup.tar.gz
```

Or via UI: **Settings → Backup → Restore → Upload backup file**.

!!! warning
    Restore overwrites the current database. Create a fresh backup before restoring.

---

## Deep Link Configuration

**Settings → Deep Links**

Configure subscription deep links for native app integration:

| Setting | Description |
|---------|-------------|
| Base URL | Panel's public URL |
| App scheme | URL scheme (e.g. `vortex://`, `clash://`) |
| Include server name | Add server identifier to links |
| QR logo | Custom logo rendered in QR code center |

---

## Update Checker

**Settings → Updates**

- Check for new VortexUI releases
- View changelog for available updates
- One-click update (calls `vortexui update` internally)
- Configure auto-check interval

---

## Config Templates

**Settings → Config Templates**

Customize the Clash/sing-box configuration output for subscriptions:

### Clash Template

Override default Clash YAML structure:

- Custom proxy groups (url-test, fallback, load-balance)
- Custom rules (block ads, direct local traffic)
- DNS configuration
- Custom routing rules

### sing-box Template

Override default sing-box JSON structure:

- Custom outbound groups
- Route rules
- DNS settings
- Experimental features

Templates support variables:

| Variable | Expands to |
|----------|-----------|
| `{PROXIES}` | List of proxy configs from user's inbounds |
| `{USERNAME}` | The user's username |
| `{SUB_URL}` | The subscription URL |

---

## Internationalization

The panel supports **8 languages** with full RTL support:

- English (EN)
- فارسی (FA)
- Türkçe (TR)
- العربية (AR)
- Русский (RU)
- 中文 (ZH)
- 日本語 (JA)
- Español (ES)

Switch language from the header dropdown. Each admin's preference is saved independently.
