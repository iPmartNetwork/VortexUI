!!! info "Translation pending"
    This page is currently available in English only.

# Changelog

Version history and migration guides for VortexUI.

---

## v1.3.1 — Latest

**Focus: Stability & Polish**

### Fixed
- ⚙️ Settings validation with translated error messages
- 🔤 `getApiErrorMessage()` now maps HTTP codes to i18n keys
- 🔌 API client fixes for edge-case responses
- 🎛 Toggle switch state corrections (Decoy, IP Guard)
- 🎯 TLS Tricks icon now opens configuration modal; `isp` field persisted

### Improved
- Translated error messages across all 8 languages
- More robust settings save flow
- Better validation feedback in forms

---

## v1.3.0

**Focus: Persistence, Auditability & Automation**

### Added
- 🔐 **Persisted Panel Settings** — Configuration moved from browser localStorage to PostgreSQL. Settings now survive browser changes and are shared across admin sessions.
- 📜 **Audit Log** — Live table at `/audit` recording every admin action with actor, action, target, timestamp, and diff view. Includes federation and ACME events.
- 🎨 **Portal Whitelabel** — Per-tenant branding for the self-service portal (logo, colors, footer, custom CSS).
- 🎟 **Referral System** — End-users share invite links from the portal. Configurable rewards for referrer and referee.
- 🔒 **Real ACME** — Let's Encrypt certificates via Cloudflare DNS-01 challenge. Auto-issue, auto-renew at T-30 days, wildcard support.
- 🌐 **Federation** — Multi-panel peer coordination with periodic health checks and user/node count synchronization.
- 💾 **Auto-Backup** — Scheduled database exports to Telegram or S3-compatible storage with retention policies.

### Improved
- IP Guard rules now hot-reload on save (no restart)
- Settings API with full CRUD
- Background workers for backup, ACME, federation

### Migration Notes
See the [Migration 1.2.x → 1.3.x](#migration-12x-13x) section below.

---

## v1.2.9

**Focus: Command Tower UI & Reseller Platform**

### Added
- 🖥 **Command Tower UI** — Merged related pages, added fleet telemetry, geo pin map on Overview.
- 💼 **Reseller Platform** — Complete sub-panel system:
  - Wallet with balance and top-up queue
  - Orders with invoices and PDF export
  - Per-reseller plans and pricing
  - Sub-admin profiles at `/settings/admins/:id`
  - Quota bars, consumption stats, policy limits
  - Wallet ledger with full history
  - Login-as (impersonate), unsuspend actions
- 📡 **Dedicated Inbounds Page** — Moved from Nodes to `/inbounds`
- 🔀 **Merged Routing Page** — `/routing?tab=packs|balancers|outbounds`
- 🛡 **Merged Security Suite** — `/evasion?tab=reality|cleanip|tls|decoy`

### Added — Node Location
- Migration `0030_nodes_location.sql` adds `region`, `country_code`, `ping_ms`, `location_auto` to nodes
- Auto-detect node location
- Geo pins and ping heatmap on dashboard

---

## v1.2.8

**Focus: Veltrix UI Refresh**

### Added
- 🎨 **Veltrix UI** — Glass design system, collapsible sidebar shell
- ⌨️ **Command Palette** — Ctrl+K fuzzy search across everything
- 🌍 **Full i18n** — Complete 8-language translations (admin + portal)
- 🛒 **Self-Service Portal & Shop** — Users buy plans from reseller shops
- 💰 **Per-Reseller Plans & Payments** — Card-to-card, crypto, ZarinPal
- 🛡 **Anti-Censorship Suite** — TLS Tricks, probing protection, decoy, DoH, WARP+
- 🖥 **Intelligent Node Fleet** — Enrollment wizard, auto-migration, mTLS
- 📊 **Advanced Analytics** — Geo-IP, top users, peak hours, world map, CSV

---

## Version Comparison

| Feature | 1.2.8 | 1.2.9 | 1.3.0 | 1.3.1 |
|---------|-------|-------|-------|-------|
| Veltrix UI | ✅ | ✅ | ✅ | ✅ |
| Command Palette | ✅ | ✅ | ✅ | ✅ |
| Anti-Censorship Suite | ✅ | ✅ | ✅ | ✅ |
| Reseller Platform | — | ✅ | ✅ | ✅ |
| Command Tower UI | — | ✅ | ✅ | ✅ |
| Node Location | — | ✅ | ✅ | ✅ |
| Persisted Settings | — | — | ✅ | ✅ |
| Audit Log | — | — | ✅ | ✅ |
| Portal Whitelabel | — | — | ✅ | ✅ |
| Referral System | — | — | ✅ | ✅ |
| Real ACME | — | — | ✅ | ✅ |
| Federation | — | — | ✅ | ✅ |
| Auto-Backup | — | — | ✅ | ✅ |
| Settings Validation | — | — | — | ✅ |
| Translated Errors | — | — | — | ✅ |

---

## Migration Guides

### Migration 1.2.x → 1.3.x {#migration-12x-13x}

> **Estimated time:** 10-15 minutes
> **Downtime:** ~2 minutes (during migration)

#### Before You Start

1. **Backup your database:**
   ```bash
   vortexui backup
   ```

2. **Note your current version:**
   ```bash
   vortexui version
   ```

#### Upgrade Steps

**Docker:**
```bash
cd VortexUI
git pull
docker compose pull
docker compose up -d
```

**Native:**
```bash
cd /opt/VortexUI
git pull origin master
go build -o vortexui ./cmd/panel
./vortexui migrate
sudo systemctl restart vortexui
```

#### Post-Upgrade

1. **Settings migration:** On first login after 1.3.0, your browser localStorage settings are automatically migrated to the database. Verify at **Settings → General**.

2. **Configure ACME (optional):** To use real certificates:
   ```env
   VORTEX_ACME_EMAIL=admin@example.com
   CLOUDFLARE_API_TOKEN=cf_token
   ```

3. **Enable Auto-Backup:** **Settings → Backup → Schedule**

4. **Review Audit Log:** Visit `/audit` to see the new logging.

#### Breaking Changes

| Change | Impact | Action |
|--------|--------|--------|
| Settings in DB | localStorage settings deprecated | Auto-migrated on first login |
| New env vars | ACME/backup optional | Add if using those features |

#### Rollback

If needed:
```bash
# Restore backup
vortexui restore backup-before-upgrade.sql.gz

# Pin previous version
git checkout v1.2.9
go build -o vortexui ./cmd/panel
sudo systemctl restart vortexui
```

---

### Migration 1.2.8 → 1.2.9

#### Database Migration

The `0030_nodes_location.sql` migration runs automatically and adds location fields to nodes. No manual action needed.

#### UI Changes

Pages have been reorganized:

| Old Location | New Location |
|--------------|--------------|
| Nodes → Inbounds tab | `/inbounds` (dedicated page) |
| Separate routing pages | `/routing?tab=...` |
| Separate security pages | `/evasion?tab=...` |

Bookmarks may need updating.

---

## Upgrade Best Practices

1. ✅ Always backup before upgrading
2. ✅ Read the changelog for breaking changes
3. ✅ Test on staging first (for production)
4. ✅ Upgrade during low-traffic windows
5. ✅ Verify after upgrade with `vortexui doctor`
6. ✅ Keep the previous version available for rollback

---

## Release Schedule

VortexUI follows semantic versioning:

- **Major (x.0.0)** — Breaking changes
- **Minor (1.x.0)** — New features, backward-compatible
- **Patch (1.3.x)** — Bug fixes, polish

Releases are announced on:
- [GitHub Releases](https://github.com/iPmartNetwork/VortexUI/releases)
- [Telegram Channel](https://t.me/vortex_ui)

---

## Full Changelog

For the complete commit-level changelog, see [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) on GitHub.
