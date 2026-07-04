# Command Tower UI (v1.2.9)

!!! info "Backend + frontend"
    Run migration **0030** and rebuild the web UI before restart.

## Highlights

| Area | What's new |
|------|------------|
| **Merged pages** | Routing, Security, Reseller Platform use `?tab=` sub-routes |
| **Settings hub** | Sidebar tabs including Admins (sudo) |
| **Reseller profile** | `/settings/admins/:id` — wallet, quota, ledger |
| **Overview** | Traffic ranges, top users + protocol, node geo/ping |
| **API** | `GET /api/admins/:id/quota`, `GET /api/admins/:id/wallet` |

See the [full English guide](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/wiki/en/17-v129-panel-redesign.md).
