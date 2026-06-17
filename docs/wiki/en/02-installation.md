# 2. Installation

!!! important
    Before installing, prepare **ports 80 and 443** (for HTTPS) and your domain DNS.

---

## Prerequisites

| Item | Docker (recommended) | Native |
|------|:--------------------:|:------:|
| OS | Linux (Ubuntu 22.04+) | Linux |
| RAM | 2 GB minimum | 2 GB+ |
| CPU | 1 vCPU | 1+ |
| Disk | 10 GB | 10 GB |
| Docker + Compose v2 | ✅ | DB/Redis only |
| Go 1.26 | — | ✅ (auto-installed) |
| Ports | 80, 443 (+ inbounds) | same |

---

## Method 1: One-Line Install (Recommended)

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

The install script is **interactive** and asks two main questions:

### 1. Install method

| Option | Description |
|--------|-------------|
| **Docker Compose** *(recommended)* | Full stack in containers: web · panel · node · PostgreSQL · Redis |
| **Native (systemd)** | Go binary as a service; DB/Redis in Docker; SPA via Caddy |

### 2. Panel access

| Option | Description |
|--------|-------------|
| **Domain + HTTPS** | Caddy obtains Let's Encrypt cert — ports 80 and 443 must be open |
| **IP + HTTP** | Custom port (e.g. 8080) |

### Non-interactive install (script/CI)

```bash
VORTEXUI_METHOD=docker \
VORTEXUI_NONINTERACTIVE=1 \
VORTEXUI_ADMIN_USER=admin \
VORTEXUI_ADMIN_PASS='your-strong-password' \
bash install.sh
```

### Install output

- Install path: `/opt/vortexui` (override with `VORTEXUI_DIR`)
- `vortexui` command in `/usr/local/bin`
- Env file: `deploy/.env` (JWT, DB password, domain)
- mTLS certs: `deploy/certs/`
- Panel URL + initial admin credentials printed in the terminal

---

## Method 2: Manual Docker Compose

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

# Generate secrets
echo "JWT_SECRET=$(openssl rand -hex 32)" >> deploy/.env
echo "DB_PASSWORD=$(openssl rand -hex 16)" >> deploy/.env
echo "SITE_ADDRESS=panel.example.com" >> deploy/.env
echo "ACME_EMAIL=admin@example.com" >> deploy/.env

make certs
docker compose --env-file deploy/.env -f deploy/compose.yml up -d --build

# Create admin
docker compose -f deploy/compose.yml exec panel \
  /usr/local/bin/panel admin create --username admin --password 'change-me' --sudo
```

### Stack services

| Service | Role |
|---------|------|
| `db` | PostgreSQL 16 + TimescaleDB |
| `redis` | Redis 7 |
| `panel` | API + local node (host network) |
| `web` | Caddy + SPA (HTTPS) |

---

## Method 3: Native Install (Development/Advanced)

```bash
git clone https://github.com/iPmartNetwork/VortexUI && cd VortexUI

docker compose up -d          # PostgreSQL + Redis
cp .env.example .env
# Fill VORTEX_JWT_SECRET with: openssl rand -hex 32

make build
make certs
make run-panel

# Another terminal — create admin
./bin/panel admin create --username admin --password 'your-password' --sudo
```

Frontend (development):

```bash
cd web && npm install && npm run dev
```

---

## Node Agent Install (Multi-Server)

For a multi-node fleet, on each separate server:

```bash
VORTEX_NODE_LISTEN=:50051 \
VORTEX_CORE=xray \
VORTEX_CORE_BIN=/usr/local/bin/xray \
VORTEX_TLS_CERT=node.crt \
VORTEX_TLS_KEY=node.key \
VORTEX_TLS_CA=ca.crt \
./bin/node
```

Then in the panel: **Nodes → Add Node** — register address and mTLS certificate.

---

## Local Node

For a single-server setup without a separate agent:

```env
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=your-public-ip-or-domain
VORTEX_CORE=xray
VORTEX_CORE_BIN=/usr/local/bin/xray
```

In Docker Compose this is enabled by default.

---

## Important Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VORTEX_HTTP_ADDR` | `:8080` | Panel HTTP address |
| `VORTEX_DATABASE_URL` | — | **Required** — PostgreSQL |
| `VORTEX_JWT_SECRET` | — | **Required** — minimum 32 bytes |
| `VORTEX_REDIS_URL` | `redis://localhost:6379/0` | Redis |
| `VORTEX_LOCAL_NODE` | `false` | In-process node |
| `VORTEX_SHARE_AUTOLIMIT` | `false` | Auto-limit on account sharing |
| `VORTEX_WEBHOOK_URL` | — | Notification webhook |
| `VORTEX_TELEGRAM_TOKEN` | — | Telegram bot token |
| `VORTEX_CF_API_TOKEN` | — | Cloudflare DNS automation |

Full list: [`.env.example`](https://github.com/iPmartNetwork/VortexUI/blob/master/.env.example)

---

## Post-Install Health Check

```bash
vortexui status
curl -s http://127.0.0.1:8080/api/health
```

Expected response: `{"status":"ok"}`

---

## Updating

```bash
vortexui update
# or
cd /opt/vortexui && git pull && docker compose -f deploy/compose.yml up -d --build
```

Re-running the install script is **safe** — secrets and DB data are preserved.
