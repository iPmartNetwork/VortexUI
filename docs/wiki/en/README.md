# 🌀 VortexUI Documentation

<div align="center">

**Next-Generation Proxy Management Panel**

*User-Centric · Core-Agnostic · Enterprise-Ready*

[![Version](https://img.shields.io/badge/version-1.3.1-7c3aed?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/releases)
[![License](https://img.shields.io/badge/license-MIT-green?style=for-the-badge)](https://github.com/iPmartNetwork/VortexUI/blob/master/LICENSE)
[![Docker](https://img.shields.io/badge/docker-ready-blue?style=for-the-badge)](https://hub.docker.com/r/ipmartnetwork/vortexui)

</div>

---

## 🚀 Quick Install

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

One command. Interactive setup. HTTPS included.

---

## 🆕 What's New in 1.3.x

> **Latest: v1.3.1** — Settings validation, translated error messages, API client fixes, toggle switch corrections.

| Feature | Since | Description |
|---------|-------|-------------|
| 🔐 Persisted Panel Settings | 1.3.0 | PostgreSQL-backed config (replaces localStorage) |
| 📜 Audit Log | 1.3.0 | Live table of every admin action at `/audit` |
| 🎨 Portal Whitelabel | 1.3.0 | Per-tenant branding for the self-service portal |
| 🎟 Referral System | 1.3.0 | End-user referral codes with rewards |
| 🔒 Real ACME | 1.3.0 | Let's Encrypt via Cloudflare DNS-01 |
| 🌐 Federation | 1.3.0 | Multi-panel peer synchronization |
| 💼 Reseller Platform | 1.2.9 | Wallet, orders, plans, profiles |
| 🖥 Command Tower UI | 1.2.9 | Merged pages, fleet telemetry, geo pins |

[Full changelog →](16-changelog.md)

---

## 📖 Documentation Map

| Section | Description |
|---------|-------------|
| [Introduction](01-introduction.md) | Architecture, feature overview, comparison |
| [Installation](02-installation.md) | One-line install, Docker, native build |
| [First Steps](03-first-steps.md) | Login, add node, create inbound, add user |
| [Dashboard](04-dashboard.md) | Widgets, analytics, monitor, command palette |
| [Users](05-user-management.md) | CRUD, quotas, subscriptions, portal, shop |
| [Nodes](06-node-management.md) | Enrollment, health, auto-migration, monitoring |
| [Network](07-network-policy.md) | Outbounds, routing packs, CDN chains, load balancers |
| [Security](08-security-administration.md) | RBAC, TLS tricks, probing protection, IP-limit |
| [Plans & Payments](09-plans-payments.md) | Per-reseller plans, payment config, wallet |
| [Notifications](10-notifications.md) | Webhooks, Telegram, quota alerts, SSE |
| [Settings](11-settings-backup.md) | Branding, whitelabel, backup, updates |
| [API Reference](12-api-reference.md) | Authentication, endpoints, OpenAPI spec |
| [Protocols](13-protocols-config.md) | 14 protocols, transports, security layers |
| [Operations](14-operations-maintenance.md) | HTTPS, Prometheus, scaling, performance |
| [Troubleshooting](15-troubleshooting-faq.md) | Common issues, debug tips, FAQ |
| [Changelog](16-changelog.md) | Version history & migration guides |
| [Menu & Usage Guide](17-menu-usage-guide.md) | Complete menu-by-menu explanation and daily workflows |

---

## ✨ Key Features

### 🔧 Engine & Infrastructure
- **Dual-core support** — Xray-core and sing-box, choose per node
- **Push delta traffic** — Restart-safe, no double-counting
- **mTLS node fleet** — Encrypted connections, auto-failover
- **Auto-migration** — Move users from unhealthy nodes automatically
- **Federation** — Sync users/nodes across multiple panels

### 🛡 Security & Anti-Censorship
- **Reality Scanner** — Discover optimal SNIs with latency scoring
- **TLS Tricks Manager** — ISP-specific profiles (fragment, mux, padding)
- **Probing protection** — Detect and block active GFW probes
- **Decoy website** — Serve fake site to probers
- **DNS-over-HTTPS** — Built-in DoH with ad/malware blocking

### 👥 User Management & Commerce
- **Self-service portal** — Login with sub token, view usage, tickets
- **Self-service shop** — Per-reseller plans with multiple payment methods
- **Smart Quota** — Progressive speed reduction (fair use tiers)
- **Family groups** — Shared data pools for multiple users
- **Referral system** — Invite codes with data/days rewards

### 💼 Reseller Platform
- **Wallet billing** — Credit system with top-up queue
- **Sub-resellers** — Create child resellers with inherited scope
- **Whitelabel** — Custom branding (logo, colors, title, footer)
- **Webhooks** — Outbound events for automation
- **Policy limits** — Max data limit, max expire, bulk restrictions

### 🎨 Frontend & UX
- **Veltrix Glass UI** — Modern glass design system
- **Command palette** — Ctrl+K fuzzy search across everything
- **Dashboard widgets** — Drag & drop, resize, customize layout
- **8 languages** — EN/FA/TR/AR/RU/ZH/JA/ES with full RTL support
- **Dark + Light** — Smooth animated theme transition

---

## 🔗 Quick Links

| Resource | Link |
|----------|------|
| GitHub Repository | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
| Telegram Channel | [@vortex_ui](https://t.me/vortex_ui) |
| OpenAPI Spec | [openapi.yaml](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| Changelog | [CHANGELOG.md](https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md) |
| Bug Reports | [GitHub Issues](https://github.com/iPmartNetwork/VortexUI/issues) |

---

## 🌍 Languages

This documentation is available in:

- 🇬🇧 **English** (current)
- 🇮🇷 [فارسی](../fa/README.md)
- 🇸🇦 [العربية](../ar/README.md)
- 🇹🇷 [Türkçe](../tr/README.md)
