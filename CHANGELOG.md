# Changelog

All notable changes to VortexUI are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.1] - 2026-06-17

### Added
- **Interactive Telegram Bot** — admin commands (/status, /users, /online, /nodes, /find, /limit, /unlimit) in addition to one-way event notifications.
- **Expiry Warning** — automatic alert 3 days before user subscriptions expire (Telegram + Webhook).
- **Admin Quota Enforcement** — non-sudo admins respect UserQuota and TrafficQuota limits on user creation.
- **Bandwidth Limit field** — per-inbound `speed_limit` (bytes/sec) for throttling user download speed.
- **Certificate Manager** — ACME-ready cert manager with domain-based issuance and caching (self-signed bootstrap, production ACME ready).
- **Cloudflare DNS Automation** — auto-create/update A records when nodes are added (`VORTEX_CF_API_TOKEN` + `VORTEX_CF_ZONE_ID`).
- **Subscription Info Page** — beautiful public HTML page at `/sub/{token}` (browser auto-detect) showing usage, QR, configs, traffic chart.
- **Traffic Chart on Sub Page** — 7-day usage bar chart via `/sub/{token}/usage` public endpoint.
- **Config Template Engine** — `ClashTemplate` and `SingboxTemplate` for customizing subscription output (DNS, routing rules, proxy groups).
- **Docker GHCR Publish** workflow — multi-arch (amd64/arm64) images to GitHub Container Registry (manual trigger).
- **Node Endpoint** field — custom tunnel/CDN/relay address per node; subscription links use it instead of real IP.

### Fixed
- **Resilient config builder** — misconfigured inbounds (missing REALITY keys, empty Shadowsocks password) are skipped instead of crashing the core.
- **gRPC keepalive** — 20s client ping + 30s server ping prevents intermediate firewalls from killing idle node connections.
- **TLS ServerName** — panel no longer requires node cert SAN to match the node name; CA-only validation for multi-node flexibility.
- **Update script** — `vortexui update` uses `git fetch + reset --hard` instead of `git pull --ff-only` (works with force-push).
- **Hysteria2/TUIC transport** — auto-lock to UDP + TLS in the frontend when these protocols are selected.

### Changed
- Reconnect backoff reduced (0.5s–15s, was 1s–30s) for faster node recovery.
- Unsupported protocols in config builders now skip (continue) instead of failing the entire build.

## [1.0.0] - 2026-06-15

First stable release.

### Core & nodes
- Core-agnostic engine layer over **Xray-core** and **sing-box**, selectable per node.
- In-process **local node** (no separate agent required) plus remote node agents.
- Panel ↔ node hub over **gRPC + mTLS**, with auto-failover and migrate-back on recovery.
- Push-based **delta traffic accounting** (restart-safe), live health monitoring, remote restart/stop core, and per-node log streaming.
- Built-in **REALITY** key generation.

### User management
- User-centric model (one identity → many protocols).
- Subscription delivery with Clash/sing-box/base64 auto-detection, QR codes, and per-format links.
- Quota enforcement + scheduled reset, device limits, and HWID allowlist.
- **Bulk add** from a shared plan/template and **import from 3x-ui / Marzban**.

### Network policy
- Outbounds (freedom/blackhole/dns + proxy chaining), routing rules, and load balancers with health-probing observatory.
- 3x-ui-style **JSON editor** for outbounds and inbounds with share-link import (vmess/vless/trojan/ss/hysteria2/wireguard).
- **GeoIP/Geosite updater** with Iran routing rules (one-click per-node refresh + `POST /api/nodes/:id/geo-update`).

### Security
- JWT auth + **TOTP 2FA**, RBAC with granular permissions.
- **API tokens** (personal access tokens) for automation.
- **Login brute-force lockout** and an **account-sharing guard** (online-IP enforcement, alert or auto-limit).
- **Audit log** of all admin mutations.

### Operations & deployment
- **Automatic HTTPS** via Caddy + Let's Encrypt (domain prompt at install).
- One-line **installer** with Docker and Native (systemd) methods, and an interactive **`vortexui`** management console (3x-ui style).
- Docker Compose stack (web · panel · node · PostgreSQL/TimescaleDB · Redis).

### Notifications & observability
- Event bus with **webhook** (HMAC-SHA256 signed) and **Telegram** notifiers.
- **Live updates over SSE** — the UI refreshes the instant something changes instead of polling.
- Real-time dashboard with aggregate traffic time-series chart.
- Transactional backup / restore.

### Frontend
- React 18 + TypeScript + Tailwind; dark + light themes; responsive.
- **8 languages** (EN/FA/TR/AR/RU/ZH/JA/ES) with full RTL for Persian and Arabic.

[Unreleased]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.1...HEAD
[1.0.1]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/iPmartNetwork/VortexUI/releases/tag/v1.0.0
