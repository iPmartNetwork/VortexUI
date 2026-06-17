<div align="center">

<img src="../assets/Logo.svg" alt="VortexUI" width="120" />

**VortexUI Wiki**

[Wiki](./README.md) · [FA](../fa/04-dashboard.md) · [AR](../ar/04-dashboard.md) · [TR](../tr/04-dashboard.md)

</div>

<div>

# 4. Dashboard (Overview)

[← First steps](./03-first-steps.md) · [Index](./README.md) · [Next: Users →](./05-user-management.md)

> [!NOTE]
> The dashboard updates via **SSE** without refresh — no polling needed.

<div align="center">

| Light | Dark |
|:-----:|:----:|
| ![Overview dashboard — live stats and traffic chart](../assets/panel/overview_light.png) | ![Overview dashboard — live stats and traffic chart](../assets/panel/overview_dark.png) |

*Overview dashboard — live stats and traffic chart*

</div>

---

## Overview

The **Overview** page is the central operations view: fleet status, traffic, active users, and recent events — all with **live updates (SSE)**.

---

## Stat Cards

| Card | Content |
|------|---------|
| **Users** | Total, active, limited, expired |
| **Traffic** | Total upload/download, time series |
| **Nodes** | Online/offline count |
| **Connections** | Active proxy connections |

---

## Traffic Chart

- Time series backed by **TimescaleDB**
- Selectable ranges (24h, 7d, 30d)
- Upload/download breakdown

---

## Live Updates (SSE)

The panel uses **Server-Sent Events**:

```
GET /api/events/stream?access_token=<JWT>
```

When an event occurs (node down, user limited, etc.) the UI updates without a page refresh.

| Event | UI effect |
|-------|-----------|
| `node.down` | Red node badge + toast |
| `user.limited` | User status updated |
| `user.ip_limit` | Account-sharing warning |
| `user.expiry_warning` | 3-day expiry notification |

> Caddy transparently proxies this stream. The token comes from the query string because `EventSource` cannot send custom headers.

---

## Prometheus / Grafana

Metrics are available on the Prometheus endpoint (for external monitoring). Details: [Chapter 14 — Operations](./14-operations-maintenance.md).

</div>
