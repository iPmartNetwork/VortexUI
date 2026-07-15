# Changelog

All notable changes to VortexUI are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.3.5] - 2026-07-15

Systemd auto-start reliability fix and graceful backup restore for legacy (pre-credentials) exports.

### Fixed
- **Systemd auto-start** — panel and node services now declare `Wants=network-online.target` and set `StartLimitIntervalSec=0` so the service always restarts after a server reboot or update, even when the PostgreSQL/Redis Docker containers take longer than usual to become ready.
- **`vortexui update`** — calls `systemctl daemon-reload` and re-enables the unit before restarting, so any service-file changes from the Git pull take effect immediately.
- **Backup restore (legacy)** — restoring a JSON backup exported before v1.3.4 (which lacked admin credentials) no longer fails the entire restore. A temporary random password is generated for each affected admin, printed to the log with the `reset immediately` hint, and the restore completes. Use `vortexui admin reset-password` to set a permanent password afterwards.

## [1.3.4] - 2026-07-15

Comprehensive backup & restore for server migration, billing/wallet history, and reseller-scoped user backups.

### Added
- **Backup v3 JSON** — supplemental tables (wallet ledger/deposits/packages, orders, billing/panel settings, reseller payment config, sub hosts, …), admin login secrets, and usage manifest (used/remaining traffic).
- **Full database backup** — `pg_dump` archive (`.tar.gz`) via `GET /api/backup?format=full` for byte-for-byte server migration.
- **Restore modes** — JSON config restore (transactional) and full DB restore (`POST /api/backup/restore?mode=full` with multipart file).
- **Manifest API** — `GET /api/backup/manifest` previews counts, traffic usage, and warnings before export/restore.
- **Optional AES-256-GCM encryption** — passphrase via `X-Backup-Passphrase` header on export/import.
- **Reseller backup v2** — export includes users, bindings, wallet ledger, orders, payment config, portal branding; **restore** via `POST /api/account/backup/users/restore`.
- **Settings UI** — backup preview panel, JSON vs full DB export, restore mode selector, encryption passphrase, traffic time-series toggle.

### Fixed
- Config restore now wipes orphan rows (traffic_points, sub_hosts, orders, …) and preserves admin passwords when credentials are included in the backup.
- **Admin login after backup restore** — restore no longer overwrites `password_hash` with empty values when backup lacks `admin_credentials`; added `panel admin reset-password` for CLI recovery.

**Release by ali**

## [1.3.3] - 2026-07-15

Multi-core Phase 1: run **Xray** and **sing-box** on the same node for A/B testing, with per-inbound engine selection.

### Added
- **Dual-core nodes** — `enabled_cores` on nodes; `CompositeDriver` on the agent runs both engines with split config sync.
- **Per-inbound core override** — `inbounds.core` routes each inbound to xray or sing-box; empty inherits the node default.
- **Migration `0042_multicore.sql`** — persists `nodes.enabled_cores` and `inbounds.core`.
- **gRPC** — `CORE_TYPE_MULTI` and `InboundSpec.core` for multi-engine sync.
- **Node agent** — `VORTEX_ENABLED_CORES=xray,singbox` with per-engine config/API env vars (`VORTEX_XRAY_CONFIG`, `VORTEX_SINGBOX_CONFIG`, …).
- **Panel UI** — enabled-core toggles on create/edit node; engine picker on inbound form when the node is dual-core; multi-core badges on the fleet view.

### Changed
- Hub resolves each inbound's effective core before sync and sends `multi` when the node runs both engines.
- Inbound validation checks the effective core against the node's enabled engines and capability matrix.

**Release by ali**

## [1.3.2] - 2026-07-11

PHASE 3 Frontend release: Performance monitoring, Security hardening, Compliance dashboard, and Overview UI enhancements.

### Added
- **Performance Monitoring Dashboard** — `/performance` with real-time cache stats, query metrics, slow query tracking, and performance alerts.
- **Security Hardening Dashboard** — `/security` with security threat monitoring, compliance status, active security policies, and threat statistics.
- **Compliance Dashboard** — `/compliance` with system compliance checklist, compliance status tracking, security policy enforcement, and recommendations.
- **Enhanced Inbounds Management** — `/inbounds` improved with search/filter, statistics cards (total/active/protocols), protocol distribution visualization, and expanded details with copy-to-clipboard.
- **Monitoring & Compliance navigation** — new section in sidebar with Performance, Security, and Compliance pages; i18n support for 8 languages.
- **Performance health hooks** — React Query integration with `/api/performance/health`, `/api/performance/queries/slow`, `/api/performance/queries/stats` endpoints.
- **Security threat hooks** — endpoints for threat monitoring, blocked threats, security scores, compliance validation, and IP reputation.

### Changed
- **Overview page** — professional UI enhancements with gradient backgrounds, better charts, improved node fleet telemetry cards, and user pool display.
- **Traffic Series Chart** — gradient styling, enhanced header, and summary stats cards (Upload/Download/Peak metrics).
- **Protocol Donut Chart** — better visual hierarchy and professional appearance.
- **Node Fleet cards** — gradient backgrounds, icon badges, enhanced load bars with color coding (green/amber/red), CPU/RAM breakdown display.
- **User Pool cards** — gradient avatars, usage progress bars with conditional coloring, usage percentage indicators, smooth hover animations.
- **Frontend routing** — lazy-loaded PHASE 3 pages with `/performance`, `/security`, `/compliance` routes.
- **API hooks structure** — security-hooks.ts with 15+ React Query hooks for Performance Optimization (3B) and Security Hardening (3D) endpoints.

### Fixed
- **TypeScript compilation** — resolved 18+ type mismatches in new pages matching actual API response types from security-hooks exports.
- **InboundsEnhanced** — proper type handling for InboundFleetRow properties with enabled status tracking.
- **Performance page** — corrected property names: `cache_hit_rate`→`cache.hit_rate`, `active_connections`→`connections_active`, `slow_queries_1h`→`slow_queries_last_hour`.
- **SecurityHardening page** — resolved undefined icon and component references; integrated useComplianceValidation hook.
- **Compliance page** — type conversion corrections and integration with useComplianceValidation hook.

### Technical Details
- **PHASE 3 backend** — fully compiled with exit code 0 (Authentication, Performance Optimization, Audit & Compliance, Security Hardening).
- **Frontend build** — 2157 modules transformed, production bundle generated without errors.
- **Lines added** — ~1,170 lines of React/TypeScript across 4 new pages + routing/navigation integration.
- **Deployment** — master branch, all TypeScript checks pass, npm run build successful.

**Release by ali**

## [1.3.1] - 2026-07-08

Patch release: persisted settings polish, portal dashboard v2, live connection stats, and UI fixes.

### Added
- **Portal dashboard v2** — subscription QR/copy, live connections & devices, quota alerts, quick actions (renew/tickets), 30-day daily usage chart, full-width portal layout aligned with admin panel.
- **Portal API endpoints** — `GET /api/portal/subscription`, `/usage`, `/online`, `/deeplink` for authenticated end-users.
- **`getApiErrorMessage()`** — maps HTTP status codes to i18n keys across Login and Settings.
- **Error i18n keys** — `errors.*` and `settings.saveFailed` in all 8 languages.
- **Frontend tests** — vitest coverage for API client body parsing and form error helpers.
- **Backend tests** — panel settings handler and normalization unit tests.

### Changed
- **Portal layout** — full-width content area; subscription and usage cards share equal height; proportional usage chart.
- Settings tabs (General, Security, Notifications, Appearance, Backup) show translated save errors.
- **Backup tab** — auto-backup save now catches API errors and shows an inline banner.
- **Lazy routes** — loading fallback uses `common.loading` i18n string.
- **Panel settings service** — normalizes trim/caps on update; merges partial payloads on get.
- **Node health poll** — `OnlineStats` queried on every health tick for live connection counts on Xray nodes.

### Fixed
- **API client** — read response body once (`text()` then parse) to avoid double-consumption bugs.
- **TLS Tricks (ISP)** — settings icon opens configure modal; `isp` column persisted so preset profiles round-trip.
- **Toggle switches** — off/on thumb position corrected (left/right) in Settings, Routing packs, and Decoy tab.
- **Live connections** — report active device IPs and connection counts; fallback from recent traffic when Xray online stats empty.
- **Startup resync** — all nodes re-provisioned on panel boot; brief unhealthy during restart triggers failover instead of stuck state.
- **Reality SNI** — allow changing applied SNI and sync `raw.reality`.
- **Card-to-card payments** — show card number in portal checkout.
- **Inbounds modal** — hooks run before early return in `NodeInboundsModal`.
- **CI Docker builds** — per-image GHA cache scopes; ignore transient cache export errors.

## [1.3.0] - 2026-07-06

Backend–frontend integration release: persisted panel settings, audit log UI, portal
whitelabel/referral, real ACME (Cloudflare DNS-01), and federation sync worker.

### Added
- **Panel settings API** — `GET/PUT /api/settings` persists general, security, appearance,
  notifications, and auto-backup options in PostgreSQL (replaces browser localStorage).
- **IP Guard runtime** — whitelist/blacklist from settings applied via middleware on save and startup.
- **Auto-backup worker** — scheduled export to Telegram/S3 driven by panel settings + `VORTEX_TELEGRAM_TOKEN`.
- **Audit Log page** — `/audit` restored with live table from `GET /api/audit`.
- **Portal referral UI** — `/portal/referral` for end-user codes and apply flow.
- **Portal branding** — portal shell loads title/logo from `GET /api/portal/branding?slug=`.
- **ACME Let's Encrypt** — DNS-01 via Cloudflare when `VORTEX_ACME_EMAIL` + CF credentials are set.
- **Federation sync worker** — periodic peer health checks and user/node count sync events.

### Changed
- Settings tabs (General, Security, Appearance, Notifications, Backup) read/write via API.
- Portal Plans page fully i18n'd (8 languages).
- Panel accent color applied from server settings after login.
- Phase 4 i18n: Family Groups, Smart Quota, Deep Links, Quota Notifications, Referrals, User Detail, and Command Palette (8 languages).
- Frontend unit tests (vitest) for `mergePanelSettings` and `hexToHslComponents`.

## [1.2.9] - 2026-07-04

Command Tower UI rollout: merged admin pages, Settings hub with reseller management,
reseller profile detail, and fleet telemetry on the dashboard.

### Added
- **Merged admin pages** — Routing & Balancers (`/routing?tab=packs|balancers|outbounds`),
  Security Suite (`/evasion?tab=reality|cleanip|tls|decoy`), Reseller Platform
  (`/wallet-billing?tab=orders|plans|wallet`).
- **Settings hub** — sidebar tab navigation (General, Security, Notifications,
  Appearance, API, Backup, Admins); Admins moved from standalone route.
- **Admins sub-tabs** — Admins list, Roles, and Reseller access matrix under
  `/settings?tab=admins&section=list|roles|access`.
- **Reseller profile page** — `/settings/admins/:id` with wallet balance, quota bars,
  consumption stats, policy limits, panel-settings access, wallet ledger, and actions
  (edit, top-up, login-as, unsuspend).
- **Dedicated Inbounds page** — `/inbounds` separated from Nodes fleet view.
- **Overview widgets (API)** — richer command-tower stats: traffic range tabs, top users
  with protocol labels, node geo/ping telemetry.
- **Node location fields** — migration `0030`: `region`, `country_code`, `ping_ms`,
  `location_auto` on nodes; editable in node UI.
- **Admin detail APIs** — `GET /api/admins/:id/quota` and `GET /api/admins/:id/wallet`
  (sudo or self).
- **Docs template** — `review/DOC-TEMPLATE.md` design reference for GitHub Pages wiki.

### Changed
- **App shell** — fixed `h-screen` layout; sticky sidebar; scrollable main content only;
  compact sidebar (236px) with VORTEXUI header, version badge, core status card, and
  Self-Service Portal shortcut.
- **Redesigned pages** — Overview, Users, Nodes, Inbounds, Tickets, Settings; tabbed
  shells match Arena/Veltrix mock layouts with live API data.
- **Route consolidation** — `/outbounds` → routing tab; `/admins` → settings tab;
  `/audit` → overview; legacy security/reseller routes redirect to merged pages.
- **Command palette** — updated targets for consolidated routes.
- **Wallet billing UI** — absorbed into Reseller Platform wallet tab (removed standalone
  `/wallet-billing` page component in favour of tab shell).

### Fixed
- Reseller names in admin/quota tables link to the new profile page for quick drill-down.

## [1.2.8] - 2026-07-04

Major frontend release: **Veltrix UI** design system across the admin panel and user
portal, with complete translations for all eight supported languages.

### Added
- **Veltrix design system** — glass-style `GlassCard`, `StatsCard`, `StatusBadge`, and
  `PageShell` components with cyan/sky accent palette and framer-motion transitions.
- **New app shell** — collapsible `AppSidebar` + sticky `AppHeader` with mini mode,
  mobile drawer, theme toggle, and language switcher (8 languages).
- **Command palette** — fuzzy navigation search via **Ctrl+K** / **⌘K** from the
  header, sidebar, and keyboard shortcut.
- **Redesigned core pages** — Overview, Users, and Nodes with live API stat cards,
  fleet health badges, and traffic leaderboards.
- **Portal UI refresh** — redesigned login, dashboard, desktop sidebar, and mobile
  bottom navigation; all portal strings wired to i18n.
- **Complete i18n (639 keys × 8 languages)** — billing, reseller payment settings,
  pending orders, shell, overview, users, nodes, and portal keys for EN/FA/TR/AR/RU/ZH/JA/ES.
- **Locale tooling** — `web/src/i18n/locale/*.json` supplements plus
  `apply-i18n-locales.mjs` and `check-i18n.mjs` scripts for translation maintenance.

### Changed
- **Admin panel rollout** — shared `Card`, `PageHeader`, `StatCard`, `Badge`, `Button`,
  and `DataTable` restyled; page-enter animations applied across all admin routes.
- **Login page** — dual-tab admin login and subscription-token portal shortcut with
  language picker and secure-connection indicator.
- **Theme tokens** — updated Tailwind palette and CSS variables for Veltrix cyan/sky
  look in dark and light modes.

### Fixed
- Hardcoded FA/EN UI strings replaced with `t()` so language switching works on
  Overview, Users, Nodes, Login, shell, and portal pages.

## [1.2.7] - 2026-06-29

Major reseller commerce release: per-reseller payment configuration, per-reseller
owned plans, self-service renewal, and payment proof/receipt uploads.

### Added
- **Self-service plan renewal** — end-users can purchase plans from their subscription
  page (`/sub/:token/shop`) to extend existing accounts. Traffic and duration stack
  additively (no reset, no new account). Payment result page at `/payment/result`.
- **Per-reseller payment configuration** — each reseller configures their own card
  number, crypto wallet addresses, and ZarinPal merchant ID from the panel. Sudo can
  also configure payment settings (since sudo sells to direct users). Gated by a new
  "billing" reseller setting that sudo grants per-reseller.
- **Per-reseller owned plans** — resellers create their own plans with their own
  pricing. Plans are scoped to their creator; a reseller's users only see that
  reseller's plans in the shop. Sudo's users see sudo's plans. New `admin_id` column
  on the `plans` table with migration `0028`.
- **Payment proof/receipt upload** — card-to-card purchases require a receipt image
  (فیش واریز); crypto purchases accept TX ID + optional transfer screenshot. Proof
  stored as base64 data URL on the order. Admins see clickable thumbnails in the
  Pending Orders review page. New `proof_image` column on `orders` (migration `0029`).
- **Manual payment methods in shop** — card-to-card and crypto options alongside
  ZarinPal in the server-side shop page and React portal, with per-method forms:
  file upload for card-to-card, TX hash + screenshot for crypto, auto-redirect for
  ZarinPal.
- **Pending Orders review page** — admins/resellers approve or reject manual payment
  orders (card-to-card, crypto) with proof image display.
- **Payment settings nav visibility** — "Payment Settings" checkbox in Edit Admin
  modal's Settings sections; payment nav items visible to sudo in the Users section
  and to resellers when billing is enabled.
- **Plan permission changes** — resellers (with `user:write`) can now create and
  delete their own plans (ownership enforced); previously required `admin:manage`.

### Changed
- `POST /api/plans` and `DELETE /api/plans/:id` permissions changed from
  `admin:manage` to `user:write` (ownership enforced in handlers).
- `PublicPlans` endpoint now returns only enabled plans owned by sudo admins.
- `SubscriptionShop` now filters plans by the user's owning admin.
- `InitPurchase` validates plan ownership against the user's admin (cross-reseller
  purchase blocked).
- Card-to-card in `InitPurchase` no longer requires `tx_id` if `proof_image` is
  provided.

### Fixed
- Portal "Purchase" button now properly initiates payment (was GET to POST endpoint).
- Subscription info "View Plans & Purchase" link now opens the shop page.
- Edit Admin modal scrollable with sticky save button (no longer requires shrinking
  the browser window).
- Plan form labels now clearly indicate what each field is (data limit in GB, duration
  in days, etc.).

### Database
- Migration `0027_reseller_payment_config.sql` — `reseller_payment_config` table.
- Migration `0028_plan_admin_id.sql` — `admin_id` column + index on `plans`, backfill
  to primary sudo admin.
- Migration `0029_order_proof_image.sql` — `proof_image` TEXT column on `orders`.

## [1.2.6] - 2026-06-17

Node operations release: enrollment wizard, live connectivity diagnostics, and CLI
health checks for panel and node installs.

### Added
- **Node enrollment wizard** — four-step UI to copy the mTLS bundle, install on a
  remote server, register the node, and run an on-demand connectivity test.
- **Node health diagnostics** — hub classifies disconnects as mTLS failure,
  unreachable agent, or core down; badges and debug bundle on the Nodes page.
- **Enrollment API** — `GET /api/nodes/enrollment`, `POST /api/nodes/:id/test`,
  and `GET /api/nodes/:id/debug` for bundle delivery and support bundles.
- **`vortexui doctor`** — CLI checks for certs, services, ports, and `/health`
  on panel, node, and docker installs; suggested after `vortexui update`.
- **Enrollment polish** — QR code alongside bundle copy; enrollment phases
  (`pending` / `connected` / `synced`); TCP vs mTLS badges (`Network OK`, `CA match`);
  automatic CA comparison on connectivity test; Telegram/webhook alert after 5+ minutes
  disconnected; migration count check in `vortexui doctor` (native panel).
- **Reseller wallet UI (minimal)** — sudo can top up reseller wallets from Admins;
  resellers export wallet ledger as CSV (statement); parent resellers can top up
  sub-reseller wallets; `GET /api/account/wallet/export`.
- **Reseller wallet billing** — wallet packages (multi-currency); ZarinPal and
  NowPayments online top-up; card-to-card and crypto with TX ID / screenshot;
  sudo approval queue; billing settings for card and crypto addresses.

### Fixed
- **Reseller wallet quota stacking** — wallet purchases now add additively onto the
  reseller's main quota (`TrafficQuota` + `UserQuota`) immediately on admin approval /
  payment completion, applied on top of the existing allowance instead of being held
  until the previous balance is consumed.
- **`vortexui doctor`** — health check now hits `/api/health` (was `/health`);
  migration check warns on extra DB records instead of false-failing when
  `goose_db_version` count exceeds embedded SQL files.

## [1.2.5] - 2026-06-17

Reseller platform release: scoped allowlists, quota modes, wallet and sub-reseller
hierarchy, whitelabel branding, outbound webhooks, policy enforcement, auto-suspend,
and full i18n across all eight panel languages.

### Added
- **Reseller allowlists** — per-admin plan, node, and inbound pickers restrict what
  resellers can sell, deploy on, and assign to users.
- **Traffic quota modes** — `allocated` (pool assigned across users) or `consumed`
  (actual usage counts against the reseller pool).
- **Reseller dashboard** — accounts, traffic assigned/consumed, users by status, top
  consumers, and expiring users summary at `/reseller-dashboard`.
- **Reseller quota alerts** — Telegram and optional webhook notifications when
  resellers approach account or traffic thresholds (`/reseller-quota-alerts`).
- **Reseller wallet** — traffic and user credits with ledger history; sudo/parent
  top-up via the Admins page.
- **Sub-resellers** — create child resellers under a parent with role and quota
  assignment from **Reseller account**.
- **Portal whitelabel** — per-reseller panel title, logo URL, accent color, portal
  slug, and footer text.
- **Outbound reseller webhook** — HMAC-SHA256 signed `user.created` / `user.deleted`
  events for automation integrations.
- **Impersonate** — sudo admins can log in as a reseller (`Login as`) for support.
- **Scoped audit log** — resellers see only their own mutating actions; sudo sees all.
- **Reseller policy limits** — max data limit, max expire days, bulk create/import and
  bulk delete toggles enforced when resellers manage users.
- **Auto-suspend** — optional suspension on IP violations (7-day window) or quota
  overage with configurable grace minutes; background worker evaluates resellers.
- **CSV export** — resellers download an owned-users report from the dashboard.
- **Bulk quota adjust** — sudo quick-adjust buttons (+50 accounts, +10/+50 GB) on the
  Admins quota usage table.
- **Reseller navigation** — dedicated sidebar section (dashboard, account, quota alerts,
  my quota) plus quota summary cards on Overview.

### Changed
- **My quota** route redirects to the reseller dashboard for a single quota home.
- **Reseller UI i18n** — dashboard, account, quota alerts, Admins, and Edit Admin
  modal translated in EN, FA, TR, AR, RU, ZH, JA, and ES.

### Fixed
- **Reseller SQL migrations** — add missing `-- +goose Up` / `-- +goose Down` directives
  to migrations `0021`–`0023` so native panel installs apply schema updates on startup.
- **Native update** — `vortexui update` rebuilds the panel with `go build -a` so embedded
  migrations are always re-baked into the binary.
- **Edit Admin modal** — fix crash when role/node/plan lists are null while editing resellers.
- **Reseller panel UX** — sudo-controlled sub-reseller creation, per-section Settings access,
  logo upload and color picker for portal branding, read-only Nodes, scoped Live Monitor and
  Analytics, and JSON backup of owned users only.

### Database
- Migrations `0021_reseller_enhancements.sql`, `0022_reseller_advanced.sql`,
  `0023_reseller_policy_suspend.sql`, `0024_reseller_features.sql`.

## [1.2.3] - 2026-06-17

### Added
- **Per-protocol capability constraints** — a single source-of-truth capability matrix
  now governs which protocols and transports each core (Xray / sing-box) supports;
  the guard, renderers, and frontend option filtering all derive from it.
- **New protocols** — `socks`, `http`, `naive` (sing-box), and `dokodemo` (xray)
  inbounds.
- **sing-box protocols** — `hysteria1`, `shadowtls`, and `anytls` support.
- **mKCP transport** — added as an Xray transport option.
- **TLS** — `alpn` configuration plus Xray `tcp` header (none/http) and `xhttp` mode
  rendering.
- **Subscription Hosts** — per-inbound host overrides (Marzban-style) projected into
  subscription share links: custom address/port/SNI/Host header/path/ALPN/fingerprint/
  security/fragment/mux per host, enable/disable plus ordering, and template variables
  (`{USERNAME}`, `{SERVER_IP}`, `{SERVER_IPV6}`, `{DATA_USAGE}`, `{DATA_LIMIT}`,
  `{DATA_LEFT}`, `{DAYS_LEFT}`, `{EXPIRE_DATE}`, `{ADMIN_USERNAME}`) in remark/address.
  New "Hosts" manager in the inbound UI.
- **New subscription output formats** — raw Xray/V2Ray JSON (`?format=xray`), Outline
  `ss://` (`?format=outline`), and plain non-base64 links for V2rayN (`?format=links`),
  with User-Agent auto-detection; existing base64/Clash/sing-box output unchanged.
- **Smart routing rule packs** — reusable named routing rule sets (built-in plus
  custom): apply a pack's rules to a node's live routing (reusing the routing engine
  plus resync) and/or embed them into Clash/sing-box subscription output; global
  default plus per-user selection. New Routing Packs management page.
- **Clean-IP scanner (Cloudflare)** — scan and score candidate CDN IPs by latency plus
  packet loss (bounded concurrency, SSRF-guarded to public IPs only), cache best-first
  results, and copy a clean IP into a subscription host. New scanner page.
- **IP-limit enforcement** — turns account-sharing detection into enforcement: warn,
  temporarily disable (auto-restore), or kill connections when a user exceeds its
  device/IP limit; configurable cooldown plus restore, with an events log
  (`kill_connections` applies to Xray nodes; sing-box degrades to temporary disable).

### Changed
- Renderer tests are now matrix-driven, covering every protocol/transport combination
  across both cores.

## [1.2.0] - 2026-06-17

The biggest release yet: **17 new features** and **24 UX improvements** transforming
VortexUI from a management panel into a complete anti-censorship platform with
self-service capabilities.

### Added — Features

#### Self-Service & Commerce
- **User Self-Service Portal** — end-users authenticate with their subscription token
  at `/portal`; view real-time usage, purchase plan renewals, and submit support
  tickets — no Telegram required. Separate portal JWT with 24h TTL.
- **Support Ticket System** — users create tickets from the portal; admins manage,
  reply, and close from `/tickets` with status filtering.
- **Family/Group Subscriptions** — shared data pools for multiple devices under one
  parent account with per-member quotas and max-member limits. Manage at `/family-groups`.
- **Invite/Referral System** — unique referral codes per user; configurable rewards
  (extra data, extra days, or discount); usage tracking; admin dashboard at `/referrals`.
- **Deep Links + QR** — custom URL scheme (`vortex://import?sub=...`) for one-tap
  subscription import into supported apps. Configurable base URL and app store
  fallbacks. Manage at `/deep-links`.
- **Quota Notifications** — notify users via Telegram or webhook when they hit
  configurable usage thresholds (50%, 80%, 90%, 100%) with cooldown. Manage at
  `/quota-notifications`.

#### Security & Anti-Censorship
- **Reality Scanner** — built-in TLS probe that scans a list of SNIs against any node,
  measures latency, scores compatibility 0-100, and caches results for quick reuse.
  Concurrent probing (10 goroutines). Admin page at `/reality-scanner`.
- **TLS Tricks Manager** — advanced ISP-specific anti-DPI profiles with one-click
  presets for Iranian ISPs (Hamrah Aval/MCI, Irancell/MTN, Mokhaberat/TCI, Shatel,
  Asiatech). Configurable fragment size/interval, mux concurrency, uTLS fingerprint,
  padding, ECH support, and auto-detect mode. Manage at `/tls-tricks`.
- **Active Probing Protection** — detect and block GFW/DPI active probes;
  configurable actions (block, honeypot, log-only); auto-expiring IP blocklist;
  Telegram alerts. Manage at `/probing-protection`.
- **Client Fingerprint Validator** — validate TLS fingerprints via JA3 hash; pre-built
  rules for Chrome/Firefox/Safari (allow) and curl/Go/Python (block); configurable
  default action for unknowns. Manage at `/fingerprint`.
- **Decoy Website** — serve a legitimate-looking website (reverse-proxy mode or static
  HTML upload) to connections that arrive without valid credentials — protects server
  identity from active probing. Configure at `/decoy-website`.
- **DNS-over-HTTPS** — built-in DoH server for subscriber privacy; upstream DNS
  configurable; ad-blocking and malware-blocking lists; custom domain blocklist; query
  logging and statistics. Configure at `/doh`.
- **Multi-Domain SNI Routing + SSL Manager** — assign multiple domains to a single
  inbound with SNI-based routing rules; auto-provision Let's Encrypt or ZeroSSL
  certificates with wildcard support; auto-renewal; cert status monitoring. Manage at
  `/sni-manager`.

#### Infrastructure & Scale
- **Smart Quota (Fair Use)** — configurable multi-tier speed reduction policies.
  Instead of hard-cutting at 100%, progressively throttle (e.g., 80% → 1 MB/s,
  100% → 512 KB/s, or block). Per-tier actions: reduce or block. Manage at `/smart-quota`.
- **CDN/Relay Chain Builder** — define multi-hop relay paths (User → CDN → Relay →
  Worker → Node) with per-hop protocol, SNI, path, and host configuration. Visual
  builder at `/relay-chains`.
- **Node Health Auto-Migration** — automatic user migration when nodes exceed CPU,
  memory, or packet-loss thresholds; configurable consecutive-failure count before
  trigger; migrate-back on recovery. Settings at `/migration`.
- **Multi-Panel Federation** — connect multiple VortexUI panels into a cluster; sync
  users and nodes across peers; SSO capability; sync event log. Manage at `/federation`.
- **Advanced Analytics** — geo-IP traffic breakdown by country, top 20 users by
  bandwidth, peak-hour distribution chart (24h heatmap), total up/down aggregates,
  and CSV export. Dashboard at `/analytics`.

### Added — UX Improvements (24)

#### Navigation & Discovery
- **Collapsible sidebar sections** — 30+ nav items organized into Dashboard, Users &
  Billing, Network & Nodes, Security, and System groups with chevron toggle.
  Auto-expands the section containing the active route.
- **Command palette (Ctrl+K / ⌘K)** — fuzzy search across all pages, users, and
  settings with keyboard navigation (↑↓ + Enter). Accessible from anywhere.
- **Keyboard shortcuts** — `N` navigates to Users, `S` opens search, `?` shows
  shortcuts help overlay.
- **Notification center** — bell icon in header with unread badge; dropdown shows
  recent alerts (SSE-powered); mark-all-read button.

#### Visual Polish
- **Skeleton loading states** — shimmer placeholders matching actual content shapes
  (cards, tables, charts) for all data-fetching states.
- **Animated page transitions** — CSS-based fade + slide on route changes.
- **Smooth theme transition** — dark ↔ light mode morphs colors over 300ms with
  rotating sun/moon icon animation.
- **Real-time gauges** — animated SVG circular gauges for CPU, RAM, and bandwidth
  with glow effects and color-coded thresholds.
- **World map heatmap** — geographic visualization of traffic origins on the Analytics
  page with pulsing data points.

#### Components
- **Professional data table** — reusable component with column sorting (asc/desc),
  fuzzy search filtering, pagination, row click handlers, and empty state.
- **Multi-step wizard** — step-by-step forms with progress indicators, validation
  per step, back/next navigation — used for inbound/node/plan creation.
- **Contextual help tooltips** — `?` icons next to complex fields with hover/focus
  explanations (positioned top/bottom/left/right).
- **Bottom sheet modal** — mobile-friendly modal that slides up from the bottom with
  drag handle and safe-area support.
- **Pull-to-refresh** — touch gesture for mobile portal pages.
- **Redesigned toast notifications** — corner position, progress bar showing remaining
  time, undo button for destructive actions.

#### Architecture
- **Error boundaries** — class component wrapping each page; catches render errors,
  shows friendly message with retry button.
- **Onboarding tour** — first-time admin walkthrough (7 steps) with spotlight effect,
  progress dots, skip button. Re-triggerable from Settings.
- **Customizable dashboard widgets** — drag & drop layout; resize (S/M/L); hide/show;
  persisted to localStorage per admin; reset to defaults.
- **Mobile-first portal layout** — dedicated bottom navigation bar, safe-area padding,
  large touch targets (44px+), swipe-friendly cards.
- **Optimistic UI updates** — mutations reflect instantly before server confirmation.
- **PWA enhancements** — improved manifest, offline-ready service worker patterns.
- **Accessibility** — aria-labels on all interactive elements, focus management,
  keyboard navigation throughout.
- **Code splitting** — lazy-loaded routes with Suspense + skeleton fallbacks.

### Changed
- Panel version bumped to **1.2.0** in sidebar footer.
- Sidebar navigation completely restructured into 5 collapsible groups.
- Theme toggle now uses smooth rotation animation between sun/moon icons.
- Portal layout available in both desktop (sidebar) and mobile (bottom nav) variants.

### Database
New tables added for v1.2.0 features:
- `tickets`, `ticket_messages` — support system
- `reality_scans` — scanner result cache
- `quota_policies` — fair-use tiers
- `relay_chains` — CDN/relay path definitions
- `decoy_sites` — decoy website config
- `traffic_geo` — geographic traffic breakdown
- `migration_events`, `migration_policy` — auto-migration
- `probing_policy`, `probe_events`, `blocked_ips` — probing protection
- `family_groups`, `family_members` — group subscriptions
- `referral_config`, `referral_codes`, `referral_events` — referral system
- `doh_config`, `doh_query_logs` — DNS-over-HTTPS
- `sni_domains`, `ssl_certificates`, `sni_routes` — SNI/SSL management
- `tls_trick_profiles` — anti-DPI profiles
- `fingerprint_policy`, `fingerprint_rules`, `fingerprint_events` — fingerprint validation
- `federation_config`, `federation_peers`, `federation_sync_events` — panel federation
- `deeplink_config` — deep link settings
- `quota_notify_config`, `quota_notify_events` — notification tracking

Indexes added for performance on high-cardinality queries (tickets by user/status,
scans by node+score, relay chains by node, probe events by IP, family members by
group, referral codes by user, DoH logs by time).

---

## [1.1.0] - 2026-06-17

### Added
- **Reseller sub-panel** — non-sudo admins only see/manage their own users (`admin_id` FK + scoped API).
- **Payment gateways** — ZarinPal (IRR) + NowPayments (crypto USDT/BTC/TON) with full purchase flow.
- **Plan system** — define subscription plans (data/duration/price), user self-purchase, order tracking.
- **Payment IPN webhook** — async verification so no payment is missed even if user doesn't redirect.
- **Evasion Profiles** — reusable anti-DPI presets: TLS fragment, mux, uTLS fingerprint (one-click).
- **WARP+ Integration** — Cloudflare WARP as outbound for clean IP / censorship bypass.
- **IP Whitelist/Blacklist** — restrict panel access by IP/CIDR.
- **WireGuard protocol** — inbound support with per-user keypairs + subscription config.
- **Geo-blocking per-inbound** — allow/block countries by ISO code.
- **Cluster Mode (HA)** — multiple panels share one DB; Redis-based leader election.
- **Auto-update** — check for new releases + download binaries from GitHub.
- **Grafana dashboard** — ready-to-import JSON template for Prometheus metrics.
- **Prometheus /metrics** — panel-wide counters (users, nodes, traffic, connections).
- **Auto-backup to Telegram/S3** — scheduled daily export.
- **User-facing Telegram bot** — subscribers authenticate with token, view usage/configs/plans.
- **Per-user notifications** — Telegram alerts for expiry warning, limit reached, reset.
- **Auto-select best node** — subscription includes url-test group for automatic failover.
- **Self-service purchase** — "View Plans & Purchase" button on subscription info page.
- **Custom branding** — panel name, accent color, logo URL, footer text.
- **PWA manifest** — installable as mobile app from browser.
- **Sub page multi-language** — auto-detects browser language (FA/AR/TR/RU/ZH) with RTL.

### Fixed
- **Core version display** — health poll reads and persists core_version from node agents.

### Changed
- Subscription output includes Auto (url-test) proxy group.
- WireGuard added to inbound protocol dropdown.

---

## [1.0.1] - 2026-06-17

### Added
- **Interactive Telegram Bot** — admin commands (/status, /users, /online, /nodes, /find, /limit, /unlimit).
- **Expiry Warning** — automatic alert 3 days before subscriptions expire.
- **Admin Quota Enforcement** — non-sudo admins respect UserQuota and TrafficQuota.
- **Bandwidth Limit** — per-inbound speed_limit (bytes/sec).
- **Certificate Manager** — ACME-ready with domain-based issuance and caching.
- **Cloudflare DNS Automation** — auto-create/update A records for nodes.
- **Subscription Info Page** — beautiful public HTML page with usage, QR, configs, traffic chart.
- **Traffic Chart** — 7-day usage bar chart on subscription page.
- **Config Template Engine** — ClashTemplate and SingboxTemplate for subscription output.
- **Docker GHCR Publish** — multi-arch images (amd64/arm64).
- **Node Endpoint field** — custom tunnel/CDN/relay address per node.

### Fixed
- Resilient config builder (skips misconfigured inbounds).
- gRPC keepalive (prevents idle connection drops).
- TLS ServerName flexibility (CA-only validation).
- Hysteria2/TUIC auto-lock to UDP + TLS in frontend.

### Changed
- Reconnect backoff reduced (0.5s–15s).
- Unsupported protocols skip instead of failing.

---

## [1.0.0] - 2026-06-15

First stable release.

### Core & Nodes
- Core-agnostic engine (Xray + sing-box), selectable per node.
- In-process local node + remote node agents over gRPC + mTLS.
- Push-based delta traffic accounting, live health monitoring.
- Built-in REALITY key generation.

### User Management
- User-centric model (one identity → many protocols).
- Subscription delivery with Clash/sing-box/base64 auto-detection.
- Quota enforcement, device limits, HWID allowlist, bulk add, import.

### Network Policy
- Outbounds, routing rules, load balancers with observatory.
- JSON editor for outbounds/inbounds with share-link import.
- GeoIP/Geosite updater with Iran routing rules.

### Security
- JWT + TOTP 2FA, RBAC, API tokens, login brute-force lockout.
- Account-sharing guard, audit log.

### Operations
- Automatic HTTPS via Caddy + Let's Encrypt.
- One-line installer + `vortexui` management console.
- Docker Compose stack.

### Notifications
- Webhook (HMAC-SHA256) + Telegram notifiers.
- Live updates over SSE.

### Frontend
- React 18 + TypeScript + Tailwind; dark + light themes; responsive; 8 languages.

---

[Unreleased]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.9...HEAD
[1.2.9]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.8...v1.2.9
[1.2.8]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.7...v1.2.8
[1.2.7]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.6...v1.2.7
[1.2.6]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.5...v1.2.6
[1.2.5]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.3...v1.2.5
[1.2.3]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.0...v1.2.3
[1.2.0]: https://github.com/iPmartNetwork/VortexUI/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/iPmartNetwork/VortexUI/releases/tag/v1.0.0
