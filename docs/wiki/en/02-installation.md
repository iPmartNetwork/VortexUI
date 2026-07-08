# Installation

> **Recommended:** Use the **one-line installer** for the fastest path to a working panel. It handles dependencies, database setup, HTTPS, and systemd services automatically.

---

## Prerequisites

| Requirement | Minimum | Recommended |
|-------------|---------|-------------|
| OS | Ubuntu 20.04 / Debian 11 | Ubuntu 22.04+ / Debian 12 |
| RAM | 1 GB | 2 GB+ |
| Disk | 10 GB | 20 GB+ (TimescaleDB grows with traffic data) |
| CPU | 1 vCPU | 2+ vCPU |
| Go (native build only) | 1.26 | 1.26 |
| Docker (container install) | 24.0+ | Latest stable |
| Domain | Optional | Recommended (for HTTPS + subscriptions) |

---

## One-Line Install

```bash
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
```

The installer will:

1. Detect your OS and architecture
2. Install dependencies (PostgreSQL, Redis, Caddy)
3. Download and build VortexUI
4. Run database migrations
5. Create a sudo admin account (interactive prompt)
6. Configure systemd services
7. Set up HTTPS via Caddy (if a domain is provided)

After completion, access the panel at `https://your-domain.com` or `http://server-ip:8080`.

---

## Docker Compose

### Quick Start

```bash
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI/deploy
cp ../.env.example .env
# Edit .env with your settings
docker compose up -d
```

### Production (with Caddy HTTPS)

```bash
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI/deploy
cp ../.env.example .env
```

Edit `.env`:

```env
VORTEX_DOMAIN=panel.example.com
VORTEX_ADMIN_USER=admin
VORTEX_ADMIN_PASS=your-secure-password
VORTEX_JWT_SECRET=random-32-byte-string
VORTEX_DB_URL=postgres://vortex:pass@db:5432/vortex?sslmode=disable
VORTEX_REDIS_URL=redis://redis:6379/0
```

Then:

```bash
docker compose up -d
```

The `deploy/compose.yml` includes: panel, web frontend, PostgreSQL + TimescaleDB, Redis, and Caddy.

---

## Native Build

### Ubuntu/Debian

```bash
# Install Go 1.26
sudo snap install go --classic
go version  # should show go1.26.x

# Install dependencies
sudo apt update && sudo apt install -y postgresql redis-server

# Clone and build
git clone https://github.com/iPmartNetwork/VortexUI.git
cd VortexUI
go build -o vortexui ./cmd/panel

# Run migrations
./vortexui migrate

# Create admin
./vortexui admin create --username admin --password your-password --sudo

# Start
./vortexui serve
```

### Other Linux

```bash
# Install Go 1.26 from official tarball
wget https://go.dev/dl/go1.26.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.26.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Then follow the same clone/build steps as Ubuntu
```

> **Note:** VortexUI requires **Go 1.26** or later. Earlier versions will fail to compile.

---

## Node Agent Setup

The node agent runs on remote servers and communicates with the panel via gRPC + mTLS.

### Enrollment Wizard (Recommended)

1. In the panel UI, go to **Nodes → Add Node**
2. The enrollment wizard generates a one-line install command
3. SSH into your remote server and paste the command
4. The agent auto-registers, exchanges certificates, and starts reporting

### Manual Install

```bash
# On the remote server
bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)
```

You'll be prompted for:
- Panel address (e.g. `https://panel.example.com`)
- Node enrollment token (generated in panel UI)

### Docker Node

```bash
docker run -d --name vortex-node \
  -e PANEL_ADDR=https://panel.example.com \
  -e NODE_TOKEN=your-enrollment-token \
  --network host \
  ghcr.io/ipmartnetwork/vortexui-node:latest
```

---

## Local Node (Single Server)

If you only need one server, use the **local node** — the proxy core runs in-process alongside the panel. No separate agent needed.

1. During installation, select "Yes" when asked about local node
2. Or later: **Nodes → Add Node → Local**
3. Choose core (Xray or sing-box)
4. The panel manages the core process directly

> **Tip:** Local node is perfect for single-server setups. For multi-server deployments, use remote nodes with the enrollment wizard.

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| VORTEX_DOMAIN | Panel domain (for HTTPS) | — |
| VORTEX_LISTEN | API listen address | :8080 |
| VORTEX_DB_URL | PostgreSQL connection string | postgres://localhost/vortex |
| VORTEX_REDIS_URL | Redis connection string | redis://localhost:6379/0 |
| VORTEX_JWT_SECRET | JWT signing key (≥32 bytes) | — (required) |
| VORTEX_ADMIN_USER | Initial admin username | — |
| VORTEX_ADMIN_PASS | Initial admin password | — |
| VORTEX_TELEGRAM_TOKEN | Telegram bot token | — |
| VORTEX_TELEGRAM_ADMIN | Admin chat ID for notifications | — |
| VORTEX_ZARINPAL_MERCHANT | ZarinPal merchant ID | — |
| VORTEX_NOWPAYMENTS_KEY | NowPayments API key | — |
| VORTEX_NOWPAYMENTS_IPN_SECRET | NowPayments IPN HMAC secret | — |
| VORTEX_BACKUP_CRON | Backup schedule (cron expression) | — |
| VORTEX_BACKUP_TELEGRAM | Send backups to Telegram | false |
| VORTEX_BACKUP_S3_BUCKET | S3 bucket for backups | — |
| VORTEX_METRICS_ENABLED | Enable Prometheus metrics | false |
| VORTEX_METRICS_LISTEN | Metrics endpoint address | :9090 |
| VORTEX_SHARE_AUTOLIMIT | Auto-limit on account sharing detection | false |

---

## CLI Management

The `vortexui` binary provides an interactive menu:

```bash
vortexui
```

```
╔══════════════════════════════════════╗
║          VortexUI Management         ║
╠══════════════════════════════════════╣
║  1) Start panel                      ║
║  2) Stop panel                       ║
║  3) Restart panel                    ║
║  4) Status                           ║
║  5) Logs (live)                      ║
║  6) Update                           ║
║  7) Admin management                 ║
║  8) Backup                           ║
║  9) Doctor (diagnostics)             ║
║  0) Exit                             ║
╚══════════════════════════════════════╝
```

Key commands:

| Command | Action |
|---------|--------|
| vortexui update | Pull latest release and restart |
| vortexui admin create | Create a new admin |
| vortexui admin reset-password | Reset admin password |
| vortexui backup | Create an immediate backup |
| vortexui doctor | Run diagnostics (DB, Redis, nodes, ports) |
| vortexui migrate | Run pending database migrations |

---

## Updating

### Auto Update (Recommended)

```bash
vortexui update
```

This pulls the latest release, rebuilds, runs migrations, and restarts.

### Manual Update (Panel Server)

```bash
cd /opt/VortexUI  # or wherever you cloned
git pull origin master
go build -o vortexui ./cmd/panel
./vortexui migrate
sudo systemctl restart vortexui
```

### Manual Update (Node Servers)

```bash
cd /opt/VortexUI-node
git pull origin master
go build -o vortex-node ./cmd/node
sudo systemctl restart vortex-node
```

### Docker Update

```bash
cd /opt/VortexUI/deploy
docker compose pull
docker compose up -d
```

---

## Post-Install Verification

After installation, verify everything is working:

1. **Panel accessible** — open `https://your-domain.com` in a browser
2. **Login works** — sign in with your admin credentials
3. **Database connected** — check Settings → System Info
4. **Node online** — if using local node, verify it shows "Online" in Nodes page
5. **Run diagnostics** — `vortexui doctor` checks all components

### Health Endpoint

The panel exposes `GET /api/health` — returns `200 OK` with component status. Use this for external monitoring (UptimeRobot, Prometheus blackbox, etc.).
