# VortexUI Documentation

Welcome to the official VortexUI guide.

Install, configure, and operate the next-generation proxy panel (Xray + sing-box). Use the **language selector** in the header to switch between English, Persian, Arabic, and Turkish.

!!! tip "Quick install"
    ```bash
    bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
    ```

## Architecture

```mermaid
flowchart TB
    subgraph Client["Clients"]
        Browser["Browser / PWA"]
        ProxyApp["Clash / sing-box / v2rayNG"]
    end
    subgraph Web["Web Layer"]
        Caddy["Caddy — HTTPS + SPA"]
    end
    subgraph Panel["Control Plane"]
        API["Panel API — Go"]
        SSE["SSE — Live Events"]
        DB[(PostgreSQL + TimescaleDB)]
        Redis[(Redis)]
    end
    subgraph Nodes["Node Fleet"]
        Local["Local Node"]
        Remote["Remote Nodes — mTLS"]
    end
    Browser --> Caddy
    ProxyApp --> Caddy
    Caddy --> API
    API --> DB
    API --> Redis
    API --> Local
    API --> Remote
```

## Useful links

| Resource | Link |
|----------|------|
| OpenAPI | [openapi.yaml on GitHub](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml) |
| Protocol examples | [protocols.md](https://github.com/iPmartNetwork/VortexUI/blob/master/docs/protocols.md) |
| Repository | [github.com/iPmartNetwork/VortexUI](https://github.com/iPmartNetwork/VortexUI) |
