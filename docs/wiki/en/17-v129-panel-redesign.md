# Command Tower UI (v1.2.9)

<div style="text-align: center; margin: 1.5rem 0;">
  <strong style="font-size: 1.25rem;">VortexUI v1.2.9</strong><br/>
  <em style="font-size: 1rem;">Merged admin pages, Settings hub, reseller profiles, and fleet telemetry</em>
</div>

!!! info "Backend + frontend release"
    Run migration **`0030_node_location_ping.sql`** before deploying the panel binary.
    Rebuild the web UI (`npm run build`) and restart the panel service.

---

## Highlights

<div class="grid cards" markdown>

- :material-view-dashboard: **Command Tower Overview**

    Live stat widgets with traffic range tabs, top users (with protocol badges), and node geo/ping telemetry.

- :material-cog: **Settings hub**

    One page for General, Security, Notifications, Appearance, API keys, Backup, and Admins — sidebar tabs on the left, panel on the right.

- :material-shield-account: **Reseller profiles**

    Click any reseller username → wallet balance, quota bars, consumption, policy limits, panel-settings access, and wallet ledger.

</div>

---

## Merged pages

Older standalone routes redirect to tabbed shells:

| Old route | New location |
|-----------|--------------|
| `/routing-packs`, `/balancers`, `/outbounds` | `/routing?tab=packs\|balancers\|outbounds` |
| `/reality-scanner`, `/clean-ip`, `/tls-tricks`, `/decoy-website` | `/evasion?tab=reality\|cleanip\|tls\|decoy` |
| `/plans`, `/pending-orders`, `/wallet-billing` (legacy) | `/wallet-billing?tab=plans\|orders\|wallet` |
| `/admins` | `/settings?tab=admins` |
| `/audit` | `/overview` |

!!! tip "Command palette"
    Press **Ctrl+K** / **⌘K** — search targets were updated for the consolidated routes.

---

## Settings → Admins

**Settings → Admins** (sudo only) has three sub-tabs:

| Tab | URL param | Purpose |
|-----|-----------|---------|
| **Admins** | `section=list` (default) | Quota usage table + admin list; usernames link to profile |
| **Roles** | `section=roles` | Create/edit permission bundles |
| **Reseller access** | `section=access` | Permission reference + matrix of enabled panel settings per reseller |

### Reseller profile

Route: **`/settings/admins/{uuid}`**

| Section | Content |
|---------|---------|
| Stats | Accounts, traffic used, pool remaining, wallet traffic/credits |
| Quota bars | User accounts + traffic (allocated or consumed mode) |
| Account info | Role, created, last login, quotas |
| Policies | Max data limit, expire days, bulk actions, auto-suspend |
| Panel settings | Which Settings tabs the reseller may use |
| Wallet ledger | Top-up history with traffic/user deltas |
| Actions | Edit, top-up wallet, login-as, unsuspend |

---

## Inbounds & fleet

- **`/inbounds`** — dedicated inbound/sub-host management (separated from Nodes fleet table).
- **`/nodes`** — fleet health, enrollment, and node edit including **region**, **country code**, and **ping ms** (migration 0030).

---

## New API endpoints

| Method | Path | Access |
|--------|------|--------|
| `GET` | `/api/admins/:id/quota` | Sudo or the admin themselves |
| `GET` | `/api/admins/:id/wallet` | Sudo or the admin themselves |

Existing `GET /api/admins/usage` (all resellers) and `POST /api/admins/:id/wallet` (top-up) are unchanged.

---

## App shell changes

- Fixed viewport layout: sidebar stays full height; only the main content scrolls.
- Compact sidebar (~236px) with **VORTEXUI** title, version badge, **Self-Service Portal** button, and **Core status** card (Xray version, online/offline).
- Search remains in the header (removed from sidebar).

---

## Upgrade

=== "Standard deploy"

    ```bash
    vortexui migrate
    go build -o vortexui-panel ./cmd/panel
    cd web && npm run build
    # copy web/dist to your static path, then:
    systemctl restart vortexui-panel
    ```

=== "Docker Compose"

    ```bash
    docker compose pull
    docker compose up -d
    ```

---

## Related docs

| Doc | Link |
|-----|------|
| Veltrix UI (v1.2.8) | [16-v128-ui-refresh.md](16-v128-ui-refresh.md) |
| Security & admins | [08-security-administration.md](08-security-administration.md) |
| Settings & backup | [11-settings-backup.md](11-settings-backup.md) |
| Changelog | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
| Docs template | [review/DOC-TEMPLATE.md](https://github.com/iPmartNetwork/VortexUI/blob/master/review/DOC-TEMPLATE.md) |
