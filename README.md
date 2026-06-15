<div align="center">
  <img src="img/Logo.svg" alt="VortexUI" width="300" />
  <p><strong>Next-generation proxy management panel</strong></p>
  <p>Core-agnostic · User-centric · Real-time</p>

  [![Release](https://img.shields.io/github/v/release/iPmartNetwork/VortexUI?style=flat-square&color=blue)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![Downloads](https://img.shields.io/github/downloads/iPmartNetwork/VortexUI/total?style=flat-square&color=green)](https://github.com/iPmartNetwork/VortexUI/releases)
  [![Stars](https://img.shields.io/github/stars/iPmartNetwork/VortexUI?style=flat-square&color=yellow)](https://github.com/iPmartNetwork/VortexUI/stargazers)
  [![License](https://img.shields.io/github/license/iPmartNetwork/VortexUI?style=flat-square)](LICENSE)
  [![CI](https://img.shields.io/github/actions/workflow/status/iPmartNetwork/VortexUI/ci.yml?style=flat-square&label=CI)](https://github.com/iPmartNetwork/VortexUI/actions)
  [![Go](https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go)](https://go.dev)
  [![TypeScript](https://img.shields.io/badge/TypeScript-5.6-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org)
  [![Docker](https://img.shields.io/docker/pulls/ipmartnetwork/vortexui?style=flat-square&logo=docker)](https://hub.docker.com/r/ipmartnetwork/vortexui)
  
  ![Visitors](https://api.visitorbadge.io/api/visitors?path=https%3A%2F%2Fgithub.com%2FiPmartNetwork%2FVortexUI&countColor=%23263759)

  <br />
  
  [Features](#-features) · [Screenshots](#-screenshots) · [Comparison](#-comparison) · [Quick Start](#-quick-start) · [Protocols](#-supported-protocols) · [Roadmap](#-roadmap) · [Contributing](#-contributing)
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

### 👥 User Management
- User-centric model (one identity → many protocols)
- Subscription: auto-detect Clash/sing-box/base64
- QR codes + per-format URLs
- Traffic accounting: delta-based, restart-safe
- Quota enforcement + scheduled reset
- Device limit + HWID allowlist
- Bulk actions (multi-select)
- Import from 3x-ui / Marzban

### 🌐 Network Policy
- Outbounds: freedom/blackhole/dns + proxy chaining
- Routing rules: domain/IP/port/protocol matchers
- Load balancers: random/roundRobin/leastPing/leastLoad
- Observatory with health probing

</td>
<td width="50%">

### 🖥 Node Fleet
- mTLS connections (panel ↔ node)
- Auto-failover + migrate-back on recovery
- Live health monitoring (CPU/RAM/Disk)
- Remote restart / stop core
- Per-node logs streaming

### 🔔 Notifications
- Event bus: user.limited, user.expired, node.down, ...
- Webhook (HMAC-SHA256 signed)
- Telegram bot notifications

### 🔐 Security
- JWT + TOTP 2FA
- API tokens (Personal Access Tokens)
- RBAC with granular permissions
- Login brute-force protection
- Online IP enforcement (account-sharing guard)
- Audit log (all admin mutations)

### 🎨 Frontend
- React 18 + TypeScript + Tailwind
- 8 languages (EN/FA/TR/AR/RU/ZH/JA/ES)
- Dark (Navy Blue) + Light (Pastel) themes
- Responsive (mobile drawer)
- Real-time dashboard with live charts

</td>
</tr>
</table>

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

| | VortexUI | 3x-ui | Marzban | Hiddify |
|:--|:--:|:--:|:--:|:--:|
| **Proxy engines** | Xray + sing-box | Xray | Xray | Xray + sing-box |
| **Data model** | User-centric | Inbound-centric | User-centric | User-centric |
| **Traffic method** | Push delta ✨ | Polling | Polling | Polling |
| **Multi-node** | ✅ mTLS + failover | ✅ | ✅ | ✅ |
| **Balancer** | ✅ | ❌ | ❌ | ❌ |
| **Outbound CRUD** | ✅ | Partial | ❌ | ❌ |
| **Routing rules** | ✅ | ❌ | ❌ | ❌ |
| **REALITY keygen** | ✅ | ✅ | ✅ | ✅ |
| **Local node** | ✅ | ✅ | ❌ | ❌ |
| **API tokens** | ✅ | ❌ | ❌ | ❌ |
| **Audit log** | ✅ | ❌ | ❌ | ❌ |
| **Anti-sharing** | ✅ IP enforce | IP limit | ❌ | ❌ |
| **Backup** | ✅ Transactional | ✅ | ✅ | ✅ |
| **Notifications** | Webhook + TG | TG | ✅ | TG |
| **Languages** | 8 | 13 | 3 | 5 |
| **Backend** | Go | Go | Python | Python |
| **Database** | PG + TimescaleDB | SQLite/PG | SQLite/Maria | SQLite |
| **Theme** | Dark + Light | Dark + Light | Dark | Dark + Light |

---

## 📡 Supported Protocols

<div align="center">

| Protocol | Inbound | Outbound | Transport |
|----------|:-------:|:--------:|:---------:|
| **VLESS** | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade |
| **VMess** | ✅ | ✅ | TCP, WS, gRPC |
| **Trojan** | ✅ | ✅ | TCP, WS, gRPC |
| **Shadowsocks** | ✅ | ✅ | TCP |
| **SOCKS** | — | ✅ | TCP |
| **HTTP** | — | ✅ | TCP |
| **Hysteria2** | 🔜 | — | UDP |
| **TUIC** | 🔜 | — | UDP |
| **WireGuard** | 🔜 | — | UDP |

</div>

**Security layers:** None, TLS, REALITY

---

## 🚀 Quick Start

### One-line Install

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

### Manual Setup

```bash
# Clone
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# Dependencies
docker compose up -d   # PostgreSQL + TimescaleDB + Redis

# Configure
cp .env.example .env
# Set VORTEX_JWT_SECRET: openssl rand -hex 32

# Build & Run
make build
make certs             # dev mTLS certificates
make run-panel

# Create admin
./bin/panel admin create --username admin --password 'your-password' --sudo
```

Open `http://your-server:8080` and login.

### Docker

```bash
# Build images
make images

# Full stack (panel + node + DB + Redis)
make stack-up
```

### Node Agent

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_TLS_CERT=node.crt VORTEX_TLS_KEY=node.key VORTEX_TLS_CA=ca.crt \
./bin/node
```

Or enable **local node** (in-process):
```env
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_HOST=your-public-ip
```

---

## 📖 Documentation

| Topic | Link |
|-------|------|
| API Reference (OpenAPI 3.0) | [`docs/openapi.yaml`](docs/openapi.yaml) |
| Environment Variables | [`.env.example`](.env.example) |
| Docker Deploy | [`deploy/`](deploy/) |
| CI/CD | [`.github/workflows/`](.github/workflows/) |

---

## 🗺 Roadmap

- [x] Core-agnostic engine (Xray + sing-box)
- [x] User-centric data model
- [x] Push-based delta traffic accounting
- [x] Auto-failover + migrate-back
- [x] Outbound/Routing/Balancer management
- [x] REALITY key generation
- [x] Webhook + Telegram notifications
- [x] Backup/Restore (transactional)
- [x] Audit log
- [x] API tokens (PAT)
- [x] Account-sharing guard (online IP enforcement)
- [x] Import from 3x-ui / Marzban
- [x] 8-language frontend + RTL
- [x] Real-time dashboard with live charts
- [ ] Hysteria2 / TUIC / WireGuard protocols
- [ ] DNS management
- [ ] Evasion profiles (fragment, fingerprint presets)
- [ ] Payment integration
- [ ] Reseller sub-panels
- [ ] Mobile app (React Native)
- [ ] Grafana integration
- [ ] Cluster mode (multi-panel HA)

---

## 💝 Support

If you find VortexUI useful, please consider:

- ⭐ **Star** this repository
- 🍴 **Fork** and contribute
- 📢 **Share** with others
- 💰 **Donate** crypto to support development

<div align="center">

| Network | Address |
|:-------:|---------|
| **USDT (TRC20)** | `TRLnjZ7YDSwjh3oay28qigEYNieGPMs6ew` |
| **BTC** | `bc1qszt4g7jdv7ev2t3pexctc07ults8nfflht3nj5` |
| **TON** | `UQAYSSSirtQ9_67ZHYUgLVLMx9Ir9vvh3vpcq2qbpit_8-Db` |

</div>

> 💡 Replace the addresses above with your actual wallet addresses.

---

## 🤝 Contributing

Contributions are welcome! Please:

1. Fork the repo
2. Create a feature branch (`git checkout -b feat/amazing`)
3. Commit (`git commit -m 'feat: add amazing feature'`)
4. Push (`git push origin feat/amazing`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

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

This project is licensed under the **AGPL-3.0** License — see the [LICENSE](LICENSE) file.

---

<div align="center">
  <br />
  <img src="img/Logo.svg" alt="VortexUI" width="200" />
  <br /><br />
  <sub>© 2026 iPmart Network. All rights reserved.</sub>
  <br /><br />
  
  **Made with ❤️ by [iPmart Network](https://github.com/iPmartNetwork)**
  
  ⭐ Star this repo if you find it useful!
</div>
