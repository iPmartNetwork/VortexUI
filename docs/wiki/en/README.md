# VortexUI Documentation

**Version: 1.4.0** — Auto-Protocol Switching & Anti-Censorship Intelligence

Welcome to the official VortexUI documentation. VortexUI is a next-generation, core-agnostic proxy management panel supporting Xray and sing-box with intelligent anti-censorship capabilities.

## Quick Navigation

| Section | Description |
|---------|-------------|
| [Introduction](01-introduction.md) | What is VortexUI, design principles |
| [Installation](02-installation.md) | Prerequisites, one-line install, Docker |
| [First Steps](03-first-steps.md) | Initial setup, first node, first user |
| [Dashboard](04-dashboard.md) | Overview, stats, traffic charts |
| [User Management](05-user-management.md) | Users, quotas, devices, subscriptions |
| [Node Management](06-node-management.md) | Fleet management, health, multi-core |
| [Network Policy](07-network-policy.md) | Routing, outbounds, balancers |
| [Security](08-security-administration.md) | TLS tricks, probing protection, admin |
| [Plans & Payments](09-plans-payments.md) | Commerce, reseller, wallet |
| [Notifications](10-notifications.md) | Webhook, Telegram, events |
| [Settings & Backup](11-settings-backup.md) | Configuration, backup/restore |
| [API Reference](12-api-reference.md) | REST API documentation |
| [Protocols](13-protocols-config.md) | Supported protocols & configuration |
| [Operations](14-operations-maintenance.md) | Maintenance, upgrades |
| [Troubleshooting](15-troubleshooting-faq.md) | FAQ, common issues |
| [Changelog](16-changelog.md) | Version history |
| [Menu Guide](17-menu-usage-guide.md) | UI navigation guide |

## What's New in v1.4.0

- **Auto-Protocol Switching** — self-healing protocol failover
- **Smart Config Engine** — per-ISP anti-DPI optimization  
- **Dynamic SNI Rotation** — daily rotation from ISP-specific pools
- **Multi-CDN Routing** — Cloudflare/ArvanCloud/Gcore clean-IP fallback
- **Smart Mux** — ISP-optimized multiplexing (h2mux/yamux)
- **Quality Score** — per-proxy scoring and auto-reordering
- **DNS Leak Prevention** — DoH + block plain DNS
- **Emergency Fallback** — last-resort outbound

## Links

- [GitHub Repository](https://github.com/iPmartNetwork/VortexUI)
- [Telegram Channel](https://t.me/vortex_ui)
