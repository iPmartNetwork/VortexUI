#!/usr/bin/env bash
#
# VortexUI installer. Provisions the panel + node + web UI + PostgreSQL/Timescale
# + Redis. Choose at runtime between two methods:
#
#   1) Docker Compose (recommended) — everything in containers.
#   2) Native (systemd)            — Go binaries as services; DB/Redis in Docker;
#                                    web served by Caddy. For users who prefer
#                                    host processes over containers.
#
#   bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)
#
# Re-running is safe: it pulls the latest code and recreates the stack without
# touching existing data or credentials.
set -euo pipefail

REPO_URL="${VORTEXUI_REPO:-https://github.com/iPmartNetwork/VortexUI}"
BRANCH="${VORTEXUI_BRANCH:-master}"
INSTALL_DIR="${VORTEXUI_DIR:-/opt/vortexui}"
WEB_PORT="${VORTEXUI_WEB_PORT:-80}"
METHOD="${VORTEXUI_METHOD:-}"   # docker | native (prompted if empty)

b=$'\e[34m'; g=$'\e[32m'; y=$'\e[33m'; r=$'\e[31m'; d=$'\e[2m'; n=$'\e[0m'
info() { echo "${b}==>${n} $*"; }
ok()   { echo "${g}✓${n} $*"; }
warn() { echo "${y}!${n} $*"; }
die()  { echo "${r}✗ $*${n}" >&2; exit 1; }

[ "$(id -u)" -eq 0 ] || die "please run as root (sudo)."

PUBLIC_HOST="$(curl -fsS https://api.ipify.org 2>/dev/null || hostname -I 2>/dev/null | awk '{print $1}' || echo 127.0.0.1)"

ensure_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    info "installing Docker…"; curl -fsSL https://get.docker.com | sh; systemctl enable --now docker || true
  fi
  docker compose version >/dev/null 2>&1 || die "Docker Compose v2 plugin not found."
}
ensure_git() {
  command -v git >/dev/null 2>&1 || { info "installing git…"; (apt-get update -y && apt-get install -y git) || yum install -y git || apk add git; }
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
    read -r -p "  email for Let's Encrypt (optional): " ACME_EMAIL
    SITE_ADDRESS="$DOMAIN"
    warn "point $DOMAIN's DNS A record to this server and open ports 80 + 443."
  else
    read -r -p "  HTTP port [${WEB_PORT}]: " p
    WEB_PORT="${p:-$WEB_PORT}"; SITE_ADDRESS=":${WEB_PORT}"
  fi
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
  ensure_docker; ensure_git; checkout; ask_access; write_env; gen_certs docker
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
  local ver=1.26.0 arch; arch="$(uname -m)"; [ "$arch" = "x86_64" ] && arch=amd64; [ "$arch" = "aarch64" ] && arch=arm64
  curl -fsSL "https://go.dev/dl/go${ver}.linux-${arch}.tar.gz" -o /tmp/go.tgz
  rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz
  export PATH="$PATH:/usr/local/go/bin"
}
deploy_native() {
  ensure_docker; ensure_git; checkout; ask_access; write_env; ensure_go; gen_certs go
  info "bringing up PostgreSQL + Redis (Docker)…"
  docker compose -f docker-compose.yml up -d

  info "building binaries…"
  /usr/local/go/bin/go build -o /usr/local/bin/vortex-panel ./cmd/panel 2>/dev/null || go build -o /usr/local/bin/vortex-panel ./cmd/panel
  go build -o /usr/local/bin/vortex-node ./cmd/node || true

  info "building web UI…"
  command -v node >/dev/null 2>&1 || { curl -fsSL https://deb.nodesource.com/setup_22.x | bash - && apt-get install -y nodejs; }
  ( cd web && npm ci && npm run build )
  mkdir -p /var/www/vortexui && cp -r web/dist/* /var/www/vortexui/

  # Native env for the panel service.
  mkdir -p /etc/vortexui
  cat > /etc/vortexui/panel.env <<EOF
VORTEX_HTTP_ADDR=:8080
VORTEX_DATABASE_URL=postgres://vortex:vortex@127.0.0.1:5432/vortex?sslmode=disable
VORTEX_REDIS_URL=redis://127.0.0.1:6379/0
VORTEX_JWT_SECRET=$JWT_SECRET
VORTEX_TLS_CERT=$INSTALL_DIR/deploy/certs/panel.crt
VORTEX_TLS_KEY=$INSTALL_DIR/deploy/certs/panel.key
VORTEX_TLS_CA=$INSTALL_DIR/deploy/certs/ca.crt
EOF
  chmod 600 /etc/vortexui/panel.env

  info "installing systemd service…"
  cat > /etc/systemd/system/vortexui-panel.service <<EOF
[Unit]
Description=VortexUI panel
After=network.target docker.service
[Service]
EnvironmentFile=/etc/vortexui/panel.env
ExecStart=/usr/local/bin/vortex-panel
Restart=always
RestartSec=3
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
	@panel path /api/* /sub/* /health
	reverse_proxy @panel 127.0.0.1:8080
	root * /var/www/vortexui
	try_files {path} /index.html
	file_server
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
if [ -z "$METHOD" ] && [ -z "${VORTEXUI_NONINTERACTIVE:-}" ]; then
  echo
  echo "  ${g}VortexUI installer${n}"
  echo "  ${d}─────────────────────${n}"
  echo "  Choose an installation method:"
  echo "   ${b}1)${n} Docker Compose   ${d}— recommended, everything in containers${n}"
  echo "   ${b}2)${n} Native (systemd) ${d}— host binaries + Caddy, DB/Redis in Docker${n}"
  read -r -p "  choose [1/2]: " m
  case "$m" in 2) METHOD=native ;; *) METHOD=docker ;; esac
fi
METHOD="${METHOD:-docker}"

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
