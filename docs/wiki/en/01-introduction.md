# Introduction

## What is VortexUI?

**VortexUI** is a next-generation proxy management panel built for operators who need:

- **Scale** — manage thousands of users across dozens of nodes
- **Resilience** — automatic failover, health monitoring, self-healing node fleet
- **Anti-censorship** — ISP-specific TLS tricks, probing protection, decoy sites, WARP+
- **Self-service** — end-users manage their own accounts, purchase plans, open tickets
- **Revenue** — per-reseller plans, multiple payment gateways, wallet billing, referral program
- **Delegation** — full reseller platform with sub-resellers, whitelabel, policy limits

Unlike inbound-centric panels (3x-ui), VortexUI uses a **user-centric model**: one user identity provides access to all assigned protocols across all nodes simultaneously.

> **Current version: 1.3.1** — Now with persisted settings, audit logging, real ACME certificates, federation, and a full reseller platform.

---

## Design Principles

| Principle | Description |
|-----------|-------------|
| **API-first** | Every UI action is backed by a documented REST endpoint |
| **Multi-tenant** | Resellers see only their scope; RBAC enforced |
| **Observable** | Audit log, Prometheus metrics, structured JSON logs |
| **Portable** | Docker Compose, Kubernetes, or bare metal |
| **Secure by default** | 2FA, IP guard, per-token scopes, mTLS |
| **Localized** | 8 languages, full RTL support (FA, AR) |

---

## Feature Overview

### Engine & Infrastructure

| Capability | Details |
|------------|---------|
| Dual-core support | Xray-core and sing-box — choose per node |
| Push delta traffic | Restart-safe, no double-counting, never loses data |
| mTLS node fleet | Encrypted connections, auto-failover, migrate-back |
| Node enrollment wizard | Four-step UI for onboarding remote nodes |
| Auto-migration | Move users from unhealthy nodes automatically |
| Federation | Sync users/nodes across multiple panels |
| Local node | In-process core — no separate agent needed |
| CDN/Relay chains | Multi-hop paths with CDN, relay, and worker hops |
| Load balancers | 4 strategies with health probing |
| Cloudflare DNS automation | Auto-manage DNS records for nodes |

### Security & Anti-Censorship

| Capability | Details |
|------------|---------|
| Reality Scanner | Discover optimal SNIs with latency scoring |
| TLS Tricks Manager | ISP-specific profiles (fragment, mux, padding) |
| Probing protection | Detect and block active GFW probes |
| Fingerprint validation | JA3-based client filtering |
| Decoy website | Serve fake site to probers (proxy or static mode) |
| DNS-over-HTTPS | Built-in DoH with ad/malware blocking |
| Evasion profiles | Reusable anti-DPI presets per country |
| WARP+ integration | Cloudflare outbound for clean IP |
| Clean-IP scanner | Find best CDN edge IPs by latency/loss scoring |
| IP-limit enforcement | Per-user concurrent IP caps with configurable actions |
| Geo-blocking | Per-inbound country restrictions |
| Account-sharing guard | Detect and act on credential sharing |

### User Management & Commerce

| Capability | Details |
|------------|---------|
| Self-service portal | Login with sub token, view usage, tickets |
| Self-service shop | Per-reseller plans with card/crypto/ZarinPal payment |
| Smart Quota | Progressive speed reduction (fair use tiers) |
| Family groups | Shared data pools for multiple users |
| Referral system | Invite codes with data/days rewards |
| Per-reseller plans | Each reseller creates their own plans and pricing |
| Payment gateways | ZarinPal (online), card-to-card (proof upload), crypto (TX hash) |
| Wallet billing | Reseller credit system with top-up queue |
| Subscription hosts | Per-inbound CDN/address overrides with template variables |
| Deep links + QR | One-tap subscription import |
| Config templates | Custom Clash/sing-box routing per user |
| Import tools | Migrate from 3x-ui or Marzban |

### Administration & Reseller Platform

| Capability | Details |
|------------|---------|
| RBAC + roles | Granular permissions per admin |
| Full reseller platform | Wallet, sub-resellers, whitelabel, webhooks, policy limits |
| Scoped allowlists | Per-reseller plan/node/inbound restrictions |
| Auto-suspend | Automatic reseller suspension on violations |
| Audit log | Every admin action tracked with diff |
| Quota notifications | Configurable thresholds for reseller alerts |
| Auto-backup | Scheduled exports to Telegram or S3 |
| Grafana metrics | Prometheus endpoint + ready dashboard |

### New in 1.3.x

| Capability | Since | Details |
|------------|-------|---------|
| Persisted settings | 1.3.0 | All panel config stored in PostgreSQL, not browser |
| Audit log | 1.3.0 | Live table of every admin mutation with diff view |
| Real ACME | 1.3.0 | Let's Encrypt certificates via Cloudflare DNS-01 |
| Federation | 1.3.0 | Multi-panel peer coordination and count sync |
| Portal referral | 1.3.0 | End-users share invite links from the portal |
| Portal whitelabel | 1.3.0 | Per-tenant branding on the self-service portal |
| Command Tower UI | 1.2.9 | Merged pages, fleet telemetry, geo pin map |
| Reseller platform | 1.2.9 | Wallet, orders, plans, sub-admin profiles |

### Frontend & UX

| Capability | Details |
|------------|---------|
| Command palette | Ctrl+K fuzzy search across everything |
| Dashboard widgets | Drag & drop, resize, customize layout |
| World map | Geographic traffic visualization |
| Real-time gauges | Animated CPU/RAM/bandwidth indicators |
| Monitor page | Live connection table (user, node, IP, protocol, duration) |
| Analytics | Geo breakdown, top users, peak hours, CSV export |
| Onboarding tour | First-time admin walkthrough |
| 8 languages | EN/FA/TR/AR/RU/ZH/JA/ES with full RTL support |
| Dark + Light | Smooth animated theme transition |
| Mobile portal | Bottom nav, pull-to-refresh, bottom sheets |

---

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  Caddy (Web Layer)     — HTTPS, SPA, reverse proxy, DoH     │
├──────────────────────────────────────────────────────────────┤
│  Panel (Go 1.26)       — REST API, SSE, gRPC hub, scheduler │
│  ├─ Auth               — JWT + TOTP + portal tokens          │
│  ├─ Services           — user, node, plan, order, analytics  │
│  ├─ Hub                — node fleet management + failover    │
│  ├─ Scanner            — Reality SNI prober + Clean-IP       │
│  ├─ Migration          — health-based user redistribution    │
│  ├─ Reseller           — wallet, plans, branding, webhooks   │
│  └─ Federation         — cross-panel sync                    │
├──────────────────────────────────────────────────────────────┤
│  PostgreSQL + TimescaleDB — data + time-series traffic       │
│  Redis                    — cache, sessions, device tracker  │
├──────────────────────────────────────────────────────────────┤
│  Node Agent (gRPC)     — remote core execution + health      │
│  Local Node            — in-process on panel host            │
└──────────────────────────────────────────────────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go 1.26, Echo, gRPC, sqlc, pgx |
| Frontend | React 18, TypeScript 5.6, Tailwind CSS, TanStack Query |
| Database | PostgreSQL 16 + TimescaleDB |
| Cache | Redis 7 |
| Proxy Cores | Xray-core, sing-box |
| Web Server | Caddy (auto HTTPS) |
| Transport | gRPC + mTLS (panel ↔ nodes) |
| Notifications | Webhook (HMAC-SHA256), Telegram Bot API |
| Monitoring | Prometheus metrics + Grafana |

---

## Comparison with Other Panels

| Feature | VortexUI 1.3.1 | 3x-ui | Marzban | Hiddify |
|---------|----------------|-------|---------|---------|
| Dual core (Xray + sing-box) | ✅ | ❌ | ❌ | ✅ |
| User-centric model | ✅ | ❌ | ✅ | ✅ |
| Push delta traffic | ✅ | polling | polling | polling |
| Node auto-migration | ✅ | ❌ | ❌ | ❌ |
| Load balancer (4 strategies) | ✅ | ❌ | ❌ | ❌ |
| Reality Scanner | ✅ | ❌ | ❌ | ❌ |
| TLS Tricks (ISP profiles) | ✅ | ❌ | ❌ | partial |
| Probing protection | ✅ | ❌ | ❌ | ❌ |
| Self-service portal | ✅ | ❌ | ❌ | ✅ |
| Per-reseller shop | ✅ | ❌ | ❌ | ❌ |
| Federation | ✅ | ❌ | ❌ | ❌ |
| Audit log | ✅ | ❌ | ❌ | ❌ |
| Real ACME (DNS-01) | ✅ | ❌ | partial | ✅ |
| Command palette (Ctrl+K) | ✅ | ❌ | ❌ | ❌ |
| Backend | Go | Go | Python | Python |
| Database | PG + TimescaleDB | SQLite | SQLite | SQLite |

---

## Supported Protocols

| Protocol | Core | Inbound | Outbound | Transport | Security |
|----------|------|---------|----------|-----------|----------|
| VLESS | Both | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, xHTTP, mKCP | None, TLS, REALITY |
| VMess | Both | ✅ | ✅ | TCP, WS, gRPC, HTTPUpgrade, mKCP | None, TLS |
| Trojan | Both | ✅ | ✅ | TCP, WS, gRPC, mKCP | TLS, REALITY |
| Shadowsocks | Both | ✅ | ✅ | TCP (+ SS-2022 multi-user) | None |
| Hysteria2 | sing-box | ✅ | ✅ | UDP (QUIC) | TLS |
| TUIC | sing-box | ✅ | ✅ | UDP (QUIC) | TLS |
| WireGuard | sing-box | ✅ | ✅ | UDP | Native |
| Hysteria (v1) | sing-box | ✅ | — | UDP | TLS |
| ShadowTLS | sing-box | ✅ | ✅ | TCP | TLS |
| AnyTLS | sing-box | ✅ | — | TCP | TLS |
| Naive | sing-box | ✅ | — | — | TLS (mandatory) |
| SOCKS | Both | ✅ | ✅ | TCP (no transport) | plaintext |
| HTTP | Both | ✅ | ✅ | TCP (no transport) | plaintext |
| Dokodemo | Xray | ✅ | — | — | plaintext |

---

## Key Terminology

| Term | Meaning |
|------|---------|
| **Panel** | The control server — API, UI, database, schedulers |
| **Node** | A server running a proxy core (Xray or sing-box) |
| **Local Node** | In-process core on the same machine as the panel |
| **Inbound** | Client-facing entry point (protocol + port + config) |
| **Outbound** | Egress path (freedom, proxy chain, WARP, blackhole) |
| **Subscription** | /sub/{token} — auto-detected config for any client app |
| **Portal** | End-user self-service web interface |
| **Shop** | Per-reseller plan purchase page (/sub/{token}/shop) |
| **Hub** | Internal component managing all node connections |
| **Federation** | Multiple panels connected for user/node sync |
| **Smart Quota** | Fair-use policy with progressive speed tiers |
| **Reseller** | Admin with scoped access, wallet, own plans/users |
| **Whitelabel** | Per-reseller branding (logo, colors, title) |

---

## Next Steps

1. **[Installation](02-installation.md)** — get VortexUI running in 5 minutes
2. **[First Steps](03-first-steps.md)** — login, add node, create first user
3. **[Dashboard](04-dashboard.md)** — explore the real-time overview
