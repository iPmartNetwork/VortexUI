# Changelog

All notable changes to VortexUI are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres
to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/iPmartNetwork/VortexUI/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/iPmartNetwork/VortexUI/releases/tag/v1.0.0
