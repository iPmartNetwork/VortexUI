VortexUI README.md Preview
👁 Preview
📄 Raw Markdown
20,063 chars · 1 lines
Copy Markdown
<div align="center">

<img src="img/Logo.svg" alt="VortexUI Logo" width="200">

# VortexUI

**Next-generation, core-agnostic proxy management panel**

`Xray + sing-box` · User-centric · Real-time · Multi-node · Anti-censorship

[![Release](https://img.shields.io/github/v/release/iPmartNetwork/VortexUI?style=flat-square&color=00b4d8&label=Release)](https://github.com/iPmartNetwork/VortexUI/releases)
[![Stars](https://img.shields.io/github/stars/iPmartNetwork/VortexUI?style=flat-square&color=f5a623)](https://github.com/iPmartNetwork/VortexUI/stargazers)
[![License](https://img.shields.io/github/license/iPmartNetwork/VortexUI?style=flat-square&color=22c55e)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/iPmartNetwork/VortexUI/ci.yml?style=flat-square&label=CI)](https://github.com/iPmartNetwork/VortexUI/actions)
[![GHCR](https://img.shields.io/badge/ghcr.io-images-2496ED?style=flat-square&logo=docker&logoColor=white)](https://github.com/iPmartNetwork/VortexUI/pkgs/container/vortexui-panel)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
[![Telegram](https://img.shields.io/badge/Telegram-@vortex__ui-26A5E4?style=flat-square&logo=telegram&logoColor=white)](https://t.me/vortex_ui)

[![Visitors](https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2FiPmartNetwork%2FVortexUI&countColor=%23263759)](https://github.com/iPmartNetwork/VortexUI)

**English** · [فارسی](README.fa.md)

<sub>

[Features](#-features) · [What's New](#-whats-new-in-129) · [Screenshots](#-screenshots) · [Comparison](#-comparison) · [Quick Start](#-quick-start) · [Protocols](#-supported-protocols) · [Docs](#-documentation) · [Roadmap](#-roadmap) · [Contributing](#-contributing)

</sub>

</div>

---

## ✨ Features

<table>
<tr>
<td width="50%" valign="top">

### 🔧 Core Engine
- **Xray-core** & **sing-box** — choose per node
- In-process local node (no separate agent)
- Hot-reload config, runtime user add/remove
- REALITY key generation built-in
- **Reality Scanner** — auto-discover best SNIs

### 👥 User Management
- User-centric model (one identity → many protocols)
- Subscription: auto-detect Clash / sing-box / base64
- **Self-service portal** (login with sub token, view usage, buy plans, open tickets)
- **Subscription Hosts** — per-inbound CDN/SNI/host overrides
- **Family/group subscriptions** (shared data pool)
- **Smart Quota** — progressive speed reduction instead of hard-cut
- **Self-service shop** (per-reseller plans + card / crypto / ZarinPal)
- **Referral system** — invite codes with rewards
- Config templates (custom Clash / sing-box routing)
- QR codes + deep links (`vortex://` scheme)
- Traffic accounting: delta-based, restart-safe
- Quota enforcement + scheduled reset
- Device limit + HWID allowlist
- Bulk actions + import from 3x-ui / Marzban

### 🌐 Network & Routing
- Outbounds: freedom / blackhole / dns + proxy chaining
- **CDN/Relay chain builder** (multi-hop paths)
- **Smart routing rule packs** (per node or embedded in subscription)
- Routing rules: domain / IP / port / protocol matchers
- **Multi-domain SNI routing** + auto SSL
- Load balancers with health probing
- GeoIP/Geosite updater with Iran rules
- **Panel Federation** — sync users/nodes across panels

</td>
<td width="50%" valign="top">

### 🖥 Node Fleet
- mTLS connections (panel ↔ node)
- **Auto-migration** — move users from unhealthy nodes
- Live health monitoring (CPU / RAM / Disk)
- Remote restart / stop core
- Custom endpoint (tunnel / CDN / relay)
- Cloudflare DNS automation
- Per-node logs streaming

### 🛡 Security & Anti-Censorship
- **TLS Tricks Manager** — ISP-specific profiles
- **Active probing protection** — detect & block GFW probes
- **Client fingerprint validator** — block curl / Go / Python
- **Decoy website** — serve fake site to probers
- **Evasion profiles** (fragment, mux, uTLS, ECH)
- **Clean-IP scanner** (Cloudflare) & **IP-limit enforcement**
- WARP+ integration
- **DNS-over-HTTPS** server (built-in, ad/malware blocking)
- IP whitelist/blacklist · Geo-blocking per inbound

### 🔐 Auth & Admin
- JWT + TOTP 2FA
- RBAC + **full reseller platform** (wallet, sub-resellers, whitelabel, webhooks)
- API tokens (PAT) · Login brute-force protection
- Account-sharing guard · Audit log
- **Support ticket system**

### 🎨 Frontend & UX
- React 18 + TypeScript + Tailwind + **framer-motion**
- **Veltrix UI** — glass cards, stat tiles, cyan/sky theme
- 8 languages with **639 translated keys** and full RTL
- Dark + Light themes · **Command palette** (Ctrl+K)
- **Customizable dashboard widgets** (drag & drop)
- **Onboarding tour** for new admins
- Real-time charts + animated gauges + **World map**
- PWA (installable mobile app)

</td>
</tr>
</table>

---

## 🆕 What's New in 1.2.9

> **Command Tower UI · merged pages · Settings hub · reseller profiles · fleet telemetry**

| Feature | Description |
|---------|-------------|
| **Merged pages** | Routing & Balancers, Security Suite, and Reseller Platform each use one route with `?tab=` sub-navigation |
| **Settings hub** | Sidebar tabs for General, Security, Appearance, API, Backup, and Admins (sudo) |
| **Reseller profiles** | Click any reseller → wallet, quota bars, consumption, ledger, policies at `/settings/admins/:id` |
| **Admins sub-tabs** | Admins list, Roles, and Reseller access matrix inside Settings |
| **Command Tower Overview** | Live widgets with traffic ranges, top users + protocol, node geo/ping |
| **Inbounds page** | Dedicated `/inbounds` view separate from node fleet |
| **Node telemetry** | Region, country code, ping ms (migration 0030) |
| **Admin APIs** | `GET /api/admins/:id/quota` and `GET /api/admins/:id/wallet` |

<details>
<summary><strong>🔽 Previous Releases (1.2.8 → 1.2)</strong></summary>

### 🆕 What's New in 1.2.8

> **Veltrix UI · complete i18n · redesigned admin + portal shell**

| Feature | Description |
|---------|-------------|
| **Veltrix design system** | Glass cards, stat tiles, status badges, page-enter animations, cyan/sky palette |
| **New app shell** | Collapsible sidebar + header with mini mode, mobile drawer, theme/language switcher |
| **Command palette** | Fuzzy page search via Ctrl+K / ⌘K |
| **Live core pages** | Overview, Users, Nodes rebuilt with real-time API stat cards and fleet health |
| **Portal refresh** | Redesigned login, dashboard, desktop sidebar, mobile bottom navigation |
| **Full i18n** | 639 keys in 8 languages — billing, reseller payment, pending orders, shell, portal |

### 🆕 What's New in 1.2.7

> **Per-reseller commerce · owned plans · payment proof · self-service renewal**

| Feature | Description |
|---------|-------------|
| **Self-service renewal** | Users purchase plans from `/sub/:token/shop`; traffic + duration stack additively |
| **Per-reseller payment config** | Each reseller sets their own card number, crypto addresses, and ZarinPal merchant |
| **Per-reseller owned plans** | Resellers create plans with custom pricing; users only see their reseller's plans |
| **Payment proof upload** | Card-to-card requires receipt image; crypto accepts TX hash + screenshot |
| **Pending order review** | Admins see proof thumbnails, approve or reject manual payments |

### 🆕 What's New in 1.2.6

> **Node enrollment wizard · wallet billing · diagnostics · doctor CLI**

| Feature | Description |
|---------|-------------|
| **Node enrollment wizard** | Four-step UI: copy mTLS bundle → install → register → connectivity test |
| **Node health diagnostics** | Classify disconnects (mTLS failure / unreachable / core down); debug bundle |
| **`vortexui doctor`** | CLI checks certs, services, ports, and `/health` for panel/node/docker |
| **Reseller wallet billing** | Multi-currency packages, ZarinPal + NowPayments, card-to-card and crypto |
| **Wallet UI** | Top-up from Admins page, CSV ledger export, parent → sub-reseller top-up |

### 🆕 What's New in 1.2.5

> **Reseller platform · wallet & sub-resellers · whitelabel · webhooks · policy limits**

| Feature | Description |
|---------|-------------|
| **Allowlists** | Per-reseller plan, node, and inbound pickers |
| **Quota modes** | Allocated vs consumed traffic pool enforcement |
| **Reseller dashboard** | Accounts, traffic pool, top consumers, expiring users, CSV export |
| **Sub-resellers** | Hierarchical child resellers with role + quota |
| **Whitelabel** | Custom panel title, logo, accent, slug, footer |
| **Auto-suspend** | IP violation and quota overage suspension worker |
| **i18n** | All reseller pages in 8 languages |

See the [v1.2.5 features guide](docs/wiki/en/18-v125-features.md) for setup details.

### 🆕 What's New in 1.2.3

> **Subscription Hosts · routing packs · clean-IP scanner · IP-limit enforcement**

| Feature | Description |
|---------|-------------|
| **Subscription Hosts** | Marzban-style per-inbound host overrides projected into subscription links |
| **New output formats** | Raw Xray/V2Ray JSON, Outline `ss://`, plain V2rayN links |
| **Smart routing rule packs** | Reusable rulesets applicable per node or embedded in Clash/sing-box subscriptions |
| **Clean-IP scanner** | Scan & score CDN candidate IPs by latency + packet loss (SSRF-protected) |
| **IP-limit enforcement** | Warn / temp-disable / disconnect when user exceeds device limit |
| **New protocols** | SOCKS, HTTP, Naive, Dokodemo, Hysteria v1, ShadowTLS, AnyTLS, mKCP transport |

### 🆕 What's New in 1.2

> **17 new features + 24 UX improvements in a single release**

<table>
<tr><td>

**🚀 Major Features:**
Self-Service Portal · Reality Scanner · Smart Quota · CDN/Relay Chain Builder · Decoy Website · Advanced Analytics · Node Auto-Migration · Active Probing Protection · Family/Group Subscriptions · Referral System · DNS-over-HTTPS · Multi-Domain SNI + SSL · TLS Tricks Manager · Client Fingerprint Validator · Multi-Panel Federation · Deep Links + QR · Quota Notifications

</td></tr>
<tr><td>

**🎨 UX Improvements (24):**
Collapsible sidebar · Command palette · Skeleton loading · Data tables · Page transitions · Code splitting · Toast notifications · Notification center · Keyboard shortcuts · Error boundaries · Animated gauges · World map · Multi-step wizard · Help tooltips · Optimistic UI · PWA · Accessibility · Theme transition · Onboarding tour · Dashboard widgets · Mobile portal · Bottom sheets · Pull-to-refresh · Safe-area support

</td></tr>
</table>

</details>

📖 Full details: [CHANGELOG.md](CHANGELOG.md) · [Documentation](https://ipmartnetwork.github.io/VortexUI/)

---

## 📸 Screenshots

<table>
<tr>
<td align="center" colspan="3"><strong>🌙 Dark Mode</strong></td>
</tr>
<tr>
<td align="center"><strong>Dashboard</strong></td>
<td align="center"><strong>Nodes</strong></td>
<td align="center"><strong>Users</strong></td>
</tr>
<tr>
<td><a href="img/panel/overview_dark.png"><img src="img/panel/overview_dark.png" alt="Dashboard Dark" width="300"></a></td>
<td><a href="img/panel/Node_dark.png"><img src="img/panel/Node_dark.png" alt="Nodes Dark" width="300"></a></td>
<td><a href="img/panel/User_dark.png"><img src="img/panel/User_dark.png" alt="Users Dark" width="300"></a></td>
</tr>
<tr>
<td align="center" colspan="3"><strong>☀️ Light Mode</strong></td>
</tr>
<tr>
<td align="center"><strong>Dashboard</strong></td>
<td align="center"><strong>Nodes</strong></td>
<td align="center"><strong>Users</strong></td>
</tr>
<tr>
<td><a href="img/panel/overview_light.png"><img src="img/panel/overview_light.png" alt="Dashboard Light" width="300"></a></td>
<td><a href="img/panel/Node_light.png"><img src="img/panel/Node_light.png" alt="Nodes Light" width="300"></a></td>
<td><a href="img/panel/User_light.png"><img src="img/panel/User_light.png" alt="Users Light" width="300"></a></td>
</tr>
</table>

---

## ⚖️ Comparison

### How VortexUI stacks up against other panels

|  | VortexUI 1.2.9 | 3x-ui | Marzban | Hiddify |
|--|----------------|-------|---------|---------|
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
| **Reseller platform** | ✅ wallet, sub-resellers, whitelabel | Partial | ✅ | Partial |
| **Payment gateways** | ✅ ZarinPal + crypto + card-to-card | ❌ | ❌ | ❌ |
| **Self-service shop** | ✅ per-reseller | ❌ | ❌ | ✅ |
| **Notifications** | Webhook + TG + portal | TG | ✅ | TG |
| **Languages** | 8 | 13 | 3 | 5 |
| **Backend** | Go | Go | Python | Python |
| **Database** | PG + TimescaleDB | SQLite/PG | SQLite/Maria | SQLite |

---

## 📡 Supported Protocols

| Protocol | Inbound | Outbound | Transport |
|----------|---------|----------|-----------|
| **VLESS** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, xHTTP, mKCP |
| **VMess** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, mKCP |
| **Trojan** | ✅ | ✅ | TCP, WS, gRPC, mKCP |
| **Shadowsocks** | ✅ | ✅ | TCP (+ SS-2022 multi-user) |
| **SOCKS** | ✅ | ✅ | TCP |
| **HTTP** | ✅ | ✅ | TCP |
| **Naive** | ✅ (sing-box) | — | TCP/TLS |
| **Dokodemo** | ✅ (xray) | — | TCP/UDP |
| **Hysteria2** | ✅ (sing-box) | — | UDP |
| **Hysteria (v1)** | ✅ (sing-box) | — | UDP |
| **TUIC** | ✅ (sing-box) | — | UDP |
| **ShadowTLS** | ✅ (sing-box) | — | TCP |
| **AnyTLS** | ✅ (sing-box) | — | TCP |
| **WireGuard** | ✅ | — | UDP |

**Subscription output:** base64 · Clash/Clash.Meta · sing-box · Xray JSON · Outline · plain links (auto-detected by client).

**Security layers:** None, TLS, REALITY (with built-in scanner)

---

## 🚀 Quick Start

### One-line Install

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

The installer asks:

1. **Method** — Docker Compose _(recommended)_ or Native (systemd)
2. **Access** — Domain + auto HTTPS (Let's Encrypt) or IP + HTTP

Then generates secrets, mTLS certs, starts the stack, creates your first admin, and installs the `vortexui` CLI.

> **💡 Non-interactive mode:**
> ```bash
> VORTEXUI_METHOD=docker VORTEXUI_NONINTERACTIVE=1 \
>   VORTEXUI_ADMIN_USER=admin VORTEXUI_ADMIN_PASS='s3cret' \
>   bash install.sh
> ```

### Management Console

After installation, type **`vortexui`** for the interactive menu:

```
$ vortexui

   1) Start            2) Stop
   3) Restart          4) Status
   5) Logs             6) Update
   7) Create admin     8) Change web port
   9) Domain / SSL    10) Settings / URL
  11) Uninstall        0) Exit
```

Or use sub-commands: `vortexui start|stop|restart|status|logs|update|admin|settings|uninstall`

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
| **Telegram** | [@vortex\_ui](https://t.me/vortex_ui) |
| **Discussions** | [GitHub Q&A](https://github.com/iPmartNetwork/VortexUI/discussions) |
| **Wiki** | [EN](docs/wiki/en/README.md) · [FA](docs/wiki/fa/README.md) · [AR](docs/wiki/ar/README.md) · [TR](docs/wiki/tr/README.md) |
| **API (OpenAPI 3.0)** | [docs/openapi.yaml](docs/openapi.yaml) |
| **Protocols** | [docs/protocols.md](docs/protocols.md) |
| **Changelog** | [CHANGELOG.md](CHANGELOG.md) |
| **Contributing** | [CONTRIBUTING.md](CONTRIBUTING.md) |

---

## 🗺 Roadmap

<details>
<summary><strong>✅ Completed (v1.0 → v1.2.9)</strong></summary>

- Core-agnostic engine (Xray + sing-box)
- User-centric data model + push delta traffic
- Multi-node with mTLS + auto-failover
- Outbound/Routing/Balancer management
- REALITY key generation + scanner
- Webhook + Telegram notifications
- Interactive Telegram bot
- Backup/Restore + auto-backup (TG/S3)
- Audit log + API tokens
- Account-sharing guard
- Import from 3x-ui / Marzban
- 8-language frontend + RTL
- Real-time dashboard (SSE)
- Automatic HTTPS (Caddy)
- One-line installer + CLI
- Hysteria2 + TUIC + WireGuard
- Reseller platform (v1.2.5)
- Payment gateways (ZarinPal + crypto)
- Evasion profiles + WARP+
- Cluster mode (HA)
- Grafana/Prometheus metrics
- Self-service portal
- Reality Scanner
- Smart Quota (fair use)
- CDN/Relay chains
- Decoy website
- Advanced analytics (geo)
- Node auto-migration
- Active probing protection
- Family/group subscriptions
- Referral system
- DNS-over-HTTPS
- Multi-domain SNI + auto SSL
- TLS Tricks (ISP profiles)
- Client fingerprint validator
- Multi-panel federation
- Deep links + QR
- Quota notifications
- Command palette + keyboard shortcuts
- Dashboard widgets + onboarding tour
- Mobile-first portal
- Command Tower UI (v1.2.9)
- Veltrix UI redesign (v1.2.8)
- Complete 8-language i18n (v1.2.8)
- Per-reseller payment configuration (v1.2.7)
- Per-reseller owned plans (v1.2.7)
- Payment proof/receipt uploads (v1.2.7)
- Node enrollment wizard (v1.2.6)
- Reseller wallet billing (v1.2.6)

</details>

### 🔜 Coming Next

- 📱 Mobile app (React Native / Flutter)
- 🤖 AI-powered anomaly detection
- 📚 Multi-language docs expansion
- ⚡ Proxy-level rate limiting per user
- 🔌 Plugin system for custom extensions
- 🌊 WebSocket transport support for sing-box

---

## 💝 Support

If VortexUI is useful to you:

- ⭐ **Star** this repository
- 🍴 **Fork** and contribute
- 📢 **Share** with your community
- 💬 **Join** [@vortex\_ui](https://t.me/vortex_ui) on Telegram

| Network | Address |
|---------|---------|
| **USDT (TRC20)** | `TRLnjZ7YDSwjh3oay28qigEYNieGPMs6ew` |
| **BTC** | `bc1qszt4g7jdv7ev2t3pexctc07ults8nfflht3nj5` |
| **TON** | `UQAYSSSirtQ9_67ZHYUgLVLMx9Ir9vvh3vpcq2qbpit_8-Db` |

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

| 🇺🇸 English | 🇮🇷 فارسی | 🇹🇷 Türkçe | 🇸🇦 العربية |
|------------|----------|-----------|-----------|
| 🇷🇺 Русский | 🇨🇳 中文 | 🇯🇵 日本語 | 🇪🇸 Español |

Full RTL support for Persian and Arabic.

---

## 📄 License

GPL-3.0 — see [LICENSE](LICENSE).

---

<div align="center">

**Built with ❤️ by [iPmart Network](https://github.com/iPmartNetwork)**

<sub>If you find VortexUI useful, please consider giving it a ⭐</sub>

</div>
