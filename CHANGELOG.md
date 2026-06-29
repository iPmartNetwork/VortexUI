# Changelog

All notable changes to VortexUI are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.2.7] - 2026-06-17

### Added
- **Self-service plan renewal** — end-users can purchase a plan from their subscription page to extend their existing account. Traffic and duration stack additively onto the current balance (no new account created, no usage reset). New server-rendered shop page at `/sub/:token/shop` with gateway selection (ZarinPal / Crypto). Payment result page at `/payment/result`.

### Fixed
- Portal "Purchase" button now properly initiates payment (was previously broken — GET to POST endpoint).
- Subscription info "View Plans & Purchase" link now opens the shop page (was linking to raw JSON).

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

[Unreleased]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.7...HEAD
[1.2.7]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.6...v1.2.7
[1.2.6]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.5...v1.2.6
[1.2.5]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.3...v1.2.5
[1.2.3]: https://github.com/iPmartNetwork/VortexUI/compare/v1.2.0...v1.2.3
[1.2.0]: https://github.com/iPmartNetwork/VortexUI/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.1...v1.1.0
[1.0.1]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/iPmartNetwork/VortexUI/releases/tag/v1.0.0
