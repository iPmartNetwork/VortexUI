#!/usr/bin/env bash
#
# VortexUI installer. Provisions the panel + node + web UI + PostgreSQL/Timescale
# + Redis. Choose at runtime between two methods:
#
#   1) Native (systemd) (recommended) — Go binaries as services; DB/Redis in
#                                       Docker; web served by Caddy.
#   2) Docker Compose                 — everything in containers.
#
#   bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
#
# Re-running is safe: it pulls the latest code and recreates the stack without
# touching existing data or credentials.
set -euo pipefail

REPO_URL="${VORTEXUI_REPO:-https://github.com/iPmartNetwork/VortexUI}"
BRANCH="${VORTEXUI_BRANCH:-master}"
VORTEXUI_VERSION="${VORTEXUI_VERSION:-1.3.4}"
INSTALL_DIR="${VORTEXUI_DIR:-/opt/vortexui}"
WEB_PORT="${VORTEXUI_WEB_PORT:-80}"
METHOD="${VORTEXUI_METHOD:-}"   # docker | native (prompted if empty)

b=$'\e[34m'; g=$'\e[32m'; y=$'\e[33m'; r=$'\e[31m'; d=$'\e[2m'; n=$'\e[0m'
info() { echo "${b}==>${n} $*"; }
ok()   { echo "${g}✓${n} $*"; }
warn() { echo "${y}!${n} $*"; }
die()  { echo "${r}✗ $*${n}" >&2; exit 1; }

# GitHub mirror for restricted networks (Iran/China). Set VORTEXUI_GH_MIRROR to
# prefix GitHub URLs, e.g. https://ghproxy.com/ or https://mirror.ghproxy.com/
GH_MIRROR="${VORTEXUI_GH_MIRROR:-}"

[ "$(id -u)" -eq 0 ] || die "please run as root (sudo)."

PUBLIC_HOST="$(curl -fsS4 https://api.ipify.org 2>/dev/null || curl -fsS4 https://icanhazip.com 2>/dev/null || curl -fsS4 https://ip.sb 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo 127.0.0.1)"

ensure_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    info "installing Docker…"
    # Try official script first, then alternative mirrors for restricted networks
    if ! curl -fsSL https://get.docker.com | sh 2>/dev/null; then
      warn "official Docker install failed — trying mirror (get.docker.com may be blocked)."
      # Alibaba Cloud mirror (accessible from most regions including Iran)
      local os_id="$(. /etc/os-release; echo "$ID")"
      local os_codename="$(. /etc/os-release; echo "$VERSION_CODENAME")"
      curl -fsSL "https://mirrors.aliyun.com/docker-ce/linux/${os_id}/gpg" | apt-key add - 2>/dev/null || true
      echo "deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/${os_id} ${os_codename} stable" \
        > /etc/apt/sources.list.d/docker.list 2>/dev/null || true
      apt-get update -y && apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin 2>/dev/null || {
        # Last resort: use the static binary install
        warn "could not install Docker via apt — trying static binaries."
        local dl_url; local docker_ver="27.3.1"
        if curl -fsSL -o /tmp/docker.tgz "https://mirrors.aliyun.com/docker-ce/linux/static/stable/$(uname -m)/docker-${docker_ver}.tgz" 2>/dev/null; then
          tar -xzf /tmp/docker.tgz -C /tmp
          # Docker static tarball has all binaries flat inside docker/ dir
          install -m 0755 /tmp/docker/* /usr/local/bin/ 2>/dev/null || true
          # Install Docker Compose plugin for the static install
          mkdir -p /usr/local/lib/docker/cli-plugins
          local compose_ver; compose_ver="$(curl -fsSL https://api.github.com/repos/docker/compose/releases/latest 2>/dev/null | grep -oP '"tag_name": "\K[^"]+' || echo "v2.32.0")"
          curl -fsSL -o /usr/local/lib/docker/cli-plugins/docker-compose \
            "${GH_MIRROR}https://github.com/docker/compose/releases/download/${compose_ver}/docker-compose-linux-$(uname -m)" 2>/dev/null
          chmod +x /usr/local/lib/docker/cli-plugins/docker-compose 2>/dev/null || true
        fi
      }
    fi
    systemctl enable --now docker || true
  fi
  docker compose version >/dev/null 2>&1 || docker-compose version >/dev/null 2>&1 || die "Docker Compose v2 plugin not found."
}
ensure_git() {
  command -v git >/dev/null 2>&1 || { info "installing git…"; (apt-get update -y && apt-get install -y git) || yum install -y git || apk add git; }
}

# Installs a pg_dump/pg_restore matching the DB's major version (pinned to
# pg16 by docker-compose.yml's timescale/timescaledb:latest-pg16 image).
# Distro repos often ship an older client (e.g. Ubuntu 22.04 → pg14), and
# pg_dump refuses to talk to a newer server, so a plain
# `apt install postgresql-client` is not enough — pull from the official
# PostgreSQL APT repo (PGDG) instead.
ensure_pg_client() {
  local want=16
  if command -v pg_dump >/dev/null 2>&1; then
    local have; have="$(pg_dump --version | grep -oE '[0-9]+' | head -1)"
    [ "$have" = "$want" ] && return 0
    info "pg_dump v$have found but DB is v$want — installing a matching client…"
  else
    info "installing postgresql-client-$want (matches the TimescaleDB pg$want image)…"
  fi
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -y && apt-get install -y postgresql-common curl ca-certificates gnupg || true
    # Try PGDG repo first, then fall back to distro's default client package
    if curl -fsSL --connect-timeout 5 https://apt.postgresql.org/pub/repos/apt/ 2>/dev/null; then
      if [ -x /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh ]; then
        yes | /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y || true
      fi
      apt-get update -y
      apt-get install -y "postgresql-client-$want" || apt-get install -y postgresql-client
    else
      info "apt.postgresql.org unreachable (network filtering) — installing distro default postgresql-client."
      apt-get install -y postgresql-client || {
        # Last resort: install from PostgreSQL APT mirror (timescale.cloud) for restricted networks
        curl -fsSL https://install.postgresql.org | sudo sh -s -- -y -p "$want" 2>/dev/null ||
        apt-get install -y "postgresql-client-${want}" 2>/dev/null ||
        apt-get install -y postgresql-client 2>/dev/null ||
        warn "could not install postgresql-client — full database backup/restore will not work until pg_dump is installed manually."
      }
    fi
  elif command -v yum >/dev/null 2>&1; then
    yum install -y "postgresql${want}" || yum install -y postgresql
  elif command -v apk >/dev/null 2>&1; then
    apk add --no-cache "postgresql${want}-client" || apk add --no-cache postgresql-client
  fi
}

checkout() {
  if [ -d "$INSTALL_DIR/.git" ]; then
    info "updating $INSTALL_DIR…"
    git -C "$INSTALL_DIR" fetch --depth 1 origin "$BRANCH"
    git -C "$INSTALL_DIR" checkout -q "$BRANCH"
    git -C "$INSTALL_DIR" reset --hard "origin/$BRANCH"
  else
    info "cloning $REPO_URL ($BRANCH)…"
    git clone --depth 1 --branch "$BRANCH" "$REPO_URL" "$INSTALL_DIR"
  fi
  cd "$INSTALL_DIR"; ok "source ready."
}

# Prompts for domain+SSL vs IP+port, sets SITE_ADDRESS / ACME_EMAIL / WEB_PORT.
ask_access() {
  SITE_ADDRESS=":${WEB_PORT}"; ACME_EMAIL=""
  [ -n "${VORTEXUI_NONINTERACTIVE:-}" ] && return
  echo
  echo "  ${b}How should the panel be reached?${n}"
  echo "   ${b}1)${n} Domain with automatic HTTPS (Let's Encrypt)  ${d}— recommended${n}"
  echo "   ${b}2)${n} IP address, plain HTTP on a port"
  read -r -p "  choose [1/2]: " mode
  if [ "$mode" = "1" ]; then
    read -r -p "  domain/subdomain (e.g. panel.example.com): " DOMAIN
    [ -n "$DOMAIN" ] || die "a domain is required for HTTPS mode."
    while [ -z "$ACME_EMAIL" ]; do
      read -r -p "  email for Let's Encrypt (required): " ACME_EMAIL
      [ -n "$ACME_EMAIL" ] || echo "  ${y}an email is required to obtain an SSL certificate${n}"
    done
    SITE_ADDRESS="$DOMAIN"
    warn "point $DOMAIN's DNS A record to this server and open ports 80 + 443."
  else
    read -r -p "  HTTP port [${WEB_PORT}]: " p
    WEB_PORT="${p:-$WEB_PORT}"; SITE_ADDRESS=":${WEB_PORT}"
  fi
}

# Sets LOCAL_CORE / LOCAL_CORE_BIN / LOCAL_V2RAY_API for the in-process panel node.
# Override with VORTEXUI_CORE=xray|singbox in non-interactive installs.
ask_local_core() {
  LOCAL_CORE="${VORTEXUI_CORE:-xray}"
  LOCAL_CORE_BIN="/usr/local/bin/xray"
  LOCAL_V2RAY_API="true"
  case "$LOCAL_CORE" in
    singbox) LOCAL_CORE_BIN="/usr/local/bin/sing-box"; LOCAL_V2RAY_API="false" ;;
  esac
  [ -n "${VORTEXUI_NONINTERACTIVE:-}" ] && return
  echo
  echo "  ${b}Local proxy engine${n} ${d}(runs on this panel server)${n}"
  echo "   ${b}1)${n} xray      ${d}— default; per-user traffic stats${n}"
  echo "   ${b}2)${n} sing-box  ${d}— hysteria/tuic/wireguard; stock binary (no v2ray_api)${n}"
  read -r -p "  choose [1/2] [1]: " cc
  case "$cc" in
    2)
      LOCAL_CORE=singbox
      LOCAL_CORE_BIN=/usr/local/bin/sing-box
      LOCAL_V2RAY_API=false
      ;;
    *)
      LOCAL_CORE=xray
      LOCAL_CORE_BIN=/usr/local/bin/xray
      LOCAL_V2RAY_API=true
      ;;
  esac
}

# Writes deploy/.env, preserving JWT/DB secrets across re-runs.
write_env() {
  ENV_FILE="deploy/.env"
  if [ ! -f "$ENV_FILE" ]; then
    info "generating secrets…"
    JWT_SECRET="$(openssl rand -hex 32 2>/dev/null || head -c32 /dev/urandom | xxd -p -c64)"
    DB_PASSWORD="$(openssl rand -hex 16 2>/dev/null || head -c16 /dev/urandom | xxd -p -c64)"
    cat > "$ENV_FILE" <<EOF
JWT_SECRET=$JWT_SECRET
DB_PASSWORD=$DB_PASSWORD
WEB_PORT=$WEB_PORT
SITE_ADDRESS=$SITE_ADDRESS
ACME_EMAIL=$ACME_EMAIL
# CORE=xray
# SINGBOX_V2RAY_API=false
EOF
    chmod 600 "$ENV_FILE"; ok "wrote $ENV_FILE."
  else
    warn "$ENV_FILE exists — updating access settings, keeping secrets."
    sed -i "/^WEB_PORT=/d;/^SITE_ADDRESS=/d;/^ACME_EMAIL=/d" "$ENV_FILE"
    printf 'WEB_PORT=%s\nSITE_ADDRESS=%s\nACME_EMAIL=%s\n' "$WEB_PORT" "$SITE_ADDRESS" "$ACME_EMAIL" >> "$ENV_FILE"
  fi
  # shellcheck disable=SC1090
  set -a; . "$ENV_FILE"; set +a
}

gen_certs() { # $1 = "go" to use host go, else dockerized go
  [ -f deploy/certs/ca.crt ] && { warn "deploy/certs present — reusing."; return; }
  info "generating mTLS certificates…"
  if [ "${1:-}" = "go" ]; then
    go run ./cmd/gencerts -out deploy/certs -san "localhost,127.0.0.1,$PUBLIC_HOST"
  else
    docker run --rm -v "$PWD":/src -w /src golang:1.26-alpine \
      go run ./cmd/gencerts -out deploy/certs -san "localhost,127.0.0.1,$PUBLIC_HOST"
  fi
  ok "certificates written."
}

access_url() {
  case "$SITE_ADDRESS" in
    :*) echo "http://$PUBLIC_HOST:$WEB_PORT" ;;
    *)  echo "https://$SITE_ADDRESS" ;;
  esac
}

# ---------------------------------------------------------------- Docker method
deploy_docker() {
  ensure_docker; ensure_git; checkout; ask_access; write_env; ask_local_core; gen_certs docker
  # Host advertised to clients in subscriptions (domain if set, else public IP).
  case "$SITE_ADDRESS" in :*|"") NODE_HOST="$PUBLIC_HOST" ;; *) NODE_HOST="$SITE_ADDRESS" ;; esac
  sed -i '/^LOCAL_NODE_HOST=/d;/^CORE=/d;/^SINGBOX_V2RAY_API=/d' deploy/.env
  {
    echo "LOCAL_NODE_HOST=$NODE_HOST"
    echo "CORE=$LOCAL_CORE"
    [ "$LOCAL_CORE" = singbox ] && echo "SINGBOX_V2RAY_API=false"
  } >> deploy/.env
  COMPOSE="docker compose --env-file deploy/.env -f deploy/compose.yml"
  info "building and starting the stack…"
  $COMPOSE up -d --build
  info "waiting for the panel…"
  for _ in $(seq 1 60); do
    $COMPOSE exec -T panel /usr/local/bin/panel admin create --help >/dev/null 2>&1 && break; sleep 2
  done
  bootstrap_admin "$COMPOSE exec -T panel /usr/local/bin/panel admin create"
  install -m 0755 scripts/vortexui /usr/local/bin/vortexui && ok "installed 'vortexui' command."
}

# ---------------------------------------------------------------- Native method
ensure_go() {
  command -v go >/dev/null 2>&1 && return
  info "installing Go toolchain…"
  local arch; arch="$(uname -m)"; [ "$arch" = "x86_64" ] && arch=amd64; [ "$arch" = "aarch64" ] && arch=arm64

  # Escape hatch for networks that block Google's download hosts: point
  # VORTEXUI_GO_URL at a reachable Go tarball (a mirror, or a file you staged).
  # curl also honors http_proxy/https_proxy automatically if you export them.
  if [ -n "${VORTEXUI_GO_URL:-}" ]; then
    info "downloading Go from VORTEXUI_GO_URL…"
    curl -fsSL "$VORTEXUI_GO_URL" -o /tmp/go.tgz \
      || die "failed to download Go from VORTEXUI_GO_URL ($VORTEXUI_GO_URL)."
    rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz
    export PATH="$PATH:/usr/local/go/bin"
    ok "Go toolchain installed (custom URL)."
    return
  fi

  # Resolve the current stable Go version dynamically (endpoint returns e.g.
  # "go1.26.3"); strip stray whitespace. Try the latest AND a known-good 1.26.x
  # fallback, each across the official download hosts. The project needs Go 1.26,
  # so the fallback stays on the 1.26 line.
  local latest; latest="$(curl -fsSL "https://go.dev/VERSION?m=text" 2>/dev/null | head -1 | tr -d '[:space:]')"
  case "$latest" in go*) ;; *) latest="" ;; esac
  local v host tgz
  for v in "$latest" go1.26.3; do
    [ -n "$v" ] || continue
    tgz="${v}.linux-${arch}.tar.gz"
    # Try official hosts first, then Chinese mirrors (for Iran and similar regions)
    for host in "https://go.dev/dl" "https://dl.google.com/go" \
                "https://mirrors.ustc.edu.cn/golang/" "https://mirrors.tuna.tsinghua.edu.cn/golang/"; do
      if curl -fsSL "${host}/${tgz}" -o /tmp/go.tgz; then
        rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz
        export PATH="$PATH:/usr/local/go/bin"
        ok "Go toolchain installed (${v})."
        return
      fi
    done
  done
  die "could not download the Go toolchain (go.dev / dl.google.com / mirrors unreachable — likely network filtering). Options: (1) export https_proxy and re-run; (2) set VORTEXUI_GO_URL to a reachable Go tarball and re-run; (3) copy /usr/local/go from a working server (tar it, scp it, extract to /usr/local), then re-run. Go downloads: https://go.dev/dl/"
}

# Download the xray-core and sing-box engines to the host and stage geo data.
install_cores() {
  local arch xarch sarch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  xarch="64";         sarch="amd64" ;;
    aarch64|arm64) xarch="arm64-v8a";  sarch="arm64" ;;
    *) warn "unknown arch '$arch' — install cores manually"; return ;;
  esac
  if ! command -v unzip >/dev/null 2>&1 || ! command -v tar >/dev/null 2>&1; then
    info "installing unzip/tar…"
    (apt-get update -y && apt-get install -y unzip tar) || yum install -y unzip tar || apk add --no-cache unzip tar || true
  fi
  command -v unzip >/dev/null 2>&1 || die "could not install 'unzip' — install it manually and re-run."
  mkdir -p /etc/vortex/assets

  if [ ! -x /usr/local/bin/xray ]; then
    info "installing xray-core…"
    curl -fsSL -o /tmp/xray.zip "${GH_MIRROR}https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-${xarch}.zip"
    unzip -o /tmp/xray.zip -d /tmp/xray-dl >/dev/null
    install -m 0755 /tmp/xray-dl/xray /usr/local/bin/xray
    cp -f /tmp/xray-dl/*.dat /etc/vortex/assets/ 2>/dev/null || true
    rm -rf /tmp/xray-dl /tmp/xray.zip
    ok "xray-core installed."
  else warn "xray already present — skipping."; fi

  if [ ! -x /usr/local/bin/sing-box ]; then
    info "installing sing-box…"
    # Keep in sync with deploy/Dockerfile SINGBOX_VERSION.
    local ver="v1.12.12"
    # Try directly first, then with GH mirror prefix
    if ! curl -fsSL -o /tmp/sb.tgz "${GH_MIRROR}https://github.com/SagerNet/sing-box/releases/download/${ver}/sing-box-${ver#v}-linux-${sarch}.tar.gz" 2>/dev/null; then
      curl -fsSL -o /tmp/sb.tgz "https://github.com/SagerNet/sing-box/releases/download/${ver}/sing-box-${ver#v}-linux-${sarch}.tar.gz"
    fi
    tar -xzf /tmp/sb.tgz -C /tmp
    install -m 0755 /tmp/sing-box-*/sing-box /usr/local/bin/sing-box
    rm -rf /tmp/sing-box-* /tmp/sb.tgz
    ok "sing-box installed."
  else
    local target="1.12.12" cur
    cur="$(/usr/local/bin/sing-box version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)"
    if [ -n "$cur" ] && [ "$(printf '%s\n' "$target" "$cur" | sort -V | tail -1)" = "$cur" ]; then
      warn "sing-box $cur already present — skipping."
    else
      info "upgrading sing-box${cur:+ (was $cur)} to $target…"
      local ver="v$target"
      curl -fsSL -o /tmp/sb.tgz "${GH_MIRROR}https://github.com/SagerNet/sing-box/releases/download/${ver}/sing-box-${target}-linux-${sarch}.tar.gz"
      tar -xzf /tmp/sb.tgz -C /tmp
      install -m 0755 /tmp/sing-box-*/sing-box /usr/local/bin/sing-box
      rm -rf /tmp/sing-box-* /tmp/sb.tgz
      ok "sing-box upgraded to $target."
    fi
  fi

  # GeoIP country database for the "Traffic by Country" analytics. Best-effort:
  # a download failure must not abort the install — the feature simply stays
  # empty until an mmdb is present at /etc/vortex/GeoLite2-Country.mmdb.
  if [ ! -f /etc/vortex/GeoLite2-Country.mmdb ]; then
    info "downloading GeoLite2-Country database…"
    if curl -fsSL -o /etc/vortex/GeoLite2-Country.mmdb \
        "${GH_MIRROR}https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-Country.mmdb"; then
      ok "GeoLite2-Country database installed."
    else
      rm -f /etc/vortex/GeoLite2-Country.mmdb
      warn "could not download GeoLite2-Country.mmdb — Traffic by Country will stay empty until you place it at /etc/vortex/GeoLite2-Country.mmdb"
    fi
  else warn "GeoLite2-Country.mmdb already present — skipping."; fi
}

deploy_native() {
  ensure_docker; ensure_git; checkout; ask_access; write_env; ask_local_core; ensure_go; gen_certs go
  info "bringing up PostgreSQL + Redis (Docker)…"
  docker compose -f docker-compose.yml up -d

  # pg_dump/pg_restore run on the host for full-DB backup/restore, so the
  # native install needs a version-matched postgres client even though the DB
  # itself lives in Docker.
  ensure_pg_client

  info "building binaries…"
  /usr/local/go/bin/go build -o /usr/local/bin/vortex-panel ./cmd/panel 2>/dev/null || go build -o /usr/local/bin/vortex-panel ./cmd/panel
  go build -o /usr/local/bin/vortex-node ./cmd/node || true

  # Proxy engines (xray + sing-box). VORTEXUI_GH_MIRROR read inside install_cores.
  install_cores

  info "building web UI…"
  command -v node >/dev/null 2>&1 || { curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs; }
  ( cd web && npm ci && npm run build )
  mkdir -p /var/www/vortexui && cp -r web/dist/* /var/www/vortexui/

  # Host advertised to clients in subscriptions (domain if set, else public IP).
  case "$SITE_ADDRESS" in :*|"") NODE_HOST="$PUBLIC_HOST" ;; *) NODE_HOST="$SITE_ADDRESS" ;; esac

  # Native env for the panel service. The panel runs an in-process local node so
  # a single server serves proxy traffic without a separate node agent.
  mkdir -p /etc/vortexui
  cat > /etc/vortexui/panel.env <<EOF
VORTEX_HTTP_ADDR=:8080
VORTEX_DATABASE_URL=postgres://vortex:vortex@127.0.0.1:5432/vortex?sslmode=disable
VORTEX_REDIS_URL=redis://127.0.0.1:6379/0
VORTEX_JWT_SECRET=$JWT_SECRET
VORTEX_TLS_CERT=$INSTALL_DIR/deploy/certs/panel.crt
VORTEX_TLS_KEY=$INSTALL_DIR/deploy/certs/panel.key
VORTEX_TLS_CA=$INSTALL_DIR/deploy/certs/ca.crt
VORTEX_LOCAL_NODE=true
VORTEX_LOCAL_NODE_NAME=local
VORTEX_LOCAL_NODE_HOST=$NODE_HOST
VORTEX_CORE=$LOCAL_CORE
VORTEX_CORE_BIN=$LOCAL_CORE_BIN
VORTEX_CORE_CONFIG=/etc/vortex/local-core.json
VORTEX_CORE_API_PORT=10085
VORTEX_SINGBOX_V2RAY_API=$LOCAL_V2RAY_API
# Stock sing-box needs VORTEX_SINGBOX_V2RAY_API=false (set above when singbox is chosen).
# Set true only with a sing-box binary built with -tags with_v2ray_api.
XRAY_LOCATION_ASSET=/etc/vortex/assets
VORTEX_GEOIP_DB=/etc/vortex/GeoLite2-Country.mmdb
EOF
  chmod 600 /etc/vortexui/panel.env

  info "installing systemd service…"
  cat > /etc/systemd/system/vortexui-panel.service <<EOF
[Unit]
Description=VortexUI panel
Wants=network-online.target
After=network-online.target docker.service
[Service]
EnvironmentFile=/etc/vortexui/panel.env
ExecStart=/usr/local/bin/vortex-panel
Restart=always
RestartSec=5
StartLimitIntervalSec=0
[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable --now vortexui-panel
  sleep 4

  info "installing Caddy reverse proxy…"
  if ! command -v caddy >/dev/null 2>&1; then
    apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl gnupg || true
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' > /etc/apt/sources.list.d/caddy-stable.list
    apt-get update -y && apt-get install -y caddy
  fi
  if [ "${SITE_ADDRESS#:}" = "$SITE_ADDRESS" ]; then CADDY_SITE="$SITE_ADDRESS"; else CADDY_SITE=":$WEB_PORT"; fi
  cat > /etc/caddy/Caddyfile <<EOF
{
	email ${ACME_EMAIL}
}
${CADDY_SITE} {
	encode gzip
	handle /api/* {
		reverse_proxy 127.0.0.1:8080
	}
	handle /sub/* {
		reverse_proxy 127.0.0.1:8080
	}
	handle /health {
		reverse_proxy 127.0.0.1:8080
	}
	handle {
		root * /var/www/vortexui
		try_files {path} /index.html
		file_server
	}
}
EOF
  systemctl restart caddy

  # The native binary reads its DB/JWT config from the environment, so load the
  # panel env before creating the admin (otherwise config validation fails).
  set -a; . /etc/vortexui/panel.env; set +a
  bootstrap_admin "/usr/local/bin/vortex-panel admin create"
  install -m 0755 scripts/vortexui /usr/local/bin/vortexui && ok "installed 'vortexui' command."
  warn "native mode: manage the panel with systemctl {start,stop,restart} vortexui-panel"
}

# ----------------------------------------------------------------- Node agent
# Installs this host as a node agent that an existing panel connects to over
# mTLS. No database/panel here — just the node binary, the proxy engines, and
# the certificates issued by the panel.
deploy_node() {
  ensure_git; checkout
  local arch sarch
  arch="$(uname -m)"; case "$arch" in x86_64|amd64) sarch=amd64 ;; aarch64|arm64) sarch=arm64 ;; *) die "unsupported arch $arch" ;; esac

  install_cores

  # GH_MIRROR is set inside install_cores above. Prefer a prebuilt release binary; fall back to building from source.
  info "installing node agent…"
  local rel; rel="$(curl -fsSL https://api.github.com/repos/iPmartNetwork/VortexUI/releases/latest | grep -oE '"tag_name": *"v[0-9.]+"' | head -1 | grep -oE 'v[0-9.]+')"
  if [ -n "$rel" ] && curl -fL -o /tmp/node.tgz "${GH_MIRROR}https://github.com/iPmartNetwork/VortexUI/releases/download/${rel}/vortexui-node-linux-${sarch}.tar.gz" 2>/dev/null; then
    tar -xzf /tmp/node.tgz -C /tmp && install -m 0755 "/tmp/vortexui-node-linux-${sarch}" /usr/local/bin/vortex-node
  else
    ensure_go; ( cd "$INSTALL_DIR" && go build -o /usr/local/bin/vortex-node ./cmd/node )
  fi

  # mTLS material issued by the panel (ca.crt + node.crt + node.key).
  mkdir -p /etc/vortexui/certs /etc/vortex/assets
  CERTDIR=/etc/vortexui/certs
  echo
  echo "  ${b}This node needs mTLS certs from your panel${n} (ca.crt, node.crt, node.key)."
  echo "   ${b}1)${n} I already uploaded the 3 cert files to a directory  ${d}— recommended${n}"
  echo "   ${b}2)${n} Paste an enrollment bundle  ${d}— on the panel run: ${g}vortexui node-bundle${n}"
  read -r -p "  choose [1/2]: " cm
  if [ "$cm" = "2" ]; then
    echo "  ${d}Paste the bundle line from the panel, then press Enter:${n}"
    read -r BUNDLE
    [ -n "$BUNDLE" ] || die "no bundle pasted — run 'vortexui node-bundle' on the panel and paste the whole line."
    printf '%s' "$BUNDLE" | base64 -d 2>/dev/null | tar -xzf - -C /etc/vortexui/certs 2>/dev/null \
      || die "invalid bundle — re-run 'vortexui node-bundle' on the panel and paste the entire line."
    CERTDIR=/etc/vortexui/certs
  else
    echo "  ${d}On the panel server the files are in:${n} ${g}<install-dir>/deploy/certs/${n} ${d}(ca.crt, node.crt, node.key).${n}"
    echo "  ${d}Upload those 3 files to this node (scp/SFTP), then enter the directory below.${n}"
    read -r -p "  directory with ca.crt/node.crt/node.key [/etc/vortexui/certs]: " d_in
    CERTDIR="${d_in:-/etc/vortexui/certs}"
  fi
  for f in ca.crt node.crt node.key; do
    [ -f "$CERTDIR/$f" ] || die "missing $CERTDIR/$f — paste a valid bundle (vortexui node-bundle) or copy the files and re-run."
  done

  read -r -p "  node listen port [50051]: " NPORT; NPORT="${NPORT:-50051}"
  read -r -p "  core engine (xray/singbox) [xray]: " NCORE; NCORE="${NCORE:-xray}"
  local cbin=/usr/local/bin/xray; [ "$NCORE" = singbox ] && cbin=/usr/local/bin/sing-box
  # The official sing-box release is built WITHOUT the with_v2ray_api tag, so the
  # experimental.v2ray_api block would make it exit with FATAL. Disable it by
  # default for sing-box nodes so the core starts out of the box (trade-off: no
  # per-user traffic stats). For stats, install a sing-box built with
  # -tags with_v2ray_api and set VORTEX_SINGBOX_V2RAY_API=true.
  local v2ray=true; [ "$NCORE" = singbox ] && v2ray=false

  cat > /etc/vortexui/node.env <<EOF
VORTEX_NODE_LISTEN=:$NPORT
VORTEX_CORE=$NCORE
VORTEX_CORE_BIN=$cbin
VORTEX_CORE_CONFIG=/etc/vortex/node-core.json
VORTEX_SINGBOX_V2RAY_API=$v2ray
VORTEX_TLS_CERT=$CERTDIR/node.crt
VORTEX_TLS_KEY=$CERTDIR/node.key
VORTEX_TLS_CA=$CERTDIR/ca.crt
XRAY_LOCATION_ASSET=/etc/vortex/assets
EOF
  chmod 600 /etc/vortexui/node.env

  cat > /etc/systemd/system/vortexui-node.service <<EOF
[Unit]
Description=VortexUI node agent
Wants=network-online.target
After=network-online.target
[Service]
EnvironmentFile=/etc/vortexui/node.env
ExecStart=/usr/local/bin/vortex-node
Restart=always
RestartSec=5
StartLimitIntervalSec=0
[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable --now vortexui-node
  install -m 0755 scripts/vortexui /usr/local/bin/vortexui
  echo
  ok "node agent running on :$NPORT (${NCORE})."
  echo "   ${b}Add it in the panel${n} → Nodes → New:"
  echo "     address: ${g}$PUBLIC_HOST:$NPORT${n}   core: ${g}$NCORE${n}"
  echo "   Manage with: ${g}vortexui {start|stop|status|logs}${n}"
}

# Bootstrap the first admin via the given 'admin create' command prefix.
# Interactively asks for a username and password (with confirmation); falls back
# to VORTEXUI_ADMIN_USER/PASS env vars, then to admin + a random password.
bootstrap_admin() { # $1 = command prefix
  if [ -f deploy/.admin-created ]; then
    warn "admin already bootstrapped (skipping)."
    ADMIN_USER="${VORTEXUI_ADMIN_USER:-admin}"; ADMIN_PASS_DISPLAY="(unchanged)"
    return
  fi

  ADMIN_USER="${VORTEXUI_ADMIN_USER:-}"
  ADMIN_PASS="${VORTEXUI_ADMIN_PASS:-}"
  ADMIN_PASS_DISPLAY=""

  if [ -z "${VORTEXUI_NONINTERACTIVE:-}" ]; then
    echo
    echo "  ${b}Create the admin account${n}"
    if [ -z "$ADMIN_USER" ]; then
      read -r -p "  admin username [admin]: " ADMIN_USER
      ADMIN_USER="${ADMIN_USER:-admin}"
    fi
    if [ -z "$ADMIN_PASS" ]; then
      while :; do
        read -r -s -p "  admin password: " ADMIN_PASS; echo
        read -r -s -p "  confirm password: " _p2; echo
        if [ -z "$ADMIN_PASS" ]; then echo "  ${y}password cannot be empty${n}"; continue; fi
        if [ "$ADMIN_PASS" != "$_p2" ]; then echo "  ${y}passwords do not match — try again${n}"; continue; fi
        break
      done
      ADMIN_PASS_DISPLAY="(the password you set)"
    fi
  fi

  ADMIN_USER="${ADMIN_USER:-admin}"
  if [ -z "$ADMIN_PASS" ]; then
    ADMIN_PASS="$(openssl rand -hex 8 2>/dev/null || echo changeme$RANDOM)"
    ADMIN_PASS_DISPLAY="$ADMIN_PASS"   # generated — show it so it isn't lost
  fi
  [ -n "$ADMIN_PASS_DISPLAY" ] || ADMIN_PASS_DISPLAY="(as provided)"

  info "creating the initial admin…"
  if $1 --username "$ADMIN_USER" --password "$ADMIN_PASS" --sudo; then
    touch deploy/.admin-created; ok "admin '$ADMIN_USER' created."
  else
    warn "admin creation failed (it may already exist) — create one later with: vortexui admin create --username U --password P --sudo"
    ADMIN_PASS_DISPLAY="(creation failed)"
  fi
}

# ----------------------------------------------------------------------- main
# ROLE: panel (full control plane + local node) or node (agent for a panel).
ROLE="${VORTEXUI_ROLE:-}"
if [ -z "$ROLE" ] && [ -z "${VORTEXUI_NONINTERACTIVE:-}" ]; then
  echo
  echo "  ${g}VortexUI installer${n} ${d}(v${VORTEXUI_VERSION})${n}"
  echo "  ${d}─────────────────────${n}"
  echo "  What are you installing on this server?"
  echo "   ${b}1)${n} Panel  ${d}— full control plane; build inbounds here (single-server ready)${n}"
  echo "   ${b}2)${n} Node   ${d}— agent only; connects to an existing panel over mTLS${n}"
  read -r -p "  choose [1/2]: " rr
  case "$rr" in 2) ROLE=node ;; *) ROLE=panel ;; esac
fi
ROLE="${ROLE:-panel}"

if [ "$ROLE" = node ]; then
  deploy_node
  exit 0
fi

# Panel role: choose the deployment method.
if [ -z "$METHOD" ] && [ -z "${VORTEXUI_NONINTERACTIVE:-}" ]; then
  echo
  echo "  Choose how to run the panel:"
  echo "   ${b}1)${n} Native (systemd) ${d}— recommended; host binaries + Caddy, DB/Redis in Docker${n}"
  echo "   ${b}2)${n} Docker Compose   ${d}— everything in containers${n}"
  read -r -p "  choose [1/2]: " m
  case "$m" in 2) METHOD=docker ;; *) METHOD=native ;; esac
fi
METHOD="${METHOD:-native}"

case "$METHOD" in
  docker) deploy_docker ;;
  native) deploy_native ;;
  *) die "unknown method: $METHOD (expected docker or native)" ;;
esac

echo
ok "VortexUI is up (${METHOD} install)."
echo "   ${b}URL:${n}      $(access_url)"
echo "   ${b}Username:${n} ${ADMIN_USER:-admin}"
echo "   ${b}Password:${n} ${ADMIN_PASS_DISPLAY:-(unchanged)}"
echo
echo "   Manage with: ${g}vortexui${n}  (interactive menu) or ${g}vortexui {start|stop|status|logs|update}${n}"
