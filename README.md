<div align="center">
  <img src="img/Logo.svg" alt="VortexUI" width="300" />
  <p><strong>Next-generation proxy management panel</strong></p>
  <p>Core-agnostic · User-centric · Real-time · Anti-censorship</p>

  [![Release](https://img.shields.io/github/v/release/iPmartNetwork/VortexUI?style=flat-square&color=blue)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![Downloads](https://img.shields.io/github/downloads/iPmartNetwork/VortexUI/total?style=flat-square&color=green)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![Stars](https://img.shields.io/github/stars/iPmartNetwork/VortexUI?style=flat-square&color=yellow)](https://github.com/iPmartNetwork/VortexUI/stargazers)
  [![License](https://img.shields.io/github/license/iPmartNetwork/VortexUI?style=flat-square)](LICENSE)
  [![CI](https://img.shields.io/github/actions/workflow/status/iPmartNetwork/VortexUI/ci.yml?style=flat-square&label=CI)](https://github.com/iPmartNetwork/VortexUI/actions)
  [![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go)](https://go.dev)
  [![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
  [![Container](https://img.shields.io/badge/ghcr.io-vortexui-2496ED?style=flat-square&logo=docker&logoColor=white)](https://github.com/iPmartNetwork/VortexUI/pkgs/container/vortexui-panel)
  [![Telegram](https://img.shields.io/badge/Telegram-@vortex__ui-26A5E4?style=flat-square&logo=telegram&logoColor=white)](https://t.me/vortex_ui)
  
  ![Visitors](https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2FiPmartNetwork%2FVortexUI&countColor=%23263759)

  <br />

  <strong>English</strong> · <a href="README.fa.md">فارسی</a>

  <br />
  
  [Features](#-features) · [What's New in 1.2](#-whats-new-in-12) · [Screenshots](#-screenshots) · [Comparison](#-comparison) · [Quick Start](#-quick-start) · [Protocols](#-supported-protocols) · [Roadmap](#-roadmap) · [Contributing](#-contributing)
</div>

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🔧 Core Engine
- **Xray-core** and **sing-box** — choose per node
- In-process local node (no separate agent needed)
- Hot-reload config, runtime user add/remove
- REALITY key generation built-in
- **Reality Scanner** — auto-discover best SNIs

### 👥 User Management
- User-centric model (one identity → many protocols)
- Subscription: auto-detect Clash/sing-box/base64
- **Self-service portal** (login with sub token, view usage, buy plans, open tickets)
- **Family/group subscriptions** (shared data pool)
- **Smart Quota** — progressive speed reduction instead of hard-cut
- **Referral system** — invite codes with rewards
- Config templates (custom Clash/sing-box routing)
- QR codes + deep links (`vortex://` scheme)
- Traffic accounting: delta-based, restart-safe
- Quota enforcement + scheduled reset
- Device limit + HWID allowlist
- Bulk actions + import from 3x-ui / Marzban

### 🌐 Network & Routing
- Outbounds: freedom/blackhole/dns + proxy chaining
- **CDN/Relay chain builder** (multi-hop paths)
- Routing rules: domain/IP/port/protocol matchers
- **Multi-domain SNI routing** + auto SSL
- Load balancers with health probing
- GeoIP/Geosite updater with Iran rules
- **Panel Federation** — sync users/nodes across panels

</td>
<td width="50%">

### 🖥 Node Fleet
- mTLS connections (panel ↔ node)
- **Auto-migration** — move users from unhealthy nodes
- Live health monitoring (CPU/RAM/Disk)
- Remote restart / stop core
- Custom endpoint (tunnel/CDN/relay)
- Cloudflare DNS automation
- Per-node logs streaming

### 🛡 Security & Anti-Censorship
- **TLS Tricks Manager** — ISP-specific profiles (Hamrah Aval, Irancell, Mokhaberat)
- **Active probing protection** — detect & block GFW probes
- **Client fingerprint validator** — block curl/Go/Python
- **Decoy website** — serve fake site to probers
- **Evasion profiles** (fragment, mux, uTLS, ECH)
- WARP+ integration
- **DNS-over-HTTPS** server (built-in, ad/malware blocking)
- IP whitelist/blacklist
- Geo-blocking per inbound

### 🔐 Auth & Admin
- JWT + TOTP 2FA
- RBAC + reseller sub-panel
- API tokens (PAT)
- Login brute-force protection
- Account-sharing guard
- Audit log
- **Support ticket system**

### 🎨 Frontend & UX
- React 18 + TypeScript + Tailwind
- 8 languages with full RTL
- Dark + Light themes (smooth transition)
- **Command palette** (Ctrl+K)
- **Customizable dashboard widgets** (drag & drop)
- **Onboarding tour** for new admins
- **Mobile-first portal** (bottom nav, pull-to-refresh)
- Real-time charts + animated gauges
- **World map** geo-visualization
- Skeleton loading states
- Keyboard shortcuts
- Error boundaries with retry
- PWA (installable mobile app)

</td>
</tr>
</table>

---

## 🆕 What's New in 1.2

<div align="center">

**17 new features + 24 UX improvements in a single release**

</div>

<details open>
<summary><strong>🚀 Major Features</strong></summary>

| Feature | Description |
|---------|-------------|
| **User Self-Service Portal** | End-users login with their sub token, view usage, buy plans, submit support tickets |
| **Reality Scanner** | Built-in TLS probe — scan SNIs, measure latency, score compatibility (0-100) |
| **Smart Quota (Fair Use)** | Progressive speed reduction at configurable thresholds instead of hard-cut |
| **CDN/Relay Chain Builder** | Define multi-hop relay paths with per-hop protocol/SNI/path config |
| **Decoy Website** | Serve fake site (reverse-proxy or static HTML) to active probers |
| **Advanced Analytics** | Geo-IP breakdown, top users, peak hours, CSV export |
| **Node Auto-Migration** | Automatic user migration when nodes become unhealthy |
| **Active Probing Protection** | Detect and block GFW/DPI probes with IP blocklist |
| **Family/Group Subscriptions** | Shared data pools for multiple devices under one parent |
| **Invite/Referral System** | Referral codes with configurable rewards (data/days/discount) |
| **DNS-over-HTTPS** | Built-in DoH server with ad/malware blocking |
| **Multi-Domain SNI + SSL** | Multiple domains per inbound, auto Let's Encrypt/ZeroSSL |
| **TLS Tricks Manager** | ISP-specific anti-DPI profiles with one-click presets |
| **Client Fingerprint Validator** | JA3-based filtering — allow Chrome/Firefox, block curl/Go |
| **Multi-Panel Federation** | Sync users and nodes across multiple VortexUI panels |
| **Deep Links + QR** | Custom URL scheme (`vortex://`) for one-tap subscription import |
| **Quota Notifications** | Telegram/webhook alerts at configurable usage thresholds |

</details>

<details>
<summary><strong>🎨 UX Improvements (24)</strong></summary>

- Collapsible sidebar sections (Dashboard, Users, Network, Security, System)
- Command palette with fuzzy search (Ctrl+K)
- Skeleton loading states (shimmer placeholders)
- Professional data table with sort, filter, pagination
- Animated page transitions (CSS-based)
- Code splitting with lazy routes
- Redesigned toast notifications (progress bar + undo)
- Notification center (bell dropdown)
- Keyboard shortcuts (N/S/?)
- Error boundaries with retry button
- Animated circular gauges (CPU/RAM/Bandwidth)
- World map heatmap (geo analytics)
- Multi-step wizard component
- Contextual help tooltips
- Optimistic UI updates
- Enhanced PWA support
- Accessibility improvements (aria, focus management)
- Smooth dark/light theme transition
- Onboarding tour for first-time users
- Customizable dashboard widgets (drag & drop)
- Mobile-first portal layout (bottom nav)
- Bottom sheet modals
- Pull-to-refresh gesture
- Safe-area support (iPhone notch)

</details>

---

## 📸 Screenshots

<details>
<summary><strong>🌙 Dark Mode</strong></summary>
<br />

| Dashboard | Nodes | Users |
|:---------:|:-----:|:-----:|
| ![Overview Dark](img/panel/overview_dark.png) | ![Nodes Dark](img/panel/Node_dark.png) | ![Users Dark](img/panel/User_dark.png) |

</details>

<details open>
<summary><strong>☀️ Light Mode</strong></summary>
<br />

| Dashboard | Nodes | Users |
|:---------:|:-----:|:-----:|
| ![Overview Light](img/panel/overview_light.png) | ![Nodes Light](img/panel/Node_light.png) | ![Users Light](img/panel/User_light.png) |

</details>

---

## ⚖️ Comparison

<div align="center">

### How VortexUI stacks up against other panels

</div>

| | VortexUI 1.2 | 3x-ui | Marzban | Hiddify |
|:--|:--:|:--:|:--:|:--:|
| **Proxy engines** | Xray + sing-box | Xray | Xray | Xray + sing-box |
| **Data model** | User-centric | Inbound-centric | User-centric | User-centric |
| **Traffic method** | Push delta | Polling | Polling | Polling |
| **Multi-node** | mTLS + auto-migration | ✅ | ✅ | ✅ |
| **Balancer** | ✅ 4 strategies | ❌ | ❌ | ❌ |
| **Outbound/Routing** | ✅ full CRUD | Partial | ❌ | ❌ |
| **Reality Scanner** | ✅ built-in | ❌ | ❌ | ❌ |
| **Anti-DPI profiles** | ✅ ISP-specific | ❌ | ❌ | ✅ |
| **Self-service portal** | ✅ | ❌ | ❌ | ✅ |
| **Family groups** | ✅ | ❌ | ❌ | ❌ |
| **Federation** | ✅ multi-panel sync | ❌ | ❌ | ❌ |
| **Referral system** | ✅ | ❌ | ❌ | ❌ |
| **Probing protection** | ✅ detect + block | ❌ | ❌ | ❌ |
| **Fingerprint validation** | ✅ JA3 | ❌ | ❌ | ❌ |
| **Decoy website** | ✅ | ❌ | ❌ | ❌ |
| **DNS-over-HTTPS** | ✅ built-in | ❌ | ❌ | ❌ |
| **Deep links** | ✅ custom scheme | ❌ | ❌ | ✅ |
| **Smart quota** | ✅ progressive | ❌ | ❌ | ❌ |
| **CDN/Relay chains** | ✅ visual builder | ❌ | ❌ | ❌ |
| **Analytics (geo)** | ✅ + CSV export | ❌ | ❌ | ❌ |
| **Notifications** | Webhook + TG + portal | TG | ✅ | TG |
| **Languages** | 8 | 13 | 3 | 5 |
| **Backend** | Go | Go | Python | Python |
| **Database** | PG + TimescaleDB | SQLite/PG | SQLite/Maria | SQLite |

---

## 📡 Supported Protocols

<div align="center">

| Protocol | Inbound | Outbound | Transport |
|----------|:-------:|:--------:|:---------:|
| **VLESS** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, xHTTP |
| **VMess** | ✅ | ✅ | TCP, WS, gRPC |
| **Trojan** | ✅ | ✅ | TCP, WS, gRPC |
| **Shadowsocks** | ✅ | ✅ | TCP |
| **SOCKS** | — | ✅ | TCP |
| **HTTP** | — | ✅ | TCP |
| **Hysteria2** | ✅ (sing-box) | — | UDP |
| **TUIC** | ✅ (sing-box) | — | UDP |
| **WireGuard** | ✅ | — | UDP |

</div>

**Security layers:** None, TLS, REALITY (with built-in scanner)

---

## 🚀 Quick Start

### One-line Install

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

The installer asks:
1. **Method** — Docker Compose *(recommended)* or Native (systemd)
2. **Access** — Domain + auto HTTPS (Let's Encrypt) or IP + HTTP

Then generates secrets, mTLS certs, starts the stack, creates your first admin, and installs the `vortexui` CLI.

Non-interactive:
```bash
VORTEXUI_METHOD=docker VORTEXUI_NONINTERACTIVE=1 \
  VORTEXUI_ADMIN_USER=admin VORTEXUI_ADMIN_PASS='s3cret' \
  bash install.sh
```

### Management Console

```text
$ vortexui

   1) Start            2) Stop
   3) Restart          4) Status
   5) Logs             6) Update
   7) Create admin     8) Change web port
   9) Domain / SSL    10) Settings / URL
  11) Uninstall        0) Exit
```

### Docker

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI
docker compose up -d
```

### Manual Build

```bash
cp .env.example .env    # edit secrets
make build              # compile Go binaries
make certs              # generate dev mTLS certs
make run-panel          # start panel
./bin/panel admin create --username admin --password 'your-pass' --sudo
```

---

## 🔒 Operations

| Feature | How |
|---------|-----|
| **Automatic HTTPS** | Caddy + Let's Encrypt — zero config renewal |
| **Live updates** | SSE push — no polling, instant UI refresh |
| **GeoIP/Geosite** | One-click Iran routing rules update per node |
| **Account-sharing guard** | Online IP enforcement + auto-limit option |
| **Auto-backup** | Scheduled exports to Telegram or S3 |
| **Prometheus metrics** | `/metrics` endpoint + Grafana dashboard |

---

## 📖 Documentation

| Topic | Link |
|-------|------|
| **Documentation site** | [ipmartnetwork.github.io/VortexUI](https://ipmartnetwork.github.io/VortexUI/) |
| **Telegram** | [@vortex_ui](https://t.me/vortex_ui) |
| **Discussions** | [GitHub Q&A](https://github.com/iPmartNetwork/VortexUI/discussions) |
| **Wiki** | [EN](docs/wiki/en/README.md) · [FA](docs/wiki/fa/README.md) · [AR](docs/wiki/ar/README.md) · [TR](docs/wiki/tr/README.md) |
| **API (OpenAPI 3.0)** | [`docs/openapi.yaml`](docs/openapi.yaml) |
| **Protocols** | [`docs/protocols.md`](docs/protocols.md) |
| **Changelog** | [`CHANGELOG.md`](CHANGELOG.md) |
| **Contributing** | [`CONTRIBUTING.md`](CONTRIBUTING.md) |

---

## 🗺 Roadmap

<details>
<summary><strong>Completed (v1.0 → v1.2)</strong></summary>

- [x] Core-agnostic engine (Xray + sing-box)
- [x] User-centric data model + push delta traffic
- [x] Multi-node with mTLS + auto-failover
- [x] Outbound/Routing/Balancer management
- [x] REALITY key generation + scanner
- [x] Webhook + Telegram notifications
- [x] Interactive Telegram bot
- [x] Backup/Restore + auto-backup (TG/S3)
- [x] Audit log + API tokens
- [x] Account-sharing guard
- [x] Import from 3x-ui / Marzban
- [x] 8-language frontend + RTL
- [x] Real-time dashboard (SSE)
- [x] Automatic HTTPS (Caddy)
- [x] One-line installer + CLI
- [x] Hysteria2 + TUIC + WireGuard
- [x] Reseller sub-panel
- [x] Payment gateways (ZarinPal + crypto)
- [x] Evasion profiles + WARP+
- [x] Cluster mode (HA)
- [x] Grafana/Prometheus metrics
- [x] Self-service portal
- [x] Reality Scanner
- [x] Smart Quota (fair use)
- [x] CDN/Relay chains
- [x] Decoy website
- [x] Advanced analytics (geo)
- [x] Node auto-migration
- [x] Active probing protection
- [x] Family/group subscriptions
- [x] Referral system
- [x] DNS-over-HTTPS
- [x] Multi-domain SNI + auto SSL
- [x] TLS Tricks (ISP profiles)
- [x] Client fingerprint validator
- [x] Multi-panel federation
- [x] Deep links + QR
- [x] Quota notifications
- [x] Command palette + keyboard shortcuts
- [x] Dashboard widgets + onboarding tour
- [x] Mobile-first portal

</details>

### Coming Next
- [ ] Mobile app (React Native / Flutter)
- [ ] AI-powered anomaly detection
- [ ] Multi-language docs expansion
- [ ] Proxy-level rate limiting per user
- [ ] Plugin system for custom extensions
- [ ] WebSocket transport support for sing-box

---

## 💝 Support

If VortexUI is useful to you:

- ⭐ **Star** this repository
- 🍴 **Fork** and contribute
- 📢 **Share** with your community
- 💬 **Join** [@vortex_ui](https://t.me/vortex_ui) on Telegram

<div align="center">

| Network | Address |
|:-------:|---------|
| **USDT (TRC20)** | `TRLnjZ7YDSwjh3oay28qigEYNieGPMs6ew` |
| **BTC** | `bc1qszt4g7jdv7ev2t3pexctc07ults8nfflht3nj5` |
| **TON** | `UQAYSSSirtQ9_67ZHYUgLVLMx9Ir9vvh3vpcq2qbpit_8-Db` |

</div>

---

## 🤝 Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/amazing`)
3. Commit (`git commit -m 'feat: add amazing feature'`)
4. Push (`git push origin feat/amazing`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 🌐 Internationalization

<div align="center">

| 🇺🇸 English | 🇮🇷 فارسی | 🇹🇷 Türkçe | 🇸🇦 العربية |
|:-----------:|:---------:|:----------:|:-----------:|
| 🇷🇺 Русский | 🇨🇳 中文 | 🇯🇵 日本語 | 🇪🇸 Español |

</div>

Full RTL support for Persian and Arabic.

---

## 📄 License

GPL-3.0 — see [LICENSE](LICENSE).

---

<div align="center">
  <br />
  <img src="img/Logo.svg" alt="VortexUI" width="200" />
  <br /><br />
  <sub>© 2026 iPmart Network. All rights reserved.</sub>
  <br /><br />
  
  **Made with ❤️ by [iPmart Network](https://github.com/iPmartNetwork)**
  
  [Telegram @vortex_ui](https://t.me/vortex_ui) · [Documentation](https://ipmartnetwork.github.io/VortexUI/)
  
  ⭐ Star this repo if you find it useful!
</div>
